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
	// wsTrackBucketSec 滑动窗口单个分桶的时间跨度（秒）
	wsTrackBucketSec = 300
	// wsTrackBuckets 分桶数量：60 × 300s = 最近 5 小时滑动窗口。
	// 桶粒度取 5 分钟：请求/攻击统计关心的是"这段时间谁在活跃"，不需要秒级精度，
	// 而 5 小时 × 10 秒要 1800 个桶，绝大多数时刻空着，纯属浪费
	wsTrackBuckets = 60
	// wsMaxTrackKeys 单个维度窗口内 key 数的硬上限（兜底）。
	// 正常情况靠分桶滑出窗口自动淘汰，只有窗口内基数异常膨胀时才触发采样淘汰。
	wsMaxTrackKeys = 20000
	// wsEvictSamples 兜底淘汰时的随机采样个数：Go map 迭代起点随机，
	// 取前 N 个即近似随机采样，避免每次扫全表。
	wsEvictSamples = 8
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

	// 来源归属地维度（GeoLite2，未加载库时为空）
	SourceCountry     string `json:"sourceCountry"`     // 国家，如：中国
	SourceCountryCode string `json:"sourceCountryCode"` // 国家代码，如 CN / US
	SourceProvince    string `json:"sourceProvince"`    // 省份
	SourceCity        string `json:"sourceCity"`        // 城市
	SourceIsp         string `json:"sourceIsp"`         // 运营商（需额外配 GeoLite2-ASN 库才有值）
	// 经纬度（地图打点用）。前端注意：0 表示这跳没定位到，别在 0,0 打点
	SourceLatitude  float64 `json:"sourceLatitude"`  // 纬度
	SourceLongitude float64 `json:"sourceLongitude"` // 经度

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
	TopPaths          []TopItem        `json:"topPaths"`          // 路径 Top10（聚合 key 已归一化）
	TopDevices        []TopItem        `json:"topDevices"`        // 设备类型 Top10
	TopSystems        []TopItem        `json:"topSystems"`        // 操作系统 Top10
	TopBrowsers       []TopItem        `json:"topBrowsers"`       // 浏览器 Top10
	TopIsps           []TopItem        `json:"topIsps"`           // 运营商 / 组织 Top10
	WafBlocks         WafStat          `json:"wafBlocks"`         // 防火墙（coraza）拦截统计，含反向代理与穿透代理
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
	Country     string `json:"country"`     // 国家，未启用 GeoLite2 时为空
	CountryCode string `json:"countryCode"` // 国家代码，如 CN / US
	Province    string `json:"province"`    // 省份 / 州
	City        string `json:"city"`        // 城市
	Isp         string `json:"isp"`         // 运营商 / 自治域组织
	// 经纬度（地图打点用），0 表示该 IP 没定位到
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
}

