# HP-Lite Docker 部署说明

## 部署完成状态

✅ Docker镜像已成功构建
✅ Docker容器已成功启动并运行
✅ 配置文件已正确映射
✅ Web界面可正常访问
✅ 前端已更新到最新版本（包含忘记密码和个人设置功能）

## 最新功能

### 新增功能入口
1. **忘记密码**: 登录页面右下角有"忘记密码？"链接
2. **个人设置**: 登录后点击右上角用户头像，下拉菜单中有"个人设置"选项

### 功能说明
- **忘记密码**: 通过邮箱验证码重置密码
- **个人设置**:
  - 邮箱设置：绑定/修改邮箱
  - 安全设置：修改登录密码
  - 账户信息：查看用户信息

## 部署信息

### 容器信息
- **容器名称**: hp-lite-server
- **镜像**: hp-lite-server:latest
- **状态**: 运行中

### 端口映射
- **9999**: 管理后台端口 (HTTP)
- **16666**: 控制指令端口 (TCP)
- **9091**: 隧道传输数据端口 (TCP + UDP)

### 配置文件映射
- **宿主机**: `C:\Users\zhangweijie\Desktop\nwct\app.yml`
- **容器内**: `/hp-lite-server/app.yml` (只读)

### 数据持久化
- **Volume名称**: hp-lite_hp-lite-data
- **挂载点**: `/hp-lite-server/data`

## 访问方式

### Web管理界面
```
http://localhost:9999
```

### 登录凭据（来自配置文件）
- **用户名**: xrilang
- **密码**: qq2686485465

## 常用命令

### 查看容器状态
```bash
docker ps | grep hp-lite-server
```

### 查看容器日志
```bash
docker logs hp-lite-server
docker logs -f hp-lite-server  # 实时查看
```

### 停止容器
```bash
docker-compose down
```

### 启动容器
```bash
docker-compose up -d
```

### 重启容器
```bash
docker-compose restart
```

### 进入容器
```bash
docker exec -it hp-lite-server sh
```

### 查看配置文件
```bash
docker exec hp-lite-server cat /hp-lite-server/app.yml
```

## 重新构建

如果需要更新代码并重新部署：

1. 构建前端和后端
```bash
cd hp-web
npm run build

cd ../hp-server-golang
go build -o target/hp-server.exe
```

2. 复制前端文件到后端
```bash
cp -r hp-web/dist/* hp-server-golang/web/static/
```

3. 重新构建Docker镜像
```bash
docker build -t hp-lite-server:latest .
```

4. 重启容器
```bash
docker-compose down
docker-compose up -d
```

## 文件说明

### Dockerfile
位于项目根目录，使用Alpine Linux作为基础镜像，包含：
- 已编译的Linux版本服务端程序
- 前端静态文件
- 必要的运行时依赖

### docker-compose.yml
位于项目根目录，定义了：
- 服务配置
- 端口映射
- 卷挂载
- 网络配置

### .dockerignore
位于项目根目录，排除不需要的文件以优化构建。

## 注意事项

1. **配置文件**: 配置文件映射为只读模式，如需修改配置，请修改宿主机上的 `C:\Users\zhangweijie\Desktop\nwct\app.yml`，然后重启容器。

2. **端口冲突**: 确保宿主机上的 9999、16666、9091 端口未被占用。

3. **数据持久化**: 数据库和其他数据文件存储在Docker volume中，即使删除容器也不会丢失。

4. **Linux二进制**: Docker镜像使用的是 `hp-lite-server-amd64` (Linux版本)，不是Windows的 `.exe` 文件。

5. **网络**: 容器使用独立的bridge网络 `hp-lite_hp-lite-network`。

## 故障排查

### 容器无法启动
```bash
# 查看详细日志
docker logs hp-lite-server

# 检查配置文件是否存在
ls -la C:\Users\zhangweijie\Desktop\nwct\app.yml
```

### 无法访问Web界面
```bash
# 检查端口映射
docker port hp-lite-server

# 检查容器是否运行
docker ps | grep hp-lite-server

# 测试端口
curl http://localhost:9999
```

### 配置文件未生效
```bash
# 检查容器内的配置文件
docker exec hp-lite-server cat /hp-lite-server/app.yml

# 重启容器使配置生效
docker-compose restart
```

## 卸载

完全删除容器、镜像和数据：

```bash
# 停止并删除容器
docker-compose down

# 删除数据卷（注意：会删除所有数据）
docker volume rm hp-lite_hp-lite-data

# 删除镜像
docker rmi hp-lite-server:latest

# 删除网络
docker network rm hp-lite_hp-lite-network
```
