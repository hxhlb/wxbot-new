//go:build windows

package loader

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"wxbot-new/internal/obfuscate"
)

// 仅支持 32 位 PE 手动映射，用于降低 DLL 被直接检测的风险。
// 实现思路：
//   1. 读取 DLL 文件并解析 DOS / NT 头
//   2. VirtualAlloc 一块内存，拷贝 Headers + Sections
//   3. 处理重定位表（IMAGE_DIRECTORY_ENTRY_BASERELOC）
//   4. 解析 Import 表并通过 LoadLibraryA / GetProcAddress 填充 IAT
//   5. 调用 TLS 回调（如有）和 DllMain(DLL_PROCESS_ATTACH)

type manualModule struct {
	base       uintptr
	size       uintptr
	entryPoint uintptr
}

// Windows 常量定义
const (
	memCommit                   = 0x1000
	memReserve                  = 0x2000
	memRelease                  = 0x8000
	pageReadWrite               = 0x04
	pageExecuteRead             = 0x20
	pageExecuteReadWrite        = 0x40
	pageReadOnly                = 0x02
	dllProcessAttach            = 1
	dllProcessDetach            = 0
	imageDirectoryEntryImport   = 1
	imageDirectoryEntryBasReloc = 5
	imageDirectoryEntryTLS      = 9
	imageSizeofBaseRelocation   = 8
	imageRelBasedHighLow        = 3
	imageOrdinalFlag32          = 0x80000000
	imageSCNMemExecute          = 0x20000000
	imageSCNMemRead             = 0x40000000
	imageSCNMemWrite            = 0x80000000
)

// PE 结构定义（32 位）

type imageDOSHeader struct {
	EMagic    uint16
	ECblp     uint16
	ECp       uint16
	ECrlc     uint16
	ECparhdr  uint16
	EMinAlloc uint16
	EMaxAlloc uint16
	ESS       uint16
	ESP       uint16
	ECsum     uint16
	EIP       uint16
	ECS       uint16
	ELfarlc   uint16
	EOverlay  uint16
	ERes      [4]uint16
	EOEMID    uint16
	EOEMInfo  uint16
	ERes2     [10]uint16
	ELfanew   int32
}

type imageFileHeader struct {
	Machine              uint16
	NumberOfSections     uint16
	TimeDateStamp        uint32
	PointerToSymbolTable uint32
	NumberOfSymbols      uint32
	SizeOfOptionalHeader uint16
	Characteristics      uint16
}

type imageDataDirectory struct {
	VirtualAddress uint32
	Size           uint32
}

type imageOptionalHeader32 struct {
	Magic                       uint16
	MajorLinkerVersion          byte
	MinorLinkerVersion          byte
	SizeOfCode                  uint32
	SizeOfInitializedData       uint32
	SizeOfUninitializedData     uint32
	AddressOfEntryPoint         uint32
	BaseOfCode                  uint32
	BaseOfData                  uint32
	ImageBase                   uint32
	SectionAlignment            uint32
	FileAlignment               uint32
	MajorOperatingSystemVersion uint16
	MinorOperatingSystemVersion uint16
	MajorImageVersion           uint16
	MinorImageVersion           uint16
	MajorSubsystemVersion       uint16
	MinorSubsystemVersion       uint16
	Win32VersionValue           uint32
	SizeOfImage                 uint32
	SizeOfHeaders               uint32
	CheckSum                    uint32
	Subsystem                   uint16
	DllCharacteristics          uint16
	SizeOfStackReserve          uint32
	SizeOfStackCommit           uint32
	SizeOfHeapReserve           uint32
	SizeOfHeapCommit            uint32
	LoaderFlags                 uint32
	NumberOfRvaAndSizes         uint32
	DataDirectory               [16]imageDataDirectory
}

type imageNTHeaders32 struct {
	Signature      uint32
	FileHeader     imageFileHeader
	OptionalHeader imageOptionalHeader32
}

type imageSectionHeader struct {
	Name                 [8]byte
	VirtualSize          uint32
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLinenumbers uint32
	NumberOfRelocations  uint16
	NumberOfLinenumbers  uint16
	Characteristics      uint32
}

type imageBaseRelocation struct {
	VirtualAddress uint32
	SizeOfBlock    uint32
}

