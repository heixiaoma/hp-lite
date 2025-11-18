package ext

import (
	"encoding/base64"
	"hp-server-lib/log"
	"io"
	"net"
	"net/http"
	"strings"
)

type HttpFwdServer struct {
	port     string
	username string
	password string
	server   *http.Server
}

func NewHttpFwdServer(port string, username string, password string) *HttpFwdServer {
	return &HttpFwdServer{
		port:     port,
		username: username,
		password: password,
	}
}

// 检查认证
func (h *HttpFwdServer) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	if h.username == "" && h.password == "" {
		return true
	}
	auth := r.Header.Get("Proxy-Authorization")
	if auth == "" {
		w.Header().Set("Proxy-Authenticate", `Basic realm="Proxy"`)
		w.WriteHeader(http.StatusProxyAuthRequired)
		return false
	}
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return false
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return false
	}
	if parts[0] == h.username && parts[1] == h.password {
		return true
	}
	w.Header().Set("Proxy-Authenticate", `Basic realm="Proxy"`)
	w.WriteHeader(http.StatusProxyAuthRequired)
	return false
}

// 处理所有请求
func (h *HttpFwdServer) handler(w http.ResponseWriter, r *http.Request) {
	if !h.checkAuth(w, r) {
		return
	}
	if r.Method == http.MethodConnect {
		h.handleConnect(w, r)
	} else {
		h.handleHTTP(w, r)
	}
}

// 处理 HTTP
func (h *HttpFwdServer) handleHTTP(w http.ResponseWriter, r *http.Request) {
	transport := http.DefaultTransport

	// 删除代理相关头
	r.RequestURI = ""
	r.Header.Del("Proxy-Authorization")

	resp, err := transport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// 复制响应
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// 处理 HTTPS (CONNECT)
func (h *HttpFwdServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// 通知客户端连接已建立
	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// 数据转发
	go io.Copy(destConn, clientConn)
	io.Copy(clientConn, destConn)
}

func (h *HttpFwdServer) Start(close func()) bool {
	addr := ":" + h.port
	h.server = &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(h.handler),
	}
	go func() {
		log.Infof("Proxy server started on %s", addr)
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(err)
		}
		defer close()
	}()
	return true
}

func (h *HttpFwdServer) Stop() {
	if h.server != nil {
		_ = h.server.Close()
		log.Info("Proxy server stopped")
	}
}
