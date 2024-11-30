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

// 动态选择证书和目标地址，根据 SNI（Server Name Indication）来进行处理
func getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	fmt.Println("SNI:", hello.ServerName)
	// 根据 SNI 选择不同的证书和目标服务器
	//key := "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDDain7vnw26NOw\nie9BRbC0E67Uy364XilnFQ1EgholJJVlnmaJeu9v848g6ITP6lVZCLR3nXrb9t1y\nbPWkYvFcYwhmq4xhLkZzfkJJlDGMI8573hkzemNrEGHC1tMNV36ykcJd+l3A+VeJ\nQN2w0gHzQPVAIfHw0fUsmtUQ0YPIdkUqe3AlVyKM6qCXvC/L8SPk9sb1KmSLshvj\nLvJzmGfNOSUN1C/kHSSiVC4kYyBZhIeFofIsS2L/N4+bt5xjc+WZOWSbsNolYmTU\n0ZqlbbcmA1C9dRIsWILPRST03ECTtWkcas8tJ2JVj+wu4DvKCY16j4Qs6aISTKBy\nSedFg4lfAgMBAAECggEAKyZ8uc2wsRFNeVyj+8ZPRBxajTuYMk78lnoUEW4nJs+C\nu8sz+iSuzYo7LzmY5i5pBT2CWC1qjTrgYF2GCnQcttlZaA52c5kbznHxYjz6ieb4\nN+rtpqveWyxlYfg852PMItND16uq2ytT9IdhzKM68gOEhFJakeJS4LQZ21pgXrm3\nkvKzOC4D+qijjUJl1XHxZd9urnQk9od4vY5sK3NkmeaiyUX3KbtzfAEd3N+2NBj1\nXytSaeUzRLpofSS92yOyBZkdR02qfeN3bX1k+nTkzju8gI8RrtABku58DdqxTwF/\naIu7o8DC9rCR5zWhJBlhpWb/x+s1rI2dN2yy92TWIQKBgQD64tAABKkErWDe2vdf\nEPjTzi06ski3mjFS2iTCMQS9bX/1YbWlCCpl0VKg6qNkRMU2w5Ju5JH7phsdo0h2\nAZGwoLL213MZLrOPRpDEmwIjSwUQfifmirVDjDA/vBhMU5hjxwAckBdqSlrHl1DF\n7XRUXaCqoEl3OjGkv0GzcNYQIQKBgQDHZeNaWnyZiBp9/tv8zsBTs3YV9jdnCdq7\n352spi3OUI8UsmCJLxkIC3Or/RI2chyqc+BgYgb0GXEfhEcnKa6IIENeTphiCdVC\nj9b0GTqdspjyAh6Y7kjsnaBmBmpejfzoAMJSUY36uYPMtBW4ijNQU3fxkJsmjvze\nwgaYODRpfwKBgQCam5m42SZbdokC7QeSsz/ULvOaf3Hmi4Qn3bzXWyPjpI49Zphs\n+ko+cq+r8Mz+Jo8uP3mHEx6PaP6+1ff6mN7ybSW8jmsksq3+9mqSbj/0BfA6CLSI\nEyS/Wq4FKOIEb2Oy4VjFQVrcqrOk2i/xuXJ95zDy1VJQwjEDqMVRUpDoYQKBgDEA\nvzDzT+/DXQ9d1N56SRXI4tpe2hq+dzz4pZ1KcbNkZOVnOQY9xt8NQW4hEZrDzHuv\nYpMNRDw1DHH8ZigfvD7D/wpsMlLVq81h4Ce5E4ix3ZiMIMzgspdD3al1Jir6pg62\nMQtd85CMivGByFzDyfyRpsZ9DUQam9Z6xHggR/EtAoGAO0crGqxbxhliGo4Owltn\nyz7V3VnV6wj1hE0ecoBsN4amJBgRXOQ5Oh6S5YSDnxtNbVd35tjGym/wpJumzGyr\n20rmjumzSPZ49u3J4YPLWxTMOx1HSNQZtr0cEIwv8F0HEIvbZ+SDx3h84HQ7K8uk\nkXb8VFf/IGcY9dbURKwGWqo=\n-----END PRIVATE KEY-----"
	//pem := "-----BEGIN CERTIFICATE-----\nMIIE6TCCA9GgAwIBAgISBHHt4OynucUya0CHi+WzLurxMA0GCSqGSIb3DQEBCwUA\nMDMxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MQwwCgYDVQQD\nEwNSMTEwHhcNMjQxMTIzMTIyMjU4WhcNMjUwMjIxMTIyMjU3WjAYMRYwFAYDVQQD\nEw1vcC5ocC5tY2xlLmNuMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA\nw2op+758NujTsInvQUWwtBOu1Mt+uF4pZxUNRIIaJSSVZZ5miXrvb/OPIOiEz+pV\nWQi0d5162/bdcmz1pGLxXGMIZquMYS5Gc35CSZQxjCPOe94ZM3pjaxBhwtbTDVd+\nspHCXfpdwPlXiUDdsNIB80D1QCHx8NH1LJrVENGDyHZFKntwJVcijOqgl7wvy/Ej\n5PbG9Spki7Ib4y7yc5hnzTklDdQv5B0kolQuJGMgWYSHhaHyLEti/zePm7ecY3Pl\nmTlkm7DaJWJk1NGapW23JgNQvXUSLFiCz0Uk9NxAk7VpHGrPLSdiVY/sLuA7ygmN\neo+ELOmiEkygcknnRYOJXwIDAQABo4ICEDCCAgwwDgYDVR0PAQH/BAQDAgWgMB0G\nA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMB0GA1Ud\nDgQWBBQZ+Em/v+5N/0YlbYzZTCPG/bs0xTAfBgNVHSMEGDAWgBTFz0ak6vTDwHps\nlcQtsF6SLybjuTBXBggrBgEFBQcBAQRLMEkwIgYIKwYBBQUHMAGGFmh0dHA6Ly9y\nMTEuby5sZW5jci5vcmcwIwYIKwYBBQUHMAKGF2h0dHA6Ly9yMTEuaS5sZW5jci5v\ncmcvMBgGA1UdEQQRMA+CDW9wLmhwLm1jbGUuY24wEwYDVR0gBAwwCjAIBgZngQwB\nAgEwggEDBgorBgEEAdZ5AgQCBIH0BIHxAO8AdgCi4wrkRe+9rZt+OO1HZ3dT14Jb\nhJTXK14bLMS5UKRH5wAAAZNZLz4GAAAEAwBHMEUCIQCEo4LmYbYhzIOAxHQExNWc\nFrBYAKOnOSaAIT0ZxI/0awIga+2eQpytB7RD96rkxtHOT5/om2LKWDRkbbMmuD52\n1z0AdQDPEVbu1S58r/OHW9lpLpvpGnFnSrAX7KwB0lt3zsw7CAAAAZNZL0YOAAAE\nAwBGMEQCIEF63LTDBoaPdctDIgNRMd7MLF5KJTYrzaiQE/viEv6SAiAqUVpuRUHg\nr5YS9aVIPYx+uQ6Jd9+WhfsLuzZGuEiIqjANBgkqhkiG9w0BAQsFAAOCAQEAcGFb\nYezHsHzck4eaRiqX7FrwIbmhMXxeQmxudyyo3ivXtIA87DZTE2ntUaHfELkM7k63\n4Q34P6F3AFj4g+uxw9oXgFyhvRMMS0qLfN5J6SFHGd1nFQUhSL2yg7WWiZmF6t/V\nfWkpK9HoqGar2BXktTPEcvceJXLKFNWARBb/gFJLuZ836cbX2Wy5d1SRd8CYOinI\nPty9aDJzW5K+8AFFu8jZeiIvQtcm42HeYQQY5ziqMHq0mdjW5RHIhpXJMijQhrDH\nJ+3VUsECJi8Y8GQ0eq7AnTFpVsg4QSKYKfptyEpSYE5w/0Z472FntqnmDCCHqUSZ\nZ5nfi9A1RQNHd4oJgg==\n-----END CERTIFICATE-----\n\n-----BEGIN CERTIFICATE-----\nMIIFBjCCAu6gAwIBAgIRAIp9PhPWLzDvI4a9KQdrNPgwDQYJKoZIhvcNAQELBQAw\nTzELMAkGA1UEBhMCVVMxKTAnBgNVBAoTIEludGVybmV0IFNlY3VyaXR5IFJlc2Vh\ncmNoIEdyb3VwMRUwEwYDVQQDEwxJU1JHIFJvb3QgWDEwHhcNMjQwMzEzMDAwMDAw\nWhcNMjcwMzEyMjM1OTU5WjAzMQswCQYDVQQGEwJVUzEWMBQGA1UEChMNTGV0J3Mg\nRW5jcnlwdDEMMAoGA1UEAxMDUjExMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB\nCgKCAQEAuoe8XBsAOcvKCs3UZxD5ATylTqVhyybKUvsVAbe5KPUoHu0nsyQYOWcJ\nDAjs4DqwO3cOvfPlOVRBDE6uQdaZdN5R2+97/1i9qLcT9t4x1fJyyXJqC4N0lZxG\nAGQUmfOx2SLZzaiSqhwmej/+71gFewiVgdtxD4774zEJuwm+UE1fj5F2PVqdnoPy\n6cRms+EGZkNIGIBloDcYmpuEMpexsr3E+BUAnSeI++JjF5ZsmydnS8TbKF5pwnnw\nSVzgJFDhxLyhBax7QG0AtMJBP6dYuC/FXJuluwme8f7rsIU5/agK70XEeOtlKsLP\nXzze41xNG/cLJyuqC0J3U095ah2H2QIDAQABo4H4MIH1MA4GA1UdDwEB/wQEAwIB\nhjAdBgNVHSUEFjAUBggrBgEFBQcDAgYIKwYBBQUHAwEwEgYDVR0TAQH/BAgwBgEB\n/wIBADAdBgNVHQ4EFgQUxc9GpOr0w8B6bJXELbBeki8m47kwHwYDVR0jBBgwFoAU\nebRZ5nu25eQBc4AIiMgaWPbpm24wMgYIKwYBBQUHAQEEJjAkMCIGCCsGAQUFBzAC\nhhZodHRwOi8veDEuaS5sZW5jci5vcmcvMBMGA1UdIAQMMAowCAYGZ4EMAQIBMCcG\nA1UdHwQgMB4wHKAaoBiGFmh0dHA6Ly94MS5jLmxlbmNyLm9yZy8wDQYJKoZIhvcN\nAQELBQADggIBAE7iiV0KAxyQOND1H/lxXPjDj7I3iHpvsCUf7b632IYGjukJhM1y\nv4Hz/MrPU0jtvfZpQtSlET41yBOykh0FX+ou1Nj4ScOt9ZmWnO8m2OG0JAtIIE38\n01S0qcYhyOE2G/93ZCkXufBL713qzXnQv5C/viOykNpKqUgxdKlEC+Hi9i2DcaR1\ne9KUwQUZRhy5j/PEdEglKg3l9dtD4tuTm7kZtB8v32oOjzHTYw+7KdzdZiw/sBtn\nUfhBPORNuay4pJxmY/WrhSMdzFO2q3Gu3MUBcdo27goYKjL9CTF8j/Zz55yctUoV\naneCWs/ajUX+HypkBTA+c8LGDLnWO2NKq0YD/pnARkAnYGPfUDoHR9gVSp/qRx+Z\nWghiDLZsMwhN1zjtSC0uBWiugF3vTNzYIEFfaPG7Ws3jDrAMMYebQ95JQ+HIBD/R\nPBuHRTBpqKlyDnkSHDHYPiNX3adPoPAcgdF3H2/W0rmoswMWgTlLn1Wu0mrks7/q\npdWfS6PJ1jty80r2VKsM/Dj3YIDfbjXKdaFU5C+8bhfJGqU3taKauuz0wHVGT3eo\n6FlWkWYtbt4pgdamlwVeZEW+LM7qZEJEsMNPrfC03APKmZsJgpWCDWOKZvkZcvjV\nuYkQ4omYCTX5ohy+knMjdOmdH9c7SpqEWBDC86fiNex+O0XOMEZSa8DA\n-----END CERTIFICATE-----\n"
	//// 这里可以根据需要加载证书
	//certificate, err := tls.X509KeyPair([]byte(pem), []byte(key))
	//if err != nil {
	//	return nil, fmt.Errorf("failed to load certificate: %v", err)
	//}
	return nil, nil
}

func StartHttpsServer() {

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
			// 根据 host 选择不同的目标代理

			log.Printf("sni" + sni)

			value, ok := service.DOMAIN_USER_INFO.Load(sni)
			if !ok {
				http.Error(w, "错误代理", http.StatusInternalServerError)
				return
			}

			info := value.(*bean.UserConfigInfo)
			target, err := url.Parse("http://127.0.0.1:" + strconv.Itoa(info.Port))
			if err != nil {
				http.Error(w, "Error parsing target URL", http.StatusInternalServerError)
				return
			}
			// 创建反向代理并转发请求
			proxy := httputil.NewSingleHostReverseProxy(target)
			proxy.ServeHTTP(w, r)
		}),
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
