# 关于androidSDK描述说明
```
# 设置环境变量ANDROID_HOME，值为android sdk的路径，我这里把android sdk放在了D:\android-sdk-windows，把ANDROID_HOME值设置为D:\android-sdk-windows就好了
go get golang.org/x/mobile/cmd/gomobile

build gomobile成功后会在$GOPATH/bin目录生成gomobile可执行程序
go build golang.org/x/mobile/cmd/gomobile
gomobile init

cd android
gomobile bind -target=android
生成aar
#gomobile build -target=android 直接生成apk了
在将aar导入android使用
```
