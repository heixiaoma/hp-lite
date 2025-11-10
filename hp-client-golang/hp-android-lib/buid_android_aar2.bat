SET ANDROID_HOME=C:\Users\admin\AppData\Local\Android\Sdk
SET ANDROID_NDK_HOME=C:\Users\admin\AppData\Local\Android\Sdk\ndk\29.0.14206865
go install golang.org/x/mobile/cmd/gomobile
go get golang.org/x/mobile/bind
gomobile init
gomobile bind -target=android -androidapi 21
