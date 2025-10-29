package http

import (
	"crypto/tls"
	"fmt"
	"hp-server-lib/bean"
	"hp-server-lib/config"
	"hp-server-lib/entity"
	"hp-server-lib/net/base"
	"hp-server-lib/service"
	"hp-server-lib/util"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	clientIP := getClientIP(r)
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

	//反向代理处理
	load, o := service.DOMAIN_REVERSE_INFO.Load(host)
	if o {
		reverse := load.(*entity.UserReverseEntity)
		if reverse.ReverseProxy == nil {

			target, err := url.Parse(*reverse.Address)
			if err != nil {
				http.Error(w, "错误URL地址", http.StatusInternalServerError)
				return
			}
			proxy := httputil.NewSingleHostReverseProxy(target)
			proxy.Transport = &http.Transport{
				MaxIdleConns:        1000,
				MaxIdleConnsPerHost: 500,
				IdleConnTimeout:     30 * time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			reverse.ReverseProxy = proxy
		}

		log.Printf("来源: %s 访问地址: http://%s%s", clientIP, host, r.URL.Path)
		reverse.ReverseProxy.ServeHTTP(w, r)
		return
	}

	//穿透代理处理
	// 根据 host 选择不同的目标代理
	value, ok := service.DOMAIN_HP_INFO.Load(host)
	if !ok {
		Error(w, DeviceNotFound(), http.StatusInternalServerError)
		return
	}
	info := value.(*bean.UserConfigInfo)

	if len(info.AllowedIps) > 0 {
		ips := info.AllowedIps
		flag := true
		for _, item := range ips {
			if util.IsIPInCIDR(clientIP, item) {
				flag = false
				break
			}
		}
		if flag {
			Error(w, Waf("您的IP地址不在防火墙白名单，无法访问此服务", clientIP), http.StatusInternalServerError)
			return
		}
	}

	if len(info.BlockedIps) > 0 {
		ips := info.BlockedIps
		for _, item := range ips {
			if util.IsIPInCIDR(clientIP, item) {
				Error(w, Waf("您的IP地址已被列入防火墙黑名单，无法访问此服务", clientIP), http.StatusInternalServerError)
				return
			}
		}
	}

	base.AddPv(info.ConfigId, 1)
	base.AddUv(info.ConfigId, r.RemoteAddr)
	if info.ReverseProxy == nil {
		if len(info.WebType) == 0 {
			info.WebType = "http"
		}
		target, err := url.Parse(info.WebType + "://127.0.0.1:" + strconv.Itoa(info.Port))
		if err != nil {
			http.Error(w, "错误URL地址", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Transport = &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 500,
			IdleConnTimeout:     30 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		info.ReverseProxy = proxy
	}
	log.Printf("来源: %s 访问地址: http://%s%s", clientIP, host, r.URL.Path)
	info.ReverseProxy.ServeHTTP(w, r)
}

func Error(w http.ResponseWriter, error string, code int) {
	h := w.Header()

	// Delete the Content-Length header, which might be for some other content.
	// Assuming the error string fits in the writer's buffer, we'll figure
	// out the correct Content-Length for it later.
	//
	// We don't delete Content-Encoding, because some middleware sets
	// Content-Encoding: gzip and wraps the ResponseWriter to compress on-the-fly.
	// See https://go.dev/issue/66343.
	h.Del("Content-Length")

	// There might be content type already set, but we reset it to
	// text/html for the error message.
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, error)
}
