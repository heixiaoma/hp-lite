export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64
go build -o hp-client-golang.exe main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build -o hp-client-golang-amd64 main.go

export CGO_ENABLED=0
export GOOS=darwin
export GOARCH=amd64
go build -o hp-client-golang-apple-amd64 main.go

export CGO_ENABLED=0
export GOOS=darwin
export GOARCH=arm64
go build -o hp-client-golang-apple-arm64 main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm64
go build -o hp-client-golang-arm64 main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm
export GOARM=7
go build -o hp-client-golang-armv7 main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=mipsle
go build -o hp-client-golang-mips main.go
