SET ANDROID_HOME=D:\android_sdk
SET ANDROID_NDK_HOME=D:\android_sdk\ndk\29.0.14206865
go install golang.org/x/mobile/cmd/gomobile
go get golang.org/x/mobile/bind
gomobile init
gomobile bind -target=android -androidapi 21
