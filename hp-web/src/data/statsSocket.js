import {baseURL} from './http'
import userInfo from './userInfo.js'

// http(s) 与 ws(s) 同源，直接换协议头
const WS_BASE = baseURL.replace(/^http/, 'ws')
const WS_PATH = '/client/ws/stats'

/* 服务端下发队列上限 1024，前端消费跟不上会被主动断开连接。
   所以本地缓冲 + 节流批量交给 UI，并且只保留最近 MAX_ROWS 条 —— 多出来的直接丢，
   绝不因为渲染慢而积压 */
const FLUSH_MS = 200
const MAX_ROWS = 200

// 指数退避：1s / 2s / 4s ... 上限 30s（文档 §6.3）
const RETRY_MIN = 1000
const RETRY_MAX = 30000

// ws 断开期间的 HTTP 兜底轮询，间隔与 summary 推送频率一致（文档 §5）
const POLL_MS = 1000

/**
 * 连接实时流量大屏 ws。
 *
 * @param {object}   opts
 * @param {Function} opts.onSummary 每秒汇总，或握手后的首帧快照
 * @param {Function} opts.onStats   节流后的 stat 明细批次（数组）
 * @param {Function} opts.onStatus  'connecting' | 'open' | 'closed' | 'error'
 * @returns {{ setFilter, dispose }}
 */
export const createStatsSocket = ({onSummary, onStats, onStatus} = {}) => {
  let ws = null
  let disposed = false
  let retry = RETRY_MIN
  let buffer = []
  let flushTimer = null
  let pollTimer = null
  let reconnectTimer = null
  let filter = {}

  const setStatus = (s, detail) => onStatus && onStatus(s, detail)

  const flush = () => {
    if (!buffer.length) return
    const batch = buffer
    buffer = []
    onStats && onStats(batch)
  }

  const startFlush = () => {
    if (!flushTimer) flushTimer = setInterval(flush, FLUSH_MS)
  }

  const stopFlush = () => {
    clearInterval(flushTimer)
    flushTimer = null
    flush()
  }

  /* 用 fetch 而不是全局 axios 实例：axios 的响应拦截器对非 200 会弹 NotifyPlugin.error，
     而 HTTP 兜底在 ws 连不上时会每秒失败一次，那样会疯狂弹窗 */
  const poll = async () => {
    try {
      const info = userInfo.getUserInfo()
      if (!info || !info.token) return
      const res = await fetch(`${baseURL}/client/ws/stats/summary`, {
        headers: {token: info.token},
      })
      const json = await res.json()
      if (!disposed && json && json.code === 200 && json.data) {
        onSummary && onSummary(json.data)
      }
    } catch (e) {
      // 兜底轮询的失败不打扰用户，等 ws 重连上就停掉
    }
  }

  const startPolling = () => {
    if (pollTimer) return
    poll()
    pollTimer = setInterval(poll, POLL_MS)
  }

  const stopPolling = () => {
    clearInterval(pollTimer)
    pollTimer = null
  }

  const sendFilter = () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({type: 'subscribe', filter}))
    }
  }

  const connect = () => {
    if (disposed) return
    const info = userInfo.getUserInfo()
    const token = info && info.token
    if (!token) {
      setStatus('error', '未获取到登录 token，请重新登录')
      return
    }

    setStatus('connecting')
    const url = `${WS_BASE}${WS_PATH}?token=${encodeURIComponent(token)}`

    let sock
    try {
      sock = new WebSocket(url)
    } catch (e) {
      scheduleReconnect()
      return
    }
    ws = sock

    sock.onopen = () => {
      retry = RETRY_MIN
      setStatus('open')
      stopPolling()
      startFlush()
      sendFilter()
    }

    sock.onmessage = (e) => {
      let msg
      try {
        msg = JSON.parse(e.data)
      } catch (err) {
        return
      }
      if (!msg || typeof msg !== 'object') return

      // 握手/鉴权失败时服务端给的是错误信封，不是 ws 帧
      if (typeof msg.code === 'number' && msg.code !== 200) {
        setStatus('error', msg.msg || '连接被拒绝')
        try {
          sock.close()
        } catch (err) { /* ignore */ }
        return
      }

      switch (msg.type) {
        case 'welcome':
          // 首帧快照，拿到就能立刻渲染大盘，不用等第一秒 summary
          onSummary && onSummary(msg.data && msg.data.summary)
          break
        case 'summary':
          onSummary && onSummary(msg.data)
          break
        case 'stat':
          buffer.push(msg.data)
          if (buffer.length > MAX_ROWS) {
            buffer.splice(0, buffer.length - MAX_ROWS)
          }
          break
        case 'error':
          setStatus('error', msg.data)
          break
        default:
          break // pong / subscribe 回执，无需处理
      }
    }

    sock.onclose = () => {
      if (ws === sock) ws = null
      stopFlush()
      scheduleReconnect()
    }

    sock.onerror = () => {
      // 统一交给 onclose 走重连
      try {
        sock.close()
      } catch (err) { /* ignore */ }
    }
  }

  const scheduleReconnect = () => {
    if (disposed) return
    setStatus('closed')
    startPolling()
    clearTimeout(reconnectTimer)
    reconnectTimer = setTimeout(connect, retry)
    retry = Math.min(retry * 2, RETRY_MAX)
  }

  connect()

  return {
    /** 订阅过滤，只作用于 stat 明细；summary 是全量大盘不受影响（文档 §4） */
    setFilter(next) {
      filter = next || {}
      sendFilter()
    },
    dispose() {
      disposed = true
      clearTimeout(reconnectTimer)
      stopFlush()
      stopPolling()
      if (ws) {
        try {
          ws.close()
        } catch (e) { /* ignore */ }
        ws = null
      }
    },
  }
}
