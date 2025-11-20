package obfuscate

import (
	"unsafe"
)

// SplitMemoryAllocation 分段申请内存 (避免大块内存特征)
// 将大块内存申请分成多次小块申请,降低检测风险
func SplitMemoryAllocation(totalSize uintptr, chunkSize uintptr) []uintptr {
	if chunkSize == 0 || chunkSize > totalSize {
		chunkSize = totalSize
	}

	chunks := make([]uintptr, 0)
	remaining := totalSize

	for remaining > 0 {
		allocSize := chunkSize
		if allocSize > remaining {
			allocSize = remaining
		}

		// 这里返回建议的分配大小,实际申请由调用者完成
		chunks = append(chunks, allocSize)
		remaining -= allocSize

		// 混淆：每次分配间隔随机延迟
		JitterDelay(10)
	}

	return chunks
}

// RandomizeMemoryLayout 随机化内存布局
// 在内存中插入随机填充,破坏固定模式
func RandomizeMemoryLayout(baseAddr uintptr, size uintptr) {
	if baseAddr == 0 || size == 0 {
		return
	}

	// 在内存的空闲区域写入随机数据
	// 注意: 只能修改不影响运行的区域
	mem := unsafe.Slice((*byte)(unsafe.Pointer(baseAddr)), int(size))

	// 填充 DOS Stub 区域 (0x40 - 0x100 通常未使用)
	for i := 0x40; i < 0x100 && i < len(mem); i++ {
		mem[i] = byte((i * 7) & 0xFF) // 使用确定性随机值
	}
}
