SET ANDROID_HOME=D:\AndroidSdk
SET ANDROID_NDK_HOME=D:\AndroidSdk\ndk\29.0.13113456
@REM go install golang.org/x/mobile/cmd/gomobile
@REM go get golang.org/x/mobile/bind
@REM gomobile init
gomobile bind -target=android -androidapi 21
