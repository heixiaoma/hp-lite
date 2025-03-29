export ANDROID_HOME=/Users/heixiaoma/Library/Android/sdk
go install golang.org/x/mobile/cmd/gomobile
go get golang.org/x/mobile/bind
gomobile init
gomobile bind -target=android
