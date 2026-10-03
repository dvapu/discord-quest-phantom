@echo off
chcp 65001 >nul
title Discord Quest Phantom
cls
if not exist .token (
    echo [!] .token file not found! Launching token setup / Chưa tìm thấy file .token! Đang mở trình cài đặt...
    call setup_token.cmd
    exit /b
)

if exist discord-quest-phantom.exe (
    discord-quest-phantom.exe
) else if exist discord-quest-completer.exe (
    discord-quest-completer.exe
) else (
    python main.py
)
pause