type imageImportDescriptor struct {
	OriginalFirstThunk uint32
	TimeDateStamp      uint32
	ForwarderChain     uint32
	Name               uint32
	FirstThunk         uint32
}

type imageTLSDirectory32 struct {
	StartAddressOfRawData uint32
	EndAddressOfRawData   uint32
	AddressOfIndex        uint32
	AddressOfCallbacks    uint32
	SizeOfZeroFill        uint32
	Characteristics       uint32
}

// 不再使用全局变量，改为运行时动态加载
// API 调用统一通过 obfuscate.CallAPI 进行

// manualLoadModule 手动映射一个 32 位 DLL 到当前进程
func manualLoadModule(path string) (*manualModule, error) {
	// 混淆：预热延迟
	obfuscate.PrewarmDelay()

	// 创建 PE 解析混淆器
	splitter := obfuscate.NewSplitPEParsingOps()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 DLL 失败: %w", err)
	}

	// 混淆：假装读取其他文件
	obfuscate.LegitimateFileOps()
	splitter.NextStep()

	if len(data) < 0x100 {
		return nil, fmt.Errorf("DLL 文件过小，可能不是有效的 PE 文件")
	}

	// 解析 DOS Header
	var dos imageDOSHeader
	if err := binary.Read(bytesReader(data, 0), binary.LittleEndian, &dos); err != nil {
		return nil, fmt.Errorf("解析 DOS Header 失败: %w", err)
	}
	splitter.NextStep()

	if dos.EMagic != 0x5A4D { // "MZ"
		return nil, fmt.Errorf("无效的 DOS 魔数")
	}

	ntOffset := int(dos.ELfanew)
	if ntOffset <= 0 || ntOffset+4 > len(data) {
		return nil, fmt.Errorf("无效的 NT Header 偏移")
	}

	splitter.NextStep()

	// 解析 NT Headers
	var nt imageNTHeaders32
	if err := binary.Read(bytesReader(data, ntOffset), binary.LittleEndian, &nt); err != nil {
		return nil, fmt.Errorf("解析 NT Header 失败: %w", err)
	}
	splitter.NextStep()

	if nt.Signature != 0x00004550 { // "PE\0\0"
		return nil, fmt.Errorf("无效的 PE 签名")
	}
	if nt.OptionalHeader.Magic != 0x10B { // PE32
		return nil, fmt.Errorf("只支持 32 位 PE (Magic=0x10B)")
	}

	splitter.NextStep()

	sizeOfImage := uintptr(nt.OptionalHeader.SizeOfImage)
	sizeOfHeaders := uintptr(nt.OptionalHeader.SizeOfHeaders)

	// 申请内存 - 先使用 PAGE_READWRITE 降低检测风险
	// 使用动态 API 加载
	base, callErr := obfuscate.CallAPI(
		obfuscate.GetKernel32(),
		obfuscate.GetVirtualAlloc(),
		0,
		sizeOfImage,
		memCommit|memReserve,
		pageReadWrite, // 修改为 RW 权限
	)
	if base == 0 {
		return nil, fmt.Errorf("VirtualAlloc 失败: %v", callErr)
	}

	mem := unsafe.Slice((*byte)(unsafe.Pointer(base)), int(sizeOfImage))

	// 拷贝 Headers
	copy(mem[:sizeOfHeaders], data[:sizeOfHeaders])
	splitter.NextStep()

	// 解析 Section Headers
	sectionOffset := ntOffset + int(unsafe.Sizeof(nt))
	sections := make([]imageSectionHeader, nt.FileHeader.NumberOfSections)
	for i := 0; i < int(nt.FileHeader.NumberOfSections); i++ {
		offset := sectionOffset + i*int(unsafe.Sizeof(imageSectionHeader{}))
		if err := binary.Read(bytesReader(data, offset), binary.LittleEndian, &sections[i]); err != nil {
			obfuscate.CallAPI(obfuscate.GetKernel32(), obfuscate.GetVirtualFree(), base, 0, memRelease)
			return nil, fmt.Errorf("解析 SectionHeader[%d] 失败: %w", i, err)
		}
		// 每解析 2 个节就混淆一次
		if i%2 == 0 {
			splitter.NextStep()
		}
	}

	// 拷贝各 Section
	for i := range sections {
		sec := &sections[i]
		if sec.SizeOfRawData == 0 {
			continue
		}
		if int(sec.PointerToRawData)+int(sec.SizeOfRawData) > len(data) {
			obfuscate.CallAPI(obfuscate.GetKernel32(), obfuscate.GetVirtualFree(), base, 0, memRelease)
			return nil, fmt.Errorf("Section[%d] 数据越界", i)
		}

		start := int(sec.VirtualAddress)
		end := start + int(sec.SizeOfRawData)
		if end > len(mem) {
			obfuscate.CallAPI(obfuscate.GetKernel32(), obfuscate.GetVirtualFree(), base, 0, memRelease)
			return nil, fmt.Errorf("Section[%d] 映射越界", i)
		}
		copy(mem[start:end], data[sec.PointerToRawData:sec.PointerToRawData+sec.SizeOfRawData])
	}

	imageBase := uintptr(nt.OptionalHeader.ImageBase)
	delta := int32(base - imageBase)

	// 处理重定位
	if nt.OptionalHeader.DataDirectory[imageDirectoryEntryBasReloc].VirtualAddress != 0 {
		if err := applyRelocations(base, &nt); err != nil {
			obfuscate.CallAPI(obfuscate.GetKernel32(), obfuscate.GetVirtualFree(), base, 0, memRelease)
			return nil, err
		}
	}

	// 解析导入表
	if nt.OptionalHeader.DataDirectory[imageDirectoryEntryImport].VirtualAddress != 0 {
		if err := resolveImports(base, &nt); err != nil {
			obfuscate.CallAPI(obfuscate.GetKernel32(), obfuscate.GetVirtualFree(), base, 0, memRelease)
			return nil, err
		}
	}

	// 处理 TLS（如有）
	if nt.OptionalHeader.DataDirectory[imageDirectoryEntryTLS].VirtualAddress != 0 {
		if err := callTLSCallbacks(base, &nt); err != nil {
			obfuscate.CallAPI(obfuscate.GetKernel32(), obfuscate.GetVirtualFree(), base, 0, memRelease)
			return nil, err
		}
	}

	// 在调用 DllMain 前，根据节属性设置正确的内存保护
	if err := applySectionProtections(base, sections); err != nil {
		obfuscate.CallAPI(obfuscate.GetKernel32(), obfuscate.GetVirtualFree(), base, 0, memRelease)
		return nil, err
	}

	// 调用 DllMain(DLL_PROCESS_ATTACH)
	entryRVA := uintptr(nt.OptionalHeader.AddressOfEntryPoint)
	entryPoint := uintptr(0)
	if entryRVA != 0 {
		entryPoint = base + entryRVA
		// 注意：DllMain 约定为 BOOL WINAPI DllMain(HINSTANCE, DWORD, LPVOID)
		ret, _, _ := syscall.Syscall(
			entryPoint,
			3,
			base,
			uintptr(dllProcessAttach),
			0,
		)
		if ret == 0 {
			// 某些 DLL 仍可能返回 0，但为保险起见视为错误
			obfuscate.CallAPI(obfuscate.GetKernel32(), obfuscate.GetVirtualFree(), base, 0, memRelease)
			return nil, fmt.Errorf("DllMain(DLL_PROCESS_ATTACH) 返回失败")
		}
	}

	_ = delta // 仅用于说明，实际已经在 applyRelocations 中使用

	return &manualModule{
		base:       base,
		size:       sizeOfImage,
		entryPoint: entryPoint,
	}, nil
}

