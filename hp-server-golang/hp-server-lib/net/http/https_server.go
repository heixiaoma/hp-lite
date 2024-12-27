package http

import (
	"crypto/tls"
	"fmt"
	"hp-server-lib/bean"
	"hp-server-lib/net/base"
	"hp-server-lib/service"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

// 根据域名返回证书和目标后端服务地址
func getCertificateAndTargetForDomain(domain string) (*tls.Certificate, error) {
	// 查找域名对应的证书和目标地址
	value, ok := service.DOMAIN_USER_INFO.Load(domain)
	if !ok {
		return nil, fmt.Errorf("域名找不到证书: %s", domain)
	}
	info := value.(*bean.UserConfigInfo)
	// 加载证书和私钥
	certificate, err := tls.X509KeyPair([]byte(info.CertificateContent), []byte(info.CertificateKey))
	if err != nil {
		return nil, fmt.Errorf("证书解析失败 %s: %v", domain, err)
	}
	return &certificate, nil
}

func StartHttpsServer() {
	// 创建一个 HTTP 多路复用器
	mux := http.NewServeMux()
	// 设置 SNI 支持的证书和目标获取函数
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			// 获取客户端请求的主机名（SNI）
			domain := clientHello.ServerName
			// 根据 SNI（域名）选择证书和目标后端服务地址
			cert, err := getCertificateAndTargetForDomain(domain)
			if err != nil {
				return nil, err
			}
			// 根据选择的目标服务动态创建反向代理
			return cert, nil
		},
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.TLS.ServerName
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
		log.Printf("来源: %s 访问地址: https://%s%s", clientIP, host, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})
	// 创建 HTTPS 服务器
	server := &http.Server{
		Addr:      ":443",
		Handler:   mux,
		TLSConfig: tlsConfig,
	}
	// 启动 HTTPS 服务
	log.Println("HTTPS代理服务启动")
	err := server.ListenAndServeTLS("", "") // 证书由 GetCertificate 动态选择
	if err != nil {
		log.Fatalf("HTTPS代理服务启动失败: %v", err)
	}

}
