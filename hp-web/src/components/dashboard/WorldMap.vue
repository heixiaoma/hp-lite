<script setup>
import {computed, onMounted, onUnmounted, ref, shallowRef, watch} from 'vue'
import {geoEquirectangular, geoGraticule10, geoPath} from 'd3-geo'
import {zoom, zoomIdentity} from 'd3-zoom'
import {select} from 'd3-selection'
// 副作用导入：给 selection 挂上 .transition()，按钮才能平滑缩放
import 'd3-transition'
import {feature} from 'topojson-client'
import countries110m from 'world-atlas/countries-110m.json'
import {asGeo, fmtNum, geoFullText, geoText, shortArea} from './format'

const props = defineProps({
  /* 一个节流批次（200ms）里新到的记录，按新→旧排列。
     只有带经纬度的记录会被画成链路，整批进来再裁剪，不逐个 ping 父组件 */
  batch: {type: Array, default: () => []},
  /* 源站节点的名字。位置不在这里配 —— 源站没有经纬度，
     它就是画布右侧那个固定的服务器图形（见 SRV_X / SRV_Y），
     所有链路直接往那个点上打 */
  server: {type: String, default: '源站节点'},
})

/* SVG 用固定坐标系而不是容器像素：d3 的投影只在加载时算一次地形路径，
   容器尺寸每变一次就得重算上百个国家的路径。固定 viewBox 交给浏览器缩放，
   后面每秒几十条的动态链路就只跟投影后的坐标打交道。
   面板不给死高度：svg 走 width:100% + height:auto，高度由 viewBox 比例自己算出来 */

/* 画布左边 MAP_W 是地图本体，右边 SRV_W 是留给虚拟服务器的一竖条，
   链路直接从来源点画到那台服务器的机箱左侧（见 buildArc）。
   这样做成一个 svg 而不是「svg + 右侧 html 卡片」：坐标只有一套，
   缩放时线和服务器永远不会错开 */

/* 投影用等距圆柱（Equirectangular）：经纬网是正交直线、球框是个正矩形，没有
   自然地球 / 罗宾森那种两侧外扩的圆弧边。整幅 2:1，正好吃掉宽屏那条行。
   想要墨卡托就把 geoEquirectangular 换成 geoMercator —— 但它接近 1:1，
   横向会缩得很小，左右留出大片空白 */
const MAP_W = 1000
const SRV_W = 190
const VIEW_W = MAP_W + SRV_W
const VIEW_H = 470

// 虚拟服务器图形的中心（地图右侧那一竖条里）
const SRV_X = MAP_W + 96
const SRV_Y = 132

// 同时挂着的链路数。超了砍最老的，高 QPS 下不会糊成一团白网
const MAX_FLIGHTS = 60
// 单个批次最多起几条链路：200ms 一批，一次十几条看不出先后也不好看
const BATCH_TAKE = 6
// 一条链路在 DOM 里留多久。必须 ≥ 最长的一发（3.4s）+ 落地后的收尾，
// 否则会出现「还没飞到目的地就被 prune 掉」的半截效果
const FLIGHT_TTL = 4200
/* 累计来源地的上限：一个点约等于一个 ~11km 网格，1w 个已经足够密。
   满了淘汰最早出现的那批（新地区还能上地图），但下面的统计不跟着回落 */
const MAX_DOTS = 10000
/* 模板渲染的那一批点按这个间隔整体换一次引用。
   1w 个点每次命中都重渲，光 diff 就是 3w 个 vnode，画面会一卡一卡的；
   改成计数照常累加、画面按秒刷新 —— 点的大小迟到 1 秒看不出来 */
const DOT_FLUSH_MS = 1000
/* 点还少的时候不走节流：刚开页面时「新地区立刻冒出来」比省那次 diff 重要，
   过了这个量级才交给定时器按秒批量刷 */
const DOT_LIVE = 300
/* 放大后最多给最近的这几千个点贴地名。
   1w 个全贴上既糊成一片，又多出 1w 个带描边的 text，拖动时会明显掉帧 */
const DOT_LABEL_MAX = 2000
/* 只有最新的这几条链路带飞行光点和线上信息。
   几十条线同时带字会互相压住，最后什么都读不出来，只给最近几条挂标签 */
const LABEL_TAKE = 8

const land = feature(countries110m, countries110m.objects.countries)
/* 只投影进左边那块：右边要留给服务器，别被地图占掉。
   fitExtent 要给 Sphere 而不是 land —— 陆地最南只到南极海岸线，
   按 land 对齐的话 ±90° 的那圈海面会顶出画布，矩形边框被上下切掉两道 */
const projection = geoEquirectangular()
    .fitExtent([[6, 6], [MAP_W - 6, VIEW_H - 6]], {type: 'Sphere'})
const pathGen = geoPath(projection)

const landPath = pathGen(land)
const spherePath = pathGen({type: 'Sphere'})
const gridPath = pathGen(geoGraticule10())

/* ---------- 链路 ---------- */

/* 命中点：链路一律收在服务器机箱的左边缘这条竖线上 */
const IMPACT_X = SRV_X - 34
const IMPACT_Y = SRV_Y

