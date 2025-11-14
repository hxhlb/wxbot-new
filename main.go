package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wxbot-new/internal/memory"
	"wxbot-new/internal/service"
)

func main() {
	// 配置日志
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)

	log.Println("====== 微信机器人服务启动 ======")

	// 1. 初始化共享内存
	memManager := memory.NewSharedMemoryManager()
	if err := memManager.CreateAndWriteSharedMemory(); err != nil {
		log.Fatalf("创建共享内存失败: %v", err)
	}
	defer memManager.Close()

	log.Println("共享内存创建成功")

	// 等待3秒
	time.Sleep(3 * time.Second)

	// 2. 配置DLL路径
	loaderPath := "./NoveLoader.dll"
	dllPath := "./NoveHelper.dll"

	// 3. 创建微信服务
	wechatService := service.NewWeChatService(loaderPath, dllPath)

	// 4. 设置信号处理器
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 5. 在goroutine中启动服务
	go func() {
		if err := wechatService.Start(); err != nil {
			log.Printf("启动微信服务失败: %v", err)
		}
	}()

	// 6. 等待信号
	sig := <-sigChan
	log.Printf("收到信号 %v，准备停止服务...", sig)

	// 7. 停止服务
	wechatService.Stop()

	log.Println("====== 微信机器人服务已停止 ======")
}
