package obfuscate

import (
	"unsafe"
)

// ErasePEHeader 擦除 PE 头特征 (防止内存扫描)
// 在 DLL 成功加载后调用,破坏 MZ/PE 签名
func ErasePEHeader(baseAddr uintptr) {
	if baseAddr == 0 {
		return
	}

	// 擦除 DOS 头的 MZ 签名
	dosHeader := (*uint16)(unsafe.Pointer(baseAddr))
	*dosHeader = 0x0000 // 原本是 0x5A4D ("MZ")

	// 擦除 PE 签名 (通常在偏移 0x3C 处指向的位置)
	e_lfanew := (*int32)(unsafe.Pointer(baseAddr + 0x3C))
	if *e_lfanew > 0 && *e_lfanew < 0x1000 {
		peSignature := (*uint32)(unsafe.Pointer(baseAddr + uintptr(*e_lfanew)))
		*peSignature = 0x00000000 // 原本是 0x00004550 ("PE\0\0")
	}

	// 可选: 擦除整个 DOS Stub (前 64 字节)
	// 但需要确保不影响运行时重定位等操作
}

// RestorePEHeader 恢复 PE 头 (如需卸载 DLL)
func RestorePEHeader(baseAddr uintptr) {
	if baseAddr == 0 {
		return
	}

	// 恢复 MZ 签名
	dosHeader := (*uint16)(unsafe.Pointer(baseAddr))
	*dosHeader = 0x5A4D

	// 恢复 PE 签名
	e_lfanew := (*int32)(unsafe.Pointer(baseAddr + 0x3C))
	if *e_lfanew > 0 && *e_lfanew < 0x1000 {
		peSignature := (*uint32)(unsafe.Pointer(baseAddr + uintptr(*e_lfanew)))
		*peSignature = 0x00004550
	}
}
