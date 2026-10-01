<template>
  <!-- 无数据时（请求失败 / 统计值全为 0）整块不渲染，描述与数据条一起隐藏 -->
  <div v-if="items.length" class="la-stats-band" :class="`la-stats-band--${theme}`">
    <p v-if="title" class="la-stats-band__title">{{ title }}</p>

    <div class="la-stats">
      <div v-for="it in items" :key="it.key" class="la-stats__item">
        <span class="la-stats__label">{{ it.label }}</span>
        <span class="la-stats__value">{{ it.display }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import {onMounted, ref} from 'vue';
import {fetchLaStats} from '../data/laStats';

const props = defineProps({
  // light：浅色底（默认）  dark：深色底（页脚那种）
  theme: {type: String, default: 'light'},
  // 数据条上方的描述文案，传空字符串可只显示数据
  title: {type: String, default: ''},
});

const items = ref([]);

const reduceMotion =
    window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false;

const easeOutExpo = t => (t === 1 ? 1 : 1 - Math.pow(2, -10 * t));

function countUp(item, delay) {
  const to = item.raw;
  if (reduceMotion || to <= 0) {
    item.display = item.value;
    return;
  }

  const duration = 1100;
  let start = null;

  const step = now => {
    if (start === null) start = now;
    const elapsed = now - start - delay;
    if (elapsed < 0) {
      requestAnimationFrame(step);
      return;
    }
    const t = Math.min(1, elapsed / duration);
    item.display = Math.round(to * easeOutExpo(t)).toLocaleString('en-US');
    if (t < 1) requestAnimationFrame(step);
    else item.display = item.value;
  };

  requestAnimationFrame(step);
}

onMounted(async () => {
  let data;
  try {
    data = await fetchLaStats();
  } catch (e) {
    console.warn('[la-stats] 统计数据加载失败：', e.message);
    return;
  }

  // 请求成功但没有数据：保持 items 为空，整块不渲染
  if (!data.length) return;

  items.value = data.map(d => ({...d, display: '0'}));
  items.value.forEach((item, i) => countUp(item, i * 70));
});
</script>

<style scoped>
/* ============ 主题变量 ============ */
.la-stats-band--dark {
  --la-bg-from: rgba(255, 255, 255, .07);
  --la-bg-to: rgba(255, 255, 255, .02);
  --la-border: rgba(255, 255, 255, .09);
  --la-inset: rgba(255, 255, 255, .07);
  --la-shadow: 0 10px 30px rgba(4, 10, 26, .28);
  --la-blur: blur(8px);
  --la-title: rgba(255, 255, 255, .45);
  --la-label: rgba(255, 255, 255, .42);
  --la-divider: rgba(255, 255, 255, .2);
  --la-num-from: #ffffff;
  --la-num-to: #a9c4ff;
  --la-highlight: rgba(139, 170, 255, .95);
}

.la-stats-band--light {
  --la-bg-from: #ffffff;
  --la-bg-to: #ffffff;
  --la-border: rgba(16, 24, 40, .08);
  --la-inset: transparent;
  --la-shadow: 0 1px 2px rgba(16, 24, 40, .04), 0 8px 22px rgba(16, 24, 40, .05);
  --la-blur: none;
  --la-title: var(--hp-text-3);
  --la-label: var(--hp-text-3);
  --la-divider: rgba(16, 24, 40, .14);
  --la-num-from: #1b2437;
  --la-num-to: #3f6cf6;
  --la-highlight: #6189fb;
}

.la-stats-band {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.la-stats-band__title {
  margin: 0;
  font-size: 13px;
  letter-spacing: .04em;
  color: var(--la-title);
}

.la-stats {
  position: relative;
  display: flex;
  align-items: stretch;
  justify-content: center;
  /* 不加限制的话会撑满 1180px 容器，7 项显得过于稀疏；
     显式给 width:100% —— 父级若是 column flex，否则会收缩到内容宽度 */
  width: 100%;
  max-width: 920px;
  margin: 0 auto;
  padding: 18px 6px;
  border-radius: 18px;
  border: 1px solid var(--la-border);
  background: linear-gradient(180deg, var(--la-bg-from) 0%, var(--la-bg-to) 100%);
  box-shadow: inset 0 1px 0 var(--la-inset), var(--la-shadow);
  backdrop-filter: var(--la-blur);
  overflow: hidden;
}

/* 顶部居中一条品牌色高光 */
.la-stats::before {
  content: '';
  position: absolute;
  left: 50%;
  top: 0;
  transform: translateX(-50%);
  width: 120px;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--la-highlight), transparent);
}

.la-stats__item {
  position: relative;
  flex: 1 1 0;
  min-width: 0;
  padding: 2px 14px;
  text-align: center;
}

/* 分隔线：中间实、两端渐隐 */
.la-stats__item + .la-stats__item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 1px;
  height: 30px;
  background: linear-gradient(180deg, transparent, var(--la-divider), transparent);
}

.la-stats__label {
  display: block;
  font-size: 11px;
  letter-spacing: .1em;
  line-height: 1.4;
  color: var(--la-label);
  white-space: nowrap;
}

.la-stats__value {
  display: block;
  margin-top: 6px;
  font-size: 23px;
  font-weight: 600;
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
  background: linear-gradient(135deg, var(--la-num-from) 0%, var(--la-num-to) 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

/* 兜底：不支持 background-clip:text 时若仍保持 color:transparent，数字会彻底看不见 */
@supports not ((-webkit-background-clip: text) or (background-clip: text)) {
  .la-stats__value {
    background: none;
    color: var(--la-num-from);
  }
}

@media (max-width: 720px) {
  .la-stats {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px 8px;
    padding: 16px 4px;
  }

  .la-stats__item + .la-stats__item::before {
    display: none;
  }

  /* 7 项两列排布，「总访问量」单独占满末行并居中 */
  .la-stats__item:last-child {
    grid-column: span 2;
  }

  .la-stats__value {
    font-size: 19px;
  }
}
</style>
