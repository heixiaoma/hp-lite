<template>
  <div class="mc">
    <!-- 无数据时给个明确的空态，而不是渲染两张空图 -->
    <div v-if="!hasData" class="mc__empty">
      <div class="mc__empty-title">暂无监控数据</div>
      <div class="mc__empty-desc">该通道还没有产生流量与访问记录</div>
    </div>

    <template v-else>
      <div class="mc__stats">
        <div class="mc__stat">
          <div class="mc__stat-label">累计上传</div>
          <div class="mc__stat-value mc__stat-value--up">{{ stats.upload }}</div>
        </div>
        <div class="mc__stat">
          <div class="mc__stat-label">累计下载</div>
          <div class="mc__stat-value mc__stat-value--down">{{ stats.download }}</div>
        </div>
        <div class="mc__stat">
          <div class="mc__stat-label">累计 PV</div>
          <div class="mc__stat-value">{{ stats.pv }} <span class="mc__stat-unit">次</span></div>
        </div>
        <div class="mc__stat">
          <div class="mc__stat-label">峰值 UV</div>
          <div class="mc__stat-value">{{ stats.uvPeak }} <span class="mc__stat-unit">人</span></div>
        </div>
      </div>

      <div class="mc__card">
        <div class="mc__head">
          <div class="mc__title">流量趋势</div>
          <div class="mc__unit">单位：{{ flowUnit.unit }}</div>
        </div>
        <div ref="flowRef" class="mc__canvas"></div>
      </div>

      <div class="mc__card">
        <div class="mc__head">
          <div class="mc__title">访问趋势</div>
          <div class="mc__unit">单位：人</div>
        </div>
        <div ref="accessRef" class="mc__canvas"></div>
      </div>

      <div class="mc__foot">共 {{ stats.count }} 个采集点 · {{ rangeText }}</div>
    </template>
  </div>
</template>

<script setup>
import * as echarts from 'echarts/core';
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent
} from 'echarts/components';
import {LineChart} from 'echarts/charts';
import {CanvasRenderer} from 'echarts/renderers';
import {computed, nextTick, onBeforeUnmount, onMounted, ref, watch} from 'vue';

echarts.use([
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  LineChart,
  CanvasRenderer
]);

const props = defineProps({
  name: {
    type: String,
    required: true
  },
  value: {
    type: Array,
    required: true
  }
});

/* ==========================================================================
   单位换算
   upload / download 是「字节」，pv / uv 是「人数」。
   字节必须自适应单位：示例数据里 upload 只有 178 字节，
   原来写死 /1024/1024 会全部显示成 0.00
   ========================================================================== */

const BYTE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];

/** 按峰值挑一个统一单位，让 y 轴刻度有意义（不能每个点各用各的单位） */
const pickUnit = (maxBytes) => {
  let i = 0;
  let v = Math.max(Number(maxBytes) || 0, 0);
  while (v >= 1024 && i < BYTE_UNITS.length - 1) {
    v /= 1024;
    i++;
  }
  return {unit: BYTE_UNITS[i], divisor: Math.pow(1024, i)};
};

/** 单个值格式化，各用各的量级（用于 tooltip 和概览，追求可读） */
const formatBytes = (bytes) => {
  const n = Math.max(Number(bytes) || 0, 0);
  if (n <= 0) return '0 B';
  const i = Math.min(Math.floor(Math.log(n) / Math.log(1024)), BYTE_UNITS.length - 1);
  const v = n / Math.pow(1024, i);
  const digits = i === 0 ? 0 : (v >= 100 ? 1 : 2);
  return `${v.toFixed(digits)} ${BYTE_UNITS[i]}`;
};

/** 折线上打点用的换算：统一除以 divisor，避免各点单位不一致 */
const scale = (bytes, divisor) => {
  const v = (Number(bytes) || 0) / divisor;
  return divisor === 1 ? v : Number(v.toFixed(3));
};

/* ==========================================================================
   时间格式
   ========================================================================== */

