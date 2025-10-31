@echo off
REM Build script for Customer Survey Desktop Application
REM Alternative to build-desktop.ps1 for systems with restricted PowerShell

echo Building Customer Survey Desktop Application...
echo.

REM Clean previous builds
if exist customer-survey.exe (
    del customer-survey.exe
    echo Cleaned previous build
)

REM Build with Windows GUI mode (no console window)
echo Compiling...
go build -ldflags="-H windowsgui" -o customer-survey.exe .\cmd\survey\main.go

if %errorlevel% equ 0 (
    echo.
    echo ✅ Build successful!
    echo    Output: customer-survey.exe
    echo.
    echo To configure webhook, create config.json with:
    echo    {"webhook_url": "https://script.google.com/macros/s/YOUR_SCRIPT_ID/exec"}
    echo.
    echo Run the application:
    echo    customer-survey.exe
) else (
    echo.
    echo ❌ Build failed!
    exit /b 1
)

pause
