package ipdb

import "errors"

// 领域错误：用于跨包判定与 HTTP 映射
var (
    ErrInvalidIP        = errors.New("invalid ip")
    ErrIPv6NotSupported = errors.New("ipv6 not supported")
    ErrNotFound         = errors.New("ip not found")
    ErrDecodeCountry    = errors.New("decode country failed")
    ErrDecodeArea       = errors.New("decode area failed")
)