const flights = ref([])
/* 累计来源地：按 ~11km 网格（经纬度各留 1 位小数）聚合计数。
   dotList 刻意不进响应式 —— 每秒几十条记录，进去了就是每次命中都触发
   1w 个点整体 diff。模板读的是 dots（dotList 的快照），由 flushDots 按秒换引用 */
const dotList = []
// key → dot：命中走 Map，别在 1w 条里 find
const dotIndex = new Map()
const dots = shallowRef([])
let dotDirty = false
const locatedTotal = ref(0)  // 累计拿到经纬度的请求
const geoMissTotal = ref(0)  // 累计没有经纬度的请求（内网 IP、未收录网段）
const errTotal = ref(0)      // 累计 4xx / 5xx
/* 覆盖了多少个国家 / 地区、出现过多少个来源网格。
   Set 不做响应式（数量单独用 ref 跟踪），免得每次命中都触发整个面板重算。
   两者都只增不减：地图上的点会被 MAX_DOTS 淘汰，统计不能跟着回落 */
const countryKeys = new Set()
const countryCount = ref(0)
const regionKeys = new Set()
const regionTotal = ref(0)
let seq = 0

/* 一条链路的路径：从来源点打到命中点，全程在画布坐标里算。
   之前是「大圆航线先落到地图上的机房坐标，再从那里拉直线进机箱」，
   一条路中间硬拐一个角，飞着飞着突然折向，很难看，整个换掉：
   源站本来就没有经纬度，它就是画布右侧那个固定图形，没必要掺球面坐标进来，
   直接在屏幕空间给一条三次贝塞尔，从头到尾一次成形 */
const buildArc = (x, y) => {
  const dx = IMPACT_X - x
  const dy = IMPACT_Y - y
  const len = Math.hypot(dx, dy) || 1

  /* 拱高：距离越远抛得越高，但压上限，跨洲的线不会飞出画布顶部；
     再用「起点到画布顶的距离」兜一次底，免得高纬度来源的弹道被 viewBox 裁掉 */
  let lift = Math.min(78, 16 + len * 0.14)
  lift = Math.max(0, Math.min(lift, Math.min(y, IMPACT_Y) - 10))

  /* 连线中垂线的法向量，统一取朝上的那一侧：
     不管从南半球还是北半球打出来，弹道都往上鼓，画面才有统一的抛物线观感 */
  let nx = -dy / len
  let ny = dx / len
  if (ny > 0) {
    nx = -nx
    ny = -ny
  }

  /* 起飞段控制点：先沿连线走 42%，再按法向抬起来 */
  const c1x = x + dx * 0.42 + nx * lift
  const c1y = y + dy * 0.42 + ny * lift
  /* 命中段控制点：锁在命中点正左方、同一条水平线上，
     于是末端切线是水平的 —— 所有弹体都齐着水平方向打进机箱左侧，落点也重合 */
  const inset = Math.min(70, Math.abs(dx) * 0.45)
  const c2x = IMPACT_X - inset

  /* 飞行时长跟着路程走：统一给 2.8s 的话，
     近距离的一发慢吞吞挪半天，跨洲的一发又一闪而过 */
  const fly = Math.round(Math.min(3400, Math.max(1500, 900 + len * 2.4)))

  const n = (v) => v.toFixed(1)
  return {d: `M${n(x)} ${n(y)} C${n(c1x)} ${n(c1y)} ${n(c2x)} ${IMPACT_Y} ${IMPACT_X} ${IMPACT_Y}`, fly}
}

/* 状态码分色：正常蓝、4xx 橙、5xx 红。
   地图上异常请求能一眼挑出来，比单纯看线密度有用 */
const classOf = (r) => {
  const c = String((r && r.statusClass) || '')
  if (c === '5xx') return 'is-err'
  if (c === '4xx') return 'is-warn'
  return ''
}

/* 挂在流动光点旁边的那行字：谁、什么结果、多快 */
/* 线上标签：只看「哪来的」，所以给国家 + 城市，不显示 IP ——
   IP 在这不够读（一闪就过），塞进去只是把地名挤掉。IP 留给悬停 tooltip */
const placeText = (r) => {
  const g = asGeo(r)
  const city = shortArea(g.city)
  // 境外请求常常没有 city，退回省 / 国家，连国家都没有才写未知
  if (!g.country && !city) return '未知来源'
  return [g.country, city || shortArea(g.province)].filter(Boolean).join(' · ')
}

const labelOf = (r) => [
  `${(r && r.sourceIp) || 'IP 未知'}`,
  `${placeText(r)}${(r && r.statusCode) ? ` · ${r.statusCode}` : ''}`,
  geoFullText(r) || '归属地未知',
  `${(r && r.method) || ''} ${(r && r.domain) || ''}${(r && r.path) || ''}`.trim(),
].filter(Boolean).join('\n')

const dotRadius = (count) => 2.2 + Math.min(2.6, Math.log10(Number(count) + 1))

