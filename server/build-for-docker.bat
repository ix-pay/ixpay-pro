@echo off
chcp 65001 >nul
echo Building ixpay-pro for Docker...

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

if not exist cmd\build mkdir cmd\build

go build -a -installsuffix cgo -ldflags "-s -w" -o cmd\build\ixpay-pro .\cmd\ixpay-pro\main.go

echo Build completed!