// manualFreeModule 卸载手动映射的模块
func manualFreeModule(m *manualModule) error {
	if m == nil || m.base == 0 {
		return nil
	}

	// 调用 DllMain(DLL_PROCESS_DETACH)
	if m.entryPoint != 0 {
		syscall.Syscall(
			m.entryPoint,
			3,
			m.base,
			uintptr(dllProcessDetach),
			0,
		)
	}

	_, err := obfuscate.CallAPI(obfuscate.GetKernel32(), obfuscate.GetVirtualFree(), m.base, 0, memRelease)
	if err != nil {
		return fmt.Errorf("VirtualFree 失败: %v", err)
	}
	return nil
}

func applyRelocations(base uintptr, nt *imageNTHeaders32) error {
	dir := nt.OptionalHeader.DataDirectory[imageDirectoryEntryBasReloc]
	if dir.VirtualAddress == 0 || dir.Size == 0 {
		return nil
	}

	imageBase := uintptr(nt.OptionalHeader.ImageBase)
	delta := int32(base - imageBase)
	if delta == 0 {
		return nil
	}

	relocAddr := base + uintptr(dir.VirtualAddress)
	relocEnd := relocAddr + uintptr(dir.Size)

	for relocAddr < relocEnd {
		block := (*imageBaseRelocation)(unsafe.Pointer(relocAddr))
		if block.SizeOfBlock < imageSizeofBaseRelocation {
			break
		}

		entryCount := (block.SizeOfBlock - imageSizeofBaseRelocation) / 2
		entryAddr := relocAddr + imageSizeofBaseRelocation

		for i := uint32(0); i < entryCount; i++ {
			entry := *(*uint16)(unsafe.Pointer(entryAddr + uintptr(i*2)))
			entryType := entry >> 12
			offset := uintptr(entry & 0x0FFF)

			if entryType == imageRelBasedHighLow {
				patchAddr := base + uintptr(block.VirtualAddress) + offset
				val := *(*uint32)(unsafe.Pointer(patchAddr))
				val += uint32(delta)
				*(*uint32)(unsafe.Pointer(patchAddr)) = val
			}
		}

		relocAddr += uintptr(block.SizeOfBlock)
	}

	return nil
}

