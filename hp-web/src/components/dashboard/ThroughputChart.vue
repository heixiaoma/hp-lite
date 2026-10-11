<script setup>
import {computed, ref} from 'vue'
import {fmtBps, fmtClock, fmtNum, useTween} from './format'

const props = defineProps({
  // 每秒一个采样点：q=请求数，b=流量字节，t=时间戳(ms)。
  // 三个字段必须落在同一条记录里，否则悬停取值会错位
  series: {type: Array, default: () => []},
})

// 趋势图窗口 60 个点，对应最近 60 秒
const SERIES_LEN = 60

const chartBoxRef = ref(null)
// 鼠标悬停命中的采样点下标，-1 表示光标不在图上
const hoverIdx = ref(-1)

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
  const d = props.series.slice(-SERIES_LEN)
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
</script>

<template>
  <!-- 根节点必须是单个 .panel：多根节点时父组件传的 class 无处可落 -->
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
</template>

<style scoped>
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
</style>
