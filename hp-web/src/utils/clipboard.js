import {h} from 'vue'
import {DialogPlugin, MessagePlugin} from 'tdesign-vue-next'

/**
 * 把文本写进剪贴板，返回是否成功。
 *
 * navigator.clipboard 需要安全上下文（HTTPS / localhost）。在 http 域名下
 * 它整个消失（不是报错，是 undefined），只能退回 execCommand。
 */
export const writeClipboard = async (text) => {
  const value = text == null ? '' : String(text)
  if (!value) return false

  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(value)
      return true
    } catch (e) {
      // 权限被拒、页面失焦时会抛。不直接放弃，继续走兜底
    }
  }

  return legacyCopy(value)
}

/* execCommand 兜底。相比「透明 textarea + select()」的写法，这里关键是：
   1. 检查 execCommand 的返回值 —— 它失败时返回 false。
      原实现无条件 return true，复制失败也会弹「已复制」，用户粘贴出来是空的
   2. 用「移出屏幕」代替 opacity: 0，部分浏览器不会选中透明元素
   3. setSelectionRange 兜住 iOS Safari —— 它对 textarea 的 select() 无效
   4. 复制完恢复用户原本的选区，不破坏页面已有的选中状态
   5. 全程同步：中间不能有 await，否则丢失用户手势，execCommand 会被浏览器拒绝 */
const legacyCopy = (value) => {
  const ta = document.createElement('textarea')
  ta.value = value
  ta.setAttribute('readonly', '')
  ta.style.position = 'absolute'
  ta.style.left = '-9999px'
  ta.style.top = `${window.pageYOffset || document.documentElement.scrollTop}px`

  document.body.appendChild(ta)

  const selection = document.getSelection()
  const previousRange = selection && selection.rangeCount > 0 ? selection.getRangeAt(0) : null

  let ok = false
  try {
    ta.select()
    ta.setSelectionRange(0, value.length)
    ok = document.execCommand('copy')
  } catch (e) {
    ok = false
  }

  document.body.removeChild(ta)

  if (selection && previousRange) {
    selection.removeAllRanges()
    selection.addRange(previousRange)
  }

  return ok
}

/* 自动复制全部失败时的最后兜底。
   这些值（设备编号 / 连接码）都是长串，表格里被省略号截断，
   不弹出来的话用户根本选不出完整内容 */
const manualCopyDialog = (value, title) => {
  DialogPlugin({
    header: title || '手动复制',
    width: '560px',
    // TDesign 的 cancelBtn 传 null / false / '' 都拦不住，会退回默认的「取消」。
    // 两个按钮都能关闭弹窗，只是「取消」在这个场景语义偏弱，去不掉
    confirmBtn: '关闭',
    body: () => h('div', null, [
      h('p', {
        style: {
          margin: '0 0 10px',
          fontSize: '12px',
          lineHeight: '1.6',
          color: '#94a3b8',
        },
      }, '自动复制失败。下方内容已全选，按 Ctrl / ⌘ + C 即可复制。'),
      h('textarea', {
        readOnly: true,
        value,
        onClick: (e) => e.target.select(),
        style: {
          width: '100%',
          minHeight: '96px',
          padding: '10px 12px',
          boxSizing: 'border-box',
          border: '1px solid #e2e8f0',
          borderRadius: '10px',
          background: '#f8fafc',
          fontFamily: "'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace",
          fontSize: '13px',
          lineHeight: '1.6',
          resize: 'vertical',
          outline: 'none',
        },
        // 弹窗挂载后再聚焦全选，直接 focus 会被 dialog 的入场动画抢走焦点
        ref: (el) => {
          if (el) setTimeout(() => {
            el.focus()
            el.select()
          }, 80)
        },
      }),
    ]),
  })
}

/**
 * 复制文本，自带提示与兜底：成功提示一次，失败弹出手动复制框。
 * @returns {Promise<boolean>} 是否自动复制成功
 */
export const copyText = async (text, label = '内容') => {
  const ok = await writeClipboard(text)
  if (ok) {
    MessagePlugin.success('已复制到剪贴板')
    return true
  }
  MessagePlugin.warning('自动复制失败，请手动复制')
  manualCopyDialog(text == null ? '' : String(text), `复制${label}`)
  return false
}
