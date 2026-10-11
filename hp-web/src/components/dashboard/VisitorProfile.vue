<script setup>
defineProps({
  // 每组 {label, items:[{key, count, pct}]}，pct 已由外层算好（带 % 的字符串，可直接当宽度用）
  groups: {type: Array, default: () => []},
})
</script>

<template>
  <!-- 根节点必须是单个 .panel：多根节点时父组件传的 class 无处可落 -->
  <div class="panel panel--profile">
    <h2>访客画像<em>近 5 小时</em></h2>
    <div class="profile">
      <div v-for="g in groups" :key="g.label" class="profile__group">
        <span class="profile__label">{{ g.label }}</span>
        <ul class="profile__list">
          <li v-for="it in g.items" :key="it.key">
            <span class="profile__name" :title="`${it.key} · ${it.count} 条`">{{ it.key }}</span>
            <b>{{ it.pct }}</b>
            <span class="profile__bar"><i :style="{width: it.pct}"/></span>
          </li>
          <li v-if="!g.items.length" class="profile__empty">暂无数据</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 设备 / 系统 / 浏览器 / 运营商四组，每组取 TOP3。2×2 排布，竖着一列会顶出面板。
   来源国家不在这里统计：服务端 topCountries 已经是全站口径，单独成榜 */
.profile {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-content: start;
  gap: 12px 16px;
}

.profile__label {
  display: block;
  font-size: 11px;
  letter-spacing: 1px;
  color: var(--c-faint);
}

.profile__list {
  margin: 4px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 6px;
}

.profile__list li {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 32px;
  align-items: center;
  gap: 3px 6px;
  font-size: 12px;
  line-height: 1.3;
}

.profile__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--c-text);
}

.profile__list b {
  font-size: 10px;
  font-weight: 600;
  text-align: right;
  color: var(--c-sub);
  font-variant-numeric: tabular-nums;
}

.profile__bar {
  grid-column: 1 / -1;
  height: 4px;
  border-radius: 2px;
  overflow: hidden;
  background: #eef2f7;
}

.profile__bar i {
  display: block;
  height: 100%;
  border-radius: 2px;
  background: linear-gradient(90deg, #0052d9, #21b7e8);
  transition: width .5s ease;
}

.profile__empty {
  font-size: 10.5px;
  color: var(--c-faint);
}

.profile::-webkit-scrollbar {
  width: 5px;
}

.profile::-webkit-scrollbar-thumb {
  background: #d8e0ea;
  border-radius: 3px;
}

@media (max-width: 1500px) {
  /* 半宽的画像面板放不下 3 列（每列只剩 ~110px），退回 2 列 */
  .profile {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  /* 手机上画像面板只有一列宽，多列会把「Windows 10/11」压成省略号 */
  .profile {
    grid-template-columns: 1fr;
  }
}
</style>
