@echo off
REM Build and push hp-lite images to DockerHub
REM User: xrilang, Version: v26.1.14

setlocal enabledelayedexpansion

echo.
echo ========================================
echo Building hp-lite Docker images
echo ========================================
echo.

REM Step 1: Build hp-client binaries
echo Step 1: Building hp-client binaries...
cd /d "C:\Users\zhangweijie\Desktop\hp-lite\hp-client-golang\hp-cli"
powershell -ExecutionPolicy Bypass -File "build-windows.ps1"
if errorlevel 1 (
    echo Error building hp-client binaries
    exit /b 1
)

REM Step 2: Build hp-server binaries
echo.
echo Step 2: Building hp-server binaries...
cd /d "C:\Users\zhangweijie\Desktop\hp-lite\hp-server-golang"
powershell -ExecutionPolicy Bypass -File "build-windows.ps1"
if errorlevel 1 (
    echo Error building hp-server binaries
    exit /b 1
)

REM Step 3: Build and push hp-client Docker image
echo.
echo Step 3: Building and pushing hp-client Docker image...
cd /d "C:\Users\zhangweijie\Desktop\hp-lite\hp-client-golang\hp-cli\docker"
docker buildx bake --push
if errorlevel 1 (
    echo Error building hp-client Docker image
    exit /b 1
)

REM Step 4: Build and push hp-server Docker image
echo.
echo Step 4: Building and pushing hp-server Docker image...
cd /d "C:\Users\zhangweijie\Desktop\hp-lite\hp-server-golang\docker"
docker buildx bake --push
if errorlevel 1 (
    echo Error building hp-server Docker image
    exit /b 1
)

echo.
echo ========================================
echo Success! Images pushed to DockerHub
echo ========================================
echo Images:
echo   - xrilang/hp-lite:latest
echo   - xrilang/hp-lite:v26.1.14
echo   - xrilang/hp-lite-server:latest
echo   - xrilang/hp-lite-server:v26.1.14
echo ========================================
