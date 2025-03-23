package http

import (
	"log"
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

func StartHttpServer() {
	mux := http.NewServeMux()
	// 使用反向代理处理所有请求
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Handler(w, r)
	})
	log.Println("HTTP代理服务启动")
	err := http.ListenAndServe(":80", mux)
	if err != nil {
		return
	}
}
