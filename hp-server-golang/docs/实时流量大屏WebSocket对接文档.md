# 实时流量大屏 WebSocket 对接文档

适用服务：`hp-server-golang`（管理端 HTTP 端口，取自 `app.yml` 的 web 端口）

| 入口 | 地址 | 说明 |
|---|---|---|
| WebSocket 推送 | `ws://{host}:{port}/client/ws/stats?token=xxx` | 主链路，实时推流 |
| HTTP 轮询兜底 | `GET /client/ws/stats/summary`（Header: `token`） | 不想用 ws 时每秒拉一次汇总 |

> WebSocket 入口**不在** `middleware.Auth` 之后，鉴权在 controller 内部完成（因为浏览器 WebSocket API 无法自定义请求头）。
> token 优先取 query 参数 `token`，取不到再读 Header `token`；token 非法/过期直接在握手阶段返回 JSON 错误。

---

## 1. 鉴权与握手

- token 来源：`util.DecodeToken` 解析出 `userId / role / 签发时间戳`
- 过期校验：`now - 签发时间戳 > TokenExpireMs` 即判定过期，需重新登录后再握手
- 握手失败返回（HTTP 200 + JSON，**不是** ws 帧）：

```json
{ "code": -2, "msg": "token已过期，请重新登录", "data": null }
```

---

## 2. 统一消息信封

**服务端下发的所有 ws 帧**都是这一个结构：

