<script setup>
import {computed, onUnmounted, ref, watch} from 'vue'
import {fmtAgo, fmtNum, fmtClock, pct, rankSorted, useNow, useTween} from './format'

const props = defineProps({
  // summary.wafBlocks：服务端每秒随大盘下发，只含聚合量
  waf: {type: Object, default: () => ({})},
})

/* 严重级别：coraza 从高到低共 8 档，后端给的是小写英文 key。
   颜色按危险程度从深红到灰，一眼能看出「是有攻击还是只是规则误报」 */
const SEVERITY = [
  {key: 'emergency', label: '紧急', color: '#7f1d1d'},
  {key: 'alert', label: '警报', color: '#b91c1c'},
  {key: 'critical', label: '严重', color: '#ef4444'},
  {key: 'error', label: '错误', color: '#f97316'},
  {key: 'warning', label: '警告', color: '#f59e0b'},
  {key: 'notice', label: '注意', color: '#0891b2'},
  {key: 'info', label: '信息', color: '#10b981'},
  {key: 'debug', label: '调试', color: '#94a3b8'},
]

/* 挂载位置：反向代理和穿透代理各挂一份 WAF，合并在一处统计 */
const SOURCES = [
  {key: 'reverse', label: '反向代理', color: '#0052d9'},
  {key: 'tunnel', label: '穿透代理', color: '#0891b2'},
]

const now = useNow(1000)

const raw = computed(() => props.waf || {})
const total = useTween(() => Number(raw.value.total) || 0)

const srcList = computed(() => {
  const m = raw.value.source || {}
  const sum = SOURCES.reduce((a, s) => a + (Number(m[s.key]) || 0), 0)
  return SOURCES.map((s) => {
    const count = Number(m[s.key]) || 0
    return {
      ...s,
      count,
      pct: sum ? `${Math.max(1, Math.round(count / sum * 100))}%` : '0%',
    }
  })
})

/* 只保留有数据的级别：8 档全列出来会有一半是 0，堆叠条也占不到宽度 */
const sevList = computed(() => {
  const m = raw.value.severity || {}
  const sum = SEVERITY.reduce((a, s) => a + (Number(m[s.key]) || 0), 0)
  return SEVERITY
      .map((s) => {
        const count = Number(m[s.key]) || 0
        return {
          ...s,
          count,
          pct: sum ? `${(count / sum * 100).toFixed(1)}%` : '0%',
          frac: sum ? count / sum : 0,
        }
      })
      .filter((s) => s.count > 0)
})

const lastMsg = computed(() => String(raw.value.lastMsg || '').trim())
const lastAt = computed(() => Number(raw.value.lastAt) || 0)
const lastText = computed(() => lastAt.value
    ? `${fmtClock(lastAt.value)} · ${fmtAgo(lastAt.value, now.value)}`
    : '暂无')

/* TOP 榜 key 可能是数字（规则 ID），sort 前统一转字符串 */
const topRules = computed(() => rankSorted(raw.value.topRules))
const topDomains = computed(() => rankSorted(raw.value.topDomains))
const topIps = computed(() => rankSorted(raw.value.topIps))

/* 命中闪红：大盘一秒一帧，没有这道脉冲的话，攻击正在发生时也不显眼 */
const hit = ref(false)
let hitTimer = null
watch(() => Number(raw.value.total) || 0, (to, from) => {
  if (to <= from) return
  hit.value = true
  clearTimeout(hitTimer)
  hitTimer = setTimeout(() => {
    hit.value = false
  }, 1200)
})
onUnmounted(() => clearTimeout(hitTimer))
</script>