func resolveImports(base uintptr, nt *imageNTHeaders32) error {
	dir := nt.OptionalHeader.DataDirectory[imageDirectoryEntryImport]
	if dir.VirtualAddress == 0 || dir.Size == 0 {
		return nil
	}

	descAddr := base + uintptr(dir.VirtualAddress)

	for {
		desc := (*imageImportDescriptor)(unsafe.Pointer(descAddr))
		if desc.Name == 0 {
			break
		}

		// 获取 DLL 名称
		name := rvaToString(base, uintptr(desc.Name))
		if name == "" {
			return fmt.Errorf("解析导入 DLL 名称失败")
		}

		mod, err := loadLibraryA(name)
		if err != nil {
			return fmt.Errorf("LoadLibraryA(%s) 失败: %w", name, err)
		}

		origThunkRVA := desc.OriginalFirstThunk
		thunkRVA := desc.FirstThunk
		if origThunkRVA == 0 {
			origThunkRVA = thunkRVA
		}

		origThunk := base + uintptr(origThunkRVA)
		thunk := base + uintptr(thunkRVA)

		for {
			orig := *(*uint32)(unsafe.Pointer(origThunk))
			if orig == 0 {
				break
			}

			var funcAddr uintptr
			var ferr error

			if (orig & imageOrdinalFlag32) != 0 {
				// 按序号导入
				ordinal := orig & 0xFFFF
				funcAddr, ferr = getProcAddressOrdinal(mod, uint16(ordinal))
			} else {
				// 按名称导入，导入名称结构为 IMAGE_IMPORT_BY_NAME
				nameRVA := uintptr(orig)
				funcName := rvaToString(base, nameRVA+2) // 跳过 Hint(WORD)
				funcAddr, ferr = getProcAddress(mod, funcName)
			}

			if ferr != nil || funcAddr == 0 {
				return fmt.Errorf("解析导入函数失败: dll=%s, err=%v", name, ferr)
			}

			*(*uintptr)(unsafe.Pointer(thunk)) = funcAddr

			origThunk += unsafe.Sizeof(uint32(0))
			thunk += unsafe.Sizeof(uintptr(0))
		}

		descAddr += unsafe.Sizeof(imageImportDescriptor{})
	}

	return nil
}

