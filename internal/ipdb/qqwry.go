package ipdb

import (
    "bytes"
    "encoding/binary"
    "errors"
    "fmt"
    "io"
    "net"
    "os"
    "sync"

    "golang.org/x/text/encoding/simplifiedchinese"
    "golang.org/x/text/transform"
)

const (
    indexEntryLen   = 7
    redirectMode1   = 0x01
    redirectMode2   = 0x02
)

// qqwryReader 封装 qqwry.dat 的二进制读取逻辑，保持线程安全。
type qqwryReader struct {
    mu   sync.RWMutex
    data []byte
}

// newReader 从指定路径加载 qqwry 数据文件。
func newReader(path string) (*qqwryReader, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("读取qqwry.dat失败: %w", err)
    }
    if len(data) < 8 {
        return nil, errors.New("qqwry.dat 文件格式不合法")
    }
    return &qqwryReader{data: data}, nil
}

// lookupRaw 返回原始的国家与区域字段（GBK 编码）。
func (r *qqwryReader) lookupRaw(ipStr string) ([]byte, []byte, error) {
    ip := net.ParseIP(ipStr)
    if ip == nil {
        return nil, nil, fmt.Errorf("%w: 无法解析IP: %s", ErrInvalidIP, ipStr)
    }
    ipv4 := ip.To4()
    if ipv4 == nil {
        return nil, nil, fmt.Errorf("%w: 当前qqwry.dat仅支持IPv4查询", ErrIPv6NotSupported)
    }

    target := binary.BigEndian.Uint32(ipv4)

    r.mu.RLock()
    defer r.mu.RUnlock()

    data := r.data
    indexStart := binary.LittleEndian.Uint32(data[:4])
    indexEnd := binary.LittleEndian.Uint32(data[4:8])
    if indexEnd <= indexStart {
        return nil, nil, errors.New("qqwry索引区异常")
    }

    total := (indexEnd-indexStart)/indexEntryLen + 1
    var recordOffset uint32

    left, right := uint32(0), total-1
    for left <= right {
        mid := (left + right) >> 1
        offset := indexStart + mid*indexEntryLen
        if int(offset)+indexEntryLen > len(data) {
            break
        }
        startIP := binary.LittleEndian.Uint32(data[offset : offset+4])
        record := offset + 4
        recordOffset = readUint24(data[record : record+3])
        if int(recordOffset)+4 > len(data) {
            break
        }
        endIP := binary.LittleEndian.Uint32(data[recordOffset : recordOffset+4])

        switch {
        case target < startIP:
            if mid == 0 {
                right = 0
                break
            }
            right = mid - 1
        case target > endIP:
            left = mid + 1
        default:
            goto FOUND
        }
    }
    return nil, nil, fmt.Errorf("%w: 未找到IP %s 的归属信息", ErrNotFound, ipStr)

FOUND:
    country, area := r.readRecord(recordOffset)
    if country == nil && area == nil {
        return nil, nil, errors.New("qqwry 记录解析失败")
    }
    return country, area, nil
}

func (r *qqwryReader) readRecord(offset uint32) ([]byte, []byte) {
    data := r.data
    if int(offset)+4 >= len(data) {
        return nil, nil
    }
    mode := data[offset+4]
    switch mode {
    case redirectMode1:
        if int(offset)+8 > len(data) {
            return nil, nil
        }
        countryOffset := readUint24(data[offset+5 : offset+8])
        if int(countryOffset) >= len(data) {
            return nil, nil
        }
        redirectMode := data[countryOffset]
        if redirectMode == redirectMode2 {
            if int(countryOffset)+4 > len(data) {
                return nil, nil
            }
            realOffset := readUint24(data[countryOffset+1 : countryOffset+4])
            country, _ := readCString(data, realOffset)
            area, _ := r.readArea(countryOffset + 4)
            return country, area
        }
        country, next := readCString(data, countryOffset)
        area, _ := r.readArea(next)
        return country, area
    case redirectMode2:
        if int(offset)+8 > len(data) {
            return nil, nil
        }
        countryOffset := readUint24(data[offset+5 : offset+8])
        country, _ := readCString(data, countryOffset)
        area, _ := r.readArea(offset + 8)
        return country, area
    default:
        country, next := readCString(data, offset+4)
        area, _ := r.readArea(next)
        return country, area
    }
}

func (r *qqwryReader) readArea(offset uint32) ([]byte, uint32) {
    data := r.data
    if int(offset) >= len(data) {
        return nil, offset
    }
    mode := data[offset]
    if mode == redirectMode1 || mode == redirectMode2 {
        if int(offset)+4 > len(data) {
            return nil, offset
        }
        areaOffset := readUint24(data[offset+1 : offset+4])
        if areaOffset == 0 {
            return nil, offset + 4
        }
        return readCString(data, areaOffset)
    }
    return readCString(data, offset)
}

func readCString(data []byte, offset uint32) ([]byte, uint32) {
    if int(offset) >= len(data) {
        return nil, offset
    }
    i := offset
    for i < uint32(len(data)) && data[i] != 0 {
        i++
    }
    return data[offset:i], i + 1
}

func readUint24(buf []byte) uint32 {
    if len(buf) < 3 {
        return 0
    }
    var val uint32
    val = uint32(buf[0])
    val |= uint32(buf[1]) << 8
    val |= uint32(buf[2]) << 16
    return val
}

func decodeGBK(b []byte) (string, error) {
    if len(b) == 0 {
        return "", nil
    }
    reader := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GBK.NewDecoder())
    converted, err := io.ReadAll(reader)
    if err != nil {
        return "", err
    }
    return string(converted), nil
}
