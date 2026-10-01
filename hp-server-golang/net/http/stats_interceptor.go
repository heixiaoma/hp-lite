package http

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"hp-server-lib/bean"
	"hp-server-lib/config"
	"hp-server-lib/entity"
	"hp-server-lib/service"
	"hp-server-lib/util"
	"hp-server-lib/web/controller"
)

// StatsInterceptor HTTP 请求统计拦截器。
// 设计要点：
//  1. 只有存在大屏 WebSocket 连接时才开启统计，无人观看时零额外开销直接透传；
//  2. 统计完成后由 controller 广播给大屏，推送链路全程非阻塞，不会拖慢代理主流程；
//  3. 包装 writer 时保留 Flusher / Hijacker 能力，避免影响 SSE、WebSocket 隧道等场景。
func StatsInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !controller.HasWsClients() {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rec := &controller.HttpStatRecord{
			Id:       nextStatId(),
			Ts:       start.UnixMilli(),
			Time:     start.Format("2006-01-02 15:04:05.000"),
			Method:   r.Method,
			Scheme:   requestScheme(r),
			Protocol: r.Proto,
		}

		// 来源维度（优先 X-Forwarded-For，和 net/http.getClientIP 保持一致）
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			rec.SourceIp = strings.TrimSpace(strings.Split(forwarded, ",")[0])
		}
		remoteIp, remotePort := splitHostPort(r.RemoteAddr)
		if rec.SourceIp == "" {
			rec.SourceIp = remoteIp
		}
		rec.SourcePort = remotePort
		rec.IpVersion = ipVersion(rec.SourceIp)

		// 来源归属地（ip2region，未启用/未命中时字段留空）
		if region, ok := util.LookupIpRegion(rec.SourceIp); ok {
			rec.SourceCountry = region.Country
			rec.SourceCountryCode = region.CountryCode
			rec.SourceProvince = region.Province
			rec.SourceCity = region.City
			rec.SourceIsp = region.Isp
		}

		// 域名与入口端口
		domain, port := splitHostPort(r.Host)
		rec.Domain = domain
		rec.HostPort = port

		// 请求行维度
		rec.Path = r.URL.Path
		rec.Query = r.URL.RawQuery
		rec.Uri = r.URL.RequestURI()
		rec.RequestContentType = r.Header.Get("Content-Type")
		rec.Referer = r.Header.Get("Referer")
		rec.UserAgent = r.UserAgent()
		rec.Browser, rec.Os, rec.Device = parseUserAgent(rec.UserAgent)

		// TLS 维度
		if r.TLS != nil {
			rec.TlsVersion = tls.VersionName(r.TLS.Version)
			rec.TlsCipher = tls.CipherSuiteName(r.TLS.CipherSuite)
		}

		// 目的维度
		rec.TargetIp, rec.TargetPort, rec.ProxyType, rec.ConfigId = resolveTarget(r)

		rec.RequestHeaderSize = estimateHeaderSize(r.Header)

		// 统计请求体实际读取字节数
		var bodyCounter *countReadCloser
		if r.Body != nil {
			bodyCounter = &countReadCloser{ReadCloser: r.Body}
			r.Body = bodyCounter
		}

		rw := &statsResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		if bodyCounter != nil {
			rec.RequestSize = bodyCounter.count
		}
		// 下游未消费请求体（如被 WAF 拦截直接返回）时，退回到声明的长度统计
		if rec.RequestSize == 0 && r.ContentLength > 0 {
			rec.RequestSize = r.ContentLength
		}
		rec.ResponseSize = rw.bodySize
		rec.ResponseHeaderSize = estimateHeaderSize(rw.Header())
		rec.TotalSize = rec.RequestSize + rec.RequestHeaderSize + rec.ResponseSize + rec.ResponseHeaderSize
		rec.DurationMs = time.Since(start).Milliseconds()
		rec.StatusCode = rw.status
		rec.StatusText = http.StatusText(rw.status)
		rec.StatusClass = statusClass(rw.status)
		rec.ContentType = rw.Header().Get("Content-Type")

		controller.BroadcastHttpStat(rec)
	})
}

// ---------------------------------------------------------------------------
// 统计包装器
// ---------------------------------------------------------------------------

// statsResponseWriter 记录状态码与响应体字节数
type statsResponseWriter struct {
	http.ResponseWriter
	status      int
	bodySize    int64
	wroteHeader bool
}

func (w *statsResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statsResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
	}
	n, err := w.ResponseWriter.Write(b)
	w.bodySize += int64(n)
	return n, err
}

