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


### 2. 下载对应版本的二进制文件

根据服务器架构，从 HP-Lite 官方渠道下载对应的服务端二进制文件，文件清单如下（与前文 “系统架构支持” 对应）：



| 服务器架构   | 适用系统                                        | 二进制文件名                 | 下载说明                               |
| ------- | ------------------------------------------- | ---------------------- | ---------------------------------- |
| X86\_64 | Windows Server 2016+/10+                    | `hp-lite-server.exe`   | 下载后保存至任意目录（如 `D:\hp-lite-server\`） |
| AMD64   | Linux（CentOS 7+/Ubuntu 18.04+）、macOS 10.15+ | `hp-lite-server-amd64` | 下载后建议保存至 `/opt/hp-lite-server/` 目录 |
| ARM64   | Linux（ARM 架构服务器 / NAS）                      | `hp-lite-server-arm64` | 下载后建议保存至 `/opt/hp-lite-server/` 目录 |


### 3. 准备并修改核心配置文件（app.yml）

二进制部署需搭配 `app.yml` 配置文件（服务端核心配置），操作步骤如下：



1. **获取配置文件**：从 HP-Lite 官方文档或压缩包中获取 `app.yml` 模板文件，将其与二进制文件放在同一目录（如 Linux 放在 `/opt/hp-lite-server/`）；

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

#### （2）云服务器（如阿里云、腾讯云）

需额外在 **云厂商安全组** 中开放端口（与服务器防火墙设置一致）：

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


###### 步骤 1：安装服务
通过 `action` 参数命令将服务写到到systemd里进行启动：

```
# AMD64 架构：
安装服务： hp-lite-server-amd64 -action install
# ARM64 架构
安装服务： hp-lite-server-arm64 -action install
```

###### 步骤 2：启动客户端

```shell
# AMD64 架构：
启动服务： hp-lite-server-amd64 -action start
# ARM64 架构
启动服务： hp-lite-server-arm64 -action start
```

###### 步骤 3：停止客户端

```shell
# AMD64 架构：
停止服务： hp-lite-server-amd64 -action stop
# ARM64 架构
启动服务： hp-lite-server-arm64 -action stop
```

###### 步骤 4：查看状态

```shell
# AMD64 架构：
停止服务： hp-lite-server-amd64 -action status
# ARM64 架构
启动服务： hp-lite-server-arm64 -action status
```

###### 步骤 5：卸载

```shell
# AMD64 架构：
停止服务： hp-lite-server-amd64 -action uninstall
# ARM64 架构
启动服务： hp-lite-server-arm64 -action uninstall
```

###### 步骤 6：查看日志

```shell
journalctl -u hp-lite-server -f
```
>注意更新文件前先停止，然后在覆盖上传