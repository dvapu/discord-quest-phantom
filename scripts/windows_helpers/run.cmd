@echo off
chcp 65001 >nul
title Discord Quest Phantom
cls
if not exist .token (
    echo [!] Chưa tìm thấy file .token! Chuyển sang trình cài đặt token...
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
