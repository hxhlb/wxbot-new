package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wxbot-new/internal/api"
	"wxbot-new/internal/config"
	"wxbot-new/internal/logging"
	"wxbot-new/internal/memory"
	"wxbot-new/internal/service"
)

func main() {
	// 配置全局日志, 统一写入按天命名的文件
	if err := logging.Init(); err != nil {
		// 日志初始化失败时, 同时输出到控制台和默认日志, 然后退出
		logBothFatalf("初始化日志失败: %v", err)
	}

	logBothln("====== WxBot 服务启动 By: Ripper ======")
	logBothln("====== 版本: v0.0.3 ======\n")
	logBothln("====== 基于微信版本 v4.1.2.17 ======\n")

	// 1. 加载配置
	configManager := config.NewManager("./config.json")
	if err := configManager.Load(); err != nil {
		logBothFatalf("加载配置失败: %v", err)
	}
	cfg := configManager.Get()
	logBothf("配置加载成功: Host=%s, Port=%d", cfg.Host, cfg.Port)

	// 2. 初始化 HTTP API 服务
	apiServer := api.NewServer(cfg.Host, cfg.Port, configManager)

	// 3. 初始化共享内存
	memManager := memory.NewSharedMemoryManager()
	if err := memManager.CreateAndWriteSharedMemory(); err != nil {
		logBothFatalf("创建共享内存失败: %v", err)
	}
	defer memManager.Close()

	logBothln("共享内存创建成功")

	// 等待3秒
	time.Sleep(3 * time.Second)

	// 4. 配置DLL路径
	loaderPath := "./NoveLoader.dll"
	dllPath := "./NoveHelper.dll"

	// 5. 创建微信服务
	wechatService := service.NewWeChatService(loaderPath, dllPath, cfg.LogRecvCallback, cfg.CallbackURLs)

	// 6. 将微信服务实例传递给 API Server
	apiServer.SetWeChatService(wechatService)

	// 启动 HTTP API 服务
	go func() {
		if err := apiServer.Start(); err != nil {
			logBothf("HTTP API 服务异常: %v", err)
		}
	}()

	// 7. 设置信号处理器
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 8. 在goroutine中启动服务
	go func() {
		if err := wechatService.Start(); err != nil {
			logBothf("启动微信服务失败: %v", err)
		}
	}()

	// 9. 等待信号
	sig := <-sigChan
	logBothf("收到信号 %v，准备停止服务...", sig)

	// 10. 停止所有服务
	apiServer.Stop()
	wechatService.Stop()

	logBothln("\n\n====== WxBot 服务已停止 ======")
}

// logBothln 在日志文件和控制台同时输出一行日志
func logBothln(v ...interface{}) {
	log.Println(v...)
	fmt.Println(v...)
}

// logBothf 在日志文件和控制台同时输出格式化日志
func logBothf(format string, v ...interface{}) {
	log.Printf(format, v...)
	fmt.Printf(format+"\n", v...)
}

// logBothFatalf 在日志文件和控制台输出格式化日志后退出程序
func logBothFatalf(format string, v ...interface{}) {
	logBothf(format, v...)
	os.Exit(1)
}