// Flush 透传给底层 writer，SSE / 流式响应不能断
func (w *statsResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack 透传，保证 WebSocket / 协议升级流量仍可用
func (w *statsResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("底层 ResponseWriter 不支持 Hijack")
}

// Unwrap 让 http.ResponseController 等标准库工具拿到原始 writer
func (w *statsResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// countReadCloser 统计请求体读取字节数
type countReadCloser struct {
	io.ReadCloser
	count int64
}

func (c *countReadCloser) Read(p []byte) (int, error) {
	n, err := c.ReadCloser.Read(p)
	c.count += int64(n)
	return n, err
}

// countWriter 用于估算头部行数
type countWriter struct {
	n int64
}

func (c *countWriter) Write(p []byte) (int, error) {
	c.n += int64(len(p))
	return len(p), nil
}

func estimateHeaderSize(h http.Header) int64 {
	cw := &countWriter{}
	_ = h.Write(cw)
	return cw.n
}

// ---------------------------------------------------------------------------
// 维度解析
// ---------------------------------------------------------------------------

var statSeq uint64

func nextStatId() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 36) + "-" + strconv.FormatUint(atomic.AddUint64(&statSeq, 1), 36)
}

// splitHostPort 安全拆分 host:port，兼容不带端口与 IPv6
func splitHostPort(addr string) (string, string) {
	if addr == "" {
		return "", ""
	}
	if host, port, err := net.SplitHostPort(addr); err == nil {
		return host, port
	}
	if strings.Count(addr, ":") > 1 {
		// 无端口的 IPv6 字面量
		return addr, ""
	}
	return addr, ""
}

func ipVersion(ip string) string {
	if ip == "" {
		return ""
	}
	if strings.Contains(ip, ":") {
		return "IPv6"
	}
	return "IPv4"
}

func requestScheme(r *http.Request) string {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func statusClass(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	case code >= 300:
		return "3xx"
	case code >= 200:
		return "2xx"
	case code >= 100:
		return "1xx"
	default:
		return "0xx"
	}
}

// resolveTarget 根据当前请求命中的链路，解析出目的 IP/端口/链路类型/配置ID
func resolveTarget(r *http.Request) (string, string, string, int) {
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}

	// 证书校验流量 -> 本地 acme 服务
	if strings.Contains(r.URL.String(), "/.well-known/acme-challenge/") {
		return "127.0.0.1", config.ConfigData.Acme.HttpPort, "acme", 0
	}

	// 反向代理链路
	if load, ok := service.DOMAIN_REVERSE_INFO.Load(host); ok {
		if reverse, ok := load.(*entity.UserReverseEntity); ok && reverse.Address != nil {
			if u, err := url.Parse(*reverse.Address); err == nil {
				ip, port := splitHostPort(u.Host)
				if port == "" {
					port = defaultPort(u.Scheme)
				}
				return ip, port, "reverse", 0
			}
		}
	}

	// 内网穿透链路
	if value, ok := service.DOMAIN_HP_INFO.Load(host); ok {
		if info, ok := value.(*bean.UserConfigInfo); ok {
			return "127.0.0.1", strconv.Itoa(info.RemotePort), "tunnel", info.ConfigId
		}
	}

	return "", "", "unknown", 0
}

func defaultPort(scheme string) string {
	switch strings.ToLower(scheme) {
	case "https":
		return "443"
	case "http":
		return "80"
	default:
		return ""
	}
}

// parseUserAgent 轻量 UA 解析：浏览器 / 操作系统 / 设备类型（不引第三方依赖）
func parseUserAgent(ua string) (string, string, string) {
	if ua == "" {
		return "未知", "未知", "未知"
	}
	lower := strings.ToLower(ua)

	device := "PC"
	switch {
	case strings.Contains(lower, "bot"), strings.Contains(lower, "spider"), strings.Contains(lower, "crawler"):
		device = "Bot"
	case strings.Contains(lower, "tablet"), strings.Contains(lower, "ipad"):
		device = "Tablet"
	case strings.Contains(lower, "mobile"), strings.Contains(lower, "android"), strings.Contains(lower, "iphone"):
		device = "Mobile"
	}

	os := "未知"
	switch {
	case strings.Contains(lower, "windows nt 10"), strings.Contains(lower, "windows nt 11"):
		os = "Windows 10/11"
	case strings.Contains(lower, "windows"):
		os = "Windows"
	case strings.Contains(lower, "iphone"), strings.Contains(lower, "ipad"), strings.Contains(lower, "mac os x"):
		os = "iOS/macOS"
	case strings.Contains(lower, "android"):
		os = "Android"
	case strings.Contains(lower, "linux"):
		os = "Linux"
	}

	browser := "未知"
	switch {
	case strings.Contains(lower, "edg/"):
		browser = "Edge"
	case strings.Contains(lower, "chrome/") && !strings.Contains(lower, "chromium"):
		browser = "Chrome"
	case strings.Contains(lower, "firefox/"):
		browser = "Firefox"
	case strings.Contains(lower, "safari/") && !strings.Contains(lower, "chrome"):
		browser = "Safari"
	case strings.Contains(lower, "postmanruntime"):
		browser = "Postman"
	case strings.Contains(lower, "curl"):
		browser = "curl"
	case strings.Contains(lower, "go-http-client"):
		browser = "GoHttpClient"
	}
	return browser, os, device
}
