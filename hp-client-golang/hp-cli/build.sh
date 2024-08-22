export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64
go build -o ./target/hp-pro.exe main.go

export CGO_ENABLED=0
export GOOS=windows
export GOARCH=386
go build -o ./target/hp-pro-386.exe main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=386
go build -o ./target/hp-pro-386 main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build -o ./target/hp-pro-amd64 main.go

export CGO_ENABLED=0
export GOOS=darwin
export GOARCH=amd64
go build -o ./target/hp-pro-apple-amd64 main.go

export CGO_ENABLED=0
export GOOS=darwin
export GOARCH=arm64
go build -o ./target/hp-pro-apple-arm64 main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm64
go build -o ./target/hp-pro-arm64 main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm
export GOARM=7
go build -o ./target/hp-pro-armv7 main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=mipsle
export GOMIPS=softfloat
go build -o ./target/hp-pro-mipsle main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=mips64le
export GOMIPS=softfloat
go build -o ./target/hp-pro-mips64le main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=mips
export GOMIPS=softfloat
go build -o ./target/hp-pro-mips main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=mips64
export GOMIPS=softfloat
go build -o ./target/hp-pro-mips64 main.go
