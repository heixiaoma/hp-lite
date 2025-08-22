export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64
go build -o ./target/hp-lite.exe main.go

export CGO_ENABLED=0
export GOOS=windows
export GOARCH=386
go build -o ./target/hp-lite-386.exe main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=386
go build -o ./target/hp-lite-386 main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build -o ./target/hp-lite-amd64 main.go

export CGO_ENABLED=0
export GOOS=darwin
export GOARCH=amd64
go build -o ./target/hp-lite-apple-amd64 main.go

export CGO_ENABLED=0
export GOOS=darwin
export GOARCH=arm64
go build -o ./target/hp-lite-apple-arm64 main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm64
go build -o ./target/hp-lite-arm64 main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm
export GOARM=7
go build -o ./target/hp-lite-armv7 main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=mipsle
export GOMIPS=softfloat
go build -o ./target/hp-lite-mipsle main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=mips64le
export GOMIPS=softfloat
go build -o ./target/hp-lite-mips64le main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=mips
export GOMIPS=softfloat
go build -o ./target/hp-lite-mips main.go

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=mips64
export GOMIPS=softfloat
go build -o ./target/hp-lite-mips64 main.go
