#  Docker 部署文档

## 一、部署前准备

### 1.1 环境要求


* **操作系统**：支持 Docker 的 Linux 发行版，需保障网络通畅（满足 “极端环境下能上网即可穿透” 的核心特性）。


#### 1.1.2 准备并修改核心配置文件（app.yml）

二进制部署需搭配 `app.yml` 配置文件（服务端核心配置），操作步骤如下：


1. **获取配置文件**：从 HP-Lite 官方文档或压缩包中获取 `app.yml` 模板文件，将其与二进制文件放在同一目录（如 Linux 放在 `/data/`）；

2. **修改关键配置**：参考前文《HP-Lite 服务端介绍》，必改以下参数（避免客户端无法连接）：

* `tunnel.ip`：改为服务器外网 IP 或已解析的域名（如 `123.45.67.89` 或 `server.hp-lite.com`）；

* `admin.password`：修改默认密码（如 `HpLite@2025`），避免安全风险；

* `acme.email`：修改Acme的邮箱，可以随意写一个，防止acme验证问题；

* 其他参数（如 `admin.port`、`cmd.port`）按需修改，确保端口未被占用。

#### 1.1.3. 防火墙处理（关键！）

根据 `app.yml` 配置，开放对应端口

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


需额外在 **云厂商安全组** 中开放端口（与服务器防火墙设置一致）：



#### 1.1.4 配置文件创建与内容


1. 在主机创建 `/data` 目录（用于存放配置文件和 `hp-lite.db` 数据库）：


```
sudo mkdir -p /data
# 编辑 app.yml 配置文件
sudo vim /data/app.yml
```

1. 将以下配置内容复制到文件中（根据实际需求修改账号、服务器 IP 等信息）：

```
admin:
  username: 'admin' #后台账号
  password: '123456' #后台密码
  port: 9090 #管理后台监听的端口（HTTP协议传输方式）

cmd:
  port: 16666 #控制指令端口，所有HP-lite 客户端需要连接这个端口（TCP传输方式）

tunnel:
  ip: '127.0.0.1' #隧道监听服务器外网的IP（记得改成你的服务器IP或者解析的域名也可以）
  port: 9091 #隧道传输数据端口，在5.0以后这个端口支持TCP和QUIC(UDP协议) 所以在开放端口时要给TCP和UDP权限
  open-domain: true #true 开启80，443端口域名转发（如果你的服务有宝塔或者nginx等，端口多半是被用了），false 关闭
acme:
  email: '232323@qq.com' #申请证书必须写一个邮箱可以随便写
  http-port: '5634' #证书验证会访问http接口，会通过80转发过来，所以这个端口不用暴露外网
```

1. 设置配置文件权限（确保 Docker 容器可读取）：


```
sudo chmod 644 /data/app.yml
```

#### 1.1.5 关键配置项提醒

* `admin.password`：默认密码安全性低，部署后需立即通过管理后台修改，防止未授权访问。

* `tunnel.ip`：必须修改为服务器**外网 IP 或已解析域名**，否则客户端无法连接隧道服务。

* `tunnel.open-domain`：若服务器已运行 Nginx、Apache 或宝塔面板，80/443 端口可能被占用，需设为 `false` 避免端口冲突。

* `acme.email`：修改Acme的邮箱，可以随意写一个，防止acme验证问题；

### 2.2 hp-lite.db 数据库核心说明

#### 2.2.1 映射路径与作用

Docker 命令中 `-v /data/data:/hp-lite-server/data` 的核心作用是**将容器内生成的&#x20;**`hp-lite.db`**&#x20;持久化到主机&#x20;**`/data/data`**&#x20;目录**，具体说明如下：


| 路径类型  | 路径地址                   | 核心作用                                          |
| ----- | ---------------------- | --------------------------------------------- |
| 主机路径  | `/data/data`           | 存储 `hp-lite.db` 数据库及 SSL 证书、运行日志，容器删除后数据不丢失   |
| 容器内路径 | `/hp-lite-server/data` | HP-Lite 服务默认数据目录，服务启动后自动生成 `hp-lite.db` 数据库文件 |


#### 2.2.2 数据库持久化的必要性


* 若未配置 `/data/data` 目录映射，容器删除后 `hp-lite.db` 会随容器销毁，所有穿透配置、设备信息将丢失，需重新配置。

* 映射后即使容器重启 / 更新，`hp-lite.db` 数据仍保留，服务可直接恢复原有配置，无需重复操作。

#### 2.2.3 目录权限配置

无需手动创建 `/data/data` 目录（Docker 启动时会自动生成），但需确保主机 `/data` 目录有读写权限，避免数据库文件无法生成：


```
sudo chmod 755 /data
```

## 三、Docker 部署命令

### 3.1 阿里云源部署（推荐，国内拉取速度快）


```
sudo docker run \
  --name hp-lite-server \  # 容器名称（自定义，便于后续运维，如：hp-lite-5.0）
  --net=host \             # 使用主机网络模式（确保端口直接映射，避免网络转发异常）
  --restart=always \       # 容器开机自启，异常退出后自动重启（保障服务稳定性）
  -d \                     # 后台运行容器（不占用当前终端）
  -v /data/app.yml:/hp-lite-server/app.yml \  # 映射配置文件（主机:容器）
  -v /data/data:/hp-lite-server/data \        # 映射数据目录（含 hp-lite.db，主机:容器）
  registry.cn-shenzhen.aliyuncs.com/heixiaoma/hp-lite-server:latest
```

