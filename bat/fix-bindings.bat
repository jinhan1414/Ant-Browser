@echo off
chcp 65001 >nul
cd /d "%~dp0.."

echo 修复 wailsjs 绑定文件...

REM 备份原文件
copy /Y "frontend\src\wailsjs\go\models.ts" "frontend\src\wailsjs\go\models.ts.bak" >nul 2>&1
copy /Y "frontend\src\wailsjs\go\main\App.js" "frontend\src\wailsjs\go\main\App.js.bak" >nul 2>&1
copy /Y "frontend\src\wailsjs\go\main\App.d.ts" "frontend\src\wailsjs\go\main\App.d.ts.bak" >nul 2>&1

REM 使用 PowerShell 修复文件
powershell -ExecutionPolicy Bypass -File "bat\fix-bindings.ps1"

if %errorlevel% equ 0 (
    echo ✓ 绑定文件修复成功
) else (
    echo ✗ 绑定文件修复失败
    pause
    exit /b 1
)
