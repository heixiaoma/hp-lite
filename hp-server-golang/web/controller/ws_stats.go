package controller

import (
	"encoding/json"
	"hp-server-lib/bean"
	"hp-server-lib/log"
	"hp-server-lib/util"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// wsWriteWait 单次写超时
	wsWriteWait = 10 * time.Second
	// wsPongWait 多久没收到客户端任何帧（含 pong）就认定连接已死
	wsPongWait = 75 * time.Second
	// wsPingPeriod 服务端心跳间隔
	wsPingPeriod = 25 * time.Second
	// wsSendBuffer 单个连接的发送缓冲区；大屏来不及消费就丢弃旧数据，绝不阻塞数据面
	wsSendBuffer = 1024
	// wsMaxMessageSize 客户端上行消息上限（只允许控制类小消息）
	wsMaxMessageSize = 4096
	// wsTopN 大屏 Top 榜单条数
	wsTopN = 10
	// wsMaxTrackKeys IP / 域名等维度最多跟踪多少个 key，防止高并发下 map 无限膨胀
	wsMaxTrackKeys = 5000
)

// wsUpgrader WebSocket 升级器。
// CheckOrigin 放开：大屏前端独立部署时会跨域握手。
var wsUpgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	ReadBufferSize:   4096,
	WriteBufferSize:  4096,
	CheckOrigin:      func(r *http.Request) bool { return true },
}

// HttpStatRecord 一条 HTTP 请求的完整统计维度，实时推送给前端大屏。
// 由 net/http 包的 StatsInterceptor 采集，经过 Gateway（本文件）广播给所有 ws 连接。
type HttpStatRecord struct {
	Id   string `json:"id"`   // 请求唯一ID
	Ts   int64  `json:"ts"`   // 时间戳（毫秒）
	Time string `json:"time"` // 格式化时间，前端可直接展示

	// 来源维度
	SourceIp   string `json:"sourceIp"`   // 来源 IP
	SourcePort string `json:"sourcePort"` // 来源端口
	IpVersion  string `json:"ipVersion"`  // IPv4 / IPv6

	// 来源归属地维度（ip2region，未加载库时为空）
	SourceCountry     string `json:"sourceCountry"`     // 国家，如：中国
	SourceCountryCode string `json:"sourceCountryCode"` // 国家代码，如 CN / US
	SourceProvince    string `json:"sourceProvince"`    // 省份
	SourceCity        string `json:"sourceCity"`        // 城市
	SourceIsp         string `json:"sourceIsp"`         // 运营商

	// 目的维度
	TargetIp   string `json:"targetIp"`   // 目的（后端）IP
	TargetPort string `json:"targetPort"` // 目的端口

	// 请求维度
	Domain    string `json:"domain"`    // 域名（Host 去掉端口）
	HostPort  string `json:"hostPort"`  // 入口端口 80 / 443
	Method    string `json:"method"`    // GET / POST ...
	Scheme    string `json:"scheme"`    // http / https
	Protocol  string `json:"protocol"`  // HTTP/1.1、HTTP/2.0
	Path      string `json:"path"`      // 路径
	Query     string `json:"query"`     // 查询串
	Uri       string `json:"uri"`       // 完整 URI
	ProxyType string `json:"proxyType"` // 链路类型：tunnel 内网穿透 / reverse 反向代理 / acme 证书校验 / unknown
	ConfigId  int    `json:"configId"`  // 命中的穿透配置ID（仅 tunnel 链路有值）

	// 响应维度
	StatusCode  int    `json:"statusCode"`  // 状态码
	StatusText  string `json:"statusText"`  // 状态码文案
	StatusClass string `json:"statusClass"` // 1xx/2xx/3xx/4xx/5xx

	// 流量与性能维度
	RequestSize        int64 `json:"requestSize"`        // 请求体字节数
	RequestHeaderSize  int64 `json:"requestHeaderSize"`  // 请求头估算字节数
	ResponseSize       int64 `json:"responseSize"`       // 响应体字节数
	ResponseHeaderSize int64 `json:"responseHeaderSize"` // 响应头估算字节数
	TotalSize          int64 `json:"totalSize"`          // 总字节数
	DurationMs         int64 `json:"durationMs"`         // 后端处理耗时（毫秒）

	// 客户端指纹维度
	ContentType        string `json:"contentType"`        // 响应 Content-Type
	RequestContentType string `json:"requestContentType"` // 请求 Content-Type
	UserAgent          string `json:"userAgent"`
	Referer            string `json:"referer"`
	Browser            string `json:"browser"` // 由 UA 解析
	Os                 string `json:"os"`      // 由 UA 解析
	Device             string `json:"device"`  // PC / Mobile / Tablet / Bot

	// TLS 维度（HTTPS 才有）
	TlsVersion string `json:"tlsVersion"`
	TlsCipher  string `json:"tlsCipher"`
}

