package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// 动态选择证书和目标地址，根据 SNI（Server Name Indication）来进行处理
func getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	fmt.Println("SNI:", hello.ServerName)
	// 根据 SNI 选择不同的证书和目标服务器
	var certFile, keyFile string
	switch hello.ServerName {
	case "example.com":
		certFile = "example.crt"
		keyFile = "example.key"
	case "anotherdomain.com":
		certFile = "anotherdomain.crt"
		keyFile = "anotherdomain.key"
	default:
		certFile = "default.crt"
		keyFile = "default.key"
	}
	// 这里可以根据需要加载证书
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %v", err)
	}
	return &certificate, nil
}

func main() {

	// 配置 HTTPS 服务器，使用 SNI 来动态选择证书
	tlsConfig := &tls.Config{
		GetCertificate: getCertificate,
	}

	// 设置要转发的地址
	// 启动 HTTPS 服务器，监听 443 端口
	server := &http.Server{
		Addr:      ":443",
		TLSConfig: tlsConfig,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// HTTPS 请求也根据 SNI 进行代理
			// 你可以根据 SNI 设置目标地址或其他逻辑
			sni := r.TLS.ServerName
			// 解析目标地址，基于 Host 或 SNI
			target := "http://default-server.com:8080" // 默认目标
			// 示例，依据 Host 来选择不同的目标地址
			switch sni {
			case "example.com":
				target = "http://192.168.100.246:5666"
			case "anotherdomain.com":
				target = "http://192.168.100.247:5666"
			}
			// 解析目标地址
			targetURL, err := url.Parse(target)
			if err != nil {
				http.Error(w, "Error parsing target URL", http.StatusInternalServerError)
				return
			}
			// 创建反向代理并转发请求
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			proxy.ServeHTTP(w, r)
		}),
	}

}
