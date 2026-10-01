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

      <div class="panel">
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
        <ul v-if="stream.length" class="stream">
          <li v-for="r in stream" :key="r.id">
            <span class="stream__time">{{ shortTime(r.time) }}</span>
            <span class="stream__method" :class="`m-${String(r.method || '').toLowerCase()}`">{{ r.method }}</span>
            <span class="stream__uri" :title="r.uri">{{ r.domain }}{{ r.path }}</span>
            <span class="stream__status" :class="`s-${r.statusClass}`">{{ r.statusCode }}</span>
            <span class="stream__dur">{{ r.durationMs }}ms</span>
            <span class="stream__size" :title="`${fmtNum(r.totalSize)} B`">{{ fmtBytes(r.totalSize) }}</span>
            <span class="stream__ip" :title="r.sourceIp">{{ r.sourceIp }}</span>
          </li>
        </ul>
        <div v-else class="empty">等待请求流入…</div>
      </div>

      <div class="panel">
        <h2>TOP 来源 IP</h2>
        <ul class="rank">
          <li v-for="(it, i) in topIps" :key="it.key">
            <span class="rank__no">{{ i + 1 }}</span>
            <span class="rank__name" :title="it.key">{{ it.key }}</span>
            <span class="rank__num">{{ fmtNum(it.count) }}</span>
            <span class="rank__bar">
              <i :style="{width: pct(it.count, topIps)}"/>
            </span>
          </li>
          <li v-if="!topIps.length" class="rank__empty">暂无数据</li>
        </ul>
      </div>

      <div class="panel">
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

// 趋势图窗口 60 个点，对应最近 60 秒；请求流只渲染最近 60 条
const SERIES_LEN = 60
const STREAM_ROWS = 60

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
const topIps = computed(() => summary.value.topIps || [])
const topPaths = computed(() => summary.value.topPaths || [])

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
  // 新的在前；只保留最近 STREAM_ROWS 条，超出的直接丢
  stream.value = batch.slice().reverse().concat(stream.value).slice(0, STREAM_ROWS)
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
.board {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
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

.panel--stream {
  grid-column: span 2;
}

.panel--chart {
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
.ring-row {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 0;
}

.ring-wrap {
  position: relative;
  flex: none;
}

.ring {
  display: block;
  width: 96px;
  height: 96px;
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

.ring__center b {
  display: block;
  font-size: 17px;
  color: var(--c-text);
}

.ring__center span {
  font-size: 10px;
  color: var(--c-faint);
}

.legend {
  flex: 1;
  min-width: 0;
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 4px;
  font-size: 10.5px;
  line-height: 1.3;
}

.legend li {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--c-sub);
}

.legend i {
  flex: none;
  width: 7px;
  height: 7px;
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
  padding-top: 8px;
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
  font-size: 9.5px;
  white-space: nowrap;
  color: var(--c-faint);
}

.ring-stat b {
  font-size: 13px;
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

/* 七列：时间 / 方法 / 地址 / 状态 / 耗时 / 流量 / 来源 IP。
   行高压到 ~20px（原来是 25px），同样高度能多露一行多的内容 */
.stream li {
  display: grid;
  grid-template-columns: 62px 40px minmax(0, 1fr) 30px 40px 48px 88px;
  align-items: center;
  gap: 7px;
  padding: 2px 0;
  line-height: 1.35;
  border-bottom: 1px solid #f1f5f9;
  animation: slidein .32s ease;
}

@keyframes slidein {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: none; }
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
.rank::-webkit-scrollbar {
  width: 5px;
}

.stream::-webkit-scrollbar-thumb,
.rank::-webkit-scrollbar-thumb {
  background: #d8e0ea;
  border-radius: 3px;
}

/* ---------- 自适应 ---------- */
/* 断点压到 900px 才拆行：KPI 拆成两行会把 board 的高度挤掉，
   环形图的图例随即溢出（1500px 的断点太保守，一般笔记本宽度就会触发） */
@media (max-width: 900px) {
  .kpi {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 1200px) {
  .board {
    grid-template-columns: repeat(2, 1fr);
    grid-template-rows: none;
    grid-auto-rows: minmax(220px, auto);
  }

  .panel--chart,
  .panel--stream {
    grid-column: span 2;
  }

  .dash {
    height: auto;
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
  .panel--stream {
    grid-column: span 1;
  }

  .kpi__value {
    font-size: 21px;
  }

  /* 请求流七列在窄屏装不下：隐掉时间 / 流量 / 来源 IP，
     留下最能说明「发生了什么」的方法 + 地址 + 状态 + 耗时 */
  .stream li {
    grid-template-columns: 40px minmax(0, 1fr) 30px 40px;
    gap: 6px;
  }

  .stream__time,
  .stream__size,
  .stream__ip {
    display: none;
  }
}
</style>