<template>
  <!-- 根节点必须是单个 .panel：多根节点时父组件传的 class 无处可落 -->
  <div class="panel panel--waf">
    <h2>
      防火墙拦截<em>coraza · 反向代理 + 穿透代理</em>
      <span class="waf__live" :class="{'is-hit': hit}"><i/>{{ hit ? '正在拦截' : '监控中' }}</span>
    </h2>

    <div class="waf">
      <!-- 概览：累计命中 + 挂载位置分布 + 最近一次命中 -->
      <div class="waf__block waf__overview">
        <div class="waf__total">
          <b>{{ fmtNum(total) }}</b>
          <span>累计规则命中</span>
        </div>
        <div v-for="s in srcList" :key="s.key" class="waf__src">
          <span>{{ s.label }}</span>
          <b>{{ fmtNum(s.count) }}</b>
          <i class="waf__bar"><u :style="{width: s.pct, background: s.color}"/></i>
        </div>
        <div class="waf__last">
          <span>最近命中</span>
          <b :title="lastMsg || '暂无'">{{ lastMsg || '暂无' }}</b>
          <em>{{ lastText }}</em>
        </div>
      </div>

      <!-- 严重级别：堆叠条看构成，图例给精确值 -->
      <div class="waf__block">
        <span class="waf__lbl">严重级别<em>累计</em></span>
        <div class="waf__stack">
          <i v-for="s in sevList" :key="s.key" :style="{width: s.pct, background: s.color}" :title="`${s.label} ${s.count}`"/>
        </div>
        <ul class="waf__legend">
          <li v-for="s in sevList" :key="s.key">
            <i :style="{background: s.color}"/>
            <span>{{ s.label }}</span>
            <b>{{ fmtNum(s.count) }}</b>
            <em>{{ s.pct }}</em>
          </li>
          <li v-if="!sevList.length" class="waf__legend-empty">暂无命中</li>
        </ul>
      </div>

      <div class="waf__block">
        <span class="waf__lbl">TOP 规则<em>近 5 小时</em></span>
        <ul class="rank">
          <li v-for="(it, i) in topRules" :key="it.key">
            <span class="rank__no">{{ i + 1 }}</span>
            <span class="rank__name">规则 {{ it.key }}</span>
            <span class="rank__num">{{ fmtNum(it.count) }}</span>
            <span class="rank__bar"><i :style="{width: pct(it.count, topRules)}"/></span>
          </li>
          <li v-if="!topRules.length" class="rank__empty">暂无数据</li>
        </ul>
      </div>

      <div class="waf__block">
        <span class="waf__lbl">TOP 被拦域名<em>近 5 小时</em></span>
        <ul class="rank">
          <li v-for="(it, i) in topDomains" :key="it.key">
            <span class="rank__no">{{ i + 1 }}</span>
            <span class="rank__name" :title="it.key">{{ it.key }}</span>
            <span class="rank__num">{{ fmtNum(it.count) }}</span>
            <span class="rank__bar"><i :style="{width: pct(it.count, topDomains)}"/></span>
          </li>
          <li v-if="!topDomains.length" class="rank__empty">暂无数据</li>
        </ul>
      </div>

      <div class="waf__block">
        <span class="waf__lbl">TOP 来源 IP<em>近 5 小时</em></span>
        <ul class="rank">
          <li v-for="(it, i) in topIps" :key="it.key">
            <span class="rank__no">{{ i + 1 }}</span>
            <span class="rank__name">{{ it.key }}</span>
            <span class="rank__num">{{ fmtNum(it.count) }}</span>
            <span class="rank__bar"><i :style="{width: pct(it.count, topIps)}"/></span>
          </li>
          <li v-if="!topIps.length" class="rank__empty">暂无数据</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 整行面板要塞「概览 + 严重级别 + 三张榜」，横向五块平分。
   每块内部自己滚动，超出部分裁掉，不给面板顶高 */
.waf {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: 1.15fr 1.1fr 1fr 1fr 1fr;
  gap: 14px;
  overflow: hidden;
}

.waf__block {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  padding-right: 14px;
  border-right: 1px dashed var(--c-border);
}

.waf__block:last-child {
  padding-right: 0;
  border-right: none;
}

.waf__lbl {
  display: block;
  margin-bottom: 8px;
  font-size: 10.5px;
  letter-spacing: 1px;
  color: var(--c-sub);
}

