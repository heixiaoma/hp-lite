<script setup>
import {fmtNum, pct} from './format'

// 来源国家 / 域名 / IP / 路径 / 防火墙 TOP 榜共用的排行列
defineProps({
  title: {type: String, default: ''},
  note: {type: String, default: ''},
  items: {type: Array, default: () => []},
})
</script>

<template>
  <!-- 根节点必须是单个 .panel：多根节点时父组件传的 class 无处可落 -->
  <div class="panel panel--top">
    <h2>{{ title }}<em>{{ note }}</em></h2>
    <ul class="rank">
      <li v-for="(it, i) in items" :key="it.key">
        <span class="rank__no">{{ i + 1 }}</span>
        <span class="rank__name" :title="`${it.key} · ${fmtNum(it.count)} 条`">{{ it.key }}</span>
        <span class="rank__num">{{ fmtNum(it.count) }}</span>
        <span v-if="it.sub" class="rank__sub" :title="it.sub">{{ it.sub }}</span>
        <span class="rank__bar">
          <i :style="{width: pct(it.count, items)}"/>
        </span>
      </li>
      <li v-if="!items.length" class="rank__empty">暂无数据</li>
    </ul>
  </div>
</template>
