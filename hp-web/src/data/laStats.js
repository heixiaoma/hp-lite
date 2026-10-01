/**
 * 51.LA 统计数据读取
 *
 * 背景：51.LA 官方对外只有「采集 SDK」（index.html 里那段 sdk.51.la）。
 * 展示侧官方提供的是「数据挂件」：https://v6-widget.51.la/v6/{站点ID}/quote.js
 * 它会自己往页面里塞 DOM，颜色/字号全部是内联 !important，基本改不动。
 *
 * 但 quote.js 是服务端按请求实时生成的，统计数值以
 *     <p><span>{序号}</span><span>{数值}</span></p>
 * 的形式直接写在 JS 源码里，且响应头带 Access-Control-Allow-Origin: *，
 * 所以这里 fetch 原文 → 正则提取 → 交给组件自己渲染，样式 100% 可控。
 */

const LA_SITE_ID = '3NkyaLWSCGchFoDV';
const WIDGET_URL = `https://v6-widget.51.la/v6/${LA_SITE_ID}/quote.js`;

// 序号 → 含义。与官方 quote.js 里的 b / N 两个常量一一对应，不要随意调整顺序
export const LA_FIELDS = [
  {key: 'online', label: '最近活跃'},
  {key: 'todayUv', label: '今日访客'},
  {key: 'todayPv', label: '今日访问'},
  {key: 'yesterdayUv', label: '昨日访客'},
  {key: 'yesterdayPv', label: '昨日访问'},
  {key: 'monthPv', label: '本月访问'},
  {key: 'totalPv', label: '总访问量'},
];

const ITEM_RE = /<p><span>(\d+)<\/span><span>([^<]+)<\/span><\/p>/g;

// 同页面反复进出首页时不重复请求
const CACHE_TTL = 5 * 60 * 1000;
let cache = null;

function formatNum(value) {
  const raw = String(value).trim();
  const n = Number(raw.replace(/[^\d.]/g, ''));
  if (!Number.isFinite(n)) return raw;
  // 与官方一致：过亿折算成「亿」
  if (n >= 1e8) return (n / 1e8).toFixed(2) + '亿';
  return n.toLocaleString('en-US');
}

export async function fetchLaStats() {
  if (cache && Date.now() - cache.at < CACHE_TTL) return cache.items;

  const res = await fetch(WIDGET_URL, {cache: 'no-store'});
  if (!res.ok) throw new Error('51.LA widget HTTP ' + res.status);

  const text = await res.text();
  const values = new Map();
  let m;
  ITEM_RE.lastIndex = 0;
  while ((m = ITEM_RE.exec(text)) !== null) {
    values.set(Number(m[1]), m[2]);
  }
  if (values.size === 0) throw new Error('51.LA widget 数据解析为空');

  // raw = 原始数值（组件用它做数字滚动），value = 最终展示文案
  const items = LA_FIELDS.map((f, i) => {
    const n = Number(String(values.get(i) ?? 0).replace(/[^\d.]/g, '')) || 0;
    return {...f, raw: n, value: formatNum(n)};
  });

  // 全部为 0 视为「没有数据」（统计未启用 / 全新站点）：返回空数组，由调用方整块隐藏。
  // 空结果不进缓存，避免刚启用统计时 5 分钟内一直不显示
  if (!items.some(it => it.raw > 0)) return [];

  cache = {at: Date.now(), items};
  return items;
}
