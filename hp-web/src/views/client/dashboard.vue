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

    <DashHeader
        :status="status"
        :status-text="statusText"
        :clock="clock"
        :is-full="isFull"
        @toggle-fullscreen="toggleFullscreen"
    />

    <KpiCards :items="kpis"/>

    <!-- 地图不进 board 栅格：那里每行是写死的 380px，地图会被压成一条。
         放在这里它才是「宽度铺满、高度跟着投影比例走」 -->
    <WorldMap :batch="mapBatch" :server="SERVER_NAME"/>

    <section class="board">
      <ThroughputChart :series="series"/>
      <StatusRing :counts="summary.statusCounts"/>
      <VisitorProfile :groups="profileGroups"/>

      <RankList title="TOP 来源国家" note="近 5 小时" :items="topCountries"/>
      <RequestStream :rows="stream" :dense="dense"/>
      <RankList title="TOP 域名" note="近 5 小时" :items="topDomains"/>
      <RankList title="TOP 来源 IP" note="近 5 小时 · 带归属地" :items="topIps"/>
      <RankList title="TOP 路径" note="近 5 小时" :items="topPaths"/>

      <WafPanel :waf="summary.wafBlocks"/>
    </section>
  </div>
</template>

<script setup>
import {computed, onMounted, onUnmounted, ref} from 'vue'
import {createStatsSocket} from '../../data/statsSocket'
import userInfo from '../../data/userInfo'
import DashHeader from '../../components/dashboard/DashHeader.vue'
import KpiCards from '../../components/dashboard/KpiCards.vue'
import WorldMap from '../../components/dashboard/WorldMap.vue'
import ThroughputChart from '../../components/dashboard/ThroughputChart.vue'
import StatusRing from '../../components/dashboard/StatusRing.vue'
import VisitorProfile from '../../components/dashboard/VisitorProfile.vue'
import RankList from '../../components/dashboard/RankList.vue'
import RequestStream from '../../components/dashboard/RequestStream.vue'
import WafPanel from '../../components/dashboard/WafPanel.vue'
import {fmtBytes, fmtNum, fmtUptime, geoFullText, rankPct, rankSorted, useTween} from './../../components/dashboard/format'
// 面板栅格、面板外壳、通用榜单样式：必须全局注入，原因见该文件头部注释
import './../../components/dashboard/dashboard.css'

/* 请求流保留最近 2000 条。行数是 DOM 规模的大头：每条 2 行约 12 个节点，
   2000 条约 2.4w 个节点，靠 .stream li 上的 content-visibility 把屏幕外的行
   跳过渲染，滚动和更新才不卡。要继续加量得先上虚拟滚动 */
const STREAM_ROWS = 2000
// 趋势图只取最近 60 个点（ThroughputChart 内部按 60s 窗口裁剪），
// 多留 1 秒缓冲，避免刚好在边界丢掉最后一个采样点
const SERIES_LEN = 61

/* 菜单里已经按角色隐藏了入口，但直接敲 URL 仍会进到这里。
   大屏是全站流量数据，这里再拦一道：非管理员不建立连接 */
const isAdmin = computed(() => userInfo.getUserInfo()?.role === 'ADMIN')

const rootRef = ref(null)
const summary = ref({})
const stream = ref([])
/* 每条新记录都跑一次入场动画，高 QPS 下会同时挂着几百个动画，帧率直接掉一半。
   批量太大时整批不做动画 —— 那个速度下本来也看不出单条滑入 */
const dense = ref(false)
// 趋势图采样点：每秒一个 {q, b, t}
const series = ref([])
/* 地图的输入：每批新到的记录（新→旧）。地图自己按经纬度取舍，
   这里只负责把批次递过去，不做任何加工 */
const mapBatch = ref([])

/* 地图右下角那个源站节点的名字。位置由地图组件自己定，
   不传经纬度 —— 源站就是画布上的图形，没有真实坐标 */
const SERVER_NAME = '源站节点'

const status = ref('connecting')
const statusDetail = ref('')
const clock = ref('')
const isFull = ref(false)

