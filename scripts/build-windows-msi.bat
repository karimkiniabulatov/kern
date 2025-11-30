@echo off
echo Building Windows MSI installer...

:: Собрать бинарники
go build -o dist/kern.exe -ldflags="-s -w -H=windowsgui" ./cmd/kern

:: Создать MSI (требуется WiX Toolset)
wix build -o dist/kern-1.2.3.msi