const pad = (n) => String(n).padStart(2, '0');

const formatFull = (ts) => {
  const d = new Date(ts);
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
      `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
};

const isSameDay = (list) => {
  if (list.length < 2) return true;
  const a = new Date(list[0].time);
  const b = new Date(list[list.length - 1].time);
  return a.getFullYear() === b.getFullYear()
      && a.getMonth() === b.getMonth()
      && a.getDate() === b.getDate();
};

/** x 轴标签要短：同一天只显示 时:分，跨天补上 月-日 */
const axisLabels = computed(() => {
  const list = sorted.value;
  const sameDay = isSameDay(list);
  return list.map(item => {
    const d = new Date(item.time);
    const hm = `${pad(d.getHours())}:${pad(d.getMinutes())}`;
    return sameDay ? hm : `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${hm}`;
  });
});

/* ==========================================================================
   数据
   ========================================================================== */

const sorted = computed(() => {
  const list = Array.isArray(props.value)
      ? props.value.filter(item => item && typeof item.time === 'number')
      : [];
  return [...list].sort((a, b) => a.time - b.time);
});

const hasData = computed(() => sorted.value.length > 0);

const stats = computed(() => {
  const list = sorted.value;
  const sum = key => list.reduce((s, item) => s + (Number(item[key]) || 0), 0);
  return {
    // PV 是浏览量，可以累加；UV 是独立访客，累加会重复计算，取峰值才有意义
    upload: formatBytes(sum('upload')),
    download: formatBytes(sum('download')),
    pv: sum('pv'),
    uvPeak: list.reduce((m, item) => Math.max(m, Number(item.uv) || 0), 0),
    count: list.length
  };
});

const flowUnit = computed(() => {
  const max = sorted.value.reduce(
      (m, item) => Math.max(m, Number(item.upload) || 0, Number(item.download) || 0), 0);
  return pickUnit(max);
});

const rangeText = computed(() => {
  const list = sorted.value;
  if (!list.length) return '';
  return `${formatFull(list[0].time)} ~ ${formatFull(list[list.length - 1].time)}`;
});

/* ==========================================================================
   图表
   ========================================================================== */

const COLOR = {
  upload: '#4b6ff6',
  download: '#10b981',
  pv: '#4b6ff6',
  uv: '#f59e0b'
};

/** 纯对象形式的渐变，不依赖 echarts.graphic（按需引入时可省一个依赖） */
const areaFill = (rgb, opacity = 0.26) => ({
  type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
  colorStops: [
    {offset: 0, color: `rgba(${rgb}, ${opacity})`},
    {offset: 1, color: `rgba(${rgb}, 0.01)`}
  ]
});

const axisCommon = {
  axisLine: {lineStyle: {color: '#e9edf5'}},
  axisTick: {show: false},
  axisLabel: {color: '#94a3b8', fontSize: 12}
};

const legendCommon = {
  top: 0,
  right: 0,
  icon: 'roundRect',
  itemWidth: 10,
  itemHeight: 10,
  itemGap: 16,
  textStyle: {color: '#64748b', fontSize: 12}
};

const tooltipCommon = {
  trigger: 'axis',
  confine: true,
  backgroundColor: 'rgba(255,255,255,.97)',
  borderColor: '#e9edf5',
  borderWidth: 1,
  padding: [10, 14],
  textStyle: {color: '#1b2437', fontSize: 12},
  extraCssText: 'border-radius:10px;box-shadow:0 8px 24px rgba(16,24,40,.12);',
  axisPointer: {type: 'line', lineStyle: {color: '#b6c8ff', width: 1}}
};

const buildSeries = (name, color, rgb, data) => ({
  name,
  type: 'line',
  smooth: 0.35,
  showSymbol: false,
  symbol: 'circle',
  symbolSize: 7,
  lineStyle: {width: 2, color},
  itemStyle: {color, borderColor: '#fff', borderWidth: 2},
  areaStyle: {color: areaFill(rgb)},
  emphasis: {focus: 'series'},
  data
});

const flowRef = ref(null);
const accessRef = ref(null);
let flowChart = null;
let accessChart = null;
let resizeObserver = null;

const buildFlowOption = () => {
  const {unit, divisor} = flowUnit.value;
  const list = sorted.value;
  // 数据点少了就别放 dataZoom，纯噪音
  const zoom = list.length > 20 ? [{
    type: 'inside',
    realtime: true,
    start: 0,
    end: 100,
    zoomOnMouseWheel: false
  }] : [];

  const formatter = (params) => {
    const arr = Array.isArray(params) ? params : [params];
    if (!arr.length) return '';
    const raw = list[arr[0].dataIndex];
    let html = '<div style="font-size:12px;color:#64748b;margin-bottom:6px">' +
        formatFull(raw.time) + '</div>';
    arr.forEach(p => {
      // 从原始记录取字节数，避免折线上只能拿到换算后的值
      const bytes = p.seriesName === '上传' ? raw.upload : raw.download;
      html += '<div style="display:flex;align-items:center;gap:8px;margin-top:4px">' +
          '<span style="width:8px;height:8px;border-radius:50%;background:' + p.color + '"></span>' +
          '<span style="flex:1;color:#64748b">' + p.seriesName + '</span>' +
          '<strong style="color:#1b2437">' + formatBytes(bytes) + '</strong>' +
          '</div>';
    });
    return html;
  };

  return {
    grid: {left: 4, right: 16, top: 40, bottom: 8, containLabel: true},
    legend: {...legendCommon, data: ['上传', '下载']},
    tooltip: {...tooltipCommon, formatter},
    dataZoom: zoom,
    xAxis: {type: 'category', boundaryGap: false, data: axisLabels.value, ...axisCommon,
      axisLabel: {...axisCommon.axisLabel, hideOverlap: true}},
    yAxis: {
      type: 'value',
      name: `流量/${unit}`,
      nameTextStyle: {color: '#94a3b8', fontSize: 12, align: 'left'},
      ...axisCommon,
      splitLine: {lineStyle: {color: '#f1f4f9', type: 'dashed'}}
    },
    // 必须传纯数值数组：echarts 会把 [{value, raw}] 当 keyed columns 解析并抛
    // "Invalid data provider"，原始值改由 formatter 按 dataIndex 回查
    series: [
      buildSeries('上传', COLOR.upload, '75,111,246',
          list.map(i => scale(i.upload, divisor))),
      buildSeries('下载', COLOR.download, '16,185,129',
          list.map(i => scale(i.download, divisor)))
    ]
  };
};

const buildAccessOption = () => {
  const list = sorted.value;
  const zoom = list.length > 20 ? [{
    type: 'inside',
    realtime: true,
    start: 0,
    end: 100,
    zoomOnMouseWheel: false
  }] : [];

  const formatter = (params) => {
    const arr = Array.isArray(params) ? params : [params];
    if (!arr.length) return '';
    const raw = list[arr[0].dataIndex];
    let html = '<div style="font-size:12px;color:#64748b;margin-bottom:6px">' +
        formatFull(raw.time) + '</div>';
    arr.forEach(p => {
      const people = p.seriesName === 'PV' ? raw.pv : raw.uv;
      html += '<div style="display:flex;align-items:center;gap:8px;margin-top:4px">' +
          '<span style="width:8px;height:8px;border-radius:50%;background:' + p.color + '"></span>' +
          '<span style="flex:1;color:#64748b">' + p.seriesName + '</span>' +
          '<strong style="color:#1b2437">' + people + ' 人</strong>' +
          '</div>';
    });
    return html;
  };

  return {
    grid: {left: 4, right: 16, top: 40, bottom: 8, containLabel: true},
    legend: {...legendCommon, data: ['PV', 'UV']},
    tooltip: {...tooltipCommon, formatter},
    dataZoom: zoom,
    xAxis: {type: 'category', boundaryGap: false, data: axisLabels.value, ...axisCommon,
      axisLabel: {...axisCommon.axisLabel, hideOverlap: true}},
    yAxis: {
      type: 'value',
      name: '人数',
      nameTextStyle: {color: '#94a3b8', fontSize: 12, align: 'left'},
      minInterval: 1,
      ...axisCommon,
      splitLine: {lineStyle: {color: '#f1f4f9', type: 'dashed'}}
    },
    series: [
      buildSeries('PV', COLOR.pv, '75,111,246', list.map(i => Number(i.pv) || 0)),
      buildSeries('UV', COLOR.uv, '245,158,11', list.map(i => Number(i.uv) || 0))
    ]
  };
};

const setupObserver = () => {
  if (resizeObserver || typeof ResizeObserver === 'undefined') return;
  resizeObserver = new ResizeObserver(() => {
    // 组件挂在 destroy-on-close 的弹窗里，首次挂载时容器宽度可能还是 0
    flowChart && flowChart.resize();
    accessChart && accessChart.resize();
  });
  if (flowRef.value) resizeObserver.observe(flowRef.value);
  if (accessRef.value) resizeObserver.observe(accessRef.value);
};

const renderAll = async () => {
  await nextTick();
  if (!hasData.value) return;

  if (flowRef.value) {
    // 复用实例，不要每次 watch 都 init（原来那样会不断泄漏实例）
    if (!flowChart || flowChart.isDisposed()) flowChart = echarts.init(flowRef.value);
    flowChart.setOption(buildFlowOption(), true);
  }
  if (accessRef.value) {
    if (!accessChart || accessChart.isDisposed()) accessChart = echarts.init(accessRef.value);
    accessChart.setOption(buildAccessOption(), true);
  }
  setupObserver();
};

onMounted(renderAll);

watch(() => [props.value, props.name], renderAll, {deep: true});

onBeforeUnmount(() => {
  resizeObserver && resizeObserver.disconnect();
  resizeObserver = null;
  flowChart && flowChart.dispose();
  accessChart && accessChart.dispose();
  flowChart = null;
  accessChart = null;
});
</script>

<style scoped>
.mc {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.mc__stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
}

.mc__stat {
  background: #f8fafc;
  border: 1px solid var(--hp-border);
  border-radius: 12px;
  padding: 12px 14px;
}

.mc__stat-label {
  font-size: 12px;
  color: var(--hp-text-3);
}

.mc__stat-value {
  margin-top: 2px;
  font-size: 18px;
  font-weight: 600;
  color: var(--hp-text);
}

.mc__stat-value--up {
  color: #2f5bef;
}

.mc__stat-value--down {
  color: #059669;
}

.mc__stat-unit {
  font-size: 12px;
  font-weight: 400;
  color: var(--hp-text-3);
}

.mc__card {
  background: #fff;
  border: 1px solid var(--hp-border);
  border-radius: 12px;
  padding: 12px 14px 6px;
}

.mc__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding: 0 4px;
}

.mc__title {
  font-size: 14px;
  font-weight: 600;
}

.mc__unit {
  font-size: 12px;
  color: var(--hp-text-3);
}

.mc__canvas {
  width: 100%;
  height: 260px;
}

.mc__foot {
  font-size: 12px;
  color: var(--hp-text-3);
  text-align: center;
}

.mc__empty {
  border: 1px dashed var(--hp-border);
  border-radius: 12px;
  padding: 48px 0;
  text-align: center;
}

.mc__empty-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--hp-text-2);
}

.mc__empty-desc {
  margin-top: 4px;
  font-size: 12px;
  color: var(--hp-text-3);
}

/* 弹窗是 width:90%，窄屏下四列会挤 */
@media (max-width: 720px) {
  .mc__stats {
    grid-template-columns: repeat(2, 1fr);
  }

  .mc__canvas {
    height: 220px;
  }
}
</style>