let socket = null
// 本地每秒收到的 stat 条数。summary.qps 是服务端 60s 平均，变化太钝，
// 这里自己按秒计数拿瞬时值（文档 §5.4 正是这个建议）
let perSec = 0
// 每秒流量字节数：累加每条 stat 的 totalSize，同样按秒归零
let bytesPerSec = 0

/* ---------- 顶层派生数据 ---------- */
const statusText = computed(() => ({
  connecting: '连接中…',
  open: '实时连接已建立',
  closed: '连接断开，重连中…',
  error: statusDetail.value || '连接异常',
}[status.value] || '初始化'))

const tQps = useTween(() => summary.value.qps || 0)
const tTotal = useTween(() => summary.value.totalRequests || 0)
const tBytes = useTween(() => summary.value.totalBytes || 0)
const tDur = useTween(() => summary.value.avgDurationMs || 0)
const tUptime = useTween(() => summary.value.uptimeSec || 0)

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
    sub: '自服务启动',
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

const topDomains = computed(() => rankSorted(summary.value.topDomains))
const topPaths = computed(() => rankSorted(summary.value.topPaths))
/* TOP 来源 IP 的归属地直接读服务端 topIps 里带的 country/province/city/isp，
   不用自己反查，免得要等对应请求进来才有数据 */
const topIps = computed(() => rankSorted(summary.value.topIps)
    .map((it) => ({...it, sub: geoFullText(it) || '归属地未知'})))
/* 服务端 topCountries 是近 5 小时窗口的国家榜，直接整榜用（面板里放不下就滚动），
   不截断成 TOP3 —— 截断后大半国家根本看不见，等于没用这份数据 */
const topCountries = computed(() => rankSorted(rankPct(summary.value.topCountries, 20)))

/* 客户端画像：设备/系统/浏览器/运营商改由服务端汇总，和其余 TOP 榜同口径
   （近 5 小时窗口），刷新页面、换设备看都是同一份数据。
   占比分母是服务端返回的榜单 count 之和；服务端只给 Top10，长尾不计入分母，
   这四个维度基数本来就小（常见取值十来个），误差可以忽略 */
const profileGroups = computed(() => [
  {label: '设备', items: rankSorted(rankPct(summary.value.topDevices, 3))},
  {label: '系统', items: rankSorted(rankPct(summary.value.topSystems, 3))},
  {label: '浏览器', items: rankSorted(rankPct(summary.value.topBrowsers, 3))},
  {label: '运营商', items: rankSorted(rankPct(summary.value.topIsps, 3))},
])

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
  /* 滚动窗口：新的插在最前面，只从尾巴上切掉最老的，窗口里始终留着 STREAM_ROWS 条。
     别写成「满了就清空再灌」—— 那样一到上限整屏会闪一下变空，滚动位置也跟着跳回顶部。
     只有单批就超过窗口（200ms 内塞满 2000 条 ≈ 1w QPS）时才会整批换掉，
     那种速度下窗口里剩下的本来也全是这一批的新数据 */
  const merged = fresh.concat(stream.value)
  stream.value = merged.length > STREAM_ROWS ? merged.slice(0, STREAM_ROWS) : merged
  // 换一个新数组引用，地图那边 watch 到就处理这一批
  mapBatch.value = fresh
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
  /* 白底主题：整套配色收敛在这几个变量里，改这里就能整体换肤。
     子组件统一读这几个变量，面板样式才能在 dashboard.css 里只写一份 */
  --c-text: #1f2d3d;
  --c-sub: #64748b;
  --c-faint: #94a3b8;
  --c-border: #e3e8ef;
  --c-card: #fff;
  --c-brand: #0052d9;

  /* board 每行的高度。加面板只会让行数变多、页面往下滚，面板自身不会被压扁。
     想整体调矮/调高改这一个值即可，不用去动每个面板。
     380px ≈ 改版前 1080p 下一行面板的实际高度 */
  --dash-row: 380px;

  position: relative;
  /* 抵消 .console__content 的 22px padding，让大屏铺满主内容区 */
  margin: -22px;
  /* 不限死高度：内容少时铺满一屏，面板多了靠 .console__main 纵向滚动 */
  min-height: calc(100vh - 96px);
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

/* ---------- 无权限遮罩 ---------- */
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
</style>