// WsMessage 推送给前端的统一信封
type WsMessage struct {
	Type string      `json:"type"` // welcome / stat / summary / pong / error
	Ts   int64       `json:"ts"`
	Data interface{} `json:"data"`
}

// SummaryData 大屏汇总面板数据（每秒推送一次）
type SummaryData struct {
	OnlineClients     int64            `json:"onlineClients"`     // 当前在线 ws 数
	TotalRequests     int64            `json:"totalRequests"`     // 累计请求数
	TotalBytes        int64            `json:"totalBytes"`        // 累计流量
	TotalRequestSize  int64            `json:"totalRequestSize"`  // 累计上行
	TotalResponseSize int64            `json:"totalResponseSize"` // 累计下行
	Qps               int64            `json:"qps"`               // 最近 60 秒平均 QPS
	AvgDurationMs     int64            `json:"avgDurationMs"`     // 平均耗时
	StatusCounts      map[string]int64 `json:"statusCounts"`      // 状态码分布 2xx/3xx/4xx/5xx
	TopCountries      []TopItem        `json:"topCountries"`      // 来源国家/地区 Top10
	TopDomains        []TopItem        `json:"topDomains"`        // 域名 Top10
	TopIps            []TopIpItem      `json:"topIps"`            // 来源 IP Top10（含归属地）
	TopPaths          []TopItem        `json:"topPaths"`          // 路径 Top10
	UptimeSec         int64            `json:"uptimeSec"`         // 服务运行时长
}

// TopItem 榜单条目
type TopItem struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// TopIpItem 来源 IP 榜单条目，附带该 IP 的归属地信息
type TopIpItem struct {
	Key         string `json:"key"`
	Count       int64  `json:"count"`
	Country     string `json:"country"`     // 国家，未启用 ip2region 时为空
	CountryCode string `json:"countryCode"` // 国家代码，如 CN / US
	Province    string `json:"province"`    // 省份 / 州
	City        string `json:"city"`        // 城市
	Isp         string `json:"isp"`         // 运营商 / 组织
}

// IpGeo 来源 IP 的归属地快照（不落库，仅用于 Top IP 榜展示）
type IpGeo struct {
	Country     string
	CountryCode string
	Province    string
	City        string
	Isp         string
}

// WsClient 一个 WebSocket 大屏连接句柄
type WsClient struct {
	Id         string          // 连接唯一ID
	Conn       *websocket.Conn // ws 句柄
	UserId     int             // 后台用户ID
	Role       string          // 角色
	RemoteAddr string          // 对端地址
	CreatedAt  time.Time       // 建立时间
	Send       chan []byte     // 待发送队列

	mu        sync.Mutex
	Filter    WsFilter // 订阅过滤条件，前端可按需订阅
	closed    bool
	onceClose sync.Once
}

// WsFilter 前端可通过上传 {"type":"subscribe","filter":{...}} 做服务端侧过滤，降低数据量
type WsFilter struct {
	Domain      string `json:"domain"`      // 只推送该域名（包含匹配，空=全部）
	SourceIp    string `json:"sourceIp"`    // 只推送该来源IP
	StatusClass string `json:"statusClass"` // 只推送该状态码分类，如 4xx
	ProxyType   string `json:"proxyType"`   // 只推送该链路类型
}

