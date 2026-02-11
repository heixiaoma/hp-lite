package http

import (
	"crypto/tls"
	"fmt"
	"hp-server-lib/bean"
	"hp-server-lib/config"
	"hp-server-lib/entity"
	"hp-server-lib/log"
	"hp-server-lib/net/base"
	"hp-server-lib/service"
	"hp-server-lib/util"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/debuglog"
	txhttp "github.com/corazawaf/coraza/v3/http"
	"github.com/corazawaf/coraza/v3/types"
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
		log.Infof("代理地址: %s %s", target, r.URL.Path)
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

			rule := service.GetRule(reverse.SafeId)
			if rule != "" {
				wafConfig := coraza.NewWAFConfig().WithDebugLogger(debuglog.Default()).WithErrorCallback(func(rule types.MatchedRule) {
					msg := rule.ErrorLog()
					log.Errorf("[防火墙拦截][%s] %s\n", rule.Rule().Severity(), msg)
				}).WithDirectives(rule)
				waf, err := coraza.NewWAF(wafConfig)
				if err != nil {
					log.Error("防火墙错误：" + err.Error())
					reverse.ReverseProxy = proxy
				} else {
					reverse.ReverseProxy = txhttp.WrapHandler(waf, proxy)
				}
			} else {
				reverse.ReverseProxy = proxy
			}
		}

		log.Infof("来源: %s 访问地址: http://%s%s", clientIP, host, r.URL.Path)
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
		err2, s, _, _ := util.ProtocolInfo(info.LocalAddress)
		if err2 != nil {
			http.Error(w, "错误URL地址", http.StatusInternalServerError)
			return
		}
		target, err := url.Parse(s + "://127.0.0.1:" + strconv.Itoa(info.RemotePort))
		if err != nil {
			http.Error(w, "错误URL地址", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
			log.Errorf("反向代理错误：%s", err.Error())
			Error(w, DeviceNotFound(), http.StatusInternalServerError)
		}
		proxy.Transport = &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 500,
			IdleConnTimeout:     30 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		//检查是否开启了防火墙
		if info.SafeType == 0 {
			info.ReverseProxy = proxy
		} else {
			rule := service.GetRule(info.SafeId)
			if rule != "" {
				wafConfig := coraza.NewWAFConfig().WithDebugLogger(debuglog.Default()).WithErrorCallback(func(rule types.MatchedRule) {
					msg := rule.ErrorLog()
					log.Errorf("[防火墙拦截][%s] %s\n", rule.Rule().Severity(), msg)
				}).WithDirectives(rule)
				waf, err := coraza.NewWAF(wafConfig)
				if err != nil {
					log.Error("防火墙错误：" + err.Error())
					info.ReverseProxy = proxy
				} else {
					info.ReverseProxy = txhttp.WrapHandler(waf, proxy)
				}
			} else {
				info.ReverseProxy = proxy
			}
		}
	}
	log.Infof("来源: %s 访问地址: http://%s%s", clientIP, host, r.URL.Path)
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
