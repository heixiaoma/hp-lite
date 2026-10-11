<script setup>
defineProps({
  status: {type: String, default: 'connecting'},
  statusText: {type: String, default: ''},
  clock: {type: String, default: ''},
  isFull: {type: Boolean, default: false},
})

defineEmits(['toggle-fullscreen'])
</script>

<template>
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
      <span class="dash__clock">{{ clock }}</span>
      <button class="dash__btn" @click="$emit('toggle-fullscreen')">
        {{ isFull ? '退出全屏' : '全屏' }}
      </button>
    </div>
  </header>
</template>

<style scoped>
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

.dash__btn:hover {
  color: var(--c-brand);
  background: #f0f5ff;
  border-color: #b8d0ff;
}

.dash__clock {
  font-size: 12px;
  color: var(--c-sub);
}

/* 顶栏一行塞不下「标题 + 连接状态 + 时钟 + 全屏」。
   不加 wrap 时 flex 会去压缩各块宽度，标题和状态就被挤成竖排了 */
@media (max-width: 1500px) {
  .dash__head {
    flex-wrap: wrap;
    gap: 10px 14px;
  }

  .brand,
  .dash__state {
    flex: none;
    white-space: nowrap;
  }

  .dash__tools {
    min-width: 0;
    margin-left: 0;
    flex-wrap: wrap;
  }

  .dash__btn,
  .dash__clock {
    white-space: nowrap;
  }
}
</style>