.waf__lbl em {
  margin-left: 5px;
  font-style: normal;
  font-size: 9.5px;
  color: var(--c-faint);
}

/* ---------- 概览 ---------- */
.waf__overview {
  gap: 8px;
}

.waf__total b {
  display: block;
  font-size: 30px;
  font-weight: 700;
  line-height: 1.1;
  font-family: 'DIN Alternate', 'SF Mono', Consolas, Menlo, monospace;
  color: var(--c-brand);
}

.waf__total span {
  font-size: 11px;
  letter-spacing: 1px;
  color: var(--c-faint);
}

.waf__src {
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: center;
  gap: 2px 8px;
  font-size: 11.5px;
}

.waf__src span {
  color: var(--c-sub);
}

.waf__src b {
  color: var(--c-text);
  font-variant-numeric: tabular-nums;
}

.waf__bar {
  grid-column: 1 / -1;
  display: block;
  height: 4px;
  border-radius: 2px;
  overflow: hidden;
  background: #eef2f7;
}

.waf__bar u {
  display: block;
  height: 100%;
  border-radius: 2px;
  transition: width .5s ease;
}

/* 最近命中摘要：贴着底边，和上面的数字拉开距离 */
.waf__last {
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid var(--c-border);
}

.waf__last span {
  display: block;
  font-size: 10px;
  letter-spacing: 1px;
  color: var(--c-faint);
}

.waf__last b {
  display: block;
  margin: 2px 0 1px;
  font-size: 11px;
  font-weight: 600;
  line-height: 1.35;
  color: var(--c-text);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.waf__last em {
  font-style: normal;
  font-size: 10px;
  color: var(--c-faint);
}

/* ---------- 状态灯 ---------- */
.waf__live {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 1px;
  color: var(--c-faint);
}

.waf__live i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #10b981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, .16);
}

/* 有新命中时变红并闪烁一下 */
.waf__live.is-hit {
  color: #dc2626;
}

.waf__live.is-hit i {
  background: #ef4444;
  box-shadow: 0 0 0 3px rgba(239, 68, 68, .2);
  animation: wafHit .6s ease-in-out infinite;
}

@keyframes wafHit {
  0%, 100% { opacity: 1; }
  50% { opacity: .25; }
}

/* ---------- 严重级别 ---------- */
/* 堆叠条：各级别按比例挤在一条里，看的是「构成」而不是精确值 */
.waf__stack {
  display: flex;
  height: 10px;
  border-radius: 5px;
  overflow: hidden;
  background: #eef2f7;
}

.waf__stack i {
  height: 100%;
  transition: width .5s ease;
}

.waf__legend {
  margin: 10px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  align-content: start;
  gap: 6px;
  font-size: 11.5px;
  overflow-y: auto;
}

.waf__legend li {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--c-sub);
}

.waf__legend i {
  flex: none;
  width: 8px;
  height: 8px;
  border-radius: 2px;
}

.waf__legend b {
  margin-left: auto;
  color: var(--c-text);
  font-variant-numeric: tabular-nums;
}

/* 占比固定占一列并右对齐，数字跳动时不会把整行挤歪 */
.waf__legend em {
  flex: none;
  width: 46px;
  text-align: right;
  font-style: normal;
  color: var(--c-faint);
  font-variant-numeric: tabular-nums;
}

.waf__legend-empty {
  font-size: 11px;
  color: var(--c-faint);
}

.waf__legend::-webkit-scrollbar {
  width: 5px;
}

.waf__legend::-webkit-scrollbar-thumb {
  background: #d8e0ea;
  border-radius: 3px;
}

/* 窄屏五列会把每块压到 200px 以下，换成按最小宽度自动换行 */
@media (max-width: 1500px) {
  .waf {
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 12px;
    overflow-y: auto;
  }

  .waf__block {
    padding-right: 0;
    border-right: none;
    min-height: 150px;
  }

  .waf__last {
    margin-top: 8px;
  }
}
</style>
