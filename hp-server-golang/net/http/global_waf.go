package http

import (
	"hp-server-lib/log"
	"hp-server-lib/service"
	"hp-server-lib/web/controller"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/debuglog"
	txhttp "github.com/corazawaf/coraza/v3/http"
	"github.com/corazawaf/coraza/v3/types"
)

/* ---------- 全局 Web 安全（coraza） ----------
   和 reverse_proxy.go 里那两份 WAF 的区别：那两份要绑定到某条穿透 / 反向代理上才生效，
   这一份是「系统级」的 —— 在 HTTP(80) / HTTPS(443) 入口各挂一次，对所有进站流量生效，
   不区分域名和归属用户，规则在后台改完即时重建，不用重启服务。

   挂载位置刻意放在 StatsInterceptor 里面：
     - 被拦下的请求同样要被统计到，大屏上才看得见「拦了多少、拦的谁」；
     - WAF 拿到的还是原始请求，不受统计包装器的影响。 */

// acmeChallengePath ACME HTTP-01 的挑战路径，全局规则必须放行它，
// 否则一条误伤的规则就能让证书续签失败，到期后整站 HTTPS 直接不可用
const acmeChallengePath = "/.well-known/acme-challenge/"

// GlobalWaf 把全局规则挂在 next 前面。没有启用任何规则时它就是一层直通。
func GlobalWaf(next http.Handler) http.Handler {
	m := &globalWafMount{next: next}
	// 启动就编译一次：别让第一个进来的请求替管理员承担编译开销
	m.rebuild()
	return m
}

// wafSnapshot 某一版规则编译出来的结果，整体替换，读的一边永远看到完整的一份
type wafSnapshot struct {
	ver     int64        // 这一份对应的规则版本号
	active  bool         // 是否真的挂着 WAF；false 时 handler 就是 next
	handler http.Handler // 包好的 handler
}

// globalWafMount 一个入口一份。
// 请求路径上只读一次原子指针，不抢锁 —— 这是所有进站流量的必经之处，
// 锁只在规则变更、需要重新编译时才会被抢一下
type globalWafMount struct {
	next http.Handler
	cur  atomic.Pointer[wafSnapshot]

	mu sync.Mutex // 只保护「检查版本 → 编译 → 换指针」这段
}

func (m *globalWafMount) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 证书挑战先放行，别让全局规则把续签卡死
	if strings.Contains(r.URL.RequestURI(), acmeChallengePath) {
		m.next.ServeHTTP(w, r)
		return
	}

	snap := m.current()
	if !snap.active {
		m.next.ServeHTTP(w, r)
		return
	}
	snap.handler.ServeHTTP(w, r)
}

// current 取当前生效的那一份快照，规则版本变了就重新编译
func (m *globalWafMount) current() *wafSnapshot {
	ver := service.GlobalWafVersion()
	if snap := m.cur.Load(); snap != nil && snap.ver == ver {
		return snap
	}
	return m.rebuild()
}

func (m *globalWafMount) rebuild() *wafSnapshot {
	ver := service.GlobalWafVersion()

	m.mu.Lock()
	defer m.mu.Unlock()
	// 进来之前可能已经被别的请求编译好了，先看一眼
	if snap := m.cur.Load(); snap != nil && snap.ver == ver {
		return snap
	}

	old := m.cur.Load()
	waf, err := buildGlobalWaf()
	if err != nil {
		/* 规则写错了不能让整站失去防护：保留上一版还能用的实例，只记日志。
		   版本号照样推进，否则之后每个请求都会重试一次注定失败的编译 */
		log.Errorf("全局Web安全规则编译失败，继续使用上一版规则：%v", err)
		snap := &wafSnapshot{ver: ver, handler: m.next}
		if old != nil {
			snap.active = old.active
			snap.handler = old.handler
		}
		m.cur.Store(snap)
		return snap
	}

	snap := &wafSnapshot{ver: ver, active: waf != nil}
	if waf == nil {
		snap.handler = m.next
	} else {
		snap.handler = txhttp.WrapHandler(waf, m.next)
	}
	m.cur.Store(snap)
	return snap
}

// buildGlobalWaf 用库里启用中的规则编译出一份 WAF。
// 返回 nil 表示当前没有启用任何规则（不是错误）
func buildGlobalWaf() (coraza.WAF, error) {
	directives, total := service.GlobalWafDirectives()
	if total == 0 {
		return nil, nil
	}

	wafConfig := coraza.NewWAFConfig().
		WithDebugLogger(debuglog.Default()).
		WithErrorCallback(func(rule types.MatchedRule) {
			msg := rule.ErrorLog()
			log.Errorf("[全局Web安全拦截][%s] %s\n", rule.Rule().Severity(), msg)
			/* 只做内存计数（总量 + 窗口榜单），不落单条明细。
			   域名这里留空：这份 WAF 挂在入口上，一次编译服务所有域名，
			   拦截回调只拿得到规则命中的信息，拿不到是被哪个域名拦的。
			   为了拿域名去把 r.URL 改成绝对地址代价太大 ——
			   REQUEST_URI 会跟着变成 http://host/xxx，管理员写的 ^/xxx 这类
			   规则会全部失效。域名维度的归属看 IP 榜和请求流，那里是逐条记的 */
			controller.RecordWafBlock(
				controller.WafSourceGlobal,
				"",
				rule.Rule().Severity().String(),
				rule.ClientIPAddress(),
				rule.Message(),
				rule.Rule().ID(),
			)
		}).
		WithDirectives(directives)

	waf, err := coraza.NewWAF(wafConfig)
	if err != nil {
		return nil, err
	}
	return waf, nil
}
