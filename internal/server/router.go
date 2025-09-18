package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ipservice/internal/ipdb"
)

// NewRouter 构建 Gin 引擎并注册全部路由。
func NewRouter(service *ipdb.Service) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// 文档路由（根路径展示 API 文档）
	registerDocRoutes(router)

	handler := &handler{service: service}

	router.GET("/health", handler.health)
	router.GET("/ip", handler.queryByClient)
	router.GET("/ip/:ip", handler.queryByPath)
	router.POST("/ip", handler.queryByBody)

	return router
}

// handler 组合领域服务，对外提供 HTTP 处理逻辑。
type handler struct {
	service *ipdb.Service
}

type ipRequest struct {
	IP string `json:"ip" binding:"required,ip"`
}

type ipResponse struct {
	IP      string   `json:"ip"`
	Country string   `json:"country"`
	Area    string   `json:"area"`
	Raw     []string `json:"raw"`
}

func (h *handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *handler) queryByPath(c *gin.Context) {
	ip := c.Param("ip")
	h.lookup(c, ip)
}

func (h *handler) queryByBody(c *gin.Context) {
	var req ipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.lookup(c, req.IP)
}

// queryByClient 根据客户端来源 IP 查询归属信息，便于直接访问接口自检。
func (h *handler) queryByClient(c *gin.Context) {
	ip, err := extractClientIPv4(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.lookup(c, ip)
}

func (h *handler) lookup(c *gin.Context, ip string) {
	result, err := h.service.Lookup(ip)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ipdb.ErrInvalidIP),
			errors.Is(err, ipdb.ErrIPv6NotSupported),
			errors.Is(err, ipdb.ErrDecodeCountry),
			errors.Is(err, ipdb.ErrDecodeArea):
			status = http.StatusBadRequest
		case errors.Is(err, ipdb.ErrNotFound):
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": toUserMessage(err)})
		return
	}

	resp := ipResponse{
		IP:      result.IP,
		Country: result.Country,
		Area:    result.Area,
		Raw:     []string{result.Country, result.Area},
	}
	c.JSON(http.StatusOK, resp)
}

func extractClientIPv4(c *gin.Context) (string, error) {
	if ip := pickIPv4(c.GetHeader("X-Forwarded-For")); ip != "" {
		return ip, nil
	}
	if ip := pickIPv4(c.GetHeader("X-Real-IP")); ip != "" {
		return ip, nil
	}

	raw := c.ClientIP()
	if raw == "" {
		return "", errors.New("无法识别客户端IP")
	}
	if ip := parseIPv4(raw); ip != "" {
		return ip, nil
	}

	return "", fmt.Errorf("当前qqwry.dat仅支持IPv4查询，检测到: %s", raw)
}

func pickIPv4(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.Split(header, ",")
	for _, part := range parts {
		if ip := parseIPv4(part); ip != "" {
			return ip
		}
	}
	return ""
}

func parseIPv4(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return ""
	}
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	return ""
}

// toUserMessage 将内部错误标准化为面向用户的中文提示，避免泄露内部标识。
func toUserMessage(err error) string {
	switch {
	case errors.Is(err, ipdb.ErrInvalidIP):
		return "无法解析 IP"
	case errors.Is(err, ipdb.ErrIPv6NotSupported):
		return "当前仅支持 IPv4 查询"
	case errors.Is(err, ipdb.ErrNotFound):
		return "未找到 IP 的归属信息"
	case errors.Is(err, ipdb.ErrDecodeCountry):
		return "国家字段编码转换失败"
	case errors.Is(err, ipdb.ErrDecodeArea):
		return "区域字段编码转换失败"
	default:
		return err.Error()
	}
}

