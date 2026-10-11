/* 大屏共用的格式化与派生逻辑。
   抽出来的原因：面板拆成子组件后，fmtNum / rankSorted 这些被 5、6 个组件用到，
   各自复制一份的话，改个口径要翻好几个文件 */

import {ref, watch, onUnmounted} from 'vue'

/* ---------- 数字 / 字节 ---------- */

export const fmtNum = (n) => Math.round(Number(n) || 0).toLocaleString('en-US')

export const fmtBytes = (b) => {
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

export const fmtBps = (b) => `${fmtBytes(b)}/s`

/* 秒数同样要先取整（缓动中间值是小数），否则秒位会拖出一长串毫秒。
   另外天/时/分级都要保留秒，才有跑秒效果 */
export const fmtUptime = (sec) => {
  const s = Math.max(0, Math.floor(Number(sec) || 0))
  const p = (n) => String(n).padStart(2, '0')
  const d = Math.floor(s / 86400)
  const h = Math.floor((s % 86400) / 3600)
  const m = Math.floor((s % 3600) / 60)
  const ss = p(s % 60)
  return d ? `${d}天${h}:${p(m)}:${ss}` : `${p(h)}:${p(m)}:${ss}`
}

/* ---------- 时间 ---------- */

/* 只取 HH:MM:SS。原来用 slice(11)，服务端时间一旦带纳秒或时区后缀
   （如 20:00:18.123456789Z）就会撑爆时间列、顶到后面的方法徽章上 */
export const shortTime = (t) => {
  const m = String(t || '').match(/(\d{2}):(\d{2}):(\d{2})/)
  return m ? m[0] : '--:--:--'
}

/* 时间戳(ms) -> HH:MM:SS */
export const fmtClock = (ms) => {
  const d = new Date(Number(ms) || 0)
  const p = (n) => String(n).padStart(2, '0')
  return `${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

/* 相对时间：大屏上「3 秒前」比绝对时间戳更能反映新鲜度 */
export const fmtAgo = (ms, now) => {
  const t = Number(ms) || 0
  if (!t) return '从未命中'
  const s = Math.max(0, Math.floor(((now || Date.now()) - t) / 1000))
  if (s < 5) return '刚刚'
  if (s < 60) return `${s} 秒前`
  if (s < 3600) return `${Math.floor(s / 60)} 分钟前`
  if (s < 86400) return `${Math.floor(s / 3600)} 小时前`
  return `${Math.floor(s / 86400)} 天前`
}

/* ---------- 来源画像 ----------
   stat 明细里带了归属地（国家/省/市/运营商）和客户端（浏览器/系统/设备），
   下面把它们还原出来 */

// 「四川省」这种全称在窄列里太占地方，去掉行政后缀只留「四川」
const ADMIN_SUFFIX = /(省|市|自治区|特别行政区)$/
export const shortArea = (s) => String(s || '').trim().replace(ADMIN_SUFFIX, '')

/* 两套字段名：stat 明细是 sourceXxx，summary 的 topIps 是 xxx。
   统一成一套，后面拼文案就不用管数据来自哪 */
export const asGeo = (o) => ({
  country: String((o && (o.sourceCountry || o.country)) || '').trim(),
  code: String((o && (o.sourceCountryCode || o.countryCode)) || '').trim(),
  province: String((o && (o.sourceProvince || o.province)) || '').trim(),
  city: String((o && (o.sourceCity || o.city)) || '').trim(),
  isp: String((o && (o.sourceIsp || o.isp)) || '').trim(),
})

/* 归属地：国内取「省·市」，境外一般没有省市，退回国家名 */
export const geoText = (r) => {
  if (!r) return '未知'
  const g = asGeo(r)
  const prov = shortArea(g.province)
  const city = shortArea(g.city)
  if (prov && city) return prov === city ? city : `${prov}·${city}`
  if (prov || city) return prov || city
  return g.country || '未知'
}

/* 国家 + 归属地 + 运营商，境外请求省市为空时 geoText 会退回国家名，别拼两遍 */
export const geoFullText = (r) => {
  const g = asGeo(r)
  const area = geoText(r)
  const parts = []
  if (g.country) parts.push(g.country)
  if (area && area !== g.country && area !== '未知') parts.push(area)
  if (g.isp) parts.push(g.isp)
  return parts.join(' · ')
}

export const clientText = (r) =>
    [r.browser, r.os, r.device].filter(Boolean).join(' · ') || '-'

/* 悬停补全：一行塞不下完整 UA，放到 title 里 */
export const metaTitle = (r) => [
  `IP ${r.sourceIp || '-'}`,
  [r.sourceCountry, r.sourceProvince, r.sourceCity].filter(Boolean).join(' '),
  r.sourceIsp,
  [r.scheme, r.protocol].filter(Boolean).join(' '),
  r.browser,
  r.os,
  r.device,
  r.userAgent,
].filter(Boolean).join(' · ')

/* ---------- 榜单 ---------- */

/* 进度条宽度：以榜内最大值为 1，最小给 2% 才看得见条 */
export const pct = (count, list) => {
  const max = (list || []).reduce((a, b) => Math.max(a, Number(b.count) || 0), 0)
  if (!max) return '0%'
  return `${Math.max(2, Math.round((Number(count) || 0) / max * 100))}%`
}

/* 榜单排序：先按次数降序，次数并列时再按名字升序。
   次数必须是第一维度 —— 榜的意义就是谁多谁在前。
   名字只做第二维度：并列项（常见的是末尾一堆 count=1）如果只按次数排，
   它们的先后全看服务端返回顺序，每次 summary 都可能不一样，看着就在乱跳；
   用名字兜底把这批钉死。
   numeric 让 IP 按段数值排（192.168.1.2 在 .10 前面），sensitivity: base 忽略大小写 */
const NAME_OPT = {numeric: true, sensitivity: 'base'}
export const rankSorted = (list) => (list || []).slice().sort((a, b) => {
  const diff = (Number(b && b.count) || 0) - (Number(a && a.count) || 0)
  return diff || String(a && a.key).localeCompare(String(b && b.key), 'zh-Hans-CN', NAME_OPT)
})

/* 服务端 TOP 榜 -> 面板用的 {key, count, pct}。
   榜单都是近 5 小时窗口，不是抽样，占比的分母就是榜内 count 之和 */
export const rankPct = (list, n) => {
  const src = list || []
  const total = src.reduce((a, b) => a + (Number(b.count) || 0), 0)
  return src.slice(0, n).map((it) => ({
    key: it.key,
    count: Number(it.count) || 0,
    pct: total ? `${Math.round(((Number(it.count) || 0) / total) * 100)}%` : '0%',
  }))
}

/* ---------- 数字滚动 ----------
   大屏的标志性效果，从旧值缓动到新值 */
export const useTween = (getter) => {
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

/* 每秒归零一次的毫秒时钟，给「最近命中 N 秒前」这类相对时间用 */
export const useNow = (interval = 1000) => {
  const now = ref(Date.now())
  let timer = setInterval(() => {
    now.value = Date.now()
  }, interval)
  onUnmounted(() => clearInterval(timer))
  return now
}
