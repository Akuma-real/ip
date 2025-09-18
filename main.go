package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "ipservice/internal/config"
    "ipservice/internal/ipdb"
    "ipservice/internal/server"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("配置加载失败: %v", err)
    }

    svc, err := ipdb.NewService(cfg.QQWryPath)
    if err != nil {
        log.Fatalf("初始化qqwry服务失败: %v", err)
    }

    router := server.NewRouter(svc)

    srv := &http.Server{
        Addr:              cfg.ListenAddr,
        Handler:           router,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       10 * time.Second,
        WriteTimeout:      15 * time.Second,
        IdleTimeout:       60 * time.Second,
    }

    // 启动 HTTP 服务
    log.Printf("服务启动，监听地址: %s，数据源: %s", cfg.ListenAddr, cfg.QQWryPath)
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("服务运行异常: %v", err)
        }
    }()

    // 监听系统信号，执行优雅关停
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    <-stop

    log.Printf("接收到退出信号，开始优雅关停...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("优雅关停失败，强制退出: %v", err)
        if err := srv.Close(); err != nil {
            log.Printf("强制关闭失败: %v", err)
        }
    }
    log.Printf("服务已关闭")
}