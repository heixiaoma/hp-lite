package http

import (
	"crypto/tls"
	"fmt"
	"hp-server-lib/bean"
	"hp-server-lib/service"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

// 模拟的证书、私钥和目标地址，实际应用中应替换为真实的内容
var certs = map[string]struct {
	certPEM string
	keyPEM  string
	target  string // 目标后端服务地址
}{
	"op.hp.mcle.cn": {
		certPEM: `-----BEGIN CERTIFICATE-----
MIIE6TCCA9GgAwIBAgISBHHt4OynucUya0CHi+WzLurxMA0GCSqGSIb3DQEBCwUA
MDMxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MQwwCgYDVQQD
EwNSMTEwHhcNMjQxMTIzMTIyMjU4WhcNMjUwMjIxMTIyMjU3WjAYMRYwFAYDVQQD
Ew1vcC5ocC5tY2xlLmNuMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
w2op+758NujTsInvQUWwtBOu1Mt+uF4pZxUNRIIaJSSVZZ5miXrvb/OPIOiEz+pV
WQi0d5162/bdcmz1pGLxXGMIZquMYS5Gc35CSZQxjCPOe94ZM3pjaxBhwtbTDVd+
spHCXfpdwPlXiUDdsNIB80D1QCHx8NH1LJrVENGDyHZFKntwJVcijOqgl7wvy/Ej
5PbG9Spki7Ib4y7yc5hnzTklDdQv5B0kolQuJGMgWYSHhaHyLEti/zePm7ecY3Pl
mTlkm7DaJWJk1NGapW23JgNQvXUSLFiCz0Uk9NxAk7VpHGrPLSdiVY/sLuA7ygmN
eo+ELOmiEkygcknnRYOJXwIDAQABo4ICEDCCAgwwDgYDVR0PAQH/BAQDAgWgMB0G
A1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMB0GA1Ud
DgQWBBQZ+Em/v+5N/0YlbYzZTCPG/bs0xTAfBgNVHSMEGDAWgBTFz0ak6vTDwHps
lcQtsF6SLybjuTBXBggrBgEFBQcBAQRLMEkwIgYIKwYBBQUHMAGGFmh0dHA6Ly9y
MTEuby5sZW5jci5vcmcwIwYIKwYBBQUHMAKGF2h0dHA6Ly9yMTEuaS5sZW5jci5v
cmcvMBgGA1UdEQQRMA+CDW9wLmhwLm1jbGUuY24wEwYDVR0gBAwwCjAIBgZngQwB
AgEwggEDBgorBgEEAdZ5AgQCBIH0BIHxAO8AdgCi4wrkRe+9rZt+OO1HZ3dT14Jb
hJTXK14bLMS5UKRH5wAAAZNZLz4GAAAEAwBHMEUCIQCEo4LmYbYhzIOAxHQExNWc
FrBYAKOnOSaAIT0ZxI/0awIga+2eQpytB7RD96rkxtHOT5/om2LKWDRkbbMmuD52
1z0AdQDPEVbu1S58r/OHW9lpLpvpGnFnSrAX7KwB0lt3zsw7CAAAAZNZL0YOAAAE
AwBGMEQCIEF63LTDBoaPdctDIgNRMd7MLF5KJTYrzaiQE/viEv6SAiAqUVpuRUHg
r5YS9aVIPYx+uQ6Jd9+WhfsLuzZGuEiIqjANBgkqhkiG9w0BAQsFAAOCAQEAcGFb
YezHsHzck4eaRiqX7FrwIbmhMXxeQmxudyyo3ivXtIA87DZTE2ntUaHfELkM7k63
4Q34P6F3AFj4g+uxw9oXgFyhvRMMS0qLfN5J6SFHGd1nFQUhSL2yg7WWiZmF6t/V
fWkpK9HoqGar2BXktTPEcvceJXLKFNWARBb/gFJLuZ836cbX2Wy5d1SRd8CYOinI
Pty9aDJzW5K+8AFFu8jZeiIvQtcm42HeYQQY5ziqMHq0mdjW5RHIhpXJMijQhrDH
J+3VUsECJi8Y8GQ0eq7AnTFpVsg4QSKYKfptyEpSYE5w/0Z472FntqnmDCCHqUSZ
Z5nfi9A1RQNHd4oJgg==
-----END CERTIFICATE-----

-----BEGIN CERTIFICATE-----
MIIFBjCCAu6gAwIBAgIRAIp9PhPWLzDvI4a9KQdrNPgwDQYJKoZIhvcNAQELBQAw
TzELMAkGA1UEBhMCVVMxKTAnBgNVBAoTIEludGVybmV0IFNlY3VyaXR5IFJlc2Vh
cmNoIEdyb3VwMRUwEwYDVQQDEwxJU1JHIFJvb3QgWDEwHhcNMjQwMzEzMDAwMDAw
WhcNMjcwMzEyMjM1OTU5WjAzMQswCQYDVQQGEwJVUzEWMBQGA1UEChMNTGV0J3Mg
RW5jcnlwdDEMMAoGA1UEAxMDUjExMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAuoe8XBsAOcvKCs3UZxD5ATylTqVhyybKUvsVAbe5KPUoHu0nsyQYOWcJ
DAjs4DqwO3cOvfPlOVRBDE6uQdaZdN5R2+97/1i9qLcT9t4x1fJyyXJqC4N0lZxG
AGQUmfOx2SLZzaiSqhwmej/+71gFewiVgdtxD4774zEJuwm+UE1fj5F2PVqdnoPy
6cRms+EGZkNIGIBloDcYmpuEMpexsr3E+BUAnSeI++JjF5ZsmydnS8TbKF5pwnnw
SVzgJFDhxLyhBax7QG0AtMJBP6dYuC/FXJuluwme8f7rsIU5/agK70XEeOtlKsLP
Xzze41xNG/cLJyuqC0J3U095ah2H2QIDAQABo4H4MIH1MA4GA1UdDwEB/wQEAwIB
hjAdBgNVHSUEFjAUBggrBgEFBQcDAgYIKwYBBQUHAwEwEgYDVR0TAQH/BAgwBgEB
/wIBADAdBgNVHQ4EFgQUxc9GpOr0w8B6bJXELbBeki8m47kwHwYDVR0jBBgwFoAU
ebRZ5nu25eQBc4AIiMgaWPbpm24wMgYIKwYBBQUHAQEEJjAkMCIGCCsGAQUFBzAC
hhZodHRwOi8veDEuaS5sZW5jci5vcmcvMBMGA1UdIAQMMAowCAYGZ4EMAQIBMCcG
A1UdHwQgMB4wHKAaoBiGFmh0dHA6Ly94MS5jLmxlbmNyLm9yZy8wDQYJKoZIhvcN
AQELBQADggIBAE7iiV0KAxyQOND1H/lxXPjDj7I3iHpvsCUf7b632IYGjukJhM1y
v4Hz/MrPU0jtvfZpQtSlET41yBOykh0FX+ou1Nj4ScOt9ZmWnO8m2OG0JAtIIE38
01S0qcYhyOE2G/93ZCkXufBL713qzXnQv5C/viOykNpKqUgxdKlEC+Hi9i2DcaR1
e9KUwQUZRhy5j/PEdEglKg3l9dtD4tuTm7kZtB8v32oOjzHTYw+7KdzdZiw/sBtn
UfhBPORNuay4pJxmY/WrhSMdzFO2q3Gu3MUBcdo27goYKjL9CTF8j/Zz55yctUoV
aneCWs/ajUX+HypkBTA+c8LGDLnWO2NKq0YD/pnARkAnYGPfUDoHR9gVSp/qRx+Z
WghiDLZsMwhN1zjtSC0uBWiugF3vTNzYIEFfaPG7Ws3jDrAMMYebQ95JQ+HIBD/R
PBuHRTBpqKlyDnkSHDHYPiNX3adPoPAcgdF3H2/W0rmoswMWgTlLn1Wu0mrks7/q
pdWfS6PJ1jty80r2VKsM/Dj3YIDfbjXKdaFU5C+8bhfJGqU3taKauuz0wHVGT3eo
6FlWkWYtbt4pgdamlwVeZEW+LM7qZEJEsMNPrfC03APKmZsJgpWCDWOKZvkZcvjV
uYkQ4omYCTX5ohy+knMjdOmdH9c7SpqEWBDC86fiNex+O0XOMEZSa8DA
-----END CERTIFICATE-----
`,
		keyPEM: `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDDain7vnw26NOw
ie9BRbC0E67Uy364XilnFQ1EgholJJVlnmaJeu9v848g6ITP6lVZCLR3nXrb9t1y
bPWkYvFcYwhmq4xhLkZzfkJJlDGMI8573hkzemNrEGHC1tMNV36ykcJd+l3A+VeJ
QN2w0gHzQPVAIfHw0fUsmtUQ0YPIdkUqe3AlVyKM6qCXvC/L8SPk9sb1KmSLshvj
LvJzmGfNOSUN1C/kHSSiVC4kYyBZhIeFofIsS2L/N4+bt5xjc+WZOWSbsNolYmTU
0ZqlbbcmA1C9dRIsWILPRST03ECTtWkcas8tJ2JVj+wu4DvKCY16j4Qs6aISTKBy
SedFg4lfAgMBAAECggEAKyZ8uc2wsRFNeVyj+8ZPRBxajTuYMk78lnoUEW4nJs+C
u8sz+iSuzYo7LzmY5i5pBT2CWC1qjTrgYF2GCnQcttlZaA52c5kbznHxYjz6ieb4
N+rtpqveWyxlYfg852PMItND16uq2ytT9IdhzKM68gOEhFJakeJS4LQZ21pgXrm3
kvKzOC4D+qijjUJl1XHxZd9urnQk9od4vY5sK3NkmeaiyUX3KbtzfAEd3N+2NBj1
XytSaeUzRLpofSS92yOyBZkdR02qfeN3bX1k+nTkzju8gI8RrtABku58DdqxTwF/
aIu7o8DC9rCR5zWhJBlhpWb/x+s1rI2dN2yy92TWIQKBgQD64tAABKkErWDe2vdf
EPjTzi06ski3mjFS2iTCMQS9bX/1YbWlCCpl0VKg6qNkRMU2w5Ju5JH7phsdo0h2
AZGwoLL213MZLrOPRpDEmwIjSwUQfifmirVDjDA/vBhMU5hjxwAckBdqSlrHl1DF
7XRUXaCqoEl3OjGkv0GzcNYQIQKBgQDHZeNaWnyZiBp9/tv8zsBTs3YV9jdnCdq7
352spi3OUI8UsmCJLxkIC3Or/RI2chyqc+BgYgb0GXEfhEcnKa6IIENeTphiCdVC
j9b0GTqdspjyAh6Y7kjsnaBmBmpejfzoAMJSUY36uYPMtBW4ijNQU3fxkJsmjvze
wgaYODRpfwKBgQCam5m42SZbdokC7QeSsz/ULvOaf3Hmi4Qn3bzXWyPjpI49Zphs
+ko+cq+r8Mz+Jo8uP3mHEx6PaP6+1ff6mN7ybSW8jmsksq3+9mqSbj/0BfA6CLSI
EyS/Wq4FKOIEb2Oy4VjFQVrcqrOk2i/xuXJ95zDy1VJQwjEDqMVRUpDoYQKBgDEA
vzDzT+/DXQ9d1N56SRXI4tpe2hq+dzz4pZ1KcbNkZOVnOQY9xt8NQW4hEZrDzHuv
YpMNRDw1DHH8ZigfvD7D/wpsMlLVq81h4Ce5E4ix3ZiMIMzgspdD3al1Jir6pg62
MQtd85CMivGByFzDyfyRpsZ9DUQam9Z6xHggR/EtAoGAO0crGqxbxhliGo4Owltn
yz7V3VnV6wj1hE0ecoBsN4amJBgRXOQ5Oh6S5YSDnxtNbVd35tjGym/wpJumzGyr
20rmjumzSPZ49u3J4YPLWxTMOx1HSNQZtr0cEIwv8F0HEIvbZ+SDx3h84HQ7K8uk
kXb8VFf/IGcY9dbURKwGWqo=
-----END PRIVATE KEY-----`,
	},
}

// 根据域名返回证书和目标后端服务地址
func getCertificateAndTargetForDomain(domain string) (*tls.Certificate, string, error) {
	// 查找域名对应的证书和目标地址
	certInfo, ok := certs[domain]
	if !ok {
		return nil, "", fmt.Errorf("no certificate or target found for domain: %s", domain)
	}

	// 加载证书和私钥
	certificate, err := tls.X509KeyPair([]byte(certInfo.certPEM), []byte(certInfo.keyPEM))
	if err != nil {
		return nil, "", fmt.Errorf("failed to load certificate for domain %s: %v", domain, err)
	}
	return &certificate, certInfo.target, nil
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
			cert, _, err := getCertificateAndTargetForDomain(domain)
			if err != nil {
				return nil, err
			}
			// 根据选择的目标服务动态创建反向代理
			return cert, nil
		},
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.TLS.ServerName
		fmt.Println("Host:", host)
		// 根据 host 选择不同的目标代理
		value, ok := service.DOMAIN_USER_INFO.Load(host)
		if !ok {
			http.Error(w, "错误代理", http.StatusInternalServerError)
			return
		}

		info := value.(*bean.UserConfigInfo)

		target, err := url.Parse("http://127.0.0.1:" + strconv.Itoa(info.Port))
		if err != nil {
			http.Error(w, "错误URL地址", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		log.Printf("Proxying request to target: %s %s", target, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})
	// 创建 HTTPS 服务器
	server := &http.Server{
		Addr:      ":443",
		Handler:   mux,
		TLSConfig: tlsConfig,
	}
	// 启动 HTTPS 服务
	fmt.Println("Starting HTTPS server on https://localhost:443")
	err := server.ListenAndServeTLS("", "") // 证书由 GetCertificate 动态选择
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}
