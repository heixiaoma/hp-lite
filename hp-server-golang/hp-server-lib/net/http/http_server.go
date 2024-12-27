package http

import (
	"hp-server-lib/bean"
	"hp-server-lib/config"
	"hp-server-lib/net/base"
	"hp-server-lib/service"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
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
		host := r.Host
		//检查是否是证书挑战
		if strings.Contains(r.URL.String(), "/.well-known/acme-challenge/") {
			target, err := url.Parse("http://127.0.0.1:" + config.ConfigData.Acme.HttpPort)
			if err != nil {
				http.Error(w, "错误URL地址", http.StatusInternalServerError)
				return
			}
			proxy := httputil.NewSingleHostReverseProxy(target)
			log.Printf("代理地址: %s %s", target, r.URL.Path)
			proxy.ServeHTTP(w, r)
			return
		}
		// 根据 host 选择不同的目标代理
		value, ok := service.DOMAIN_USER_INFO.Load(host)
		if !ok {
			http.Error(w, "设备不在线", http.StatusInternalServerError)
			return
		}
		info := value.(*bean.UserConfigInfo)
		clientIP := getClientIP(r)
		if strings.Compare(info.ProxyVersion, "V3") == 0 {
			// 获取当前的 X-Forwarded-For 头部（如果有）
			xfwd := r.Header.Get("X-Forwarded-For")
			if xfwd != "" {
				// 如果已有 X-Forwarded-For，添加当前客户端 IP 地址到头部中
				xfwd = xfwd + ", " + clientIP
			} else {
				// 如果没有 X-Forwarded-For 头部，直接设置客户端 IP 地址
				xfwd = clientIP
			}
			// 动态设置 X-Forwarded-For 头部
			r.Header.Set("X-Forwarded-For", xfwd)
		}

		base.AddPv(info.ConfigId, 1)
		base.AddUv(info.ConfigId, r.RemoteAddr)
		target, err := url.Parse("http://127.0.0.1:" + strconv.Itoa(info.Port))
		if err != nil {
			http.Error(w, "错误URL地址", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		log.Printf("来源: %s 访问地址: http://%s%s", clientIP, host, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})
	log.Println("HTTP代理服务启动")
	err := http.ListenAndServe(":80", mux)
	if err != nil {
		return
	}
}
