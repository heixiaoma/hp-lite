<template>
  <div ref="rootRef" class="dash">
    <div class="dash__bg"/>
    <div class="dash__scan"/>

    <div v-if="!isAdmin" class="dash__denied">
      <div class="denied__box">
        <span class="denied__mark">◈</span>
        <b>无访问权限</b>
        <span>流量天眼仅对管理员开放</span>
      </div>
    </div>

    <header class="dash__head">
      <div class="brand">
        <span class="brand__mark">◈</span>
        <div>
          <h1>流量天眼</h1>
          <p>REAL-TIME TRAFFIC MONITOR</p>
        </div>
      </div>

      <div class="dash__state" :class="`is-${status}`">
        <i/>
        <span>{{ statusText }}</span>
      </div>

      <div class="dash__tools">
        <div class="dash__filters">
          <button
              v-for="f in FILTERS"
              :key="f.value"
              :class="{on: filter === f.value}"
              @click="applyFilter(f.value)"
          >{{ f.label }}</button>
        </div>
        <span class="dash__clock">{{ clock }}</span>
        <button class="dash__btn" @click="toggleFullscreen">
          {{ isFull ? '退出全屏' : '全屏' }}
        </button>
      </div>
    </header>

    <section class="kpi">
      <div v-for="k in kpis" :key="k.label" class="kpi__card">
        <span class="kpi__label">{{ k.label }}</span>
        <span class="kpi__value" :style="{color: k.color}">{{ k.text }}</span>
        <span class="kpi__sub">{{ k.sub }}</span>
      </div>
    </section>

    <section class="board">
      <div class="panel panel--chart">
        <h2>
          实时吞吐<em>60s 窗口</em>
          <span class="chart__lg">
            <b class="is-qps"><i></i>REQ/S <u>{{ fmtNum(tQpsNow) }}</u></b>
            <b class="is-flux"><i></i>流量 <u>{{ fmtBps(tBpsNow) }}</u></b>
          </span>
        </h2>
        <div
            class="chart-box"
            ref="chartBoxRef"
            @mousemove="onChartMove"
            @mouseleave="onChartLeave"
        >
          <!-- 两条线各按自己的峰值缩放，所以峰值要分别标出来 -->
          <span class="chart__peak is-qps">峰值 {{ fmtNum(qpsPeak) }}</span>
          <span class="chart__peak is-flux">峰值 {{ fmtBps(bpsPeak) }}</span>
          <svg class="chart" viewBox="0 0 100 100" preserveAspectRatio="none">
            <defs>
              <linearGradient id="dashArea" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#0052d9" stop-opacity=".22"/>
                <stop offset="100%" stop-color="#0052d9" stop-opacity="0"/>
              </linearGradient>
            </defs>
            <path :d="areaPath" fill="url(#dashArea)"/>
            <path :d="linePath" class="chart__line" vector-effect="non-scaling-stroke"/>
            <path :d="flowPath" class="chart__flow" vector-effect="non-scaling-stroke"/>
          </svg>

          <!-- 十字线和取值圆点用 HTML 而不是 svg：preserveAspectRatio="none"
               会把 x 轴拉伸好几倍，svg 里的圆会被压成椭圆 -->
          <template v-if="hover">
            <span class="chart__cross" :style="{left: hover.x + '%'}"></span>
            <i class="chart__dot is-qps" :style="{left: hover.x + '%', top: hover.qy + '%'}"></i>
            <i class="chart__dot is-flux" :style="{left: hover.x + '%', top: hover.by + '%'}"></i>
            <div class="chart__tip" :style="tipStyle">
              <b>{{ hover.time }}<em>{{ hover.ago }}s前</em></b>
              <span class="is-qps"><i></i>REQ/S <u>{{ fmtNum(hover.qps) }}</u></span>
              <span class="is-flux"><i></i>流量 <u>{{ fmtBps(hover.bps) }}</u></span>
            </div>
          </template>
        </div>
        <div class="chart__axis">
          <span>-60s</span><span>-30s</span><span>now</span>
        </div>
      </div>

      <div class="panel panel--ring">
        <h2>状态码分布<em>累计</em></h2>
        <!-- 横排：面板宽度有余、高度紧张，图例放右侧才能在底部腾出指标条 -->
        <div class="ring-row">
          <div class="ring-wrap">
            <svg class="ring" viewBox="0 0 100 100">
              <g transform="rotate(-90 50 50)">
                <circle class="ring__track" cx="50" cy="50" r="38"/>
                <circle
                    v-for="s in ringSegs"
                    :key="s.cls"
                    class="ring__seg"
                    cx="50" cy="50" r="38"
                    :stroke="s.color"
                    :stroke-dasharray="s.dash"
                    :stroke-dashoffset="s.offset"
                />
              </g>
            </svg>
            <div class="ring__center">
              <b>{{ fmtNum(statusTotal) }}</b>
              <span>总请求</span>
            </div>
          </div>
          <ul class="legend">
            <li v-for="s in ringSegs" :key="s.cls">
              <i :style="{background: s.color}"/>
              <span>{{ s.cls }}</span>
              <b>{{ fmtNum(s.count) }}</b>
              <em>{{ s.pct }}</em>
            </li>
          </ul>
        </div>
        <div class="ring-stat">
          <span><i>成功率</i><b>{{ okRate }}</b></span>
          <span><i>4xx</i><b>{{ fmtNum(c4xx) }}</b></span>
          <span><i>5xx</i><b>{{ fmtNum(c5xx) }}</b></span>
        </div>
      </div>

      <div class="panel panel--profile">
        <h2>访客画像<em>本地最近 {{ pool.length }} 条</em></h2>
        <div class="profile">
          <div v-for="g in profileGroups" :key="g.label" class="profile__group">
            <span class="profile__label">{{ g.label }}</span>
            <ul class="profile__list">
              <li v-for="it in g.items" :key="it.key">
                <span class="profile__name" :title="`${it.key} · ${fmtNum(it.count)} 条`">{{ it.key }}</span>
                <b>{{ it.pct }}</b>
                <span class="profile__bar"><i :style="{width: it.pct}"/></span>
              </li>
              <li v-if="!g.items.length" class="profile__empty">暂无数据</li>
            </ul>
          </div>
        </div>
      </div>

      <div class="panel panel--top">
        <h2>TOP 来源国家<em>累计</em></h2>
        <ul class="rank">
          <li v-for="(it, i) in topCountries" :key="it.key">
            <span class="rank__no">{{ i + 1 }}</span>
            <span class="rank__name" :title="`${it.key} · ${fmtNum(it.count)} 条`">{{ it.key }}</span>
            <span class="rank__num">{{ fmtNum(it.count) }}</span>
            <span class="rank__bar">
              <i :style="{width: pct(it.count, topCountries)}"/>
            </span>
          </li>
          <li v-if="!topCountries.length" class="rank__empty">暂无数据</li>
        </ul>
      </div>

      <div class="panel panel--top">
        <h2>TOP 域名</h2>
        <ul class="rank">
          <li v-for="(it, i) in topDomains" :key="it.key">
            <span class="rank__no">{{ i + 1 }}</span>
            <span class="rank__name" :title="it.key">{{ it.key }}</span>
            <span class="rank__num">{{ fmtNum(it.count) }}</span>
            <span class="rank__bar">
              <i :style="{width: pct(it.count, topDomains)}"/>
            </span>
          </li>
          <li v-if="!topDomains.length" class="rank__empty">暂无数据</li>
        </ul>
      </div>

      <div class="panel panel--stream">
        <h2>
          实时请求流<em>最近 {{ stream.length }} 条 · 均 {{ streamStat.avg }}ms
          · 异常 {{ streamStat.bad }} · {{ fmtBytes(streamStat.bytes) }}</em>
        </h2>
        <ul v-if="stream.length" class="stream" :class="{'is-dense': dense}">
          <li v-for="r in stream" :key="r.id">
            <div class="stream__row">
              <span class="stream__time">{{ shortTime(r.time) }}</span>
              <span class="stream__method" :class="`m-${String(r.method || '').toLowerCase()}`">{{ r.method }}</span>
              <span class="stream__uri" :title="r.uri">{{ r.domain }}{{ r.path }}</span>
              <span class="stream__status" :class="`s-${r.statusClass}`">{{ r.statusCode }}</span>
              <span class="stream__dur">{{ r.durationMs }}ms</span>
              <span class="stream__size" :title="`${fmtNum(r.totalSize)} B`">{{ fmtBytes(r.totalSize) }}</span>
            </div>
            <!-- 第二行是来源画像：归属地 + 运营商 + 客户端，完整 UA 放在 title 里 -->
            <div class="stream__meta" :title="metaTitle(r)">
              <span class="stream__flag">{{ r.sourceCountryCode || '--' }}</span>
              <span class="stream__ip">{{ r.sourceIp }}</span>
              <span class="stream__geo">{{ geoText(r) }}</span>
              <!-- 境外 IP 常常没有运营商，空着比挂一个「-」干净 -->
              <span v-if="r.sourceIsp" class="stream__isp">{{ r.sourceIsp }}</span>
              <span v-if="r.scheme" class="stream__scheme">{{ r.scheme }}</span>
              <span v-if="r.protocol" class="stream__proto">{{ r.protocol }}</span>
              <span class="stream__ua">{{ clientText(r) }}</span>
            </div>
          </li>
        </ul>
        <div v-else class="empty">等待请求流入…</div>
      </div>

      <div class="panel panel--top">
        <h2>TOP 来源 IP<em>带归属地</em></h2>
        <ul class="rank">
          <li v-for="(it, i) in topIps" :key="it.key">
            <span class="rank__no">{{ i + 1 }}</span>
            <span class="rank__name" :title="it.key">{{ it.key }}</span>
            <span class="rank__num">{{ fmtNum(it.count) }}</span>
            <span class="rank__sub" :title="ipGeoText(it)">{{ ipGeoText(it) || '归属地未知' }}</span>
            <span class="rank__bar">
              <i :style="{width: pct(it.count, topIps)}"/>
            </span>
          </li>
          <li v-if="!topIps.length" class="rank__empty">暂无数据</li>
        </ul>
      </div>

      <div class="panel panel--top">
        <h2>TOP 路径</h2>
        <ul class="rank">
          <li v-for="(it, i) in topPaths" :key="it.key">
            <span class="rank__no">{{ i + 1 }}</span>
            <span class="rank__name" :title="it.key">{{ it.key }}</span>
            <span class="rank__num">{{ fmtNum(it.count) }}</span>
            <span class="rank__bar">
              <i :style="{width: pct(it.count, topPaths)}"/>
            </span>
          </li>
          <li v-if="!topPaths.length" class="rank__empty">暂无数据</li>
        </ul>
      </div>
    </section>
  </div>