func (f WsFilter) match(rec *HttpStatRecord) bool {
	if rec == nil {
		return false
	}
	if f.Domain != "" && !strings.Contains(rec.Domain, f.Domain) {
		return false
	}
	if f.SourceIp != "" && !strings.Contains(rec.SourceIp, f.SourceIp) {
		return false
	}
	if f.StatusClass != "" && rec.StatusClass != f.StatusClass {
		return false
	}
	if f.ProxyType != "" && rec.ProxyType != f.ProxyType {
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// 连接数组：用数组维护所有 ws 句柄
// ---------------------------------------------------------------------------

var (
	wsClients   = make([]*WsClient, 0, 16) // 所有在线 ws 句柄
	wsClientsMu sync.RWMutex
	wsClientSeq uint64
	wsPushOnce  sync.Once
	wsStartAt   = time.Now()
)

// StartWsPushLoop 启动汇总推送协程（首个大屏连上时才会启动）
func StartWsPushLoop() {
	wsPushOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if !HasWsClients() {
					continue
				}
				data, err := json.Marshal(WsMessage{Type: "summary", Ts: time.Now().UnixMilli(), Data: BuildSummary()})
				if err != nil {
					continue
				}
				dispatch(data)
			}
		}()
	})
}

// HasWsClients 是否存在在线大屏连接。
// 数据面的拦截器靠这个判断是否需要统计——没人看时零开销直通。
func HasWsClients() bool {
	wsClientsMu.RLock()
	defer wsClientsMu.RUnlock()
	return len(wsClients) > 0
}

// WsClientCount 当前在线连接数
func WsClientCount() int {
	wsClientsMu.RLock()
	defer wsClientsMu.RUnlock()
	return len(wsClients)
}

func addWsClient(c *WsClient) {
	wsClientsMu.Lock()
	wsClients = append(wsClients, c)
	wsClientsMu.Unlock()
	StartWsPushLoop()
	log.Infof("大屏WebSocket已连接 id:%s user:%d 当前在线:%d", c.Id, c.UserId, WsClientCount())
}

func removeWsClient(c *WsClient) {
	c.onceClose.Do(func() {
		c.mu.Lock()
		c.closed = true
		c.mu.Unlock()
		close(c.Send)
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
	})
	wsClientsMu.Lock()
	for i, item := range wsClients {
		if item == c {
			wsClients = append(wsClients[:i], wsClients[i+1:]...)
			break
		}
	}
	wsClientsMu.Unlock()
	log.Infof("大屏WebSocket已断开 id:%s 当前在线:%d", c.Id, WsClientCount())
}

// BroadcastHttpStat 把一条请求统计推送给所有匹配订阅条件的大屏连接。
// 全程非阻塞：前端消费慢就丢弃，绝不能拖慢代理主流程。
func BroadcastHttpStat(rec *HttpStatRecord) {
	if rec == nil || !HasWsClients() {
		return
	}
	data, err := json.Marshal(WsMessage{Type: "stat", Ts: time.Now().UnixMilli(), Data: rec})
	if err != nil {
		return
	}
	aggregate(rec)
	dispatch(data)
}

// dispatch 按订阅条件分发，遇到慢客户端直接断开，避免拖垮数据面
func dispatch(data []byte) {
	wsClientsMu.RLock()
	clients := make([]*WsClient, len(wsClients))
	copy(clients, wsClients)
	wsClientsMu.RUnlock()

	for _, c := range clients {
		if c.trySend(data) {
			continue
		}
		// 队列满或连接已关闭：直接回收，不能阻塞 http 请求
		log.Errorf("大屏WebSocket推送异常（队列满或已关闭），断开连接 id:%s", c.Id)
		go removeWsClient(c)
	}
}

// ---------------------------------------------------------------------------
// 写协程 / 读协程
// ---------------------------------------------------------------------------

// trySend 安全投递：连接若已关闭必须返回 false，否则向已关闭的 channel 发送会 panic
func (c *WsClient) trySend(data []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return false
	}
	select {
	case c.Send <- data:
		return true
	default:
		return false
	}
}

func (c *WsClient) writePump() {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *WsClient) readPump() {
	defer removeWsClient(c)
	c.Conn.SetReadLimit(wsMaxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(wsPongWait))
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
		c.handleMessage(msg)
	}
}

