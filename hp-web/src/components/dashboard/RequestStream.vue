<script setup>
import {computed} from 'vue'
import {clientText, fmtBytes, fmtNum, geoText, metaTitle, shortTime} from './format'

const props = defineProps({
  rows: {type: Array, default: () => []},
  dense: {type: Boolean, default: false},
})

/* 请求流窗口内汇总：标题原来只有「最近 N 条」，信息量太少，
   补上窗口均耗时、异常数、流量合计 */
const streamStat = computed(() => {
  const d = props.rows
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
</script>

<template>
  <!-- 根节点必须是单个 .panel：多根节点时父组件传的 class 无处可落 -->
  <div class="panel panel--stream">
    <h2>
      实时请求流<em>最近 {{ rows.length }} 条 · 均 {{ streamStat.avg }}ms
      · 异常 {{ streamStat.bad }} · {{ fmtBytes(streamStat.bytes) }}</em>
    </h2>
    <ul v-if="rows.length" class="stream" :class="{'is-dense': dense}">
      <li v-for="r in rows" :key="r.id">
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
</template>

<style scoped>
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

.stream::-webkit-scrollbar {
  width: 5px;
}

.stream::-webkit-scrollbar-thumb {
  background: #d8e0ea;
  border-radius: 3px;
}

@media (max-width: 768px) {
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
}
</style>
