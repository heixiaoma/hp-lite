package http

import (
	"hp-server-lib/log"
	"net"
	"net/http"
	"strings"
)

func getClientIP(r *http.Request) string {
	// 如果请求头中包含 X-Forwarded-For，返回其第一个 IP 地址
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For 是以逗号分隔的 IP 地址列表，第一个是客户端的真实 IP
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	// 如果没有 X-Forwarded-For 头部，则使用 RemoteAddr 获取客户端 IP
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// httpProxyServer 保存 HTTP 代理服务句柄，StopHttpServer 时用于释放 80 端口
var httpProxyServer *http.Server

func StartHttpServer() {
	mux := http.NewServeMux()
	// 使用反向代理处理所有请求
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Handler(w, r)
	})
	log.Info("HTTP代理服务启动")
	// StatsInterceptor：存在大屏 WebSocket 连接时统计请求全维度数据并实时推送，
	// 无人观看时零开销直通
	httpProxyServer = &http.Server{Addr: ":80", Handler: StatsInterceptor(mux)}
	err := httpProxyServer.ListenAndServe()
	if err != nil {
		// 不要 os.Exit(1)：一个端口冲突会顺手把 DB、QUIC、TCP、Web 后台全带走。
		// 返回后外层 goroutine 自然结束，由 main 的 WaitGroup 收尾。
		log.Errorf("HTTP代理服务启动失败: %v", err)
	}
}

// StopHttpServer 关闭 HTTP 代理服务，Stop 由服务停止流程调用
func StopHttpServer() {
	if httpProxyServer != nil {
		_ = httpProxyServer.Close()
		httpProxyServer = nil
	}
}
