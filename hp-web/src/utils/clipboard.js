/**
 * 复制文本到剪贴板，成功返回 true。
 *
 * navigator.clipboard 只在安全上下文（HTTPS 或 localhost）可用，
 * 而后台管理页常被部署在 http 域名下，所以要留 execCommand 兜底。
 * 提示消息交给调用方，这里只关心有没有复制成功。
 */
export const copyText = async (text) => {
  if (!text) return false
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
    } else {
      const ta = document.createElement('textarea')
      ta.value = text
      ta.style.position = 'fixed'
      ta.style.opacity = '0'
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
    }
    return true
  } catch (e) {
    return false
  }
}