/* 把 dotList 的当前状态拍一份给模板。
   计数是每次命中都加的，但半径和标签只在这里算一次 ——
   1w 个点每次都重渲，光 diff 就是几万个 vnode，画面会一卡一卡的 */
const flushDots = () => {
  if (!dotDirty) return
  dotDirty = false
  for (let i = 0; i < dotList.length; i++) {
    const d = dotList[i]
    d.r = dotRadius(d.count)
    // 只给最近的这批点留标签，见 DOT_LABEL_MAX
    d.lab = i >= dotList.length - DOT_LABEL_MAX
  }
  // 换一个新数组，shallowRef 才会通知模板重渲
  dots.value = dotList.slice()
}
const dotTimer = setInterval(flushDots, DOT_FLUSH_MS)

const bumpDot = (key, lat, lon, r) => {
  const hit = dotIndex.get(key)
  if (hit) {
    hit.count++
    dotDirty = true
    return
  }
  const p = projection([lon, lat])
  if (!p) return
  const dot = {
    key,
    x: p[0],
    y: p[1],
    count: 1,
    r: dotRadius(1),
    lab: true,
    label: geoFullText(r) || '归属地未知',
    // 放大后贴在点旁边的短名：带运营商的长串在地图上放不下
    short: placeText(r),
  }
  dotList.push(dot)
  dotIndex.set(key, dot)
  // 满了砍最早那批，保证新出现的地区还能上地图
  if (dotList.length > MAX_DOTS) {
    for (const gone of dotList.splice(0, dotList.length - MAX_DOTS)) dotIndex.delete(gone.key)
  }
  dotDirty = true
  // 点还少的时候立刻刷新：这时候「新地区冒出来」比省那次 diff 重要
  if (dotList.length <= DOT_LIVE) flushDots()
}

const addBatch = (batch) => {
  if (!Array.isArray(batch) || !batch.length) return
  const now = Date.now()
  const cells = new Set()
  let taken = 0

  for (const r of batch) {
    const lat = Number(r && r.sourceLatitude)
    const lon = Number(r && r.sourceLongitude)
    // GeoLite2 对内网 IP / 未收录网段返回 0,0 —— 那是大西洋几内亚湾外海，别打上去
    if (!lat || !lon) {
      geoMissTotal.value++
      continue
    }
    locatedTotal.value++
    if (classOf(r)) errTotal.value++
    const country = String((r && r.sourceCountry) || '').trim()
    if (country && !countryKeys.has(country)) {
      countryKeys.add(country)
      countryCount.value = countryKeys.size
    }

    // 网格 key 在这里算一次：地图聚合、累计计数、本批去重三处共用
    const cell = `${lat.toFixed(1)}_${lon.toFixed(1)}`
    if (!regionKeys.has(cell)) {
      regionKeys.add(cell)
      regionTotal.value = regionKeys.size
    }
    bumpDot(cell, lat, lon, r)

    if (taken >= BATCH_TAKE || cells.has(cell)) continue
    const p = projection([lon, lat])
    if (!p) continue
    cells.add(cell)
    taken++
    const {d, fly} = buildArc(p[0], p[1])
    flights.value.push({
      id: ++seq,
      born: now,
      d,
      // 这一发的飞行时长，交给 CSS 里的 animation-duration
      fly,
      x: p[0],
      y: p[1],
      klass: classOf(r),
      ip: (r && r.sourceIp) || '-',
      area: geoText(r),
      text: placeText(r),
      label: labelOf(r),
    })
  }

  if (flights.value.length > MAX_FLIGHTS) {
    flights.value.splice(0, flights.value.length - MAX_FLIGHTS)
  }
}

watch(() => props.batch, addBatch)

/* 到点的链路统一收掉。放在一个定时器里批量清，
   而不是每条起一个 timeout —— 高 QPS 下那是几百个定时器 */
const prune = () => {
  const now = Date.now()
  const alive = flights.value.filter((f) => now - f.born < FLIGHT_TTL)
  if (alive.length !== flights.value.length) flights.value = alive
}
const pruneTimer = setInterval(prune, 600)
onUnmounted(() => {
  clearInterval(pruneTimer)
  clearInterval(dotTimer)
})

/* 最近几条链路才有资格挂线上信息；这里顺带标出来传给模板 */
const labelledFrom = computed(() => Math.max(0, flights.value.length - LABEL_TAKE))

const errRate = computed(() => {
  if (!locatedTotal.value) return '0.0'
  return (errTotal.value / locatedTotal.value * 100).toFixed(1)
})

/* 服务器图形下面那几行数字。
   五项全是累计值，只增不减：抵达请求 / 内网跳过 / 异常占比累加自每条 stat；
   覆盖国家、来源地区记的是「出现过多少个」，用 Set 去重 —— 地图上的点被
   MAX_DOTS 淘汰后不会让这两个数往回掉，所以这里读的是 regionTotal 而不是 dots.length */
const srvStats = computed(() => [
  {k: '抵达请求', v: fmtNum(locatedTotal.value)},
  {k: '来源地区', v: fmtNum(regionTotal.value)},
  {k: '覆盖国家', v: fmtNum(countryCount.value)},
  {k: '内网跳过', v: fmtNum(geoMissTotal.value)},
  {k: '异常占比', v: `${errRate.value}%`, warn: errRate.value !== '0.0'},
])

