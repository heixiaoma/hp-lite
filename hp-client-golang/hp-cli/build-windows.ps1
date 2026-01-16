# Build hp-lite for multiple architectures
$ErrorActionPreference = "Stop"

Write-Host "Building hp-lite binaries for multiple architectures..." -ForegroundColor Green

# Create target directory
New-Item -ItemType Directory -Force -Path "./target" | Out-Null

# Linux amd64
Write-Host "Building for linux/amd64..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ./target/hp-lite-amd64 main.go

# Linux 386
Write-Host "Building for linux/386..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "386"
go build -o ./target/hp-lite-386 main.go

# Linux arm64
Write-Host "Building for linux/arm64..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "arm64"
go build -o ./target/hp-lite-arm64 main.go

# Linux armv7
Write-Host "Building for linux/arm/v7..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "arm"
$env:GOARM = "7"
go build -o ./target/hp-lite-armv7 main.go

Write-Host "Build complete! Binaries are in ./target/" -ForegroundColor Green
