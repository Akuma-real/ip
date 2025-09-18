package main

import (
    "log"

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

    log.Printf("服务启动，监听地址: %s，数据源: %s", cfg.ListenAddr, cfg.QQWryPath)
    if err := router.Run(cfg.ListenAddr); err != nil {
        log.Fatalf("服务运行异常: %v", err)
    }
}
