@echo off
setlocal enabledelayedexpansion

echo Building kern for Windows...

:: Проверяем наличие Go
go version >nul 2>&1
if errorlevel 1 (
    echo Error: Go is not installed or not in PATH
    exit /b 1
)

:: Создаем директорию для сборки
if not exist "dist\windows" mkdir "dist\windows"

:: Собираем основную версию
echo Building main executable...
go build -o "dist\windows\kern.exe" -ldflags="-s -w" ./cmd/kern

:: Собираем версию для службы Windows
echo Building service version...
go build -o "dist\windows\kern-service.exe" -ldflags="-s -w -H=windowsgui" ./cmd/kern

:: Копируем файлы конфигурации
if not exist "dist\windows\config" mkdir "dist\windows\config"
copy "internal\i18n\active.en.json" "dist\windows\config\" 2>nul || echo No language files found
copy "internal\i18n\active.ru.json" "dist\windows\config\" 2>nul || echo No language files found

:: Создаем установочный скрипт
echo Creating installation script...
(
echo @echo off
echo echo Installing kern for Windows...
echo.
echo :: Создаем директорию в Program Files
echo if not exist "%%PROGRAMFILES%%\kern" mkdir "%%PROGRAMFILES%%\kern"
echo.
echo :: Копируем файлы
echo copy "kern.exe" "%%PROGRAMFILES%%\kern\"
echo copy "kern-service.exe" "%%PROGRAMFILES%%\kern\"
echo if exist config xcopy config "%%PROGRAMFILES%%\kern\config\" /E /I
echo.
echo :: Добавляем в PATH
echo setx PATH "%%PATH%%;%%PROGRAMFILES%%\kern" /M
echo.
echo echo Installation complete!
echo echo Usage:
echo echo   kern --help
echo echo   kern --remote
echo echo   kern-service --daemon
) > "dist\windows\install.bat"

:: Создаем README для Windows
(
echo # kern for Windows
echo.
echo ## Installation
echo 1. Run 'install.bat' as Administrator
echo 2. Or manually add kern.exe to your PATH
echo.
echo ## Usage Modes
echo.
echo ### Interactive Mode (PowerShell/CMD)
echo kern --cpu --mem
echo kern --all
echo.
echo ### Service Mode
echo kern-service --daemon
echo kern --remote
echo.
echo ### API Access
echo Start API server: kern --remote
echo Access: http://localhost:28126/api/cpu
echo.
echo ## Troubleshooting
echo - Run in PowerShell for best experience
echo - Use WSL for full TUI functionality
echo - Service mode works in background
) > "dist\windows\README.md"

echo.
echo Build complete!
echo Files are in: dist\windows\
echo.
echo For full TUI experience, use Windows Terminal or PowerShell
echo For service mode, use kern-service.exe