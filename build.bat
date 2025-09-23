@echo off
setlocal enabledelayedexpansion

REM 跨平台打包脚本 (Windows 版本)
REM 为不同操作系统和架构编译 sim-sms-forward

echo 开始跨平台构建...

REM 项目信息
set PROJECT_NAME=sim-sms-forward
set VERSION=v1.0.0
set BINARY_NAME=sim-sms-forward

REM 获取构建时间
for /f "tokens=1-5 delims=/ " %%a in ('date /t') do set build_date=%%c-%%a-%%b
for /f "tokens=1-2 delims=: " %%a in ('time /t') do set build_time=%%a:%%b
set BUILD_TIME=%build_date% %build_time%

REM 输出目录
set OUTPUT_DIR=dist
if exist %OUTPUT_DIR% rmdir /s /q %OUTPUT_DIR%
mkdir %OUTPUT_DIR%

REM 构建信息
set LDFLAGS=-s -w

echo 开始构建 %PROJECT_NAME% %VERSION%
echo 构建时间: %BUILD_TIME%
echo ======================================

REM 定义平台和架构
set PLATFORMS=linux/amd64 linux/arm64 linux/arm windows/amd64 windows/arm64 darwin/amd64 darwin/arm64

for %%p in (%PLATFORMS%) do (
    for /f "tokens=1,2 delims=/" %%a in ("%%p") do (
        set GOOS=%%a
        set GOARCH=%%b
        
        REM 设置输出文件名
        set output_name=%BINARY_NAME%-!GOOS!-!GOARCH!
        if "!GOOS!"=="windows" set output_name=!output_name!.exe
        
        set output_path=%OUTPUT_DIR%\!output_name!
        
        echo 构建 !GOOS!/!GOARCH!...
        
        REM 编译
        set GOOS=!GOOS!
        set GOARCH=!GOARCH!
        go build -ldflags="!LDFLAGS!" -o "!output_path!" main.go
        
        if !errorlevel! equ 0 (
            echo   成功: !output_path!
        ) else (
            echo   失败: !GOOS!/!GOARCH!
            exit /b 1
        )
    )
)

echo ======================================
echo 所有平台构建完成！
echo.
echo 构建文件列表:
dir %OUTPUT_DIR%

echo.
echo 打包完成！输出目录: %OUTPUT_DIR%

pause