</template>

<script setup>
import {computed, onMounted, onUnmounted, ref, watch} from 'vue';
import {createStatsSocket} from '../../data/statsSocket';
import userInfo from '../../data/userInfo';

const FILTERS = [
  {label: '全部', value: ''},
  {label: '2xx', value: '2xx'},
  {label: '4xx', value: '4xx'},
  {label: '5xx', value: '5xx'},
]

// 趋势图窗口 60 个点，对应最近 60 秒
const SERIES_LEN = 60
/* 请求流保留最近 2000 条。行数是 DOM 规模的大头：每条 2 行约 12 个节点，
   2000 条约 2.4w 个节点，靠 .stream li 上的 content-visibility 把屏幕外的行
   跳过渲染，滚动和更新才不卡。要继续加量得先上虚拟滚动 */
const STREAM_ROWS = 2000
/* 画像统计窗口：设备/系统/浏览器/运营商只在 stat 明细里、summary 不汇总，得自己数。
   500 条足够稳住榜首；不跟请求流一样存 2000 条，是因为这批数据每条进来都要重算四遍
   TOP（成本随条数线性涨），而分布的精度过了 500 条基本不再变化 */
const POOL_ROWS = 500

// 环形图分段色：白底需要更饱和的深色才看得清，深色主题的荧光色在白底上会发飘
const STATUS_CLASSES = [
  {cls: '1xx', color: '#8b5cf6'},
  {cls: '2xx', color: '#10b981'},
  {cls: '3xx', color: '#0891b2'},
  {cls: '4xx', color: '#f59e0b'},
  {cls: '5xx', color: '#ef4444'},
]

/* 菜单里已经按角色隐藏了入口，但直接敲 URL 仍会进到这里。
   大屏是全站流量数据，这里再拦一道：非管理员不建立连接 */
const isAdmin = computed(() => userInfo.getUserInfo()?.role === 'ADMIN')

const rootRef = ref(null)
const chartBoxRef = ref(null)
// 鼠标悬停命中的采样点下标，-1 表示光标不在图上
const hoverIdx = ref(-1)
const summary = ref({})
const stream = ref([])
// 只用于统计画像、不参与渲染，所以不放在 stream 里（stream 会被裁到 60 条）
const pool = ref([])
/* 每条新记录都跑一次入场动画，高 QPS 下会同时挂着几百个动画，帧率直接掉一半。
   批量太大时整批不做动画 —— 那个速度下本来也看不出单条滑入 */
const dense = ref(false)
// 每秒一个采样点：q=请求数，b=流量字节，t=时间戳(ms)。
// 三个字段必须落在同一条记录里，否则悬停取值会错位
const series = ref([])
const status = ref('connecting')
const statusDetail = ref('')
const filter = ref('')
const clock = ref('')
const isFull = ref(false)

let socket = null
// 本地每秒收到的 stat 条数。summary.qps 是服务端 60s 平均，变化太钝，
// 这里自己按秒计数拿瞬时值（文档 §6.4 正是这个建议）
let perSec = 0
// 每秒流量字节数：累加每条 stat 的 totalSize，同样按秒归零
let bytesPerSec = 0

/* ---------- 工具 ---------- */
const fmtNum = (n) => Math.round(Number(n) || 0).toLocaleString('en-US')

const fmtBytes = (b) => {
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  let n = Math.max(0, Number(b) || 0)
  let i = 0
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024
    i++
  }
  /* 字节档必须取整：useTween 缓动过程中传进来的是小数，
     直接渲染会变成 "123.4567890123 B/s" 这种又长又没意义的串 */
  return `${i === 0 ? Math.round(n) : n.toFixed(n < 10 ? 1 : 0)} ${units[i]}`
}

