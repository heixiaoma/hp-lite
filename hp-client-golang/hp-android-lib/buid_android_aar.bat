SET ANDROID_HOME=D:\AndroidSdk
SET ANDROID_NDK_HOME=D:\AndroidSdk\ndk\29.0.13113456
go install golang.org/x/mobile/cmd/gomobile
go get golang.org/x/mobile/bind
gomobile init
gomobile bind -target=android -androidapi 21