/* 面板角落的「最近来源」：地图上的线一闪就过，这里留个文字版 */
const feed = computed(() => flights.value.slice(-4).reverse()
    .map((f) => ({key: f.id, ip: f.ip, area: f.area, klass: f.klass})))

/* ---------- 缩放 / 平移 ----------
   手势交给 d3-zoom：滚轮缩放、拖拽平移、双击放大、触屏双指捏合，
   这些都是现成的，自己写一遍既啰嗦又容易漏掉边界情况。
   d3 只算变换（不碰 DOM），算出来存到 scene 里交给模板渲染 —— Vue 管渲染，
   d3 管数学，两边不打架 */

// 倍数上限：国界数据是 110m 精度，再放大看到的是矢量本身的锯齿而不是更多细节
const MAX_K = 8
const DOT_LABEL_K = 2.6   // 放大到这个倍率才开始给来源点贴地名

const svgRef = ref(null)
// 视图变换：x / y 是平移量，k 是倍数
const scene = ref(zoomIdentity)
const panning = ref(false)
const dotLabels = computed(() => scene.value.k >= DOT_LABEL_K)

let zoomBehavior = null

onMounted(() => {
  zoomBehavior = zoom()
      .scaleExtent([1, MAX_K])
      /* 不许把地图拖出画布留白。d3 对 SVG 取的是 viewBox 作为 extent，
         所以这里填的和画布是同一套坐标 */
      .translateExtent([[0, 0], [VIEW_W, VIEW_H]])
      .on('zoom', (e) => { scene.value = e.transform })
      /* 滚轮也会派发 start/end，但那时候鼠标没按住，改光标会很怪 */
      .on('start', (e) => { panning.value = e.sourceEvent && e.sourceEvent.type !== 'wheel' })
      .on('end', () => { panning.value = false })

  select(svgRef.value).call(zoomBehavior)
})

onUnmounted(() => {
  // 不清监听的话组件销毁后 d3 还在观望一个已经不存在的节点
  select(svgRef.value).on('.zoom', null)
})

/* 自己塞给 d3 的变换不会被 translateExtent 夹住，先按同样的规则夹一遍：
   可见区间是 [-x/k, (VIEW_W - x)/k]，要求它落在 [0, VIEW_W] 内 */
const clampPane = (v, k, size) => Math.min(0, Math.max(size * (1 - k), v))

/* 按钮 / 双击都走 d3 的过渡，直接跳会显得很突兀 */
const transition = (dur) => select(svgRef.value).transition().duration(dur)

const zoomStep = (mul) => transition(260).call(zoomBehavior.scaleBy, mul)
const resetView = () => transition(320).call(zoomBehavior.transform, zoomIdentity)

/* 点某个来源点直接聚焦过去 —— 标记密集的时候自己拖过去太费劲 */
const focusDot = (d) => {
  const k = 6
  const x = clampPane(VIEW_W / 2 - k * d.x, k, VIEW_W)
  const y = clampPane(VIEW_H / 2 - k * d.y, k, VIEW_H)
  transition(420).call(zoomBehavior.transform, zoomIdentity.translate(x, y).scale(k))
}

/* ---------- 链路横跨两层 ----------
   地图那层在缩放平移，服务器那层钉死不动，链路夹在中间：
   起点（来源城市）必须跟着地图走，终点必须永远落在服务器机箱上那个固定的入口点。
   所以这条 path 不能像地形那样丢进缩放组里了事，只能按当前视图每条重算。
   applyX / applyY 就是那个视图变换本身（k*x + tx）—— 内容坐标 → 当前视图坐标 */
const drawnFlights = computed(() => flights.value.map((f) => {
  const x = scene.value.applyX(f.x)
  const y = scene.value.applyY(f.y)
  return {...f, x, y, d: buildArc(x, y).d}
}))
</script>