const fmtBps = (b) => `${fmtBytes(b)}/s`

/* 秒数同样要先取整（缓动中间值是小数），否则秒位会拖出一长串毫秒。
   另外天/时/分级都要保留秒，才有跑秒效果 */
const fmtUptime = (sec) => {
  const s = Math.max(0, Math.floor(Number(sec) || 0))
  const p = (n) => String(n).padStart(2, '0')
  const d = Math.floor(s / 86400)
  const h = Math.floor((s % 86400) / 3600)
  const m = Math.floor((s % 3600) / 60)
  const ss = p(s % 60)
  return d ? `${d}天${h}:${p(m)}:${ss}` : `${p(h)}:${p(m)}:${ss}`
}

/* 只取 HH:MM:SS。原来用 slice(11)，服务端时间一旦带纳秒或时区后缀
   （如 20:00:18.123456789Z）就会撑爆时间列、顶到后面的方法徽章上 */
const shortTime = (t) => {
  const m = String(t || '').match(/(\d{2}):(\d{2}):(\d{2})/)
  return m ? m[0] : '--:--:--'
}

/* 时间戳 -> HH:MM:SS，悬停时显示采样点时间用 */
const fmtClock = (ms) => {
  const d = new Date(Number(ms) || 0)
  const p = (n) => String(n).padStart(2, '0')
  return `${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

/* ---------- 来源画像解析 ----------
   stat 明细里带了归属地（国家/省/市/运营商）和客户端（浏览器/系统/设备），
   但大屏之前只用到了 IP，这些字段全被丢掉了。下面把它们还原出来 */

// 「四川省」这种全称在窄列里太占地方，去掉行政后缀只留「四川」
const ADMIN_SUFFIX = /(省|市|自治区|特别行政区)$/
const shortArea = (s) => String(s || '').trim().replace(ADMIN_SUFFIX, '')

/* 两套字段名：stat 明细是 sourceXxx，summary 的 topIps 是 xxx。
   统一成一套，后面拼文案就不用管数据来自哪 */
const asGeo = (o) => ({
  country: String((o && (o.sourceCountry || o.country)) || '').trim(),
  code: String((o && (o.sourceCountryCode || o.countryCode)) || '').trim(),
  province: String((o && (o.sourceProvince || o.province)) || '').trim(),
  city: String((o && (o.sourceCity || o.city)) || '').trim(),
  isp: String((o && (o.sourceIsp || o.isp)) || '').trim(),
})

/* 归属地：国内取「省·市」，境外一般没有省市，退回国家名 */
const geoText = (r) => {
  if (!r) return '未知'
  const g = asGeo(r)
  const prov = shortArea(g.province)
  const city = shortArea(g.city)
  if (prov && city) return prov === city ? city : `${prov}·${city}`
  if (prov || city) return prov || city
  return g.country || '未知'
}

/* 国家 + 归属地 + 运营商，境外请求省市为空时 geoText 会退回国家名，别拼两遍 */
const geoFullText = (r) => {
  const g = asGeo(r)
  const area = geoText(r)
  const parts = []
  if (g.country) parts.push(g.country)
  if (area && area !== g.country && area !== '未知') parts.push(area)
  if (g.isp) parts.push(g.isp)
  return parts.join(' · ')
}

const clientText = (r) =>
    [r.browser, r.os, r.device].filter(Boolean).join(' · ') || '-'

/* 悬停补全：一行塞不下完整 UA，放到 title 里 */
const metaTitle = (r) => [
  `IP ${r.sourceIp || '-'}`,
  [r.sourceCountry, r.sourceProvince, r.sourceCity].filter(Boolean).join(' '),
  r.sourceIsp,
  [r.scheme, r.protocol].filter(Boolean).join(' '),
  r.browser,
  r.os,
  r.device,
  r.userAgent,
].filter(Boolean).join(' · ')

const pct = (count, list) => {
  const max = (list || []).reduce((a, b) => Math.max(a, Number(b.count) || 0), 0)
  if (!max) return '0%'
  return `${Math.max(2, Math.round((Number(count) || 0) / max * 100))}%`
}

/* 数字滚动：大屏的标志性效果，从旧值缓动到新值 */
const useTween = (getter) => {
  const val = ref(0)
  let raf = 0
  watch(getter, (to) => {
    const from = val.value
    const target = Number(to) || 0
    const t0 = performance.now()
    cancelAnimationFrame(raf)
    const step = (now) => {
      const p = Math.min(1, (now - t0) / 700)
      val.value = from + (target - from) * (1 - Math.pow(1 - p, 3))
      if (p < 1) raf = requestAnimationFrame(step)
    }
    raf = requestAnimationFrame(step)
  }, {immediate: true})
  onUnmounted(() => cancelAnimationFrame(raf))
  return val
}

const tQps = useTween(() => summary.value.qps || 0)
const tTotal = useTween(() => summary.value.totalRequests || 0)
const tBytes = useTween(() => summary.value.totalBytes || 0)
const tDur = useTween(() => summary.value.avgDurationMs || 0)
const tUptime = useTween(() => summary.value.uptimeSec || 0)

/* ---------- 派生数据 ---------- */
const statusText = computed(() => ({
  connecting: '连接中…',
  open: '实时连接已建立',
  closed: '连接断开，重连中…',
  error: statusDetail.value || '连接异常',
}[status.value] || '初始化'))

const kpis = computed(() => [
  {
    label: '实时 QPS',
    text: fmtNum(tQps.value),
    sub: '60s 滑动均值',
    color: '#0052d9',
  },
  {
    label: '累计请求',
    text: fmtNum(tTotal.value),
    sub: '自首个大屏接入',
    color: '#10b981',
  },
  {
    label: '总流量',
    text: fmtBytes(tBytes.value),
    sub: `↑ ${fmtBytes(summary.value.totalRequestSize)}  ↓ ${fmtBytes(summary.value.totalResponseSize)}`,
    color: '#8b5cf6',
  },
  {
    label: '平均耗时',
    text: `${Math.round(tDur.value)} ms`,
    sub: '全量平均',
    color: '#f59e0b',
  },
  {
    label: '运行时长',
    text: fmtUptime(tUptime.value),
    sub: `${Number(summary.value.onlineClients) || 0} 个连接在线`,
    color: '#0891b2',
  },
])

const topDomains = computed(() => summary.value.topDomains || [])
const topIps = computed(() => rankSorted(summary.value.topIps))
const topPaths = computed(() => rankSorted(summary.value.topPaths))
/* 服务端 topCountries 是全量累计的国家榜，直接整榜用（面板里放不下就滚动），
   不截断成 TOP3 —— 截断后大半国家根本看不见，等于没用这份数据 */
const topCountries = computed(() => rankSorted(rankPct(summary.value.topCountries, 20)))

/* 按字段取 TOP n。
   分母用「识别出该字段的条数」而不是池子总量：字段为空的请求压根没进榜，
   拿全量当分母会把占比整体压低（比如大量爬虫没有 UA） */
const topBy = (field, n) => {
  const m = new Map()
  for (const r of pool.value) {
    const k = String((r && r[field]) || '').trim()
    if (!k) continue
    m.set(k, (m.get(k) || 0) + 1)
  }
  const total = [...m.values()].reduce((a, b) => a + b, 0)
  return [...m.entries()]
      .sort((a, b) => b[1] - a[1])
      .slice(0, n)
      .map(([key, count]) => ({
        key,
        count,
        pct: total ? `${Math.round((count / total) * 100)}%` : '0%',
      }))
}

/* 这四项服务端 summary 都不汇总，只能从本地池子里统计；
   来源国家服务端给了 topCountries，单独成榜（见 topCountries） */
const profileGroups = computed(() => [
  {label: '设备', items: topBy('device', 3)},
  {label: '系统', items: topBy('os', 3)},
  {label: '浏览器', items: topBy('browser', 3)},
  {label: '运营商', items: topBy('sourceIsp', 3)},
])

/* TOP 来源 IP 的归属地直接读服务端 topIps 里带的 country/province/city/isp。
   之前靠本地池子反查，样本只有 500 条且要等对应请求进来才有数据；
   服务端是累计统计，既准又不用等 */
const ipGeoText = (it) => geoFullText(it)

/* 榜单排序：先按次数降序，次数并列时再按名字升序。
   次数必须是第一维度 —— 榜的意义就是谁多谁在前。
   名字只做第二维度：并列项（常见的是末尾一堆 count=1）如果只按次数排，
   它们的先后全看服务端返回顺序，每次 summary 都可能不一样，看着就在乱跳；
   用名字兜底把这批钉死。
   numeric 让 IP 按段数值排（192.168.1.2 在 .10 前面），sensitivity: base 忽略大小写 */
const NAME_OPT = {numeric: true, sensitivity: 'base'}
const rankSorted = (list) => (list || []).slice().sort((a, b) => {
  const diff = (Number(b && b.count) || 0) - (Number(a && a.count) || 0)
  return diff || String(a && a.key).localeCompare(String(b && b.key), 'zh-Hans-CN', NAME_OPT)
})

/* 服务端 TOP 榜 -> 面板用的 {key, count, pct}。国家榜按全量累计，不是抽样 */
const rankPct = (list, n) => {
  const src = list || []
  const total = src.reduce((a, b) => a + (Number(b.count) || 0), 0)
  return src.slice(0, n).map((it) => ({
    key: it.key,
    count: Number(it.count) || 0,
    pct: total ? `${Math.round(((Number(it.count) || 0) / total) * 100)}%` : '0%',
  }))
}

const statusTotal = computed(() => {
  const c = summary.value.statusCounts || {}
  return Object.values(c).reduce((a, b) => a + (Number(b) || 0), 0)
})

// 环形图：周长固定，每段用 dasharray 表示占比，dashoffset 累加偏移起点
const RING_C = 2 * Math.PI * 38
const ringSegs = computed(() => {
  const c = summary.value.statusCounts || {}
  const total = statusTotal.value
  let acc = 0
  return STATUS_CLASSES.map((s) => {
    const count = Number(c[s.cls]) || 0
    const frac = total ? count / total : 0
    const seg = {
      cls: s.cls,
      color: s.color,
      count,
      pct: total ? `${(frac * 100).toFixed(1)}%` : '0.0%',
      dash: `${(frac * RING_C).toFixed(3)} ${(RING_C - frac * RING_C).toFixed(3)}`,
      offset: `${(-acc * RING_C).toFixed(3)}`,
    }
    acc += frac
    return seg
  })
})

/* 成功率与异常数：光看占比图看不出「健康度」，补几个派生指标 */
const c4xx = computed(() => Number((summary.value.statusCounts || {})['4xx']) || 0)
const c5xx = computed(() => Number((summary.value.statusCounts || {})['5xx']) || 0)
const badCount = computed(() => c4xx.value + c5xx.value)

const okRate = computed(() => {
  const t = statusTotal.value
  return t ? `${(((t - badCount.value) / t) * 100).toFixed(1)}%` : '--'
})

/* 请求流窗口内汇总：标题原来只有「最近 N 条」，信息量太少，
   补上窗口均耗时、异常数、流量合计 */
const streamStat = computed(() => {
  const d = stream.value
  if (!d.length) return {avg: 0, bad: 0, bytes: 0}
  let dur = 0
  let bad = 0
  let bytes = 0
  for (const r of d) {
    dur += Number(r.durationMs) || 0
    const c = String(r.statusClass || '')
    if (c === '4xx' || c === '5xx') bad++
    bytes += Number(r.totalSize) || 0
  }
  return {avg: Math.round(dur / d.length), bad, bytes}
})

/* 折线图：REQ/S 与流量两条曲线共用 x 轴，但各自按自己的峰值独立缩放 ——
   两者量级差好几个数量级，共用一个 y 轴会把流量压成一条贴底的直线 */
const toPoints = (d) => {
  // peak 是真实峰值（用于刻度文字），缩放用的分母至少给 1 避免除零
  const peak = d.reduce((a, b) => Math.max(a, Number(b) || 0), 0)
  const max = Math.max(1, peak)
  return {
    peak,
    pts: d.map((v, i) => ({
      x: (i / (SERIES_LEN - 1)) * 100,
      y: 94 - (v / max) * 84,
    })),
  }
}

/* 开局数据不满时前面补空点，避免线条横向拉伸。
   补出来的点按 1s 间隔往前推时间戳，这样悬停到开局那段也有时间可读 */
const padded = computed(() => {
  const d = series.value.slice(-SERIES_LEN)
  const head = d.length ? d[0].t : Date.now()
  const pad = []
  for (let i = d.length; i < SERIES_LEN; i++) {
    pad.push({q: 0, b: 0, t: head - (SERIES_LEN - i) * 1000})
  }
  return pad.concat(d)
})

const qpsChart = computed(() => toPoints(padded.value.map((p) => p.q)))
const bpsChart = computed(() => toPoints(padded.value.map((p) => p.b)))

const toLine = (pts) => (pts.length
    ? 'M ' + pts.map((pt) => `${pt.x.toFixed(2)},${pt.y.toFixed(2)}`).join(' L ')
    : '')

const linePath = computed(() => toLine(qpsChart.value.pts))
const flowPath = computed(() => toLine(bpsChart.value.pts))

const areaPath = computed(() => {
  const p = qpsChart.value.pts
  if (!p.length) return ''
  return `M 0,100 L ` + p.map((pt) => `${pt.x.toFixed(2)},${pt.y.toFixed(2)}`).join(' L ') + ' L 100,100 Z'
})

/* 光有曲线看不出量级，图例给当前值、图上给窗口峰值 */
const qpsPeak = computed(() => qpsChart.value.peak)
const bpsPeak = computed(() => bpsChart.value.peak)
const lastPoint = computed(() => padded.value[padded.value.length - 1] || {q: 0, b: 0, t: 0})
const tQpsNow = useTween(() => lastPoint.value.q)
const tBpsNow = useTween(() => lastPoint.value.b)

/* ---------- 悬停取值 ----------
   光标横向位置换算成采样点下标，再取该点的 REQ/S、流量和时间 */
const onChartMove = (e) => {
  const box = chartBoxRef.value
  if (!box) return
  const r = box.getBoundingClientRect()
  if (!r.width) return
  const ratio = (e.clientX - r.left) / r.width
  hoverIdx.value = Math.min(SERIES_LEN - 1, Math.max(0, Math.round(ratio * (SERIES_LEN - 1))))
}

const onChartLeave = () => {
  hoverIdx.value = -1
}

const hover = computed(() => {
  const i = hoverIdx.value
  const d = padded.value
  if (i < 0 || !d[i]) return null
  const p = d[i]
  return {
    idx: i,
    x: (i / (SERIES_LEN - 1)) * 100,
    qps: p.q,
    bps: p.b,
    time: fmtClock(p.t),
    ago: SERIES_LEN - 1 - i,
    qy: qpsChart.value.pts[i].y,
    by: bpsChart.value.pts[i].y,
  }
})

/* tooltip 贴边时翻转对齐，否则会被面板裁掉 */
const tipStyle = computed(() => {
  const h = hover.value
  if (!h) return null
  const t = h.x < 18 ? 'translateX(0)' : (h.x > 82 ? 'translateX(-100%)' : 'translateX(-50%)')
  return {left: `${h.x}%`, transform: t}
})

/* ---------- 数据接入 ---------- */
const onSummary = (data) => {
  summary.value = data || {}
  series.value.push({q: perSec, b: bytesPerSec, t: Date.now()})
  if (series.value.length > SERIES_LEN) series.value.shift()
  perSec = 0
  bytesPerSec = 0
}

const onStats = (batch) => {
  perSec += batch.length
  // 每条 stat 的 totalSize 就是这一跳的完整流量（请求+响应，含头）
  bytesPerSec += batch.reduce((a, r) => a + (Number(r && r.totalSize) || 0), 0)
  const fresh = batch.slice().reverse()
  // 200ms 一批，> 12 条即 > 60 QPS：这个速度下相邻记录看不出先后，动画纯属浪费
  dense.value = fresh.length > 12
  // 新的在前；只保留最近 STREAM_ROWS 条，超出的直接丢
  stream.value = fresh.concat(stream.value).slice(0, STREAM_ROWS)
  // 画像池同样只留最近 POOL_ROWS 条，够统计又不至于一直涨
  pool.value = fresh.concat(pool.value).slice(0, POOL_ROWS)
}

const applyFilter = (v) => {
  filter.value = v
  // 过滤只作用于 stat 明细，summary 是全量大盘不受影响
  socket && socket.setFilter(v ? {statusClass: v} : {})
}

/* ---------- 全屏 ---------- */
const toggleFullscreen = () => {
  const el = rootRef.value
  if (!el) return
  if (document.fullscreenElement) {
    document.exitFullscreen()
  } else {
    el.requestFullscreen && el.requestFullscreen()
  }
}

const onFullscreenChange = () => {
  isFull.value = !!document.fullscreenElement
}

let clockTimer = null

onMounted(() => {
  const tick = () => {
    clock.value = new Date().toLocaleString('zh-CN', {hour12: false})
  }
  tick()
  clockTimer = setInterval(tick, 1000)

  document.addEventListener('fullscreenchange', onFullscreenChange)

  if (!isAdmin.value) {
    status.value = 'error'
    statusDetail.value = '仅限管理员访问'
    return
  }

  socket = createStatsSocket({
    onSummary,
    onStats,
    onStatus: (s, detail) => {
      status.value = s
      statusDetail.value = detail || ''
    },
  })
})

onUnmounted(() => {
  clearInterval(clockTimer)
  document.removeEventListener('fullscreenchange', onFullscreenChange)
  socket && socket.dispose()
  socket = null
})
</script>

<style scoped>
.dash {
  /* 白底主题：整套配色收敛在这几个变量里，改这里就能整体换肤 */
  --c-text: #1f2d3d;
  --c-sub: #64748b;
  --c-faint: #94a3b8;
  --c-border: #e3e8ef;
  --c-card: #fff;
  --c-brand: #0052d9;

  position: relative;
  /* 抵消 .console__content 的 22px padding，让大屏铺满主内容区 */
  margin: -22px;
  height: calc(100vh - 96px);
  min-height: 560px;
  padding: 14px 18px 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  border-radius: var(--hp-radius);
  overflow: hidden;
  color: var(--c-text);
  font-variant-numeric: tabular-nums;
  background: linear-gradient(180deg, #fafcfe 0%, #f1f5fb 100%);
}

/* ---------- 背景装饰 ---------- */
.dash__bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image:
      linear-gradient(rgba(0, 82, 217, .05) 1px, transparent 1px),
      linear-gradient(90deg, rgba(0, 82, 217, .05) 1px, transparent 1px);
  background-size: 40px 40px;
}

.dash__scan {
  position: absolute;
  left: 0;
  right: 0;
  height: 130px;
  pointer-events: none;
  background: linear-gradient(180deg, transparent, rgba(0, 82, 217, .05), transparent);
  animation: scan 8s linear infinite;
}

@keyframes scan {
  0% { top: -130px; }
  100% { top: 100%; }
}

.dash__denied {
  position: absolute;
  inset: 0;
  z-index: 5;
  display: grid;
  place-items: center;
  background: rgba(255, 255, 255, .94);
}

.denied__box {
  text-align: center;
  line-height: 2;
}

.denied__mark {
  display: block;
  font-size: 40px;
  color: #ef4444;
}

.denied__box b {
  display: block;
  font-size: 18px;
  letter-spacing: 3px;
  color: var(--c-text);
}

.denied__box span:last-child {
  font-size: 12px;
  letter-spacing: 1px;
  color: var(--c-sub);
}

/* ---------- 顶栏 ---------- */
.dash__head {
  position: relative;
  display: flex;
  align-items: center;
  gap: 18px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--c-border);
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand__mark {
  font-size: 26px;
  line-height: 1;
  color: var(--c-brand);
  animation: pulse 2.8s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: .45; }
}

.brand h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 5px;
  color: var(--c-text);
}

.brand p {
  margin: 2px 0 0;
  font-size: 10px;
  letter-spacing: 3px;
  color: var(--c-faint);
}

.dash__state {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 12px;
  color: var(--c-sub);
}

.dash__state i {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--c-faint);
}

/* 白底上发光会糊成一团，改成柔和光圈表达状态 */
.dash__state.is-open i {
  background: #10b981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, .18);
  animation: blink 1.5s ease-in-out infinite;
}

.dash__state.is-connecting i,
.dash__state.is-closed i {
  background: #f59e0b;
  box-shadow: 0 0 0 3px rgba(245, 158, 11, .18);
  animation: blink .8s ease-in-out infinite;
}

.dash__state.is-closed i {
  animation: none;
}

.dash__state.is-error i {
  background: #ef4444;
  box-shadow: 0 0 0 3px rgba(239, 68, 68, .18);
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: .3; }
}

.dash__tools {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 12px;
}

.dash__filters {
  display: flex;
  gap: 4px;
}

.dash__filters button,
.dash__btn {
  padding: 3px 11px;
  font-size: 11px;
  color: var(--c-sub);
  background: var(--c-card);
  border: 1px solid var(--c-border);
  border-radius: 6px;
  cursor: pointer;
  transition: color .2s, background .2s, border-color .2s;
}

.dash__filters button:hover,
.dash__btn:hover {
  color: var(--c-brand);
  background: #f0f5ff;
  border-color: #b8d0ff;
}

.dash__filters button.on {
  color: #fff;
  font-weight: 600;
  background: var(--c-brand);
  border-color: var(--c-brand);
  box-shadow: 0 2px 8px rgba(0, 82, 217, .28);
}

.dash__clock {
  font-size: 12px;
  color: var(--c-sub);
}

/* ---------- KPI ---------- */
.kpi {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
}

.kpi__card {
  position: relative;
  padding: 10px 14px;
  border: 1px solid var(--c-border);
  border-radius: 10px;
  overflow: hidden;
  background: var(--c-card);
  box-shadow: 0 2px 10px rgba(31, 45, 61, .05);
}

/* 顶部一道品牌色高亮，用来替代深色主题的霓虹描边 */
.kpi__card::after {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 2px;
  background: linear-gradient(90deg, var(--c-brand), rgba(0, 82, 217, .12));
}

.kpi__label {
  font-size: 11px;
  letter-spacing: 1px;
  color: var(--c-sub);
}

.kpi__value {
  display: block;
  margin: 3px 0 2px;
  font-size: 25px;
  font-weight: 700;
  line-height: 1.15;
  font-family: 'DIN Alternate', 'SF Mono', Consolas, Menlo, monospace;
}

.kpi__sub {
  font-size: 10px;
  color: var(--c-faint);
}

/* ---------- 面板 ---------- */
/* 12 列 2 行。用 12 而不是 6，是因为各面板需要的宽度差得比较多：
   6 列时状态码只能给 1/6（太窄，环和图例挤不下）或 2/6（太宽，内容撑不满、留白一堆）。
   上行：吞吐 4 / 状态码 3 / 访客画像 3 / 来源国家 2；
   下行：请求流 6（半宽，两行信息才排得开）+ 域名 / IP / 路径 各 2 */
.board {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  grid-template-rows: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.panel {
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: 0;
  /* 面板高度由 grid 决定，内容超了就地裁掉。
     不加这句时环形图的图例会溢出面板，压到相邻面板上 */
  overflow: hidden;
  padding: 10px 12px;
  border: 1px solid var(--c-border);
  border-radius: 10px;
  background: var(--c-card);
  box-shadow: 0 2px 10px rgba(31, 45, 61, .05);
}

/* 四角装饰角标 */
.panel::before,
.panel::after {
  content: '';
  position: absolute;
  width: 10px;
  height: 10px;
  border: 1px solid var(--c-brand);
  opacity: .28;
}

.panel::before {
  left: -1px;
  top: -1px;
  border-right: none;
  border-bottom: none;
}

.panel::after {
  right: -1px;
  bottom: -1px;
  border-left: none;
  border-top: none;
}

.panel--chart {
  grid-column: span 4;
}

.panel--ring {
  grid-column: span 3;
}

/* 画像只剩 4 组（国家单独成榜了），3 列即可；省下的 2 列给 TOP 来源国家 */
.panel--profile {
  grid-column: span 3;
}

.panel--stream {
  grid-column: span 6;
}

.panel--top {
  grid-column: span 2;
}

.panel h2 {
  margin: 0 0 8px;
  display: flex;
  align-items: baseline;
  gap: 8px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 2px;
  color: var(--c-text);
}

.panel h2 em {
  font-style: normal;
  font-size: 10px;
  letter-spacing: 1px;
  color: var(--c-faint);
}

/* ---------- 折线图 ---------- */
.chart-box {
  position: relative;
  flex: 1;
  min-height: 0;
}

/* 参考线画在绘图区而不是整个面板，才能和曲线对齐 */
.chart-box::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image: linear-gradient(180deg, #dfe7f0 1px, transparent 1px);
  background-size: 100% 25%;
}

.chart {
  display: block;
  width: 100%;
  height: 100%;
}

/* 白底上线条发光会发灰发脏，只保留干净的实色描边 */
.chart__line {
  fill: none;
  stroke: var(--c-brand);
  stroke-width: 2;
  stroke-linejoin: round;
  stroke-linecap: round;
}

/* 流量曲线：虚线 + 紫色，和 REQ/S 的实线蓝区分开 */
.chart__flow {
  fill: none;
  stroke: #7c4dff;
  stroke-width: 1.6;
  stroke-linejoin: round;
  stroke-linecap: round;
  stroke-dasharray: 4 2.5;
}

/* 图例：右侧对齐，带当前值 */
.chart__lg {
  margin-left: auto;
  display: flex;
  flex-wrap: wrap;
  gap: 4px 12px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0;
}

.chart__lg b {
  display: flex;
  align-items: center;
  gap: 4px;
}

.chart__lg i {
  width: 10px;
  height: 2px;
  border-radius: 1px;
}

.chart__lg u {
  text-decoration: none;
  font-variant-numeric: tabular-nums;
}

.chart__lg .is-qps,
.chart__peak.is-qps {
  color: var(--c-brand);
}

.chart__lg .is-flux,
.chart__peak.is-flux {
  color: #7c4dff;
}

.chart__lg .is-qps i {
  background: var(--c-brand);
}

.chart__lg .is-flux i {
  background: #7c4dff;
}

/* 峰值贴在绘图区顶部两角；半透明白底保证压在曲线上也能看清 */
.chart__peak {
  position: absolute;
  top: 0;
  z-index: 1;
  padding: 0 3px;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0;
  border-radius: 2px;
  background: rgba(255, 255, 255, .82);
}

.chart__peak.is-qps {
  left: 0;
}

.chart__peak.is-flux {
  right: 0;
}

/* ---------- 悬停取值 ---------- */
.chart-box:hover {
  cursor: crosshair;
}

.chart__cross {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 1px;
  margin-left: -.5px;
  pointer-events: none;
  background: rgba(31, 45, 61, .38);
}

.chart__dot {
  position: absolute;
  width: 7px;
  height: 7px;
  margin: -3.5px 0 0 -3.5px;
  border-radius: 50%;
  border: 1.5px solid #fff;
  pointer-events: none;
}

.chart__dot.is-qps {
  background: var(--c-brand);
}

.chart__dot.is-flux {
  background: #7c4dff;
}

.chart__tip {
  position: absolute;
  top: 0;
  z-index: 2;
  display: flex;
  flex-direction: column;
  gap: 1px;
  padding: 4px 7px;
  font-size: 10.5px;
  font-weight: 600;
  line-height: 1.45;
  letter-spacing: 0;
  white-space: nowrap;
  border: 1px solid var(--c-border);
  border-radius: 4px;
  background: rgba(255, 255, 255, .97);
  box-shadow: 0 3px 10px rgba(31, 45, 61, .12);
  pointer-events: none;
}

.chart__tip b {
  display: flex;
  align-items: baseline;
  gap: 4px;
  color: var(--c-text);
}

.chart__tip em {
  font-style: normal;
  font-weight: 400;
  color: var(--c-faint);
}

.chart__tip span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.chart__tip i {
  width: 8px;
  height: 2px;
  border-radius: 1px;
}

.chart__tip u {
  text-decoration: none;
  font-variant-numeric: tabular-nums;
}

.chart__tip .is-qps {
  color: var(--c-brand);
}

.chart__tip .is-flux {
  color: #7c4dff;
}

.chart__tip .is-qps i {
  background: var(--c-brand);
}

.chart__tip .is-flux i {
  background: #7c4dff;
}

.chart__axis {
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  color: var(--c-faint);
}

/* ---------- 环形图 ----------
   横排布局：环形图固定宽，图例吃掉剩余宽度，
   这样底部还能留一整条给成功率 / 4xx / 5xx */
/* 吃掉标题以下的全部高度，环和图例才能跟着面板一起长大，
   否则面板一高就露出一大截空白（之前环写死 96px，240px 的面板里空着一半） */
.ring-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-height: 0;
}

/* 宽高都交给外层，svg 自己按 viewBox 居中缩放（不变形）。
   面板矮的时候由高度兜住，不会把图例挤出去 */
.ring-wrap {
  position: relative;
  flex: none;
  width: 48%;
  max-width: 240px;
  height: 100%;
  min-height: 0;
}

.ring {
  display: block;
  width: 100%;
  height: 100%;
}

.ring__track {
  fill: none;
  stroke: #eef2f7;
  stroke-width: 9;
}

.ring__seg {
  fill: none;
  stroke-width: 9;
  transition: stroke-dasharray .6s ease, stroke-dashoffset .6s ease;
}

.ring__center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  line-height: 1.2;
}

/* 环大了，中心的数字也跟着放大才压得住 */
.ring__center b {
  display: block;
  font-size: 20px;
  color: var(--c-text);
}

.ring__center span {
  font-size: 11px;
  color: var(--c-faint);
}

/* align-content: center 让图例在剩下的高度里居中，行距也拉开一点 */
.legend {
  flex: 1;
  min-width: 0;
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  align-content: center;
  gap: 8px;
  font-size: 12.5px;
  line-height: 1.3;
}

/* 行内留白把图例撑起来，环那侧才不会显得孤零零的 */
.legend li {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 0;
  color: var(--c-sub);
}

.legend i {
  flex: none;
  width: 8px;
  height: 8px;
  border-radius: 2px;
}

.legend b {
  margin-left: auto;
  color: var(--c-text);
  font-variant-numeric: tabular-nums;
}

/* 占比固定占一列并右对齐，数字跳动时不会把整行挤歪 */
.legend em {
  flex: none;
  width: 32px;
  text-align: right;
  font-style: normal;
  color: var(--c-faint);
  font-variant-numeric: tabular-nums;
}

/* 底部指标条：margin-top:auto 让它贴着面板底边，把剩余空间填满 */
.ring-stat {
  display: flex;
  gap: 6px;
  margin-top: auto;
  padding-top: 10px;
  border-top: 1px solid var(--c-border);
}

.ring-stat > span {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.ring-stat i {
  font-style: normal;
  font-size: 10.5px;
  white-space: nowrap;
  color: var(--c-faint);
}

.ring-stat b {
  font-size: 15px;
  font-weight: 700;
  font-family: 'DIN Alternate', 'SF Mono', Consolas, Menlo, monospace;
  font-variant-numeric: tabular-nums;
  color: var(--c-text);
}

/* ---------- 排行 ---------- */
.rank {
  margin: 0;
  padding: 0 2px 0 0;
  list-style: none;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 12px;
}

.rank li {
  display: grid;
  grid-template-columns: 17px minmax(0, 1fr) auto;
  align-items: center;
  gap: 3px 8px;
}

.rank__no {
  display: grid;
  place-items: center;
  width: 17px;
  height: 17px;
  font-size: 10px;
  font-weight: 700;
  color: var(--c-sub);
  background: #eef2f7;
  border-radius: 3px;
}

.rank li:nth-child(1) .rank__no {
  color: #fff;
  background: var(--c-brand);
}

.rank li:nth-child(2) .rank__no {
  color: #fff;
  background: #10b981;
}

.rank li:nth-child(3) .rank__no {
  color: #fff;
  background: #f59e0b;
}

.rank__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--c-text);
}

.rank__num {
  color: var(--c-sub);
}

/* 归属地副标题占第二行，和进度条一起排在名称列下面 */
.rank__sub {
  grid-column: 2 / -1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 10px;
  color: var(--c-faint);
}

.rank__bar {
  grid-column: 1 / -1;
  height: 3px;
  border-radius: 2px;
  overflow: hidden;
  background: #eef2f7;
}

.rank__bar i {
  display: block;
  height: 100%;
  border-radius: 2px;
  background: linear-gradient(90deg, #0052d9, #21b7e8);
  transition: width .5s ease;
}

.rank__empty {
  display: block;
  color: var(--c-faint);
  font-size: 11px;
}

/* ---------- 访客画像 ----------
   设备 / 系统 / 浏览器 / 运营商四组（sourceCountry 不在服务端 summary 里，
   来源国家走独立的 TOP 榜），每组取 TOP3。2×2 排布，竖着一列会顶出面板 */
.profile {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-content: start;
  gap: 12px 16px;
}

.profile__label {
  display: block;
  font-size: 11px;
  letter-spacing: 1px;
  color: var(--c-faint);
}

.profile__list {
  margin: 4px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 6px;
}

.profile__list li {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 32px;
  align-items: center;
  gap: 3px 6px;
  font-size: 12px;
  line-height: 1.3;
}

.profile__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--c-text);
}

.profile__list b {
  font-size: 10px;
  font-weight: 600;
  text-align: right;
  color: var(--c-sub);
  font-variant-numeric: tabular-nums;
}

.profile__bar {
  grid-column: 1 / -1;
  height: 4px;
  border-radius: 2px;
  overflow: hidden;
  background: #eef2f7;
}

.profile__bar i {
  display: block;
  height: 100%;
  border-radius: 2px;
  background: linear-gradient(90deg, #0052d9, #21b7e8);
  transition: width .5s ease;
}

.profile__empty {
  font-size: 10.5px;
  color: var(--c-faint);
}

/* ---------- 请求流 ---------- */
.stream {
  margin: 0;
  padding: 0 2px 0 0;
  list-style: none;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  font-size: 11.5px;
  font-family: 'SF Mono', Consolas, Menlo, monospace;
}

/* 一条记录两行：第一行是请求本身，第二行是来源画像。
   归属地/运营商/客户端原本就随 stat 带回来了，硬塞进同一行会把地址列压得没法看 */
/* 屏幕外的行整块跳过渲染 —— 2000 条全部参与布局会拖垮滚动。
   contain-intrinsic-size 给未渲染行一个预估高度（约等于两行的实际高度），
   滚动条长度才不会一跳一跳 */
.stream li {
  padding: 2px 0;
  border-bottom: 1px solid #f1f5f9;
  animation: slidein .32s ease;
  content-visibility: auto;
  contain-intrinsic-size: auto 34px;
}

/* 时间 / 方法 / 地址 / 状态 / 耗时 / 流量 */
.stream__row {
  display: grid;
  grid-template-columns: 62px 40px minmax(0, 1fr) 30px 40px 48px;
  align-items: center;
  gap: 7px;
  line-height: 1.35;
}

/* 左侧一道竖线把副行挂在主行下面，视觉上仍是一条记录 */
.stream__meta {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin: 1px 0 0 4px;
  padding-left: 6px;
  border-left: 2px solid var(--c-border);
  font-size: 10px;
  line-height: 1.4;
  color: var(--c-faint);
}

/* 每一项都可能超宽，统一省略号，免得某一项把整行撑破 */
.stream__meta > span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 国家代码徽章：一眼区分境内外 */
.stream__flag {
  flex: none;
  padding: 0 4px;
  border-radius: 3px;
  font-size: 9px;
  font-weight: 700;
  letter-spacing: .5px;
  color: var(--c-brand);
  background: rgba(0, 82, 217, .1);
}

.stream__geo {
  color: var(--c-sub);
}

/* 运营商用暖色底，和灰色的 IP / 归属地拉开层次 */
.stream__isp {
  flex: none;
  padding: 0 4px;
  border-radius: 3px;
  color: #b45309;
  background: rgba(245, 158, 11, .12);
}

/* scheme：https 用青色标出来，明文 http 一眼能看出来用的是灰底 */
.stream__scheme {
  flex: none;
  padding: 0 4px;
  border-radius: 3px;
  color: #0f766e;
  background: rgba(13, 148, 136, .12);
}

/* 协议版本（HTTP/2.0 等），中性灰，不抢主信息的注意力 */
.stream__proto {
  flex: none;
  padding: 0 4px;
  border-radius: 3px;
  color: var(--c-sub);
  background: #eef2f7;
}

/* 客户端信息吃掉副行剩余宽度，并靠右对齐 */
.stream__ua {
  margin-left: auto;
  color: var(--c-sub);
}

@keyframes slidein {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: none; }
}

/* 高 QPS 下关掉入场动画（dense 由每次批量的条数决定） */
.stream.is-dense li {
  animation: none;
}

/* 时间列自己再留一点右白，保证和方法徽章之间有明确间隔 */
.stream__time {
  padding-right: 6px;
  white-space: nowrap;
  color: var(--c-faint);
}

.stream__method {
  padding: 1px 0;
  font-size: 10px;
  font-weight: 700;
  text-align: center;
  border-radius: 3px;
  color: var(--c-sub);
  background: #eef2f7;
}

.m-get {
  color: #0a7f5a;
  background: rgba(16, 185, 129, .13);
}

.m-post {
  color: #b45309;
  background: rgba(245, 158, 11, .15);
}

.m-put {
  color: #0052d9;
  background: rgba(0, 82, 217, .1);
}

.m-delete {
  color: #dc2626;
  background: rgba(239, 68, 68, .12);
}

.stream__uri {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--c-text);
}

.stream__status {
  font-weight: 700;
  text-align: center;
}

.s-2xx { color: #0a7f5a; }
.s-3xx { color: #0891b2; }
.s-4xx { color: #b45309; }

/* 5xx 是最该被看见的，白底上用浅红底衬托 */
.s-5xx {
  color: #dc2626;
  background: rgba(239, 68, 68, .1);
  border-radius: 3px;
}

.stream__dur {
  text-align: right;
  color: var(--c-sub);
}

/* 单条请求的流量（请求+响应含头），之前这条数据拿到了却没往表里放 */
.stream__size {
  text-align: right;
  color: var(--c-sub);
  font-variant-numeric: tabular-nums;
}

.stream__ip {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--c-faint);
}

.empty {
  flex: 1;
  display: grid;
  place-items: center;
  font-size: 12px;
  letter-spacing: 2px;
  color: var(--c-faint);
}

.stream::-webkit-scrollbar,
.rank::-webkit-scrollbar,
.profile::-webkit-scrollbar {
  width: 5px;
}

.stream::-webkit-scrollbar-thumb,
.rank::-webkit-scrollbar-thumb,
.profile::-webkit-scrollbar-thumb {
  background: #d8e0ea;
  border-radius: 3px;
}

/* ---------- 自适应 ---------- */
/* KPI 拆行会把 board 的高度挤掉，所以断点压到 900px 以下才拆行 */
@media (max-width: 900px) {
  .kpi {
    grid-template-columns: repeat(3, 1fr);
  }
}

/* 12 列要排得开得 1500px 起步：再窄下去，「TOP 来源 IP」这类只占 2/12 的面板
   连一条 IPv4 + 归属地都放不下（1500px 时 2/12 ≈ 195px，刚好够）。
   1500px 以下换成 2 列自适应高度，页面纵向滚动 */
@media (max-width: 1500px) {
  .board {
    grid-template-columns: repeat(2, 1fr);
    grid-template-rows: none;
    grid-auto-rows: minmax(220px, auto);
  }

  /* 只有图和数据流需要整行；环形图/画像/三个 TOP 榜在半宽里排得下，
     全部拉成整行会让页面长出一倍 */
  .panel--chart,
  .panel--stream {
    grid-column: span 2;
  }

  .panel--ring,
  .panel--profile,
  .panel--top {
    grid-column: span 1;
  }

  /* 半宽的画像面板放不下 3 列（每列只剩 ~110px），退回 2 列 */
  .profile {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .dash {
    height: auto;
  }

  /* 整行的图表没有兄弟面板撑着高度，会被 flex 撑得很高，给个上限 */
  .panel--chart {
    max-height: 340px;
  }

  /* 顶栏一行塞不下「标题 + 连接状态 + 筛选 + 时钟 + 全屏」。
     不加 wrap 时 flex 会去压缩各块宽度，标题和状态就被挤成竖排了 */
  .dash__head {
    flex-wrap: wrap;
    gap: 10px 14px;
  }

  .brand,
  .dash__state {
    flex: none;
    white-space: nowrap;
  }

  .dash__tools {
    min-width: 0;
    margin-left: 0;
    flex-wrap: wrap;
  }

  .dash__filters button,
  .dash__btn,
  .dash__clock {
    white-space: nowrap;
  }
}

/* 手机上 2 列会把面板压到 160px 左右：状态码面板是「环 + 图例」横排，
   这个宽度装不下，会被面板的 overflow: hidden 裁掉。这里直接单列铺满 */
@media (max-width: 768px) {
  .kpi {
    grid-template-columns: repeat(2, 1fr);
  }

  .board {
    grid-template-columns: 1fr;
  }

  .panel--chart,
  .panel--ring,
  .panel--profile,
  .panel--stream,
  .panel--top {
    grid-column: span 1;
  }

  .kpi__value {
    font-size: 21px;
  }

  /* 主行在窄屏只留「发生了什么」：方法 + 地址 + 状态 + 耗时 */
  .stream__row {
    grid-template-columns: 40px minmax(0, 1fr) 30px 40px;
    gap: 6px;
  }

  .stream__time,
  .stream__size {
    display: none;
  }

  /* 副行留归属地 + 运营商；浏览器/系统/设备太宽，窄屏交给「访客画像」面板看 */
  .stream__ua {
    display: none;
  }

  /* 手机上画像面板只有一列宽，多列会把「Windows 10/11」压成省略号 */
  .profile {
    grid-template-columns: 1fr;
  }
}
</style>
