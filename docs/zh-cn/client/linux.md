
#### Linux 启动客户端（两种方式）

HP-Lite 客户端支持 “前台启动”（便于实时查看日志，适合测试）和 “后台启动”（生产环境推荐，不占用终端）。

##### （1）前台启动（测试用）

直接执行二进制文件，日志会实时输出到终端（关闭终端则服务停止）：



```
# AMD64 架构
./hp-lite-amd64 -c=连接码
# ARM64 架构（若使用）
# ./hp-lite-arm64 -c=连接码
```


##### （2）后台启动（生产环境推荐）

通过 `nohup` 命令将服务挂在后台运行，日志输出到指定文件（关闭终端不影响服务）：



```
# AMD64 架构：日志输出到 hp-lite.log 文件
nohup ./hp-lite-amd64 -c=连接码 > hp-lite.log 2>&1 &
# ARM64 架构（若使用）
nohup ./hp-lite-arm64 -c=连接码 > hp-lite.log 2>&1 &
```



* 命令说明：`nohup` 表示忽略挂起信号，`> hp-lite.log` 表示将标准输出重定向到日志文件，`2>&1` 表示将错误输出也重定向到日志文件，`&` 表示后台运行。

#### 步骤 4：验证服务是否启动成功

通过以下命令查询服务进程是否存在，或访问管理后台确认：

##### （1）查询服务进程



```
# 查找 HP-Lite 服务进程
ps aux | grep hp-lite
```



* 若输出类似以下内容，说明服务已启动（`PID` 为进程号，后续停止服务需用到）：



```
root      1234  0.0  0.5 123456  7890 ?        Ss   10:00   0:00 ./hp-lite-amd64
```


#### 步骤 5：停止服务

若需停止服务，需先找到进程号（PID），再通过 `kill` 命令终止：



```
# 1. 查找进程号（PID）
ps aux | grep hp-lite
# 2. 终止进程（将 1234 替换为实际 PID）
kill 1234
# 若进程无法正常终止，可强制终止（谨慎使用）
kill -9 1234
```

#### 步骤 6：设置开机自启（生产环境推荐）

为避免服务器重启后服务需要手动启动，可通过 `systemd` 配置开机自启（以 CentOS 7+ 为例）：



1. 创建系统服务文件：



```
vim /etc/systemd/system/hp-lite.service
```



1. 粘贴以下内容（需根据实际路径和文件名修改）：



```
[Unit]
Description=HP-Lite Service
After=network.target  # 网络启动后再启动服务
[Service]
Type=simple
# 部署目录（替换为实际目录）
WorkingDirectory=/opt/hp-lite/
# 启动命令（替换为实际文件名）
ExecStart=/opt/hp-lite/hp-lite-amd64 -c=连接码
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
systemctl enable hp-lite.service
# 启动服务
systemctl start hp-lite.service
# 验证开机自启是否设置成功
systemctl is-enabled hp-lite.service  # 输出 enabled 表示成功
```


1. 后续可通过 `systemd` 命令管理服务（更便捷）：



```
# 启动服务
systemctl start hp-lite.service
# 停止服务
systemctl stop hp-lite.service
# 重启服务
systemctl restart hp-lite.service
# 查看服务状态
systemctl status hp-lite.service
```