<template>
  <div class="panel panel--map">
    <h2>
      全球实时访问态势<em>来源 → {{ server }}</em>
      <span class="map__lg">
        <b class="is-ok"><i></i>正常</b>
        <b class="is-warn"><i></i>4xx</b>
        <b class="is-err"><i></i>5xx</b>
      </span>
    </h2>

    <div class="map-body">
      <div class="map-box">
        <svg
            ref="svgRef"
            class="map"
            :class="{'is-panning': panning}"
            :viewBox="`0 0 ${VIEW_W} ${VIEW_H}`"
            preserveAspectRatio="xMidYMid meet"
        >
          <!-- 缩放只作用于地图本体：地形 + 来源点。
               服务器和链路都留在这层之外 —— 服务器是画面上的固定设施，不跟着地图动；
               链路则横跨两层，起点要跟着地图走、终点还得钉在服务器上，
               所以那几条 path 只能按当前视图重算（见 drawnFlights） -->
          <defs>
            <!-- 放大后地形会横向溢出到右边那一竖条，在这里切掉，别糊到服务器上 -->
            <clipPath id="mapClip">
              <rect x="0" y="0" :width="MAP_W" :height="VIEW_H"/>
            </clipPath>
          </defs>

          <!-- clip-path 必须挂在没有 transform 的那层：写在同一个 <g> 上，
               裁剪框会跟着变换一起缩放平移，等于没裁 -->
          <g clip-path="url(#mapClip)">
            <g :transform="`translate(${scene.x} ${scene.y}) scale(${scene.k})`">
              <path class="map__sphere" :d="spherePath"/>
              <path class="map__grid" :d="gridPath" vector-effect="non-scaling-stroke"/>
              <path class="map__land" :d="landPath" vector-effect="non-scaling-stroke"/>

              <!-- 累计来源地：线会淡出，点留着，能看出这台机器长期被哪些地区访问。
                   每个点自己套一层 1/k 的反向缩放 ——
                   放大时位置和距离跟着放大，但点本身在屏幕上大小不变，
                   标记才会一直是「一颗点」而不是摊成一个饼 -->
              <g class="map__dots">
                <g
                    v-for="d in dots"
                    :key="d.key"
                    :transform="`translate(${d.x} ${d.y}) scale(${(1 / scene.k).toFixed(3)})`"
                    @click="focusDot(d)"
                >
                  <circle class="map__dot" :r="d.r">
                    <title>{{ d.label }} · {{ fmtNum(d.count) }} 次（点击定位）</title>
                  </circle>
                  <text v-if="dotLabels && d.lab" class="map__dotlabel" x="7" y="-5">{{ d.short }}</text>
                </g>
              </g>
            </g>
          </g>

          <!-- 虚拟服务器：画在地图右侧那一竖条里，不参与缩放 ——
               它是画面上固定的一个设施，不是地图上的一个地理位置。
               所有链路仍然接进它的机箱左侧（见 drawnFlights） -->
          <g class="srv">
            <rect class="srv__wave" :x="SRV_X - 62" :y="SRV_Y - 74" width="124" height="148" rx="14"/>
            <!-- 命中点：所有链路都打在这，一圈持续的撞击波纹 -->
            <circle class="srv__impact" :cx="SRV_X - 34" :cy="SRV_Y" r="5"/>
            <rect class="srv__body" :x="SRV_X - 28" :y="SRV_Y - 38" width="56" height="76" rx="7"/>

            <!-- 机架槽位 + 指示灯：绿色闪一下，看着像在收包 -->
            <rect class="srv__slot" :x="SRV_X - 20" :y="SRV_Y - 28" width="40" height="13" rx="2.5"/>
            <rect class="srv__slot" :x="SRV_X - 20" :y="SRV_Y - 9" width="40" height="13" rx="2.5"/>
            <rect class="srv__slot" :x="SRV_X - 20" :y="SRV_Y + 10" width="40" height="13" rx="2.5"/>
            <circle class="srv__led" :cx="SRV_X - 13" :cy="SRV_Y - 21.5" r="2"/>
            <circle class="srv__led is-b" :cx="SRV_X - 13" :cy="SRV_Y - 2.5" r="2"/>
            <circle class="srv__led is-c" :cx="SRV_X - 13" :cy="SRV_Y + 16.5" r="2"/>

            <text class="srv__name" :x="SRV_X" :y="SRV_Y + 62">{{ server }}</text>

            <g class="srv__stats">
              <g v-for="(s, i) in srvStats" :key="s.k" :transform="`translate(0 ${SRV_Y + 96 + i * 15})`">
                <text class="srv__k" :x="SRV_X - 40" y="0">{{ s.k }}</text>
                <text class="srv__v" :class="{'is-warn': s.warn}" :x="SRV_X + 40" y="0">{{ s.v }}</text>
              </g>
            </g>
          </g>

          <!-- 链路层自己不做任何 transform：起点已经按当前视图换算过（drawnFlights），
               终点固定落在服务器入口上，所以地图怎么缩放平移，这一头跟着走、那一头不动。
               这里没有 /k 反向缩放 —— 地图那层的缩放已经不作用于它了 -->
          <!-- 每条链路：底层一条粗的半透明「光晕」+ 上层虚线，
               虚线靠 flowing dash 表示数据正往源站走。
               --fly 是这一发的飞行时长，由 JS 按路程算出后交给 CSS -->
          <g
              v-for="(f, i) in drawnFlights"
              :key="f.id"
              class="map__flight"
              :class="f.klass"
              :style="{'--fly': `${f.fly}ms`}"
          >
            <path :d="f.d" class="map__halo" vector-effect="non-scaling-stroke"/>
            <path :d="f.d" class="map__arc" vector-effect="non-scaling-stroke"/>
            <!-- 发射点：出膛那一圈冲击波 -->
            <circle :cx="f.x" :cy="f.y" r="2.6" class="map__ping"/>
            <circle :cx="f.x" :cy="f.y" r="1.9" class="map__from"/>
            <title>{{ f.label }}</title>

            <!-- 弹体：offset-rotate:auto 让这一组跟着路径切线掉头，
                 所以尾焰画在本地 -x 方向就永远是往后拖的「弹尾」，不用逐帧算角度。
                 上面那条虚线只是航路底下淡淡的参考线，这里才是飞出去的那发 -->
            <g
                v-if="i >= labelledFrom"
                class="war"
                :style="{'offset-path': `path('${f.d}')`}"
            >
              <path class="war__flame" d="M-34 0 L-9 -3.6 L0 -2.2 L0 2.2 L-9 3.6 Z"/>
              <circle class="war__glow" r="7"/>
              <circle class="war__core" r="2.8"/>
              <circle class="war__tip" r="1.2"/>
              <!-- 命中闪光：延迟到快抵达时才炸，跟着弹体走所以位置一定在落点上 -->
              <circle class="war__boom" r="6"/>
            </g>

            <!-- 标签单独一组：跟着同一个动画跑，但 offset-rotate:0deg，
                 文字始终水平，不会跟着弹体翻过去 -->
            <g
                v-if="i >= labelledFrom"
                class="pkt"
                :style="{'offset-path': `path('${f.d}')`}"
            >
              <text class="pkt__txt" x="9" y="-5">{{ f.text }}</text>
            </g>
          </g>
        </svg>

        <!-- 缩放控件：滚轮坏了 / 纯键盘的设备也得能放大 -->
        <div class="map__ctrl">
          <button type="button" title="放大" @click="zoomStep(1.7)">＋</button>
          <b>{{ Math.round(scene.k * 100) }}%</b>
          <button type="button" title="缩小" @click="zoomStep(1 / 1.7)">－</button>
          <button type="button" class="map__rst" title="回到整幅地图" @click="resetView">复位</button>
        </div>

        <p class="map__hint">滚轮缩放 · 拖拽平移 · 双击放大 · 点蓝点聚焦</p>

        <p v-if="!locatedTotal" class="map__empty">等待带经纬度的请求…（内网 IP 无归属地，不会出现在图上）</p>

        <div v-if="feed.length" class="map__feed">
          <b>最近来源</b>
          <p v-for="r in feed" :key="r.key" :class="r.klass"><i></i>{{ r.ip }}<em>{{ r.area }}</em></p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.map-body {
  min-height: 0;
}

