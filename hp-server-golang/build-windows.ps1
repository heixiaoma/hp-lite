# Build hp-lite-server for multiple architectures
$ErrorActionPreference = "Stop"

Write-Host "Building hp-lite-server binaries for multiple architectures..." -ForegroundColor Green

# Create target directory
New-Item -ItemType Directory -Force -Path "./target" | Out-Null

# Linux amd64
Write-Host "Building for linux/amd64..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ./target/hp-lite-server-amd64 main.go

# Linux arm64
Write-Host "Building for linux/arm64..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "arm64"
go build -o ./target/hp-lite-server-arm64 main.go

# Linux armv7
Write-Host "Building for linux/arm/v7..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "arm"
$env:GOARM = "7"
go build -o ./target/hp-lite-server-armv7 main.go

Write-Host "Build complete! Binaries are in ./target/" -ForegroundColor Green