// IpGeo 来源 IP 的归属地快照（不落库，仅用于 Top IP 榜展示）
type IpGeo struct {
	Country     string
	CountryCode string
	Province    string
	City        string
	Isp         string
	Latitude    float64
	Longitude   float64
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
	closed    bool
	onceClose sync.Once
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

// RecordHttpStat 记录一条请求统计，并在有大屏在线时广播出去。
// 与在线状态解耦：无论有没有人看都会累计（大屏连上来就能看到断连期间的累计数据）；
// 只有存在在线连接时才做 JSON 序列化与推送，无人观看时省掉这部分开销。
// 推送全程非阻塞：前端消费慢就丢弃，绝不能拖慢代理主流程。
func RecordHttpStat(rec *HttpStatRecord) {
	if rec == nil {
		return
	}
	aggregate(rec)
	if !HasWsClients() {
		return
	}
	data, err := json.Marshal(WsMessage{Type: "stat", Ts: time.Now().UnixMilli(), Data: rec})
	if err != nil {
		return
	}
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

// handleMessage 处理前端上行控制消息：ping
func (c *WsClient) handleMessage(msg []byte) {
	var payload struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg, &payload); err != nil {
		return
	}
	switch payload.Type {
	case "ping":
		c.send(WsMessage{Type: "pong", Ts: time.Now().UnixMilli(), Data: nil})
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

	aggMu           sync.Mutex
	aggStatusCounts = map[string]int64{}
	// 以下维度走滑动窗口：窗口外的 key 自动淘汰，新热点永远进得了榜。
	// path 尤其容易被带 ID 的 URL 打爆，其余维度则是基数天然不可控
	aggCountryCounts = newSlidingCounts()
	aggDomainCounts  = newSlidingCounts()
	aggIpCounts      = newSlidingCounts()
	aggPathCounts    = newSlidingCounts()
	// 客户端画像：UA 解析结果与运营商，同样五分钟窗口
	aggDeviceCounts  = newSlidingCounts()
	aggOsCounts      = newSlidingCounts()
	aggBrowserCounts = newSlidingCounts()
	aggIspCounts     = newSlidingCounts()
	// aggIpGeo IP -> 归属地，给 Top IP 榜补充来源信息；随 IP 滑出窗口一并清理
	aggIpGeo      = map[string]IpGeo{}
	aggQpsBuckets [60]struct {
		sec   int64
		count int64
	}
)

func init() {
	// IP 滑出窗口时联动清理归属地缓存，避免 aggIpGeo 单方面堆积
	aggIpCounts.onEvict = func(ip string) { delete(aggIpGeo, ip) }
}

// ---------------------------------------------------------------------------
// 滑动窗口计数
// ---------------------------------------------------------------------------

// slidingCounts 最近 wsTrackBuckets × wsTrackBucketSec 秒的滑动窗口计数器。
// 请求统计与防火墙统计共用这一个窗口长度（当前 5 小时）。
//
// 时间被切成固定数量的桶环形复用：每推进一格，就把最老那个桶的计数从 total 里扣掉，
// 扣到 0 即自动删除——淘汰是窗口滑出的自然结果，key 不会永久堆积，也不会出现
// "撞到上限后新 key 永远进不来" 的情况。
//
// total 恒等于窗口内各 key 的计数，Top 榜直接读它即可，无需遍历所有桶。
// 注意：所有方法都要求调用方持有各自的锁（请求统计 aggMu、防火墙 wafMu）。
type slidingCounts struct {
	buckets [wsTrackBuckets]map[string]int64
	curIdx  int              // 当前桶下标
	curEnd  int64            // 当前桶的截止时刻（unix 秒，不含）
	total   map[string]int64 // 窗口内累计，Top 榜读这张表
	onEvict func(key string) // key 被淘汰时的回调，用于联动清理附属表
}

func newSlidingCounts() *slidingCounts {
	s := &slidingCounts{total: make(map[string]int64)}
	for i := range s.buckets {
		s.buckets[i] = make(map[string]int64)
	}
	s.curIdx = 0
	s.curEnd = time.Now().Unix() + wsTrackBucketSec
	return s
}

// rotate 推进到 now 所在的桶，并把滑出窗口的桶从 total 中扣除。
func (s *slidingCounts) rotate(now int64) {
	if now < s.curEnd {
		return
	}
	span := int64(wsTrackBuckets) * wsTrackBucketSec
	if now >= s.curEnd+span {
		// 时间跳跃超过整个窗口（长时间无统计 / 时钟调整）：整表重置
		for i := range s.buckets {
			s.buckets[i] = make(map[string]int64)
		}
		s.total = make(map[string]int64)
		s.curIdx = 0
		s.curEnd = now + wsTrackBucketSec
		return
	}
	for now >= s.curEnd {
		s.curIdx = (s.curIdx + 1) % wsTrackBuckets
		old := s.buckets[s.curIdx]
		for k, v := range old {
			if remain := s.total[k] - v; remain > 0 {
				s.total[k] = remain
			} else {
				delete(s.total, k)
				if s.onEvict != nil {
					s.onEvict(k)
				}
			}
		}
		// 清空复用，避免每次 rotate 都重新分配一次 map
		for k := range old {
			delete(old, k)
		}
		s.curEnd += wsTrackBucketSec
	}
}

// add 记录一次命中
func (s *slidingCounts) add(now int64, key string) {
	if key == "" {
		return
	}
	s.rotate(now)
	if _, ok := s.total[key]; !ok && len(s.total) >= wsMaxTrackKeys {
		// 兜底：窗口内基数异常膨胀，淘汰一个低计数 key，保证新热点能进榜
		s.evictOne(key)
	}
	s.buckets[s.curIdx][key]++
	s.total[key]++
}

// evictOne 硬上限兜底：随机采样若干 key，淘汰计数最小的那个。
// Go 的 map 迭代起点是随机的，取前 wsEvictSamples 个即近似随机采样，开销 O(1)。
func (s *slidingCounts) evictOne(exclude string) {
	minKey := ""
	var minVal int64 = -1
	n := 0
	for k, v := range s.total {
		if k == exclude {
			continue
		}
		if minVal < 0 || v < minVal {
			minKey, minVal = k, v
		}
		n++
		if n >= wsEvictSamples {
			break
		}
	}
	if minKey == "" {
		return
	}
	delete(s.total, minKey)
	if s.onEvict != nil {
		s.onEvict(minKey)
	}
}

// topN 窗口内计数最高的 n 个 key
func (s *slidingCounts) topN(n int) []TopItem {
	items := make([]TopItem, 0, len(s.total))
	for k, v := range s.total {
		items = append(items, TopItem{Key: k, Count: v})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Count > items[j].Count })
	if len(items) > n {
		items = items[:n]
	}
	return items
}

// topIps 窗口内计数最高的 n 个 IP，并附上归属地
func (s *slidingCounts) topIps(geo map[string]IpGeo, n int) []TopIpItem {
	items := make([]TopIpItem, 0, len(s.total))
	for k, v := range s.total {
		item := TopIpItem{Key: k, Count: v}
		if g, ok := geo[k]; ok {
			item.Country = g.Country
			item.CountryCode = g.CountryCode
			item.Province = g.Province
			item.City = g.City
			item.Isp = g.Isp
			item.Latitude = g.Latitude
			item.Longitude = g.Longitude
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Count > items[j].Count })
	if len(items) > n {
		items = items[:n]
	}
	return items
}

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

	// 状态码只有 5 种取值，基数可控，直接全量累计
	incCount(aggStatusCounts, rec.StatusClass)
	// 其余维度走滑动窗口：窗口外的 key 自动淘汰，新热点永远能进榜
	aggCountryCounts.add(sec, rec.SourceCountry)
	aggDomainCounts.add(sec, rec.Domain)
	aggIpCounts.add(sec, rec.SourceIp)
	aggPathCounts.add(sec, normalizePath(rec.Path))
	// 客户端画像：挪到服务端统计，前端不用再从明细里抽样
	aggDeviceCounts.add(sec, rec.Device)
	aggOsCounts.add(sec, rec.Os)
	aggBrowserCounts.add(sec, rec.Browser)
	aggIspCounts.add(sec, rec.SourceIsp)
	if rec.SourceIp != "" {
		// 归属地每次刷新：同一个 IP 换出口/库更新后展示的是最新结果
		aggIpGeo[rec.SourceIp] = IpGeo{
			Country:     rec.SourceCountry,
			CountryCode: rec.SourceCountryCode,
			Province:    rec.SourceProvince,
			City:        rec.SourceCity,
			Isp:         rec.SourceIsp,
			Latitude:    rec.SourceLatitude,
			Longitude:   rec.SourceLongitude,
		}
	}
	aggMu.Unlock()
}

