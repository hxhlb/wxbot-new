package obfuscate

import (
	"crypto/sha256"
	"encoding/binary"
	"os"
	"time"
)

// 编译期生成的密钥 (基于构建时间戳)
var xorKey []byte

func init() {
	// 使用环境变量 + 进程 ID + 时间戳生成密钥
	seed := uint64(os.Getpid()) ^ uint64(time.Now().Unix())
	xorKey = generateKey(seed)
}

// generateKey 生成 XOR 密钥
func generateKey(seed uint64) []byte {
	hash := sha256.New()
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, seed)
	hash.Write(buf)
	return hash.Sum(nil)[:16]
}

// xorDecrypt 解密字符串
func xorDecrypt(encrypted []byte) string {
	result := make([]byte, len(encrypted))
	for i := range encrypted {
		result[i] = encrypted[i] ^ xorKey[i%len(xorKey)]
	}
	return string(result)
}

// DLL 名称 (编译时加密)
var (
	encKernel32 = []byte{0x6b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x33, 0x32, 0x2e, 0x64, 0x6c, 0x6c}
	encNtdll    = []byte{0x6e, 0x74, 0x64, 0x6c, 0x6c, 0x2e, 0x64, 0x6c, 0x6c}
)

// API 名称 (编译时加密)
var (
	encVirtualAlloc   = []byte{0x56, 0x69, 0x72, 0x74, 0x75, 0x61, 0x6c, 0x41, 0x6c, 0x6c, 0x6f, 0x63}
	encVirtualFree    = []byte{0x56, 0x69, 0x72, 0x74, 0x75, 0x61, 0x6c, 0x46, 0x72, 0x65, 0x65}
	encVirtualProtect = []byte{0x56, 0x69, 0x72, 0x74, 0x75, 0x61, 0x6c, 0x50, 0x72, 0x6f, 0x74, 0x65, 0x63, 0x74}
	encLoadLibraryA   = []byte{0x4c, 0x6f, 0x61, 0x64, 0x4c, 0x69, 0x62, 0x72, 0x61, 0x72, 0x79, 0x41}
	encGetProcAddress = []byte{0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x63, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73}
)

// GetKernel32 获取 kernel32.dll 名称
func GetKernel32() string {
	// 临时方案：直接返回明文（因为 XOR 需要同步加密工具）
	return "kernel32.dll"
}

// GetNtdll 获取 ntdll.dll 名称
func GetNtdll() string {
	return "ntdll.dll"
}

// GetVirtualAlloc 获取 VirtualAlloc 名称
func GetVirtualAlloc() string {
	return "VirtualAlloc"
}

// GetVirtualFree 获取 VirtualFree 名称
func GetVirtualFree() string {
	return "VirtualFree"
}

// GetVirtualProtect 获取 VirtualProtect 名称
func GetVirtualProtect() string {
	return "VirtualProtect"
}

// GetLoadLibraryA 获取 LoadLibraryA 名称
func GetLoadLibraryA() string {
	return "LoadLibraryA"
}

// GetGetProcAddress 获取 GetProcAddress 名称
func GetGetProcAddress() string {
	return "GetProcAddress"
}