// handleMessage 处理前端上行控制消息：ping / subscribe
func (c *WsClient) handleMessage(msg []byte) {
	var payload struct {
		Type   string   `json:"type"`
		Filter WsFilter `json:"filter"`
	}
	if err := json.Unmarshal(msg, &payload); err != nil {
		return
	}
	switch payload.Type {
	case "ping":
		c.send(WsMessage{Type: "pong", Ts: time.Now().UnixMilli(), Data: nil})
	case "subscribe":
		c.mu.Lock()
		c.Filter = payload.Filter
		c.mu.Unlock()
		c.send(WsMessage{Type: "subscribe", Ts: time.Now().UnixMilli(), Data: payload.Filter})
	}
}

func (c *WsClient) send(m WsMessage) {
	data, err := json.Marshal(m)
	if err != nil {
		return
	}
	c.trySend(data)
}

// ---------------------------------------------------------------------------
// HTTP 入口：websocket 握手 + 汇总查询接口
// ---------------------------------------------------------------------------

type WsStatsController struct {
}

// Stats WebSocket 大屏数据推送入口：ws://host:端口/client/ws/stats?token=xxx
// 浏览器 WebSocket API 不能自定义请求头，所以 token 走 query 参数（同时也兼容 header）。
func (receiver WsStatsController) Stats(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("token")
	}
	userId, role, ts, err := util.DecodeToken(token)
	if err != nil {
		writeWsErr(w, "token非法，请重新登录")
		return
	}
	now := time.Now().UnixMilli()
	if age := now - ts; age < 0 || age > bean.TokenExpireMs {
		writeWsErr(w, "token已过期，请重新登录")
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("大屏WebSocket升级失败：%v", err)
		return
	}

	seq := atomic.AddUint64(&wsClientSeq, 1)
	client := &WsClient{
		Id:         strconv.FormatInt(time.Now().UnixMilli(), 36) + "-" + strconv.FormatUint(seq, 36),
		Conn:       conn,
		UserId:     userId,
		Role:       role,
		RemoteAddr: r.RemoteAddr,
		CreatedAt:  time.Now(),
		Send:       make(chan []byte, wsSendBuffer),
	}
	addWsClient(client)

	go client.writePump()
	// 握手完成后先发一条欢迎/快照消息，前端可立刻渲染大盘
	client.send(WsMessage{Type: "welcome", Ts: time.Now().UnixMilli(), Data: map[string]interface{}{
		"clientId": client.Id,
		"summary":  BuildSummary(),
	}})
	client.readPump()
}

// Summary 当前汇总数据（HTTP 轮询兜底，不想用 ws 的前端可以直接拉）
func (receiver WsStatsController) Summary(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(bean.ResOk(BuildSummary()))
}

func writeWsErr(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(bean.ResErrorCode(-2, msg))
}

// ---------------------------------------------------------------------------
// 汇总指标计算
// ---------------------------------------------------------------------------

var (
	aggTotalRequests int64
	aggTotalReqSize  int64
	aggTotalRespSize int64
	aggTotalDuration int64

	aggMu            sync.Mutex
	aggStatusCounts  = map[string]int64{}
	aggCountryCounts = map[string]int64{}
	aggDomainCounts  = map[string]int64{}
	aggIpCounts      = map[string]int64{}
	aggPathCounts    = map[string]int64{}
	// aggIpGeo IP -> 归属地，给 Top IP 榜补充来源信息
	aggIpGeo      = map[string]IpGeo{}
	aggQpsBuckets [60]struct {
		sec   int64
		count int64
	}
)

