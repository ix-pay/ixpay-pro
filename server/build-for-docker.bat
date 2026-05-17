@echo off
echo Building ixpay-pro for Docker...

REM 设置 CGO_ENABLED=0 和 GOOS=linux
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

REM 创建输出目录
mkdir cmd\build

REM 编译
go build -a -installsuffix cgo -ldflags "-s -w" -o cmd\build\ixpay-pro .\internal\main.go

echo Build completed!
