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
    "sourceIsp": "Alibaba (US) Technology Co., Ltd.",
    "sourceLatitude": 32.0617,
    "sourceLongitude": 118.7778,

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
| `sourceCountry` | string | 来源国家，如 `中国` / `United States`；未启用 GeoLite2 或库里查不到时为空，内网为 `局域网`、回环为 `本机` |
| `sourceCountryCode` | string | 两位国家代码，如 `CN` / `US`，前端可用于渲染国旗 |
| `sourceProvince` | string | 省份 / 州 |
| `sourceCity` | string | 城市 |
| `sourceIsp` | string | 运营商 / 自治域组织，如 `Alibaba (US) Technology Co., Ltd.` / `Google LLC`；**需要额外配 GeoLite2-ASN 库才有值**，否则恒为空 |
| `sourceLatitude` | float64 | **纬度**，地图打点用；`0` 表示这跳没定位到（见 [§8](#8-ip-归属地geolite2)） |
| `sourceLongitude` | float64 | **经度**，同为 `0` 时表示没定位到 |
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
    "topIps":     [{ "key": "112.80.248.75", "count": 8321, "country": "中国", "countryCode": "CN", "province": "江苏省", "city": "南京市", "isp": "Alibaba (US) Technology Co., Ltd.", "latitude": 30.2943, "longitude": 120.1663 }],
    "topPaths":   [{ "key": "/api/user/info", "count": 6634 }],
    "uptimeSec": 86400
  }
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `onlineClients` | int64 | 当前在线的大屏 ws 连接数 |
| `totalRequests` | int64 | 累计请求数（自**服务启动**开始计数） |
| `totalBytes` | int64 | 累计总流量 = 上行 + 下行 |
| `totalRequestSize` | int64 | 累计上行字节 |
| `totalResponseSize` | int64 | 累计下行字节 |
| `qps` | int64 | **最近 60 秒滑动窗口的平均 QPS**（整数除法，低流量时可能为 0） |
| `avgDurationMs` | int64 | 全量平均耗时 |
| `statusCounts` | map | 状态码分类分布（**全量累计**），key 为 `1xx`~`5xx`，缺失的类别即无数据 |
| `topCountries` | TopItem[] | 来源国家 Top10（来自 GeoLite2，未启用时为空数组） |
| `topDomains` | TopItem[] | 域名 Top10，按 count 降序 |
| `topIps` | TopIpItem[] | 来源 IP Top10，带归属地 |
| `topPaths` | TopItem[] | 路径 Top10（聚合 key 已做归一化，见下文） |
| `topDevices` | TopItem[] | 设备类型 Top10：`PC` / `Mobile` / `Tablet` / `Bot` |
| `topSystems` | TopItem[] | 操作系统 Top10，如 `Windows 10/11` / `iOS/macOS` / `Android` / `Linux` |
| `topBrowsers` | TopItem[] | 浏览器 Top10，如 `Chrome` / `Edge` / `Firefox` / `curl` |
| `topIsps` | TopItem[] | 运营商 / 自治域组织 Top10（需配 GeoLite2-ASN 库，否则空数组） |
| `wafBlocks` | object | 防火墙（coraza）拦截统计，结构见 [3.3](#33-wafblocks-防火墙拦截统计) |
| `uptimeSec` | int64 | ws 服务启动至今的秒数 |

> **统计口径**：所有 `top*` 榜单（含客户端画像四项，以及 `wafBlocks` 的三张 Top 榜）
> 都是**最近 5 小时滑动窗口**（60 个 5 分钟分桶，每 5 分钟向前推进一格），
> 窗口外的 key 自动淘汰，因此：
> - 榜单反映的是"近 5 小时谁在活跃"，不再是秒级当下的瞬时热点；
>   窗口刻意放这么长，是因为攻击/扫描流量往往稀疏 —— 短窗口每次只截到零星几次、
>   一滑就清零，"同一个 IP 连着扫了几小时"反而看不出来；
> - 热点退出是**随窗口平滑滑出**，不是定时清零，榜单不会出现瞬间掉到 0 的跳变；
> - 分桶粒度是 5 分钟，所以热点"消失"的速度也是以 5 分钟为台阶，不是秒级回落。
>
> 内存上限：单个维度窗口内 key 数超过 **2 万**时采样淘汰最小的 key 给新 key 让位。
> 5 小时窗口下 IP / 路径这类高基数维度在大流量站点容易顶到这个上限，
> 此时 Top 榜是有界近似 —— 新 key 仍能进榜，但低频长尾可能被提前挤掉。
>
> 与之相对，`totalRequests` / `totalBytes` / `totalRequestSize` / `totalResponseSize` /
> `statusCounts` / `avgDurationMs` 是**进程启动至今的全量累计**，不随窗口变化。
>
> **路径归一化**：`topPaths` 的 key 会把动态段替换成 `*`，避免带 ID 的 URL 打爆榜单：
> `/api/user/12345` → `/api/user/*`，`/f/<32位哈希>` → `/f/*`（纯数字段、UUID、
> 长度 ≥20 的十六进制串都会被替换）。**`stat` 明细里的 `path` 仍是原始完整路径**，不受影响。

`TopItem`：`{ "key": string, "count": int64 }`

`TopIpItem`（仅 `topIps`）：在 `TopItem` 基础上多 7 个归属地字段
`{ "key", "count", "country", "countryCode", "province", "city", "isp", "latitude", "longitude" }`
归属地来自该 IP 最近一次请求的解析结果；未启用 GeoLite2 时字符串字段为空、`latitude` / `longitude` 为 `0`。
前端拿经纬度画散点前记得先过滤掉 `latitude == 0 && longitude == 0` 的条目。

### 3.3 `wafBlocks` 防火墙拦截统计

反向代理与穿透代理各自挂了一份 coraza，命中 `SafeId` 对应规则时在错误回调里累计，
两边的数据**合并统计**，用 `source` 字段区分是哪一段拦下的。**只存聚合量，不存任何单条明细**。

```json
"wafBlocks": {
  "total": 1832,
  "source": { "reverse": 1200, "tunnel": 632 },
  "severity": { "critical": 1204, "error": 520, "warning": 108 },
  "topDomains": [{ "key": "example.com", "count": 900 }],
  "topIps":     [{ "key": "203.0.113.9", "count": 640 }],
  "topRules":   [{ "key": "949110", "count": 700 }],
  "lastAt": 1759300001000,
  "lastMsg": "穿透代理 · [critical] 规则949110 · Inbound Anomaly Score Exceeded · example.com · 203.0.113.9"
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `total` | int64 | 累计**规则命中次数**（一条请求命中多条规则会计多次），全量累计不清零 |
| `source` | map | 按挂载位置的累计分布：`reverse` = 反向代理，`tunnel` = 穿透代理 |
| `severity` | map | 按严重级别的累计分布，key 为 coraza 的 `emergency` / `alert` / `critical` / `error` / `warning` / `notice` / `info` / `debug` |
| `topDomains` | TopItem[] | 近 **5 小时**被拦最多的域名 Top10 |
| `topIps` | TopItem[] | 近 **5 小时**被拦最多的来源 IP Top10 |
| `topRules` | TopItem[] | 近 **5 小时**命中最多的规则 Top10，key 为规则 ID |
| `lastAt` | int64 | 最近一次命中的毫秒时间戳，`0` 表示从未命中 |
| `lastMsg` | string | 最近一次命中的摘要。**不含 URI / 请求体**，攻击载荷不会在内存里留存 |

> 口径：`total` / `source` / `severity` 是进程启动至今累计；三个 Top 榜是 **5 小时滑动窗口**
> （60 个 5 分钟分桶，与请求统计的 `top*` 同口径）。
>
> 窗口为什么要 5 小时？异常流量本来就稀疏 —— 扫描器常常隔几分钟才来一次，
> 短窗口每次只截到零星几次、一滑就清零，"同一个 IP 连着扫了三小时"
> 这种最该看见的信息反而看不见。放宽到 5 小时后慢速攻击也能攒够计数留在榜上，
> 而快轰型攻击照旧一眼可见。
>
> 内存上限托底：单个维度窗口内 key 数超过 **2 万**时采样淘汰最小的 key，
> 不会无界增长；代价是挤到上限时 IP 榜变成**有界近似** —— 极端分散的低频探测源可能被提前挤掉。
>
> 注意统计的是**规则命中**而非"被拦截的请求数"——coraza 的 error callback 每条命中规则都会触发一次，
> 包括未配置阻断动作的规则。想知道真正被拦了多少请求，得看同时段的 `5xx` / `403` 数量。
>
> 三个 Top 榜是两类代理**共用**的一张表，想单独看某一类请用 `source` 判断，
> 再叠加 `topDomains` 里的域名做交叉（域名本身就是按 `host` 分的，反代和隧道的域名不同）。

### 3.4 pong

```json
{ "type": "pong", "ts": 1759300005000, "data": null }
```

---

## 4. HTTP 兜底接口

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

## 5. 对接必须知道的 4 个行为约定

1. **统计与在线状态解耦**：数据面拦截器**始终**统计每一条请求（不管有没有人看），`totalRequests` 等累计值从**服务启动**开始算；大屏随时连上都能看到断连期间的累计数据。只有 `stat` 明细的序列化与推送在无人在线时跳过（省掉这部分开销）。服务重启后累计值归零（无持久化）。
2. **`stat` 逐条实时推送**：只在有连接时推送，且连接上之后才开始推当前请求（历史请求不会补推，历史数据看 `summary`）。建议前端做**节流渲染**（例如本地缓冲，每 200ms 刷一次列表 DOM），不要每条都触发 React 重渲染。
3. **发送队列 1024，满了会被断开**：单个连接待发队列上限 `wsSendBuffer = 1024`。如果前端消费跟不上导致队列写满，服务端会记录错误日志并**主动断开该连接**，不会阻塞代理主流程。服务端不做内容过滤（stat 明细全量下发），所以务必在本地：**限制列表长度**（如只保留最近 200 条）+ **节流渲染**（每 200ms 刷一次 DOM）。被断开后请做**指数退避重连**（建议 1s / 2s / 4s，上限 30s）。
4. **`qps` 是 60 秒平均不是瞬时值**：最近 60 秒总请求不足 60 时会算出 `0`。需要瞬时速率，请用自己收到的 `stat` 条数在本地按秒计数。

---

## 6. 前端接入示例

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

## 7. 服务端实现位置（便于联调）

| 文件 | 作用 |
|---|---|
| `web/web_server.go:103-108` | 路由注册（ws 入口 + HTTP 兜底） |
| `web/controller/ws_stats.go` | 连接管理、上行控制消息、定时推送、汇总计算 |
| `net/http/stats_interceptor.go` | 数据面埋点，`BroadcastHttpStat` 调用点 |
| `util/ipgeo.go` | IP 归属地查询（MaxMind GeoLite2 mmdb） |

---

## 8. IP 归属地（MaxMind GeoLite2）

归属地字段（`sourceCountry` / `sourceCountryCode` / `sourceProvince` / `sourceCity` / `sourceIsp` / `sourceLatitude` / `sourceLongitude`）
由服务端 [geoip2-golang](https://github.com/oschwald/geoip2-golang) 解析 MaxMind 的 mmdb 库得出，
**未启用时字符串字段为空、经纬度为 `0`**，不影响其它字段。

### 启用方式

数据文件已**内嵌进程序**，无需任何配置、无需外部文件：

| 文件 | 必需 | 提供字段 |
|---|---|---|
| `web/static/data/GeoLite2-City.mmdb` | 是（约 50MB） | 国家 / 国家代码 / 省 / 城市 + **经纬度**，IPv4 与 IPv6 通用 |
| `web/static/data/GeoLite2-ASN.mmdb` | 否（约 9MB） | 运营商 / 自治域组织（`sourceIsp`、`topIsps`） |

通过 `web/web_server.go` 的 `//go:embed static` 打包，启动时由
`util.InitIpGeoFromFS(web.StaticFS(), "static/data/GeoLite2-City.mmdb", "static/data/GeoLite2-ASN.mmdb")` 载入内存。
服务启动日志出现 `IP 归属地库加载成功：static/data/GeoLite2-City.mmdb|类型:GeoLite2-City` 即生效。

更新库文件：替换 `web/static/data/` 下的 mmdb 后重新编译即可
（MaxMind 官网注册后可免费下载；注意是 **City** 和 **ASN** 两个独立产品，`GeoLite2-Country` 不带经纬度，会导致 `City()` 查询报错）。

### 关于运营商

GeoLite2-City **不含** ISP 信息，只配 City 库时 `sourceIsp` 恒为空、`topIsps` 为空数组
（原来的 ip2region 库自带运营商，这是换库后唯一少掉的维度）。
想要保留运营商打码，额外下载 `GeoLite2-ASN.mmdb` 丢进同一个目录重新编译即可，
缺失时启动不会报错，只是没有这一项。

ASN 给的是 MaxMind 登记的**自治域组织全称**（如 `Alibaba (US) Technology Co., Ltd.` / `Google LLC` / `Zenlayer Inc`），
不是 ip2region 那种"联通"级别的国内运营商简称；

两点要注意：

- 名字很长，窄列里会撑爆 —— 建议前端给运营商徽章加个 `max-width`（如 120px）+ 省略号，
  完整名字已经在 `.stream__meta` 的 `title` 里（`'metaTitle'()` 的第三项就是 ISP）。
- 少数自治域没登记组织名时，服务端退回 `AS{号}`。

> 上面这些取自 2026-10-11 那批库的真实查询结果。

### 行为约定

- 整块 mmdb 字节直接当作查找数据用（零 IO 查询）+ IP 结果缓存（上限 10 万条，满了整体清空），
  单次查询几十微秒，**查不到的结果同样会缓存**，避免漏掉的 IP 每来一次都白跑一遍树查找。
- 国家 / 省 / 城市优先取 `zh-CN` 名，没有才兜底到 `en`。
- `127.0.0.1` 等国家/省份/城市统一为 `本机`；`10.x` / `172.16.x` / `192.168.x` 等内网地址为 `局域网`，不查库。
- 部分 IP（如 Cloudflare 的 `1.1.1.1`）本身没有城市级定位，`latitude` / `longitude` 会是 `0`，
  **前端画散点图前务必过滤掉双 0 的条目**，否则会全部堆到几内亚湾（0°,0°）。
- 库缺失或格式错误只打一条错误日志，不影响服务启动。
- 注意：mmdb 放在 `web/static/data/` 下会随静态目录一起被 `StaticController` 对外提供（即 `/data/GeoLite2-City.mmdb` 可被匿名下载）。若不想暴露，把文件移到非静态目录（例如 `web/data/`）并单独加一个 `//go:embed data`，再改 `main.go` 里的路径即可。