func aggregate(rec *HttpStatRecord) {
	atomic.AddInt64(&aggTotalRequests, 1)
	atomic.AddInt64(&aggTotalReqSize, rec.RequestSize+rec.RequestHeaderSize)
	atomic.AddInt64(&aggTotalRespSize, rec.ResponseSize+rec.ResponseHeaderSize)
	atomic.AddInt64(&aggTotalDuration, rec.DurationMs)

	sec := time.Now().Unix()
	aggMu.Lock()
	bucket := &aggQpsBuckets[sec%60]
	if bucket.sec != sec {
		bucket.sec = sec
		bucket.count = 0
	}
	bucket.count++

	incCount(aggStatusCounts, rec.StatusClass)
	if rec.SourceCountry != "" && len(aggCountryCounts) < wsMaxTrackKeys {
		incCount(aggCountryCounts, rec.SourceCountry)
	}
	if rec.Domain != "" && len(aggDomainCounts) < wsMaxTrackKeys {
		incCount(aggDomainCounts, rec.Domain)
	}
	if rec.SourceIp != "" {
		_, tracked := aggIpCounts[rec.SourceIp]
		// 已跟踪过的 IP 继续累加；新 IP 受 wsMaxTrackKeys 限制，防止高并发下 map 无限膨胀
		if tracked || len(aggIpCounts) < wsMaxTrackKeys {
			incCount(aggIpCounts, rec.SourceIp)
			// 归属地每次刷新：同一个 IP 换出口/库更新后展示的是最新结果
			aggIpGeo[rec.SourceIp] = IpGeo{
				Country:     rec.SourceCountry,
				CountryCode: rec.SourceCountryCode,
				Province:    rec.SourceProvince,
				City:        rec.SourceCity,
				Isp:         rec.SourceIsp,
			}
		}
	}
	if rec.Path != "" && len(aggPathCounts) < wsMaxTrackKeys {
		incCount(aggPathCounts, rec.Path)
	}
	aggMu.Unlock()
}

func incCount(m map[string]int64, key string) {
	if key == "" {
		return
	}
	m[key]++
}

// BuildSummary 生成大屏汇总数据
func BuildSummary() SummaryData {
	aggMu.Lock()
	defer aggMu.Unlock()

	totalRequests := atomic.LoadInt64(&aggTotalRequests)
	totalReq := atomic.LoadInt64(&aggTotalReqSize)
	totalResp := atomic.LoadInt64(&aggTotalRespSize)
	totalDuration := atomic.LoadInt64(&aggTotalDuration)

	// 最近 60 秒滑动窗口 QPS
	now := time.Now().Unix()
	var sum int64
	for i := 0; i < len(aggQpsBuckets); i++ {
		b := &aggQpsBuckets[i]
		if b.sec > 0 && now-b.sec < 60 {
			sum += b.count
		}
	}

	var avg int64
	if totalRequests > 0 {
		avg = totalDuration / totalRequests
	}

	return SummaryData{
		OnlineClients:     int64(WsClientCount()),
		TotalRequests:     totalRequests,
		TotalBytes:        totalReq + totalResp,
		TotalRequestSize:  totalReq,
		TotalResponseSize: totalResp,
		Qps:               sum / 60,
		AvgDurationMs:     avg,
		StatusCounts:      copyCounts(aggStatusCounts),
		TopCountries:      topN(aggCountryCounts, wsTopN),
		TopDomains:        topN(aggDomainCounts, wsTopN),
		TopIps:            topNIps(aggIpCounts, aggIpGeo, wsTopN),
		TopPaths:          topN(aggPathCounts, wsTopN),
		UptimeSec:         int64(time.Since(wsStartAt).Seconds()),
	}
}

func copyCounts(m map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// topNIps 生成 Top IP 榜，并附上每个 IP 的归属地
func topNIps(m map[string]int64, geo map[string]IpGeo, n int) []TopIpItem {
	items := make([]TopIpItem, 0, len(m))
	for k, v := range m {
		item := TopIpItem{Key: k, Count: v}
		if g, ok := geo[k]; ok {
			item.Country = g.Country
			item.CountryCode = g.CountryCode
			item.Province = g.Province
			item.City = g.City
			item.Isp = g.Isp
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Count > items[j].Count })
	if len(items) > n {
		items = items[:n]
	}
	return items
}

func topN(m map[string]int64, n int) []TopItem {
	items := make([]TopItem, 0, len(m))
	for k, v := range m {
		items = append(items, TopItem{Key: k, Count: v})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Count > items[j].Count })
	if len(items) > n {
		items = items[:n]
	}
	return items
}
