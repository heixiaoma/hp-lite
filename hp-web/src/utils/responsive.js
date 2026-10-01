import {computed, onMounted, onUnmounted, ref} from 'vue'

/* 全站共用一个窗口宽度 ref：每个页面各自 addEventListener 会白养十几个监听器，
   而且首屏时各自的初始值还可能不一致 */
const MOBILE_BP = 768

const vw = ref(typeof window === 'undefined' ? 1280 : window.innerWidth)
let users = 0

const onResize = () => {
  vw.value = window.innerWidth
}

/**
 * 视口宽度小于断点时返回 true。
 * 用法：const isMobile = useIsMobile(); 模板里直接判断即可。
 */
export function useIsMobile(bp = MOBILE_BP) {
  onMounted(() => {
    if (users++ === 0) window.addEventListener('resize', onResize)
    onResize()
  })
  onUnmounted(() => {
    if (--users === 0) window.removeEventListener('resize', onResize)
  })
  return computed(() => vw.value < bp)
}

/**
 * 表格列在手机上的取舍：标了 mobile: false 的列会在窄屏隐藏，
 * 只保留「能认出这一行」和「能操作这一行」的列。
 * 剩下的宽度万一还是不够，外层有横向滚动兜底，不会出现内容被裁掉又滚不到的情况。
 */
export function useResponsiveColumns(allColumns) {
  const isMobile = useIsMobile()
  return computed(() => {
    if (!isMobile.value) return allColumns
    // 窄屏双管齐下：去掉次要列，并且把剩下的列写死宽度也去掉，
    // 这样浏览器按容器重新分配列宽、单元格换行，多数页面根本不用横向滚
    return allColumns
        .filter((c) => c.mobile !== false)
        .map(({width, ...rest}) => rest)
  })
}