```json
{ "type": "summary", "ts": 1759300000000, "data": {} }
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `type` | string | `welcome` / `stat` / `summary` / `pong` / `subscribe` / `error` |
| `ts` | int64 | 服务端生成时间戳（毫秒） |
| `data` | object \| array \| null | 载荷，随 `type` 变化 |

**前端上行**帧（唯一结构）：

```json
{ "type": "ping" }
{ "type": "subscribe", "filter": { "domain": "example.com", "statusClass": "4xx" } }
```

服务端心跳：每 **25s** 主动发一个 ws Ping；读超时 **75s**。建议前端被动接收即可，主动 `ping` 只是可选的心跳。

---

## 3. 下行消息一览

| type | 触发时机 | 频率 | data 结构 |
|---|---|---|---|
| `welcome` | 握手成功后立即下发一次 | 每条连接 1 次 | `WelcomeData` |
| `stat` | 每个 HTTP 请求处理完成时立即下发 | 随流量（每秒 N 条） | `HttpStatRecord` |
| `summary` | 后台定时器 | **每秒 1 次**（仅当有在线连接） | `SummaryData` |
| `pong` | 收到上行 `ping` | 按需 | `null` |
| `subscribe` | 收到上行 `subscribe` 后回执 | 按需 | `WsFilter` |
| `error` | 异常场景 | 按需 | 字符串 |

### 3.1 welcome（首帧快照）

```json
{
  "type": "welcome",
  "ts": 1759300000000,
  "data": {
    "clientId": "m1k2n3p-1",
    "summary": { "...SummaryData 完整结构，见 3.3..." }
  }
}
```

拿到这帧就可以立刻渲染大盘，不用等第一秒 `summary`。

### 3.2 stat（单条请求明细）

**注意：这条是事件驱动的，不是定时推送，QPS 越高每秒条数越多。**

```json
{
  "type": "stat",
  "ts": 1759300000123,
  "data": {
    "id": "m1k2n3q-9",
    "ts": 1759300000123,
    "time": "2026-10-01 12:00:00.123",

    "sourceIp": "112.80.248.75",
    "sourcePort": "54321",
    "ipVersion": "IPv4",
    "sourceCountry": "中国",
    "sourceCountryCode": "CN",
    "sourceProvince": "江苏省",
    "sourceCity": "南京市",
    "sourceIsp": "联通",

    "targetIp": "192.168.1.10",
    "targetPort": "8080",

    "domain": "example.com",
    "hostPort": "443",
    "method": "GET",
    "scheme": "https",
    "protocol": "HTTP/2.0",
    "path": "/api/user/info",
    "query": "id=1",
    "uri": "/api/user/info?id=1",
    "proxyType": "reverse",
    "configId": 0,

    "statusCode": 200,
    "statusText": "OK",
    "statusClass": "2xx",

    "requestSize": 0,
    "requestHeaderSize": 428,
    "responseSize": 1536,
    "responseHeaderSize": 236,
    "totalSize": 2200,
    "durationMs": 37,

    "contentType": "application/json; charset=utf-8",
    "requestContentType": "",
    "userAgent": "Mozilla/5.0 ...",
    "referer": "https://example.com/",
    "browser": "Chrome",
    "os": "Windows",
    "device": "PC",

    "tlsVersion": "TLS 1.3",
    "tlsCipher": "TLS_AES_128_GCM_SHA256"
  }
}
```

#### HttpStatRecord 字段说明

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | 请求唯一 ID |
| `ts` | int64 | 服务端时间戳（毫秒） |
| `time` | string | 已格式化时间，可直接展示 |
| `sourceIp` | string | 客户端 IP（`X-Forwarded-For` 优先） |
| `sourcePort` | string | 客户端端口，取不到为空串 |
| `ipVersion` | string | `IPv4` / `IPv6` |
| `sourceCountry` | string | 来源国家，如 `中国` / `United States`；未启用 ip2region 或库里查不到时为空，内网为 `局域网`、回环为 `本机` |
| `sourceCountryCode` | string | 两位国家代码，如 `CN` / `US`，前端可用于渲染国旗；v3 库才有，v2 库为空 |
| `sourceProvince` | string | 省份 / 州 |
| `sourceCity` | string | 城市 |
| `sourceIsp` | string | 运营商 / 组织，如 `联通` / `Google LLC` |
| `targetIp` | string | 后端（被代理）IP |
| `targetPort` | string | 后端端口 |
| `domain` | string | Host 去掉端口 |
| `hostPort` | string | 入口监听端口，如 `80` / `443` |
| `method` | string | `GET` / `POST` ... |
| `scheme` | string | `http` / `https` |
| `protocol` | string | `HTTP/1.1` / `HTTP/2.0` |
| `path` | string | 路径部分（不含 query） |
| `query` | string | 查询串（不含 `?`） |
| `uri` | string | path + query |
| `proxyType` | string | `tunnel` 内网穿透 / `reverse` 反向代理 / `acme` 证书校验 / `unknown` |
| `configId` | int | 命中的穿透配置 ID，仅 `tunnel` 链路有值，其余为 `0` |
| `statusCode` | int | HTTP 状态码 |
| `statusText` | string | 状态码文案，如 `OK` / `Not Found` |
| `statusClass` | string | `1xx` / `2xx` / `3xx` / `4xx` / `5xx` |
| `requestSize` | int64 | 请求体字节数 |
| `requestHeaderSize` | int64 | 请求头估算字节数 |
| `responseSize` | int64 | 响应体字节数 |
| `responseHeaderSize` | int64 | 响应头估算字节数 |
| `totalSize` | int64 | 上述四项之和 |
| `durationMs` | int64 | 后端处理耗时（毫秒） |
| `contentType` | string | 响应 `Content-Type` |
| `requestContentType` | string | 请求 `Content-Type` |
| `userAgent` | string | 原始 UA |
| `referer` | string | 来源页 |
| `browser` | string | UA 解析出的浏览器 |
| `os` | string | UA 解析出的操作系统 |
| `device` | string | `PC` / `Mobile` / `Tablet` / `Bot` |
| `tlsVersion` | string | HTTPS 才有值，如 `TLS 1.3` |
| `tlsCipher` | string | HTTPS 才有值 |

### 3.3 summary（每秒汇总）

```json
{
  "type": "summary",
  "ts": 1759300001000,
  "data": {
    "onlineClients": 2,
    "totalRequests": 158423,
    "totalBytes": 982731264,
    "totalRequestSize": 102388224,
    "totalResponseSize": 880343040,
    "qps": 2640,
    "avgDurationMs": 42,
    "statusCounts": { "2xx": 150000, "3xx": 3200, "4xx": 4800, "5xx": 423 },
    "topCountries": [{ "key": "中国", "count": 120000 }, { "key": "United States", "count": 20431 }],
    "topDomains": [{ "key": "example.com", "count": 98231 }],
    "topIps":     [{ "key": "112.80.248.75", "count": 8321 }],
    "topPaths":   [{ "key": "/api/user/info", "count": 6634 }],
    "uptimeSec": 86400
  }
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `onlineClients` | int64 | 当前在线的大屏 ws 连接数 |
| `totalRequests` | int64 | 累计请求数（自**第一个大屏连上**开始计数） |
| `totalBytes` | int64 | 累计总流量 = 上行 + 下行 |
| `totalRequestSize` | int64 | 累计上行字节 |
| `totalResponseSize` | int64 | 累计下行字节 |
| `qps` | int64 | **最近 60 秒滑动窗口的平均 QPS**（整数除法，低流量时可能为 0） |
| `avgDurationMs` | int64 | 全量平均耗时 |
| `statusCounts` | map | 状态码分类分布，key 为 `1xx`~`5xx`，缺失的类别即无数据 |
| `topCountries` | TopItem[] | 来源国家 Top10（来自 ip2region，未启用时为空数组） |
| `topDomains` | TopItem[] | 域名 Top10，按 count 降序 |
| `topIps` | TopItem[] | 来源 IP Top10 |
| `topPaths` | TopItem[] | 路径 Top10 |
| `uptimeSec` | int64 | ws 服务启动至今的秒数 |

`TopItem`：`{ "key": string, "count": int64 }`

### 3.4 pong / subscribe

```json
{ "type": "pong", "ts": 1759300005000, "data": null }
```

```json
{
  "type": "subscribe",
  "ts": 1759300005000,
  "data": { "domain": "example.com", "sourceIp": "", "statusClass": "4xx", "proxyType": "" }
}
```

回执内容是**服务端最终生效的过滤条件**，前端可用来确认设置成功。

---

## 4. 订阅过滤（建议大屏必用）

上行：

```json
{ "type": "subscribe", "filter": { "statusClass": "4xx" } }
```

| 字段 | 类型 | 匹配规则 |
|---|---|---|
| `domain` | string | 包含匹配 `strings.Contains`，空 = 不限 |
| `sourceIp` | string | 包含匹配，空 = 不限 |
| `statusClass` | string | **精确匹配**，如 `4xx`，空 = 不限 |
| `proxyType` | string | **精确匹配**，如 `tunnel`，空 = 不限 |

留空 `{}` 即恢复为"接收全部"。过滤只作用于 `stat` 明细，`summary` 是全量大盘，不受影响。

---

## 5. HTTP 兜底接口

```
GET /client/ws/stats/summary
Header: token: xxx
```

返回（走统一的 `ResData` 信封）：

```json
{
  "code": 200,
  "msg": "操作成功",
  "data": { "...SummaryData，见 3.3..." }
}
```

建议轮询间隔 1s，与 ws 的 `summary` 频率一致。

---

## 6. 对接必须知道的 4 个行为约定

1. **无人在线时不统计**：数据面拦截器先判断 `HasWsClients()`，没人连 ws 就直接透传，零额外开销。因此 `totalRequests` 等指标是从**第一个大屏连上那一刻**开始累计的，服务重启或所有连接断开重连后会重新计数。
2. **`stat` 逐条实时推送**：不走批处理。建议前端做**节流渲染**（例如本地缓冲，每 200ms 刷一次列表 DOM），不要每条都触发 React 重渲染。
3. **发送队列 1024，满了会被断开**：单个连接待发队列上限 `wsSendBuffer = 1024`。如果前端消费跟不上导致队列写满，服务端会记录错误日志并**主动断开该连接**，不会阻塞代理主流程。所以务必：加订阅过滤 + 限制本地列表长度（如只保留最近 200 条）+ 节流渲染。被断开后请做**指数退避重连**（建议 1s / 2s / 4s，上限 30s）。
4. **`qps` 是 60 秒平均不是瞬时值**：最近 60 秒总请求不足 60 时会算出 `0`。需要瞬时速率，请用自己收到的 `stat` 条数在本地按秒计数。

---

## 7. 前端接入示例

```js
let retry = 1000
let ws

function connect() {
  ws = new WebSocket(`ws://${location.host}/client/ws/stats?token=${token}`)

  ws.onopen = () => {
    retry = 1000
    // 只要异常请求，降低数据量
    ws.send(JSON.stringify({ type: 'subscribe', filter: { statusClass: '4xx' } }))
  }

  ws.onmessage = (e) => {
    const msg = JSON.parse(e.data)
    switch (msg.type) {
      case 'welcome':
        renderSummary(msg.data.summary)   // 首帧快照，立即渲染大盘
        clientId = msg.data.clientId
        break
      case 'summary':
        renderSummary(msg.data)           // 每秒一次
        break
      case 'stat':
        buffer.push(msg.data)             // 高频，入缓冲后节流渲染，别直接 setState
        break
      case 'pong':
      case 'subscribe':
        break
      case 'error':
        console.error(msg.data)
        break
    }
  }

  ws.onclose = () => {
    setTimeout(connect, retry)            // 指数退避
    retry = Math.min(retry * 2, 30000)
  }

  ws.onerror = () => ws.close()
}

