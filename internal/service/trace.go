package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

var traceCounter uint64

// GenerateTraceID 生成唯一的追踪ID
// 格式: timestamp-counter-random
func GenerateTraceID() string {
	// 时间戳(秒)
	timestamp := time.Now().Unix()

	// 递增计数器
	counter := atomic.AddUint64(&traceCounter, 1)

	// 4字节随机数
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("%d-%d-%s", timestamp, counter, randomHex)
}
