package server

import (
    "net/http"
    "os"
    "strings"

    "github.com/gin-gonic/gin"
    blackfriday "github.com/russross/blackfriday/v2"
)

func registerDocRoutes(router *gin.Engine) {
    router.GET("/", func(c *gin.Context) {
        // 动态文档：根据当前 URL 生成可点击链接与 curl 示例
        renderDynamicDocs(c)
    })

    router.GET("/docs", func(c *gin.Context) {
        md, err := os.ReadFile("docs/api_usage.md")
        if err != nil {
            c.String(http.StatusInternalServerError, "读取文档失败: %v", err)
            return
        }
        renderMarkdown(c, "IP 服务 · API 使用说明", md)
    })

}

func renderMarkdown(c *gin.Context, title string, md []byte) {
    html := blackfriday.Run(md)
    page := []byte(`<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>` + title + `</title>
<style>
body{font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;line-height:1.6;margin:0;padding:32px;color:#222;background:#fff}
main{max-width: 860px; margin: 0 auto}
h1,h2,h3{line-height:1.25}
code, pre{background:#f6f8fa;border-radius:6px}
pre{padding:12px; overflow:auto}
code{padding:2px 4px}
a{color:#0969da; text-decoration:none}
a:hover{text-decoration:underline}
hr{border:none;border-top:1px solid #eee;margin:24px 0}
</style>
</head>
<body>
<main>` + string(html) + `<hr/><p>IP地址位置数据由<a href="https://www.cz88.net" target="_blank" rel="noopener noreferrer">纯真CZ88</a>提供支持</p></main>
</body>
</html>`)
    c.Data(http.StatusOK, "text/html; charset=utf-8", page)
}

func baseURL(c *gin.Context) string {
    // 优先读取反向代理头，其次回退到请求信息
    scheme := "http"
    if p := c.GetHeader("X-Forwarded-Proto"); p != "" {
        scheme = strings.TrimSpace(strings.Split(p, ",")[0])
    } else if c.Request.TLS != nil {
        scheme = "https"
    }

    host := c.GetHeader("X-Forwarded-Host")
    if host == "" {
        host = c.Request.Host
    } else {
        host = strings.TrimSpace(strings.Split(host, ",")[0])
    }
    if host == "" {
        host = "localhost"
    }
    return scheme + "://" + host
}

func renderDynamicDocs(c *gin.Context) {
    base := baseURL(c)
    exampleIP := "8.8.8.8"

    // 动态 HTML：展示可点击链接、curl 示例与在线试用
    html := "<!doctype html>" +
        "<html lang=\"zh-CN\"><head><meta charset=\"utf-8\" />" +
        "<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\" />" +
        "<title>IP 服务 · 动态文档</title>" +
        "<style>body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,'Noto Sans','PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;line-height:1.6;margin:0;padding:32px;color:#222}main{max-width:860px;margin:0 auto}code,pre{background:#f6f8fa;border-radius:6px}pre{padding:12px;overflow:auto}code{padding:2px 4px}a{color:#0969da;text-decoration:none}a:hover{text-decoration:underline}input,button{font-size:14px;padding:6px 10px;margin:2px}label{display:block;margin-top:12px}hr{border:none;border-top:1px solid #eee;margin:24px 0}</style>" +
        "</head><body><main>" +
        "<h1>IP 服务 · 动态文档</h1>" +
        "<p>当前基准地址：<code>" + base + "</code></p>" +

        "<h2>快速链接</h2><ul>" +
        "<li><a href='" + base + "/health' target='_blank'>GET /health</a></li>" +
        "<li><a href='" + base + "/ip' target='_blank'>GET /ip</a></li>" +
        "<li><a href='" + base + "/ip/" + exampleIP + "' target='_blank'>GET /ip/" + exampleIP + "</a></li>" +
        "</ul>" +

        "<h2>curl 示例</h2><pre><code>curl -s \"" + base + "/health\"\n" +
        "curl -s \"" + base + "/ip\"\n" +
        "curl -s \"" + base + "/ip/" + exampleIP + "\"\n" +
        "curl -s -H 'Content-Type: application/json' -d '{\"ip\":\"" + exampleIP + "\"}' \"" + base + "/ip\"" +
        "</code></pre>" +

        "<h2>在线试用</h2>" +
        "<div><label>目标 IPv4： <input id='ip' placeholder='例如 8.8.8.8' value='" + exampleIP + "'/></label>" +
        "<button onclick=\"tryGet()\">GET /ip/{ip}</button>" +
        "<button onclick=\"tryPost()\">POST /ip</button></div>" +
        "<pre id='out' style='min-height:120px'></pre>" +

        "<script>const base='" + base + "';\n" +
        "function show(o){document.getElementById('out').textContent=typeof o==='string'?o:JSON.stringify(o,null,2)}\n" +
        "async function tryGet(){const ip=document.getElementById('ip').value.trim(); if(!ip){show('请输入 IP');return;} try{const r=await fetch(base+'/ip/'+encodeURIComponent(ip)); show(await r.json());}catch(e){show(String(e))}}\n" +
        "async function tryPost(){const ip=document.getElementById('ip').value.trim(); if(!ip){show('请输入 IP');return;} try{const r=await fetch(base+'/ip',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({ip})}); show(await r.json());}catch(e){show(String(e))}}\n" +
        "</script>" +

        "<hr/><p>更多静态说明请见 <a href='" + base + "/docs' target='_blank'>/docs</a>。</p>" +
        "<p>IP地址位置数据由<a href='https://www.cz88.net' target='_blank' rel='noopener noreferrer'>纯真CZ88</a>提供支持</p>" +
        "</main></body></html>"

    c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
