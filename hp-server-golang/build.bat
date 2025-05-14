SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
go build -o ./target/hp-lite-server.exe main.go

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o ./target/hp-lite-server-amd64 main.go

set CGO_ENABLED=0
set GOOS=darwin
set GOARCH=amd64
go build -o ./target/hp-lite-server-apple-amd64 main.go

set CGO_ENABLED=0
set GOOS=darwin
set GOARCH=arm64
go build -o ./target/hp-lite-server-apple-arm64 main.go

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm64
go build -o ./target/hp-lite-server-arm64 main.go

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm
set GOARM=7
go build -o ./target/hp-lite-server-armv7 main.go
