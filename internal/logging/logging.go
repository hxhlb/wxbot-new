package logging

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// Init 初始化全局日志配置:
// - 日志格式: 日期、时间、短文件名
// - 输出位置: ./logs/wxbot-YYYY-MM-DD.log
// - 不再输出到控制台
func Init() error {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	filename := "wxbot-" + time.Now().Format("2006-01-02") + ".log"
	logPath := filepath.Join(logDir, filename)

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// 不关闭 file, 由进程退出时操作系统回收, 避免在运行期间出现写入无效的问题
	log.SetOutput(file)

	return nil
}
