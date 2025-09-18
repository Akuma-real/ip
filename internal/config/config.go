package config

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"
)

const (
    envListen     = "IP_API_LISTEN"
    envQQwryPath  = "IP_API_QQWRY_PATH"
    defaultListen = ":8080"
    defaultData   = "qqwry.dat"
)

// Config 表示服务运行时所需的核心配置。
type Config struct {
    ListenAddr string
    QQWryPath  string
}

// Load 从环境变量读取配置并补全默认值，同时校验关键依赖是否存在。
func Load() (*Config, error) {
    cfg := &Config{
        ListenAddr: getOrDefault(envListen, defaultListen),
        QQWryPath:  resolvePath(getOrDefault(envQQwryPath, defaultData)),
    }

    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    return cfg, nil
}

// Validate 对配置进行完整性校验。
func (c *Config) Validate() error {
    if c.ListenAddr == "" {
        return errors.New("监听地址不能为空")
    }
    if c.QQWryPath == "" {
        return errors.New("qqwry.dat 路径不能为空")
    }
    if _, err := os.Stat(c.QQWryPath); err != nil {
        if errors.Is(err, os.ErrNotExist) {
            return fmt.Errorf("未找到qqwry.dat: %s", c.QQWryPath)
        }
        return fmt.Errorf("无法读取qqwry.dat: %w", err)
    }
    return nil
}

func getOrDefault(key, def string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return def
}

func resolvePath(p string) string {
    if filepath.IsAbs(p) {
        return p
    }
    abs, err := filepath.Abs(p)
    if err != nil {
        return p
    }
    return abs
}