/* ---------- 地图右侧的虚拟服务器 ---------- */
.srv__wave {
  fill: none;
  stroke: var(--c-brand);
  stroke-width: 1;
  opacity: .5;
  /* 不写这句，scale 会以整个 viewBox 的左上角为原点，波纹会飞出去 */
  transform-box: fill-box;
  transform-origin: center;
  animation: srvWave 2.6s ease-out infinite;
}

@keyframes srvWave {
  0% { transform: scale(.82); opacity: .55; }
  100% { transform: scale(1.18); opacity: 0; }
}

/* 撞击点：一波波往外炸 */
.srv__impact {
  fill: none;
  stroke: var(--c-brand);
  stroke-width: 1;
  transform-box: fill-box;
  transform-origin: center;
  animation: srvImpact 1.2s ease-out infinite;
}

@keyframes srvImpact {
  0% { transform: scale(.4); opacity: .85; }
  100% { transform: scale(3.2); opacity: 0; }
}

.srv__body {
  fill: #fff;
  stroke: var(--c-brand);
  stroke-width: 1.4;
}

.srv__slot {
  fill: #eef3fa;
  stroke: #d5e2f2;
}

/* 三盏灯错开闪，收一个包亮一盏的感觉 */
.srv__led {
  fill: #10b981;
  opacity: .35;
  animation: srvLed 1.4s steps(1, end) infinite;
}

.srv__led.is-b { animation-delay: .46s; }
.srv__led.is-c { animation-delay: .92s; }

@keyframes srvLed {
  0%, 45% { opacity: 1; }
  46%, 100% { opacity: .3; }
}

/* svg 里的字会随 viewBox 缩放：viewBox 1180 宽实际渲染约 1.5~1.9 倍，
   所以这里写的是 viewBox 单位，屏幕上看到大约是它的 1.6 倍 */
.srv__name {
  font-size: 9.5px;
  font-weight: 700;
  letter-spacing: .6px;
  text-anchor: middle;
  fill: var(--c-text);
}



.srv__stats text {
  font-size: 7.5px;
  font-variant-numeric: tabular-nums;
}

.srv__k {
  fill: var(--c-sub);
}

.srv__v {
  text-anchor: end;
  font-weight: 600;
  fill: var(--c-text);
}

.srv__v.is-warn {
  fill: #dc2626;
}

/* ---------- 地图 ---------- */
.map-box {
  position: relative;
  flex: 1;
  min-width: 0;
}

/* height:auto —— 高度由 viewBox 比例推出来，面板不再被锁死。
   max-height 只是超宽屏下的兜底：再高就压到下面板了 */
.map {
  display: block;
  width: 100%;
  height: auto;
  max-height: 78vh;
  cursor: grab;
  /* 不写这句，触屏上的拖拽会被浏览器抢去滚页面 */
  touch-action: none;
}

.map.is-panning {
  cursor: grabbing;
}