func callTLSCallbacks(base uintptr, nt *imageNTHeaders32) error {
	dir := nt.OptionalHeader.DataDirectory[imageDirectoryEntryTLS]
	if dir.VirtualAddress == 0 || dir.Size == 0 {
		return nil
	}

	tlsDir := (*imageTLSDirectory32)(unsafe.Pointer(base + uintptr(dir.VirtualAddress)))
	callbacksAddr := uintptr(tlsDir.AddressOfCallbacks)
	if callbacksAddr == 0 {
		return nil
	}

	for {
		cb := *(*uintptr)(unsafe.Pointer(callbacksAddr))
		if cb == 0 {
			break
		}
		syscall.Syscall(cb, 3, base, uintptr(dllProcessAttach), 0)
		callbacksAddr += unsafe.Sizeof(uintptr(0))
	}
	return nil
}

func loadLibraryA(name string) (uintptr, error) {
	if name == "" {
		return 0, fmt.Errorf("空 DLL 名称")
	}
	b := append([]byte(name), 0)
	mod, err := obfuscate.CallAPI(
		obfuscate.GetKernel32(),
		obfuscate.GetLoadLibraryA(),
		uintptr(unsafe.Pointer(&b[0])),
	)
	if mod == 0 {
		return 0, fmt.Errorf("LoadLibraryA(%s) 失败: %v", name, err)
	}
	return mod, nil
}

func getProcAddress(mod uintptr, name string) (uintptr, error) {
	if name == "" {
		return 0, fmt.Errorf("空函数名")
	}
	b := append([]byte(name), 0)
	addr, err := obfuscate.CallAPI(
		obfuscate.GetKernel32(),
		obfuscate.GetGetProcAddress(),
		mod,
		uintptr(unsafe.Pointer(&b[0])),
	)
	if addr == 0 {
		return 0, fmt.Errorf("GetProcAddress(%s) 失败: %v", name, err)
	}
	return addr, nil
}

func getProcAddressOrdinal(mod uintptr, ordinal uint16) (uintptr, error) {
	addr, err := obfuscate.CallAPI(
		obfuscate.GetKernel32(),
		obfuscate.GetGetProcAddress(),
		mod,
		uintptr(ordinal),
	)
	if addr == 0 {
		return 0, fmt.Errorf("GetProcAddress(ordinal=%d) 失败: %v", ordinal, err)
	}
	return addr, nil
}

// bytesReader 从给定偏移创建只读 Reader，便于 binary.Read
func bytesReader(b []byte, offset int) *bytes.Reader {
	if offset < 0 || offset > len(b) {
		return bytes.NewReader(nil)
	}
	return bytes.NewReader(b[offset:])
}

func rvaToString(base uintptr, rva uintptr) string {
	if rva == 0 {
		return ""
	}

	ptr := base + rva
	// 读取 null 结尾字符串
	var buf []byte
	for {
		ch := *(*byte)(unsafe.Pointer(ptr))
		if ch == 0 {
			break
		}
		buf = append(buf, ch)
		ptr++
	}
	return string(buf)
}

// applySectionProtections 根据节属性设置内存保护
func applySectionProtections(base uintptr, sections []imageSectionHeader) error {
	for i := range sections {
		sec := &sections[i]
		if sec.VirtualSize == 0 {
			continue
		}

		// 计算该节的地址和大小
		sectionAddr := base + uintptr(sec.VirtualAddress)
		sectionSize := uintptr(sec.VirtualSize)

		// 根据节特征确定保护属性
		var protect uint32
		characteristics := sec.Characteristics

		isExecutable := (characteristics & imageSCNMemExecute) != 0
		isWritable := (characteristics & imageSCNMemWrite) != 0
		isReadable := (characteristics & imageSCNMemRead) != 0

		if isExecutable {
			if isWritable {
				protect = pageExecuteReadWrite
			} else {
				protect = pageExecuteRead // 可执行节通常是只读+执行
			}
		} else if isWritable {
			protect = pageReadWrite
		} else if isReadable {
			protect = pageReadOnly
		} else {
			protect = pageReadOnly // 默认只读
		}

		// 调用 VirtualProtect 修改保护属性
		var oldProtect uint32
		ret, err := obfuscate.CallAPI(
			obfuscate.GetKernel32(),
			obfuscate.GetVirtualProtect(),
			sectionAddr,
			sectionSize,
			uintptr(protect),
			uintptr(unsafe.Pointer(&oldProtect)),
		)

		if ret == 0 {
			return fmt.Errorf("VirtualProtect 失败 (Section[%d]): %v", i, err)
		}
	}

	return nil
}
