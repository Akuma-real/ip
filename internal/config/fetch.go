package config

import (
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

// ensureQQWryFile 确保本地存在 qqwry.dat；若不存在且提供了 URL，则尝试下载。
func ensureQQWryFile(path, url string) error {
    if path == "" {
        return errors.New("qqwry.dat 路径不能为空")
    }
    if _, err := os.Stat(path); err == nil {
        return nil
    } else if !errors.Is(err, os.ErrNotExist) {
        return fmt.Errorf("检查qqwry.dat失败: %w", err)
    }

    if url == "" {
        return fmt.Errorf("缺少下载地址，请设置 %s 或手动放置数据文件", envQQwryURL)
    }

    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return fmt.Errorf("创建数据目录失败: %w", err)
    }

    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        return fmt.Errorf("构造下载请求失败: %w", err)
    }
    req.Header.Set("User-Agent", "ipservice/1.0 (+https://github.com)")

    client := &http.Client{Timeout: 60 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("下载qqwry.dat失败: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("下载qqwry.dat失败: HTTP %d", resp.StatusCode)
    }

    tmp, err := os.CreateTemp(dir, "qqwry-*.dat.tmp")
    if err != nil {
        return fmt.Errorf("创建临时文件失败: %w", err)
    }
    defer func() {
        tmp.Close()
        os.Remove(tmp.Name())
    }()

    if _, err := io.Copy(tmp, resp.Body); err != nil {
        return fmt.Errorf("写入临时文件失败: %w", err)
    }
    if err := tmp.Close(); err != nil {
        return fmt.Errorf("关闭临时文件失败: %w", err)
    }

    if err := os.Rename(tmp.Name(), path); err != nil {
        return fmt.Errorf("移动数据文件失败: %w", err)
    }
    return nil
}