connect()

// stat 缓冲每 200ms 刷一次 UI，列表只保留最近 200 条
setInterval(() => {
  if (buffer.length) {
    renderList(buffer.splice(0, buffer.length).slice(-200))
  }
}, 200)
```

---

## 8. 服务端实现位置（便于联调）

| 文件 | 作用 |
|---|---|
| `web/web_server.go:103-108` | 路由注册（ws 入口 + HTTP 兜底） |
| `web/controller/ws_stats.go` | 连接管理、上行控制消息、定时推送、汇总计算 |
| `net/http/stats_interceptor.go` | 数据面埋点，`BroadcastHttpStat` 调用点 |
| `util/ip2region.go` | IP 归属地查询（ip2region xdb） |

---

## 9. IP 归属地（ip2region）

归属地字段（`sourceCountry` / `sourceCountryCode` / `sourceProvince` / `sourceCity` / `sourceIsp`）由服务端 ip2region 库解析，**未启用时这些字段为空字符串**，不影响其它字段。

### 启用方式

数据文件已**内嵌进程序**，无需任何配置、无需外部文件：

- `web/static/data/ip2region_v4.xdb`（IPv4 库，约 10MB）
- `web/static/data/ip2region_v6.xdb`（IPv6 库，约 35MB，可选，不存在会自动跳过）

通过 `web/web_server.go` 的 `//go:embed static` 打包，启动时由 `util.InitIp2RegionFromFS(web.StaticFS(), ...)` 载入内存。
服务启动日志出现 `ip2region 加载成功：static/data/ip2region_v4.xdb|版本:IPv4` 即生效。

更新库文件：替换 `web/static/data/` 下的 xdb 后重新编译即可（官方仓库 `lionsoul2014/ip2region` 的 `data/` 目录）。

### 行为约定

- 整库载入内存（零 IO 查询）+ IP 结果缓存（上限 10 万条，满了整体清空），单次查询微秒级，不阻塞代理主流程。
- `127.0.0.1` 等国家/省份/城市统一为 `本机`；`10.x` / `172.16.x` / `192.168.x` 等内网地址为 `局域网`，不查库。
- 未内嵌 IPv6 库时，IPv6 来源的归属地为空（启动日志会提示"未内嵌数据文件，跳过"）。
- 库缺失或格式错误只打一条错误日志，不影响服务启动。
- 注意：xdb 放在 `web/static/data/` 下会随静态目录一起被 `StaticController` 对外提供（即 `/data/ip2region_v4.xdb` 可被匿名下载）。若不想暴露，把文件移到非静态目录（例如 `web/data/`）并单独加一个 `//go:embed data`，再改 `main.go` 里的路径即可。
