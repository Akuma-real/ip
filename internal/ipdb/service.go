package ipdb

import (
    "fmt"
    "strings"
    "sync"
)

// Result 表示一次 IP 查询的标准化结果。
type Result struct {
    IP      string
    Country string
    Area    string
}

// Service 管理 qqwry 数据的加载与查询，并提供线程安全的对外接口。
type Service struct {
    path string

    mu     sync.RWMutex
    reader *qqwryReader
}

// NewService 创建服务实例并加载数据。
func NewService(path string) (*Service, error) {
    reader, err := newReader(path)
    if err != nil {
        return nil, err
    }
    return &Service{path: path, reader: reader}, nil
}

// Lookup 返回指定 IP 的归属地信息。
func (s *Service) Lookup(ip string) (Result, error) {
    s.mu.RLock()
    reader := s.reader
    s.mu.RUnlock()

    if reader == nil {
        return Result{}, fmt.Errorf("qqwry 数据尚未加载")
    }

    countryRaw, areaRaw, err := reader.lookupRaw(ip)
    if err != nil {
        return Result{}, err
    }

    country, err := decodeGBK(countryRaw)
    if err != nil {
        return Result{}, fmt.Errorf("%w: 国家字段编码转换失败: %v", ErrDecodeCountry, err)
    }
    area, err := decodeGBK(areaRaw)
    if err != nil {
        return Result{}, fmt.Errorf("%w: 区域字段编码转换失败: %v", ErrDecodeArea, err)
    }

    country = normalize(country)
    area = normalize(area)

    return Result{
        IP:      ip,
        Country: country,
        Area:    area,
    }, nil
}

// Reload 重新加载数据文件，便于热更新。
func (s *Service) Reload() error {
    reader, err := newReader(s.path)
    if err != nil {
        return err
    }

    s.mu.Lock()
    s.reader = reader
    s.mu.Unlock()
    return nil
}

func normalize(value string) string {
    cleaned := strings.TrimSpace(value)
    if cleaned == "" || strings.EqualFold(cleaned, "CZ88.NET") {
        return ""
    }
    return cleaned
}
