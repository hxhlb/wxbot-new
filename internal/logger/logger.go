package logger

import (
    "io"
    "log"
    "os"
)

// Setup 初始化日志到文件和标准输出
func Setup(filepath string) error {
    f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return err
    }
    mw := io.MultiWriter(os.Stdout, f)
    log.SetOutput(mw)
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    return nil
}