/* 平面 / 扁平：纯色填充，不上渐变 */
.map__sphere {
  fill: #f7fafd;
  stroke: #dbe4ef;
}

.map__grid {
  fill: none;
  stroke: #dfe8f3;
}

.map__land {
  fill: #e9eff6;
  stroke: #cfdcea;
}

.map__dot {
  fill: #0052d9;
  opacity: .32;
  /* 每个点都可点（点了聚焦过去），给个手型提示 */
  cursor: pointer;
  transition: opacity .15s;
}

.map__dots > g:hover .map__dot {
  opacity: .85;
}

/* 放大后贴在点旁边的地名。字号写在 <g> 的 scale(1/k) 里被抵消掉了，
   所以这里 7.5px 就是最终屏幕上看到的大小 */
.map__dotlabel {
  font-size: 7.5px;
  fill: var(--c-sub);
  paint-order: stroke;
  stroke: rgba(255, 255, 255, .9);
  stroke-width: 2.5px;
  pointer-events: none;
}

/* ---------- 访问链路 ---------- */
/* 整条链路的生命周期交给父组统一管：淡入 → 打完 → 淡出。
   这里刻意不用固定时长：飞行时长是每发按路程算的（1.5s ~ 3.4s），
   淡出若按统一的 4.2s 收，短程航班落地后还剩两秒空线干挂着。
   所以第二段动画的延迟挂在 --fly 上，谁落地谁收摊。
   FLIGHT_TTL 只管什么时候从 DOM 里删掉，那时候它早就透明了 */
.map__flight {
  animation: flightIn .24s ease-out both, flightOut .7s ease-out both;
  animation-delay: 0s, calc(var(--fly) * .92);
}

@keyframes flightIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes flightOut {
  from { opacity: 1; }
  to { opacity: 0; }
}

/* 底层光晕：粗一点、透明度低，给虚线垫出一层发光的感觉 */
.map__halo {
  fill: none;
  stroke: var(--c-brand);
  stroke-width: 3.6;
  stroke-linecap: round;
  opacity: .13;
}

/* 虚线的流动方向（path 起点=来源，终点=源站）靠 stroke-dashoffset 递减实现，
   看起来就是数据包往源站走。
   亮度压在 .5：这条虚线只是「航路参考线」，亮过头会和弹尾抢视线 */
.map__arc {
  fill: none;
  stroke: var(--c-brand);
  stroke-width: 1.3;
  stroke-linecap: round;
  stroke-dasharray: 4 3;
  opacity: .5;
  animation: arcFlow .9s linear infinite;
}

@keyframes arcFlow {
  to { stroke-dashoffset: -14; }
}

/* 发射冲击波也只放一次：导弹就一发，起点一直冒圈会很吵 */
.map__ping {
  fill: none;
  stroke: var(--c-brand);
  stroke-width: 1;
  transform-box: fill-box;
  transform-origin: center;
  animation: ping 1.1s ease-out forwards;
}

@keyframes ping {
  0% { transform: scale(.5); opacity: .9; }
  100% { transform: scale(3.6); opacity: 0; }
}

.map__from {
  fill: var(--c-brand);
}

/* ---------- 导弹：沿路径飞 + 自动掉头 + 尾焰 ---------- */
/* offset-path 让这一组沿这条链路飞；
   offset-rotate:auto 是关键 —— 元素会跟着路径切线自动转向，
   于是尾焰只要画在本地 -x 方向，看上去永远拖在屁股后面。
   跑不动的浏览器上它会停在起点，那里本来就有 ping 圈，不会变成脏点 */
/* 一发一发地打：不再 infinite ——
   之前是 2.8s 一圈反复跑，而链路只活 4.2s，第二圈还没到头就被删了，
   看着就是「走到一半凭空消失」。
   现在一趟跑完全程就停，收尾交给父组的淡出动画 */
.war,
.pkt {
  animation: pktRun var(--fly) linear forwards;
}

.war {
  offset-rotate: auto;
}

/* 标签不跟着转，否则字会翻过来 */
.pkt {
  offset-rotate: 0deg;
}

@keyframes pktRun {
  0% { offset-distance: 0%; opacity: 0; }
  8% { opacity: 1; }
  /* 88% 走完全程，剩下 12% 停在命中点上淡出 —— 命中的视觉反馈交给 war__boom */
  88% { offset-distance: 100%; opacity: 1; }
  100% { offset-distance: 100%; opacity: 0; }
}

/* 尾焰：一个从头到尾收细的火舌，长 34 个单位（屏幕上约 50px） */
.war__flame {
  fill: var(--c-brand);
  opacity: .22;
}

.war__glow {
  fill: var(--c-brand);
  opacity: .16;
  /* 不写这句，scale 会以整个 viewBox 的左上角为原点，光团会飞出去而不是原地亮 */
  transform-box: fill-box;
  transform-origin: center;
  animation: warGlow .6s ease-in-out infinite alternate;
}

@keyframes warGlow {
  0% { transform: scale(.8); opacity: .3; }
  100% { transform: scale(1.5); opacity: .12; }
}

