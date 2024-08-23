# HP-Lite 2.0内网穿透


### 项目介绍
项目1.0使用的TCP传输 项目2.0升级调整QUIC传输和HP-PRO版本同款传输机制，HP-PRO是商业版本拥有更多的功能和稳定性，内置了关键字过滤，实名认证，微信小程序管理端，集群部署，数据统计，限流套件，智能黑名单防护功能，套餐支付功能，服务节点的动态上架下架管理，等诸多的商业功能
如果你是非常感兴趣，可以通过QQ群进行交流
QQ群1：767474069

#### HP Lite介绍
HP-Lite2.0是一个单机方案
我们采用的是数据转发实现 稳定性可靠性是有保证的即便是极端的环境只要能上网就能实现穿透。
我们支持TCP和UDP协议，针对 http/https ws/wss 协议做了大量的优化工作可以更加灵活的控制。让用户使用更佳舒服简单。


### 服务端
- 下载2.0的jar 使用java1.8+版本运行
- 配置说明 app.yml
- app.yml文件放在和jar同目录下即可，不然会默认启动配置
```yaml
admin:
  username: 'admin' #后台账号
  password: '123456' #后台密码
  port: 9090 #管理后台监听的端口

cmd:
  port: 6666 #控制指令端口，所有HP-lite 客户端需要连接这个端口

tunnel:
  ip: '127.0.0.1' #隧道监听服务器真实的IP（记得改成你的服务器IP或者解析的域名也可以）
  port: 9091 #隧道传输数据端口，这个端口用来传输数据的，
  open-domain: true #true 开启80，443端口域名转发（如果你的服务有宝塔或者nginx等，端口多半是被用了），false 关闭

log: info #日志级别，不用动
```


### B站安装教程
https://www.bilibili.com/video/BV1pAW7eMEPX


### 客户端运行方式
##### docker
```shell
# 通过 docker run 运行容器
sudo docker run --restart=always -d  -e server=xxx.com穿透服务:6666 -e deviceId=32位的设备ID registry.cn-shenzhen.aliyuncs.com/hserver/hp-lite:latest
# 通过 docker run 运行容器 ARM
sudo docker run --restart=always -d  -e server=xxx.com穿透服务:6666 -e deviceId=32位的设备ID registry.cn-shenzhen.aliyuncs.com/hserver/hp-lite:latest-arm64
```
##### Linux或者win
```shell
chmod -R 777 ./hp-lite-amd64
./hp-lite-amd64 -server=xxx.com穿透服务:6666 -deviceId=32位的设备ID 
```


## 运行截图
<img src="https://gitee.com/HServer/hp-lite/raw/main/doc/img/img.png"  />
<img src="https://gitee.com/HServer/hp-lite/raw/main/doc/img/img_1.png"  />
<img src="https://gitee.com/HServer/hp-lite/raw/main/doc/img/img_2.png"  />
<img src="https://gitee.com/HServer/hp-lite/raw/main/doc/img/img_3.png"  />
<img src="https://gitee.com/HServer/hp-lite/raw/main/doc/img/img_4.png"  />

