package http

import (
	"crypto/tls"
	"fmt"
	"hp-server-lib/entity"
	"hp-server-lib/log"
	"hp-server-lib/service"
	"net/http"
	"os"
)

// 根据域名返回证书和目标后端服务地址
func getCertificateAndTargetForDomain(domain string) (*tls.Certificate, error) {
	// 查找域名对应的证书和目标地址
	value, ok := service.DOMAIN_INFO.Load(domain)
	if !ok {
		return nil, fmt.Errorf("域名找不到证书: %s", domain)
	}
	info := value.(*entity.UserDomainEntity)
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
		Handler(w, r)
	})
	// 创建 HTTPS 服务器
	server := &http.Server{
		Addr:      ":443",
		Handler:   mux,
		TLSConfig: tlsConfig,
	}
	// 启动 HTTPS 服务
	log.Info("HTTPS代理服务启动")
	err := server.ListenAndServeTLS("", "") // 证书由 GetCertificate 动态选择
	if err != nil {
		log.Errorf("HTTPS代理服务启动失败: %v", err)
		os.Exit(1)
	}

}
