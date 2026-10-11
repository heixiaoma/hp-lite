<script setup>
import {computed} from 'vue'
import {fmtNum} from './format'

const props = defineProps({
  counts: {type: Object, default: () => ({})},
})

// 环形图分段色：白底需要更饱和的深色才看得清，深色主题的荧光色在白底上会发飘
const STATUS_CLASSES = [
  {cls: '1xx', color: '#8b5cf6'},
  {cls: '2xx', color: '#10b981'},
  {cls: '3xx', color: '#0891b2'},
  {cls: '4xx', color: '#f59e0b'},
  {cls: '5xx', color: '#ef4444'},
]

// 环形图：周长固定，每段用 dasharray 表示占比，dashoffset 累加偏移起点
const RING_C = 2 * Math.PI * 38

const statusTotal = computed(() => {
  const c = props.counts || {}
  return Object.values(c).reduce((a, b) => a + (Number(b) || 0), 0)
})

const ringSegs = computed(() => {
  const c = props.counts || {}
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
const c4xx = computed(() => Number(props.counts['4xx']) || 0)
const c5xx = computed(() => Number(props.counts['5xx']) || 0)
const badCount = computed(() => c4xx.value + c5xx.value)

const okRate = computed(() => {
  const t = statusTotal.value
  return t ? `${(((t - badCount.value) / t) * 100).toFixed(1)}%` : '--'
})
</script>

<template>
  <!-- 根节点必须是单个 .panel：多根节点时父组件传的 class 无处可落 -->
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
</template>

<style scoped>
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
</style>
