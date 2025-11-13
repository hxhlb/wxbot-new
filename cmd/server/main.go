package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"

    "wxbot-new/internal/httpapi"
    "wxbot-new/internal/logger"
    "wxbot-new/internal/wxsvc"
)

func main() {
    // 初始化日志（同时输出到文件和控制台）
    if err := logger.Setup("wechat_service.log"); err != nil {
        log.Fatalf("初始化日志失败: %v", err)
    }

    // 读取配置
    loaderPath := getEnv("LOADER_PATH", "./NoveLoader.dll")
    dllPath := getEnv("DLL_PATH", "./NoveHelper.dll")
    apiHost := getEnv("API_HOST", "0.0.0.0")
    apiPortStr := getEnv("API_PORT", "5000")
    apiPort, _ := strconv.Atoi(apiPortStr)

    // 构建服务（内部会根据平台选择具体实现）
    svc := wxsvc.NewService(loaderPath, dllPath)

    // 启动服务（后台运行）
    go func() {
        if err := svc.Start(); err != nil {
            log.Printf("启动服务失败: %v", err)
        }
    }()

    // HTTP API
    mux := http.NewServeMux()
    httpapi.RegisterRoutes(mux, svc)
    srv := &http.Server{Addr: apiHost + ":" + strconv.Itoa(apiPort), Handler: mux}

    // 优雅退出
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-quit
        log.Println("收到退出信号，正在关闭...")
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = srv.Shutdown(ctx)
        svc.Stop()
    }()

    log.Printf("HTTP API 启动于 http://%s:%d", apiHost, apiPort)
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("HTTP 服务异常: %v", err)
    }
}

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

