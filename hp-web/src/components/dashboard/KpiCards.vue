<script setup>
defineProps({
  items: {type: Array, default: () => []},
})
</script>

<template>
  <section class="kpi">
    <div v-for="k in items" :key="k.label" class="kpi__card">
      <span class="kpi__label">{{ k.label }}</span>
      <span class="kpi__value" :style="{color: k.color}">{{ k.text }}</span>
      <span class="kpi__sub">{{ k.sub }}</span>
    </div>
  </section>
</template>

<style scoped>
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
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* KPI 拆行会把 board 的高度挤掉，所以断点压到 900px 以下才拆行 */
@media (max-width: 900px) {
  .kpi {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .kpi {
    grid-template-columns: repeat(2, 1fr);
  }

  .kpi__value {
    font-size: 21px;
  }
}
</style>
