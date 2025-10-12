#  宝塔部署



## 二进制部署

### 1. 在宝塔文件里创建相关目录
<img src="../../tech/bt/img.png" />

### 2. 设置上传和修改配置文件
<img src="../../tech/bt/img_4.png" />

### 进行部署
<img src="../../tech/bt/img_2.png" />


## docker部署
### 1. 设置上传和修改配置文件
<img src="../../tech/bt/img_4.png" />

### 2.创建容器进行运行
<img src="../../tech/bt/img_5.png" />

* **以 阿里云源 为例**：

```
docker run --name hp-lite-server --net=host --restart=always -d  -v /www/wwwroot/hp/app.yml:/hp-lite-server/app.yml  -v  /www/wwwroot/hp/data:/hp-lite-server/data  registry.cn-shenzhen.aliyuncs.com/heixiaoma/hp-lite-server:latest
```

* **以 docker源 为例**：

```
docker run --name hp-lite-server --net=host --restart=always -d  -v /www/wwwroot/hp/app.yml:/hp-lite-server/app.yml  -v  /www/wwwroot/hp/data:/hp-lite-server/data  heixiaoma/hp-lite-server:latest
```


## 疑难
- 关闭宝塔的防火墙，或者放开相关端口
- 关闭云厂商的安全策略组，或者添加相关规则
- 注意宝塔安装的nginx和apache程序是否占用了80和443端口，app.yml的`open-domain`域名转发需要进行关闭