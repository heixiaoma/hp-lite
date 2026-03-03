
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

###### 步骤 1：安装服务
通过 `action` 参数命令将服务写到到systemd里进行启动：

```
# AMD64 架构：
安装服务： hp-lite-amd64 -c "你的连接码" -action install
# ARM64 架构
安装服务： hp-lite-arm64 -c "你的连接码" -action install
```

###### 步骤 2：启动客户端

```shell
# AMD64 架构：
启动服务： hp-lite-amd64 -action start
# ARM64 架构
启动服务： hp-lite-arm64 -action start
```

###### 步骤 3：停止客户端

```shell
# AMD64 架构：
停止服务： hp-lite-amd64 -action stop
# ARM64 架构
启动服务： hp-lite-arm64 -action stop
```

###### 步骤 4：查看状态

```shell
# AMD64 架构：
停止服务： hp-lite-amd64 -action status
# ARM64 架构
启动服务： hp-lite-arm64 -action status
```

###### 步骤 5：卸载

```shell
# AMD64 架构：
停止服务： hp-lite-amd64 -action uninstall
# ARM64 架构
启动服务： hp-lite-arm64 -action uninstall
```

###### 步骤 6：查看日志

```shell
journalctl -u hp-lite -f
```

>注意更新文件前先停止，然后在覆盖上传