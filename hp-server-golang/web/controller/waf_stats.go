package controller

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// WafSource 防火墙挂载位置。全局入口、反向代理、穿透代理各挂一份 WAF，
// 统计合并在一处，靠这个字段区分是链路上的哪一段拦下的。
const (
	WafSourceReverse = "reverse" // 反向代理（绑在域名上）
	WafSourceTunnel  = "tunnel"  // 穿透代理（绑在穿透配置上）
	WafSourceGlobal  = "global"  // 全局（HTTP/HTTPS 入口统一挂载，不区分域名）
)

// wafSourceLabel 来源的展示名
var wafSourceLabel = map[string]string{
	WafSourceReverse: "反向代理",
	WafSourceTunnel:  "穿透代理",
	WafSourceGlobal:  "全局防护",
}

// WafStat 防火墙（coraza）实时拦截统计。
//
// 只保留聚合量，不存任何单条明细 —— 拦截回调是在代理主流程里同步执行的，
// 每多花一点时间都会直接压到每一条被拦的请求上，攻击者批量试探时尤其明显。
//
// 两条口径：
//   - Total / Severity / Source 是进程启动至今的累计值，不随窗口变化；
//   - TopDomains / TopIps / TopRules 和请求统计的 top* 同口径：
//     最近 5 小时滑动窗口，窗口外自动衰减淘汰（见 ws_stats.go 的 slidingCounts）。
type WafStat struct {
	Total      int64            `json:"total"`      // 累计规则命中次数
	Source     map[string]int64 `json:"source"`     // 按挂载位置的累计分布：reverse / tunnel
	Severity   map[string]int64 `json:"severity"`   // 按严重级别的累计分布，key 为 emergency~debug
	TopDomains []TopItem        `json:"topDomains"` // 近 5 小时被拦最多的域名 Top10
	TopIps     []TopItem        `json:"topIps"`     // 近 5 小时被拦最多的来源 IP Top10
	TopRules   []TopItem        `json:"topRules"`   // 近 5 小时命中最多的规则 Top10，key 为规则ID
	LastAt     int64            `json:"lastAt"`     // 最近一次命中的时间戳（毫秒），0 表示从未命中
	LastMsg    string           `json:"lastMsg"`    // 最近一次命中的摘要
}

var (
	wafTotal int64
	// 严重级别 8 种、挂载位置 3 种，基数固定，直接全量累计
	wafSeverity = map[string]int64{}
	wafSource   = map[string]int64{}
	// 域名 / IP / 规则三个维度基数不可控，走滑动窗口，和请求统计同一套 5 小时窗口。
	// 窗口长是因为异常流量稀疏：扫描器隔几分钟才来一次，短窗口每次只截到零星几次、
	// 一滑就清零，"同一个 IP 连着扫了几小时"这种最该看见的信息反而看不见。
	// 内存上限托底：窗口内 key 数超过 wsMaxTrackKeys（2w）时采样淘汰最小的 key，
	// 拉长窗口只是让慢攻击更容易攒够计数，不会无界增长。
	wafDomains = newSlidingCounts()
	wafIps     = newSlidingCounts()
	wafRules   = newSlidingCounts()
	wafLastAt  int64
	wafLastMsg string
	wafMu      sync.Mutex
)

// RecordWafBlock 记录一次防火墙规则命中。
// 传基本类型而不是 MatchedRule：controller 不该依赖 coraza，采集点的职责到此为止。
func RecordWafBlock(source, domain, severity, clientIp, msg string, ruleId int) {
	atomic.AddInt64(&wafTotal, 1)

	nowMs := time.Now().UnixMilli()
	summary := buildWafMsg(source, severity, msg, domain, clientIp, ruleId)

	wafMu.Lock()
	defer wafMu.Unlock()
	if source == "" {
		source = "unknown"
	}
	if severity == "" {
		severity = "unknown"
	}
	wafSource[source]++
	wafSeverity[severity]++
	sec := nowMs / 1000
	// 挂在入口上的那份全局 WAF 拿不到域名（见 net/http/global_waf.go），
	// 空域名不进域名榜，免得榜单上多一个空 key
	if domain != "" {
		wafDomains.add(sec, domain)
	}
	wafIps.add(sec, clientIp)
	wafRules.add(sec, strconv.Itoa(ruleId))
	wafLastAt = nowMs
	wafLastMsg = summary
}

// BuildWafStat 生成防火墙拦截面板数据
func BuildWafStat() WafStat {
	now := time.Now().Unix()
	wafMu.Lock()
	defer wafMu.Unlock()

	// 推进窗口：即使已经没人触发规则，榜单也会随时间自然衰减
	wafDomains.rotate(now)
	wafIps.rotate(now)
	wafRules.rotate(now)

	return WafStat{
		Total:      atomic.LoadInt64(&wafTotal),
		Source:     copyCounts(wafSource),
		Severity:   copyCounts(wafSeverity),
		TopDomains: wafDomains.topN(wsTopN),
		TopIps:     wafIps.topN(wsTopN),
		TopRules:   wafRules.topN(wsTopN),
		LastAt:     wafLastAt,
		LastMsg:    wafLastMsg,
	}
}

// buildWafMsg 拼一条人能看懂的命中摘要。
// 只留定位必需的信息，不带 URI / 请求体 —— 攻击载荷没必要在内存里再存一份
func buildWafMsg(source, severity, msg, domain, clientIp string, ruleId int) string {
	parts := make([]string, 0, 6)
	if label := wafSourceLabel[source]; label != "" {
		parts = append(parts, label)
	}
	parts = append(parts, "["+severity+"]")
	if ruleId > 0 {
		parts = append(parts, "规则"+strconv.Itoa(ruleId))
	}
	if s := shortText(msg, 100); s != "" {
		parts = append(parts, s)
	}
	if domain != "" {
		parts = append(parts, domain)
	}
	if clientIp != "" {
		parts = append(parts, clientIp)
	}
	return strings.Join(parts, " · ")
}

func shortText(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