### 3.2 Docker 官方源部署（适合海外服务器）

```
sudo docker run \
  --name hp-lite-server \
  --net=host \
  --restart=always \
  -d \
  -v /data/app.yml:/hp-lite-server/app.yml \
  -v /data/data:/hp-lite-server/data \
  heixiaoma/hp-lite-server:latest
```

### 3.3 部署验证



1. 确认容器正常运行：



```
sudo docker ps | grep hp-lite-server

# 输出含 "Up" 表示运行正常，如："Up 5 minutes"
```



1. 确认 `hp-lite.db` 已生成：



```
ls /data/data | grep hp-lite.db

# 输出 "hp-lite.db" 表示数据库文件创建成功
```



1. 访问管理后台验证：

* 打开浏览器，输入 `http://服务器外网IP:9090`，使用 `app.yml` 中配置的 `admin.username` 和 `admin.password` 登录，能正常进入后台即部署成功。

## 四、容器日常运维操作

### 4.1 查看服务状态与日志

#### 4.1.1 查看容器运行状态



```
# 查看容器是否运行（输出 "running" 表示正常）
sudo docker inspect --format '{{.State.Status}}' hp-lite-server
# 查看容器资源占用（CPU、内存、网络流量）
sudo docker stats hp-lite-server
```

#### 4.1.2 查看服务日志（排查错误）



```
# 实时查看日志（按 Ctrl+C 退出，可定位连接失败、配置错误等问题）
sudo docker logs -f hp-lite-server
# 查看最近 200 行日志（聚焦近期操作）
sudo docker logs --tail 200 hp-lite-server
```

### 4.2 配置修改与服务重启

若修改 `app.yml` 配置（如更换管理密码、调整端口），需重启容器使配置生效：



```
# 修改 app.yml 后执行重启
sudo docker restart hp-lite-server
# 验证重启结果（查看日志中是否有 "server started successfully" 等成功信息）
sudo docker logs -f hp-lite-server --tail 50
```

### 4.3 服务更新（版本升级）

当 HP-Lite 发布新版本时，更新流程如下（`hp-lite.db`**&#x20;数据会保留，无需重新配置**）：



1. 停止并删除旧容器：



```
sudo docker stop hp-lite-server && sudo docker rm hp-lite-server
```



1. 拉取最新镜像（以阿里云源为例）：



```
sudo docker pull registry.cn-shenzhen.aliyuncs.com/heixiaoma/hp-lite-server:latest
```



1. 重新执行 **第三章** 的部署命令（配置文件和 `hp-lite.db` 会自动复用）。

### 4.4 服务停止与数据备份 / 删除

#### 4.4.1 仅停止服务（可重启恢复）



```
sudo docker stop hp-lite-server
# 重启服务（如需恢复）
sudo docker start hp-lite-server
```

#### 4.4.2 hp-lite.db 数据备份（推荐）

为防止数据丢失，建议定期备份 `hp-lite.db`：



```
# 备份到 /data/backup 目录（需先创建备份目录）
sudo mkdir -p /data/backup
sudo cp /data/data/hp-lite.db /data/backup/hp-lite.db\_\$(date +%Y%m%d)
# 示例：备份文件名为 hp-lite.db\_20240520（含日期，便于区分版本）
```

#### 4.4.3 彻底删除服务与数据（谨慎操作）



```
# 停止并删除容器
sudo docker stop hp-lite-server && sudo docker rm hp-lite-server
# 【谨慎】删除配置文件和 hp-lite.db 数据库（彻底卸载）
sudo rm -rf /data
```

## 五、常见问题排查

### 5.1 容器启动失败（状态为 exited）



* **原因 1**：`app.yml` 格式错误（YAML 对缩进敏感，如空格数不统一）。

  解决：使用 `yamllint` 工具校验格式：



```
sudo pip install yamllint  # 安装校验工具

yamllint /data/app.yml     # 查看错误提示并修正（如缩进错误、引号不匹配）
```



* **原因 2**：端口被占用（如 9090 端口已被其他服务使用）。

  解决：查看端口占用并修改 `app.yml` 端口：



```
# 查看 9090 端口占用情况（替换端口号可查其他端口）
sudo netstat -tulpn | grep 9090
# 若有占用，修改 app.yml 中对应端口（如将 admin.port 改为 9092）
```

### 5.2 管理后台无法访问（http:// 服务器 IP:9090）



* 检查服务器安全组是否开放 9090 端口（TCP）。

* 确认容器已正常运行：`sudo docker ps | grep hp-lite-server`。

* 查看日志定位问题：`sudo docker logs hp-lite-server`，确认管理后台是否启动成功。

### 5.3 hp-lite.db 未生成（/data/data 目录下无该文件）



* **原因**：`/data` 目录权限不足，容器无法写入数据。

  解决：重新设置权限并重启容器：



```
sudo chmod 755 /data

sudo docker restart hp-lite-server
```

通过以上步骤，可完成 HP-Lite 服务端的 Docker 部署，并确保 `hp-lite.db` 中穿透配置、域名配置等关键数据持久化，保障服务稳定运行及数据安全。
