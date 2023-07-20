# HP-Lite内网穿透

#### 介绍
HP-Lite是一个单机方案
我们采用的是数据转发实现 稳定性可靠性是有保证的即便是极端的环境只要能上网就能实现穿透。
我们支持TCP和UDP协议，针对 http/https ws/wss 协议做了大量的优化工作可以更加灵活的控制。让用户使用更佳舒服简单。

### 运行方式
##### docker
```json
# 通过 docker run 运行容器
sudo docker run -P -d  -e server=xxx.com穿透服务 deviceId=10-36位的自定义设备ID registry.cn-shenzhen.aliyuncs.com/hserver/hp-lite:latest
# 通过 docker run 运行容器 ARM
sudo docker run -P -d  -e server=xxx.com穿透服务 deviceId=10-36位的自定义设备ID registry.cn-shenzhen.aliyuncs.com/hserver/hp-lite:latest-arm64
```
##### Linux或者win
```json
chmod -R 777 ./hp-client-golang-amd64 
./hp-client-golang-amd64 -server=xxx.com穿透服务 -deviceId=10-36位的自定义设备ID 
```


## 运行截图
<img src="https://gitee.com/HServer/hp-lite/raw/main/doc/img/img.png"  />
<img src="https://gitee.com/HServer/hp-lite/raw/main/doc/img/img_1.png"  />
<img src="https://gitee.com/HServer/hp-lite/raw/main/doc/img/img_2.png"  />
<img src="https://gitee.com/HServer/hp-lite/raw/main/doc/img/img_3.png"  />
<img src="https://gitee.com/HServer/hp-lite/raw/main/doc/img/img_4.png"  />

