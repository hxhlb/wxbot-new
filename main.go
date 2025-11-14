package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wxbot-new/internal/api"
	"wxbot-new/internal/config"
	"wxbot-new/internal/memory"
	"wxbot-new/internal/service"
)

func main() {
	// 配置日志
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)

	log.Println("====== 微信机器人服务启动 ======")

	// 1. 加载配置
	configManager := config.NewManager("./config.json")
	if err := configManager.Load(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	cfg := configManager.Get()
	log.Printf("配置加载成功: Host=%s, Port=%d", cfg.Host, cfg.Port)

	// 2. 初始化 HTTP API 服务
	apiServer := api.NewServer(cfg.Host, cfg.Port, configManager)

	// 3. 初始化共享内存
	memManager := memory.NewSharedMemoryManager()
	if err := memManager.CreateAndWriteSharedMemory(); err != nil {
		log.Fatalf("创建共享内存失败: %v", err)
	}
	defer memManager.Close()

	log.Println("共享内存创建成功")

	// 等待3秒
	time.Sleep(3 * time.Second)

	// 4. 配置DLL路径
	loaderPath := "./NoveLoader.dll"
	dllPath := "./NoveHelper.dll"

	// 5. 创建微信服务
	wechatService := service.NewWeChatService(loaderPath, dllPath)

	// 6. 将微信服务实例传递给 API Server
	apiServer.SetWeChatService(wechatService)

	// 启动 HTTP API 服务
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Printf("HTTP API 服务异常: %v", err)
		}
	}()

	// 7. 设置信号处理器
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 8. 在goroutine中启动服务
	go func() {
		if err := wechatService.Start(); err != nil {
			log.Printf("启动微信服务失败: %v", err)
		}
	}()

	// 9. 等待信号
	sig := <-sigChan
	log.Printf("收到信号 %v，准备停止服务...", sig)

	// 10. 停止所有服务
	apiServer.Stop()
	wechatService.Stop()

	log.Println("====== 微信机器人服务已停止 ======")
}
