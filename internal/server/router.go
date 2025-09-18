package server

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"

    "ipservice/internal/ipdb"
)

// NewRouter 构建 Gin 引擎并注册全部路由。
func NewRouter(service *ipdb.Service) *gin.Engine {
    router := gin.New()
    router.Use(gin.Logger(), gin.Recovery())

    handler := &handler{service: service}

    router.GET("/health", handler.health)
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
    IP       string `json:"ip"`
    Country  string `json:"country"`
    Area     string `json:"area"`
    Raw      []string `json:"raw"`
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

func (h *handler) lookup(c *gin.Context, ip string) {
    result, err := h.service.Lookup(ip)
    if err != nil {
        status := http.StatusInternalServerError
        switch {
        case isBadRequest(err):
            status = http.StatusBadRequest
        case isNotFound(err):
            status = http.StatusNotFound
        }
        c.JSON(status, gin.H{"error": err.Error()})
        return
    }

    resp := ipResponse{
        IP:      result.IP,
        Country: result.Country,
        Area:    result.Area,
        Raw: []string{result.Country, result.Area},
    }
    c.JSON(http.StatusOK, resp)
}

func isBadRequest(err error) bool {
    if err == nil {
        return false
    }
    return containsAny(err.Error(), "无法解析IP", "仅支持IPv4", "编码转换失败")
}

func isNotFound(err error) bool {
    if err == nil {
        return false
    }
    return containsAny(err.Error(), "未找到IP")
}

func containsAny(msg string, subs ...string) bool {
    for _, sub := range subs {
        if sub != "" && strings.Contains(msg, sub) {
            return true
        }
    }
    return false
}
