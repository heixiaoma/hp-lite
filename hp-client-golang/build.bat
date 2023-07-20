SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
go build -o hp-client-golang.exe main.go

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o hp-client-golang-amd64 main.go

set CGO_ENABLED=0
set GOOS=darwin
set GOARCH=amd64
go build -o hp-client-golang-apple-amd64 main.go

set CGO_ENABLED=0
set GOOS=darwin
set GOARCH=arm64
go build -o hp-client-golang-apple-arm64 main.go

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm64
go build -o hp-client-golang-arm64 main.go

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm
set GOARM=7
go build -o hp-client-golang-armv7 main.go

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=mipsle
go build -o hp-client-golang-mips main.go
