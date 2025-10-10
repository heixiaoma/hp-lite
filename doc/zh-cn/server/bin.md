# HP-Lite 服务端二进制方式部署教程

二进制部署是 HP-Lite 服务端最直接的部署方式，无需依赖容器或可视化面板，仅通过 “下载二进制文件 + 配置文件” 即可启动服务，适合熟悉命令行操作的用户，同时具备部署速度快、性能损耗低的优势。本教程将详细介绍从环境检查到服务运维（启动 / 停止 / 日志查看）的完整流程。

## 一、部署前准备

在开始部署前，请确保已完成以下准备工作，避免因环境缺失导致部署失败：

### 1. 确认服务器架构与操作系统

二进制文件需与服务器架构严格匹配，先通过命令确认服务器信息：



* **Linux/macOS 系统**：执行以下命令查看架构（重点关注 `x86_64` 或 `aarch64`，对应 AMD64 或 ARM64 架构）：



```
# 查看系统架构
uname -m
# 查看操作系统版本（可选，确认是否符合要求）
cat /etc/os-release  # Linux
sw_vers              # macOS
```



* **Windows 系统**：右键 “此电脑”→“属性”，在 “系统” 栏查看 “系统类型”（如 “64 位操作系统，基于 x64 的处理器”，对应 X86\_64 架构）。

### 2. 下载对应版本的二进制文件

根据服务器架构，从 HP-Lite 官方渠道下载对应的服务端二进制文件，文件清单如下（与前文 “系统架构支持” 对应）：



| 服务器架构   | 适用系统                                        | 二进制文件名                 | 下载说明                               |
| ------- | ------------------------------------------- | ---------------------- | ---------------------------------- |
| X86\_64 | Windows Server 2016+/10+                    | `hp-lite-server.exe`   | 下载后保存至任意目录（如 `D:\hp-lite-server\`） |
| AMD64   | Linux（CentOS 7+/Ubuntu 18.04+）、macOS 10.15+ | `hp-lite-server-amd64` | 下载后建议保存至 `/opt/hp-lite-server/` 目录 |
| ARM64   | Linux（ARM 架构服务器 / NAS）                      | `hp-lite-server-arm64` | 下载后建议保存至 `/opt/hp-lite-server/` 目录 |


### 3. 准备并修改核心配置文件（app.yml）

二进制部署需搭配 `app.yml` 配置文件（服务端核心配置），操作步骤如下：



1. **获取配置文件**：从 HP-Lite 官方文档或压缩包中获取 `app.yml` 模板文件，将其与二进制文件放在同一目录（如 Linux 放在 `/opt/hp-lite-server/`，Windows 放在 `D:\hp-lite-server\`）；

2. **修改关键配置**：参考前文《HP-Lite 服务端介绍》，必改以下参数（避免客户端无法连接）：

* `tunnel.ip`：改为服务器外网 IP 或已解析的域名（如 `123.45.67.89` 或 `server.hp-lite.com`）；

* `admin.password`：修改默认密码（如 `HpLite@2025`），避免安全风险；

* 其他参数（如 `admin.port`、`cmd.port`）按需修改，确保端口未被占用。

### 4. 开放必要端口（关键！）

根据 `app.yml` 配置，开放对应端口（需同时开放 TCP+UDP 协议，隧道端口支持双协议），操作方式分系统如下：

#### （1）Linux 系统

* **以 CentOS 为例**：

```
# 关闭 firewalld 防火墙
systemctl stop firewalld
# 禁止防火墙开机启动（可选，生产环境建议后续按需开放端口）
systemctl disable firewalld
```
* **以 Ubuntu 为例**：


```
# 禁止 ufw 开机启动​
sudo systemctl disable ufw​
# 验证是否禁止成功（输出 disabled 表示成功）​
sudo systemctl is-enabled ufw
```


#### （2）Windows 系统


* **临时关闭防火墙**（测试阶段）：

  控制面板 → 系统和安全 → Windows Defender 防火墙 → 点击 “关闭 Windows Defender 防火墙”（分别关闭 “专用网络” 和 “公用网络”）；

#### （3）云服务器（如阿里云、腾讯云）

需额外在 **云厂商安全组** 中开放端口（与服务器防火墙设置一致）：



* 登录云厂商控制台 → 找到对应服务器 → 安全组 → 配置规则 → 入站规则 → 添加规则 → 协议（TCP+UDP）、授权对象（0.0.0.0/0 表示允许所有 IP 访问，生产环境建议限制特定 IP）→ 保存。

## 二、二进制部署详细步骤

### 1. Linux 系统部署

#### 步骤 1：进入部署目录

打开终端，切换到二进制文件和 `app.yml` 所在的目录（以 `/opt/hp-lite-server/` 为例）：



```
cd /opt/hp-lite-server/
```

#### 步骤 2：赋予二进制文件执行权限

新下载的二进制文件默认可能没有执行权限，需通过命令赋予：



```
# 对 AMD64 架构文件
chmod +x hp-lite-server-amd64
# 对 ARM64 架构文件（若使用）
# chmod +x hp-lite-server-arm64
```

#### 步骤 3：启动服务（两种方式）

HP-Lite 服务支持 “前台启动”（便于实时查看日志，适合测试）和 “后台启动”（生产环境推荐，不占用终端）。

##### （1）前台启动（测试用）

直接执行二进制文件，日志会实时输出到终端（关闭终端则服务停止）：



```
# AMD64 架构
./hp-lite-server-amd64
# ARM64 架构（若使用）
# ./hp-lite-server-arm64
```



* 启动成功标志：终端输出类似 `admin server started on :9090`（管理后台启动）、`cmd server started on :16666`（控制端口启动）、`tunnel server started on 123.45.67.89:9091`（隧道端口启动）的日志。

##### （2）后台启动（生产环境推荐）

通过 `nohup` 命令将服务挂在后台运行，日志输出到指定文件（关闭终端不影响服务）：



```
# AMD64 架构：日志输出到 hp-lite-server.log 文件
nohup ./hp-lite-server-amd64 > hp-lite-server.log 2>&1 &
# ARM64 架构（若使用）
nohup ./hp-lite-server-arm64 > hp-lite-server.log 2>&1 &
```



* 命令说明：`nohup` 表示忽略挂起信号，`> hp-lite-server.log` 表示将标准输出重定向到日志文件，`2>&1` 表示将错误输出也重定向到日志文件，`&` 表示后台运行。

#### 步骤 4：验证服务是否启动成功

通过以下命令查询服务进程是否存在，或访问管理后台确认：

##### （1）查询服务进程



```
# 查找 HP-Lite 服务进程
ps aux | grep hp-lite-server
```



* 若输出类似以下内容，说明服务已启动（`PID` 为进程号，后续停止服务需用到）：



```
root      1234  0.0  0.5 123456  7890 ?        Ss   10:00   0:00 ./hp-lite-server-amd64
```

##### （2）访问管理后台

打开浏览器，输入 `http://服务器外网IP:admin.port`（如 `http://123.45.67.89:9090`），若能看到登录页面，输入 `app.yml` 中配置的 `admin.username` 和 `admin.password` 登录成功，说明服务启动正常。

