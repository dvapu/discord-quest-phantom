@echo off
chcp 65001 >nul
title Setup Discord Token - Discord Quest Phantom
cls
echo ==================================================================
echo    👻 DISCORD QUEST PHANTOM - QUICK TOKEN SETUP / THIẾT LẬP TOKEN
echo ==================================================================
echo.
echo [EN] How to get your Discord Token:
echo   1. Open https://discord.com/app in your browser and log in.
echo   2. Press F12 -> Go to Console tab (type 'allow pasting' if blocked).
echo   3. Run the extraction script from README and copy the token string.
echo.
echo [VI] Hướng dẫn lấy token:
echo   1. Mở trình duyệt vào https://discord.com/app và đăng nhập.
echo   2. Nhấn F12 -> Vào tab Console (gõ 'allow pasting' nếu bị chặn dán).
echo   3. Dán đoạn mã trích xuất token và sao chép chuỗi token nhận được.
echo.
echo ==================================================================
set /p TOKEN="Paste Discord Token here / Dán Token vào đây rồi ấn Enter: "

if "%TOKEN%"=="" (
    echo.
    echo [!] Token cannot be empty / Token không được để trống!
    pause
    exit /b
)

:: Strip surrounding quotes
set TOKEN=%TOKEN:"=%

echo %TOKEN%> .token

echo.
echo [OK] Token saved successfully to .token / Đã lưu token thành công!
echo.
set /p RUN_NOW="Start tool now? / Khởi động tool ngay bây giờ? (y/n): "
if /i "%RUN_NOW%"=="y" (
    if exist discord-quest-phantom.exe (
        start discord-quest-phantom.exe
    ) else if exist discord-quest-completer.exe (
        start discord-quest-completer.exe
    ) else (
        python main.py
    )
)