func incCount(m map[string]int64, key string) {
	if key == "" {
		return
	}
	m[key]++
}

// ---------------------------------------------------------------------------
// 路径归一化
// ---------------------------------------------------------------------------

// normalizePath 把路径里的动态段（ID / UUID / 哈希）替换成 *，
// 避免 /api/user/12345 这类 URL 每换一个 ID 就产生一个新 key，把 Top 榜打爆。
// 只作用于聚合维度；推送给前端的 stat 明细仍然是原始完整路径。
func normalizePath(p string) string {
	if p == "" {
		return ""
	}
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	if p == "" || p == "/" {
		return p
	}
	segs := strings.Split(p, "/")
	for i, seg := range segs {
		if isDynamicSeg(seg) {
			segs[i] = "*"
		}
	}
	return strings.Join(segs, "/")
}

// isDynamicSeg 判断路径段是否为动态值
func isDynamicSeg(s string) bool {
	switch {
	case s == "":
		return false
	case isAllDigits(s):
		// 纯数字：/api/user/12345 中的 12345
		return true
	case len(s) == 36 && strings.Count(s, "-") == 4:
		// UUID
		return true
	case len(s) >= 20 && isHex(s):
		// 长十六进制串：ObjectId / MD5 / SHA1 / SHA256 / 各类随机 ID
		return true
	}
	return false
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isHex(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			continue
		}
		return false
	}
	return true
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

	// 推进所有滑动窗口：即使当前没有流量，榜单也会随时间自然衰减
	aggCountryCounts.rotate(now)
	aggDomainCounts.rotate(now)
	aggIpCounts.rotate(now)
	aggPathCounts.rotate(now)
	aggDeviceCounts.rotate(now)
	aggOsCounts.rotate(now)
	aggBrowserCounts.rotate(now)
	aggIspCounts.rotate(now)

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
		TopCountries:      aggCountryCounts.topN(wsTopN),
		TopDomains:        aggDomainCounts.topN(wsTopN),
		TopIps:            aggIpCounts.topIps(aggIpGeo, wsTopN),
		TopPaths:          aggPathCounts.topN(wsTopN),
		TopDevices:        aggDeviceCounts.topN(wsTopN),
		TopSystems:        aggOsCounts.topN(wsTopN),
		TopBrowsers:       aggBrowserCounts.topN(wsTopN),
		TopIsps:           aggIspCounts.topN(wsTopN),
		WafBlocks:         BuildWafStat(),
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