#### 步骤 5：停止服务

若需停止服务，需先找到进程号（PID），再通过 `kill` 命令终止：



```
# 1. 查找进程号（PID）
ps aux | grep hp-lite-server
# 2. 终止进程（将 1234 替换为实际 PID）
kill 1234
# 若进程无法正常终止，可强制终止（谨慎使用）
kill -9 1234
```

#### 步骤 6：设置开机自启（生产环境推荐）

为避免服务器重启后服务需要手动启动，可通过 `systemd` 配置开机自启（以 CentOS 7+ 为例）：



1. 创建系统服务文件：



```
vim /etc/systemd/system/hp-lite-server.service
```



1. 粘贴以下内容（需根据实际路径和文件名修改）：



```
[Unit]
Description=HP-Lite Server Service
After=network.target  # 网络启动后再启动服务
[Service]
Type=simple
# 部署目录（替换为实际目录）
WorkingDirectory=/opt/hp-lite-server/
# 启动命令（替换为实际文件名）
ExecStart=/opt/hp-lite-server/hp-lite-server-amd64
# 重启策略：服务异常退出后自动重启
Restart=always
RestartSec=3  # 重启间隔 3 秒
[Install]
WantedBy=multi-user.target
```



1. 保存并退出：按 `Esc`，输入 `:wq`，回车。

2. 启用并启动服务：



```
# 重新加载 systemd 配置
systemctl daemon-reload
# 设置开机自启
systemctl enable hp-lite-server.service
# 启动服务
systemctl start hp-lite-server.service
# 验证开机自启是否设置成功
systemctl is-enabled hp-lite-server.service  # 输出 enabled 表示成功
```



1. 后续可通过 `systemd` 命令管理服务（更便捷）：



```
# 启动服务
systemctl start hp-lite-server.service
# 停止服务
systemctl stop hp-lite-server.service
# 重启服务
systemctl restart hp-lite-server.service
# 查看服务状态
systemctl status hp-lite-server.service
```


## 三、服务运维：状态查询与日志查看

### 1. 服务状态查询

#### （1）Linux 系统



* 若用 `systemd` 管理（推荐）：



```
systemctl status hp-lite-server.service
```



* 输出 `active (running)` 表示服务正常运行；

* 输出 `inactive (dead)` 表示服务已停止；

* 输出 `failed` 表示服务启动失败（需查看日志排查）。

- 若用 `nohup` 后台启动：



```
ps aux | grep hp-lite-server
```



* 若能找到进程，说明服务运行；若找不到，说明服务已停止。



#### （1）Linux/macOS 系统



* 若用 `nohup` 后台启动（日志输出到 `hp-lite-server.log`）：



```
# 实时查看最新日志（按 Ctrl+C 退出）
tail -f hp-lite-server.log
# 查看全部日志
cat hp-lite-server.log
# 查看日志最后 100 行（适合日志文件较大时）
tail -n 100 hp-lite-server.log
# 搜索包含“error”的日志（排查错误）
grep "error" hp-lite-server.log
```



* 若用 `systemd` 管理（日志由 `journalctl` 记录）：



```
# 查看 HP-Lite 服务的所有日志
journalctl -u hp-lite-server.service
# 实时查看最新日志（按 Ctrl+C 退出）
journalctl -u hp-lite-server.service -f
```