.war__core {
  fill: var(--c-brand);
  stroke: #fff;
  stroke-width: 1;
}

/* 弹尖：比身子亮一档，用来强调方向 */
.war__tip {
  fill: #fff;
}

/* 命中闪光：延迟到这一发走完全程的那一刻（pktRun 在 88% 抵达）才炸开，
   给「到达」一个明确的收尾，而不是弹头无声无息地淡掉 */
.war__boom {
  fill: none;
  stroke: var(--c-brand);
  stroke-width: 1.4;
  opacity: 0;
  transform-box: fill-box;
  transform-origin: center;
  animation: warBoom .5s ease-out forwards;
  animation-delay: calc(var(--fly) * .87);
}

@keyframes warBoom {
  0% { transform: scale(.3); opacity: .95; }
  100% { transform: scale(2.8); opacity: 0; }
}

/* 文字不受 non-scaling-stroke 保护，会被放大约 1.6 倍，这里写小一号；
   白描边保证压在国界上也能读 */
.pkt__txt {
  font-size: 8px;
  font-weight: 600;
  fill: var(--c-brand);
  paint-order: stroke;
  stroke: rgba(255, 255, 255, .92);
  stroke-width: 3.5px;
}

/* 异常链路：4xx 橙、5xx 红，正常链路保持蓝 */
.map__flight.is-warn .map__halo,
.map__flight.is-warn .map__arc,
.map__flight.is-warn .map__ping,
.map__flight.is-warn .war__core {
  stroke: #f59e0b;
}

.map__flight.is-warn .map__from,
.map__flight.is-warn .war__flame,
.map__flight.is-warn .war__glow,
.map__flight.is-warn .war__core {
  fill: #f59e0b;
}

.map__flight.is-warn .pkt__txt {
  fill: #d97706;
}

.map__flight.is-err .map__halo,
.map__flight.is-err .map__arc,
.map__flight.is-err .map__ping,
.map__flight.is-err .war__core {
  stroke: #ef4444;
}

.map__flight.is-err .map__from,
.map__flight.is-err .war__flame,
.map__flight.is-err .war__glow,
.map__flight.is-err .war__core {
  fill: #ef4444;
}

.map__flight.is-err .pkt__txt {
  fill: #dc2626;
}

/* ---------- 叠加层 ---------- */
/* 缩放控件挂在右上角；底下是 feed（左下），中间那条是空状态提示（正下方） */
.map__ctrl {
  position: absolute;
  right: 2px;
  top: 2px;
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 3px;
  background: rgba(255, 255, 255, .86);
  border: 1px solid var(--c-border);
  border-radius: 8px;
}

.map__ctrl button {
  width: 22px;
  height: 22px;
  font-size: 13px;
  line-height: 1;
  color: var(--c-sub);
  background: transparent;
  border: 0;
  border-radius: 6px;
  cursor: pointer;
  transition: color .15s, background .15s;
}

.map__ctrl button:hover {
  color: var(--c-brand);
  background: #f0f5ff;
}

.map__ctrl b {
  min-width: 34px;
  font-size: 11px;
  font-variant-numeric: tabular-nums;
  text-align: center;
  color: var(--c-text);
}

.map__ctrl .map__rst {
  width: auto;
  padding: 0 6px;
  font-size: 11px;
}

/* 操作提示：这几个手势不写出来的话基本没人知道 */
.map__hint {
  position: absolute;
  right: 6px;
  bottom: 2px;
  margin: 0;
  font-size: 10px;
  color: var(--c-faint);
  pointer-events: none;
}

.map__empty {
  position: absolute;
  left: 50%;
  bottom: 10px;
  transform: translateX(-50%);
  margin: 0;
  font-size: 11px;
  color: var(--c-faint);
}

/* 右下角现在是服务器，最近来源挪到左下 */
.map__feed {
  position: absolute;
  left: 2px;
  bottom: 2px;
  display: grid;
  gap: 3px;
  padding: 6px 9px;
  font-size: 11px;
  background: rgba(255, 255, 255, .82);
  border: 1px solid var(--c-border);
  border-radius: 8px;
}

.map__feed b {
  color: var(--c-faint);
  font-weight: 600;
  letter-spacing: 1px;
}

.map__feed p {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 5px;
  color: var(--c-text);
}

/* <｜hy_place▁holder▁no▁813｜>的小圆点跟链路同色，一眼看出状态码分类 */
.map__feed i {
  width: 6px;
  height: 6px;
  flex: none;
  border-radius: 50%;
  background: var(--c-brand);
}

.map__feed .is-warn i { background: #f59e0b; }
.map__feed .is-err i { background: #ef4444; }

.map__feed em {
  font-style: normal;
  color: var(--c-faint);
}

.map__lg {
  margin-left: auto;
  display: flex;
  gap: 10px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0;
}

.map__lg b {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--c-sub);
}

.map__lg i {
  width: 12px;
  height: 2px;
  border-radius: 1px;
  background: var(--c-brand);
}

.map__lg .is-warn i { background: #f59e0b; }
.map__lg .is-err i { background: #ef4444; }


</style>
