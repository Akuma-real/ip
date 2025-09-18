package config

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

const (
    envListen     = "IP_API_LISTEN"
    envQQwryPath  = "IP_API_QQWRY_PATH"
    envQQwryURL   = "IP_API_QQWRY_URL"
    envAutoFetch  = "IP_API_AUTO_FETCH"

    defaultListen    = ":8080"
    defaultData      = "qqwry.dat"
    defaultDataURL   = "https://github.com/metowolf/qqwry.dat/releases/latest/download/qqwry.dat"
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

    // 若启用自动获取，则在校验前尝试从远端下载缺失的数据文件
    if isTruthy(getOrDefault(envAutoFetch, "true")) {
        if err := ensureQQWryFile(cfg.QQWryPath, getOrDefault(envQQwryURL, defaultDataURL)); err != nil {
            return nil, err
        }
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

func isTruthy(v string) bool {
    switch strings.ToLower(strings.TrimSpace(v)) {
    case "1", "true", "yes", "on", "y", "t":
        return true
    default:
        return false
    }
}
