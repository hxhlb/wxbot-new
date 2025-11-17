package loader

import (
	"fmt"
	"unsafe"
)

// PE文件常量
const (
	IMAGE_DOS_SIGNATURE      = 0x5A4D     // MZ
	IMAGE_NT_SIGNATURE       = 0x00004550 // PE00
	IMAGE_FILE_MACHINE_I386  = 0x014c
	IMAGE_FILE_MACHINE_AMD64 = 0x8664
)

// DOS Header
type ImageDOSHeader struct {
	Magic  uint16
	_      [29]uint16
	LfaNew int32
}

// NT Headers (32位)
type ImageNTHeaders32 struct {
	Signature      uint32
	FileHeader     ImageFileHeader
	OptionalHeader ImageOptionalHeader32
}

// File Header
type ImageFileHeader struct {
	Machine              uint16
	NumberOfSections     uint16
	TimeDateStamp        uint32
	PointerToSymbolTable uint32
	NumberOfSymbols      uint32
	SizeOfOptionalHeader uint16
	Characteristics      uint16
}

// Optional Header (32位)
type ImageOptionalHeader32 struct {
	Magic                       uint16
	MajorLinkerVersion          uint8
	MinorLinkerVersion          uint8
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
	DataDirectory               [16]ImageDataDirectory
}

// Data Directory
type ImageDataDirectory struct {
	VirtualAddress uint32
	Size           uint32
}

// Section Header
type ImageSectionHeader struct {
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

// Import Descriptor
type ImageImportDescriptor struct {
	OriginalFirstThunk uint32
	TimeDateStamp      uint32
	ForwarderChain     uint32
	Name               uint32
	FirstThunk         uint32
}

// Base Relocation
type ImageBaseRelocation struct {
	VirtualAddress uint32
	SizeOfBlock    uint32
}

// PE信息
type PEInfo struct {
	ImageBase       uintptr
	EntryPoint      uint32
	SizeOfImage     uint32
	SizeOfHeaders   uint32
	Sections        []SectionInfo
	ImportDirectory *ImageDataDirectory
	RelocDirectory  *ImageDataDirectory
	Is64Bit         bool
}

// Section信息
type SectionInfo struct {
	Name             string
	VirtualAddress   uint32
	VirtualSize      uint32
	PointerToRawData uint32
	SizeOfRawData    uint32
	Characteristics  uint32
	Data             []byte
}

// PEParser PE解析器
type PEParser struct {
	data []byte
}

// NewPEParser 创建PE解析器
func NewPEParser(data []byte) *PEParser {
	return &PEParser{data: data}
}

// Parse 解析PE文件
func (p *PEParser) Parse() (*PEInfo, error) {
	if len(p.data) < int(unsafe.Sizeof(ImageDOSHeader{})) {
		return nil, fmt.Errorf("文件太小，不是有效的PE文件")
	}

	// 解析DOS头
	dosHeader := (*ImageDOSHeader)(unsafe.Pointer(&p.data[0]))
	if dosHeader.Magic != IMAGE_DOS_SIGNATURE {
		return nil, fmt.Errorf("无效的DOS签名: 0x%X", dosHeader.Magic)
	}

	// 解析NT头
	ntHeaderOffset := dosHeader.LfaNew
	if ntHeaderOffset < 0 || int(ntHeaderOffset) >= len(p.data) {
		return nil, fmt.Errorf("无效的NT头偏移: %d", ntHeaderOffset)
	}

	ntHeaders := (*ImageNTHeaders32)(unsafe.Pointer(&p.data[ntHeaderOffset]))
	if ntHeaders.Signature != IMAGE_NT_SIGNATURE {
		return nil, fmt.Errorf("无效的PE签名: 0x%X", ntHeaders.Signature)
	}

	// 检查是否为32位
	if ntHeaders.FileHeader.Machine != IMAGE_FILE_MACHINE_I386 {
		return nil, fmt.Errorf("仅支持32位DLL (x86), 当前架构: 0x%X", ntHeaders.FileHeader.Machine)
	}

	peInfo := &PEInfo{
		ImageBase:     uintptr(ntHeaders.OptionalHeader.ImageBase),
		EntryPoint:    ntHeaders.OptionalHeader.AddressOfEntryPoint,
		SizeOfImage:   ntHeaders.OptionalHeader.SizeOfImage,
		SizeOfHeaders: ntHeaders.OptionalHeader.SizeOfHeaders,
		Is64Bit:       false,
	}

	// 获取导入表和重定位表
	if ntHeaders.OptionalHeader.NumberOfRvaAndSizes > 1 {
		peInfo.ImportDirectory = &ntHeaders.OptionalHeader.DataDirectory[1] // IMAGE_DIRECTORY_ENTRY_IMPORT
	}
	if ntHeaders.OptionalHeader.NumberOfRvaAndSizes > 5 {
		peInfo.RelocDirectory = &ntHeaders.OptionalHeader.DataDirectory[5] // IMAGE_DIRECTORY_ENTRY_BASERELOC
	}

	// 解析节表
	sectionHeaderOffset := ntHeaderOffset + 4 + int32(unsafe.Sizeof(ImageFileHeader{})) +
		int32(ntHeaders.FileHeader.SizeOfOptionalHeader)

	for i := uint16(0); i < ntHeaders.FileHeader.NumberOfSections; i++ {
		offset := sectionHeaderOffset + int32(i)*int32(unsafe.Sizeof(ImageSectionHeader{}))
		section := (*ImageSectionHeader)(unsafe.Pointer(&p.data[offset]))

		sectionInfo := SectionInfo{
			Name:             string(section.Name[:]),
			VirtualAddress:   section.VirtualAddress,
			VirtualSize:      section.VirtualSize,
			PointerToRawData: section.PointerToRawData,
			SizeOfRawData:    section.SizeOfRawData,
			Characteristics:  section.Characteristics,
		}

		// 读取节数据
		if section.PointerToRawData > 0 && section.SizeOfRawData > 0 {
			end := section.PointerToRawData + section.SizeOfRawData
			if int(end) <= len(p.data) {
				sectionInfo.Data = p.data[section.PointerToRawData:end]
			}
		}

		peInfo.Sections = append(peInfo.Sections, sectionInfo)
	}

	return peInfo, nil
}

// GetRawData 获取原始数据
func (p *PEParser) GetRawData() []byte {
	return p.data
}

// RVAToOffset 将RVA转换为文件偏移
func (p *PEParser) RVAToOffset(rva uint32, sections []SectionInfo) uint32 {
	for _, section := range sections {
		if rva >= section.VirtualAddress && rva < section.VirtualAddress+section.SizeOfRawData {
			return rva - section.VirtualAddress + section.PointerToRawData
		}
	}
	return rva
}
