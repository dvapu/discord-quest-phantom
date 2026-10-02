@echo off
chcp 65001 >nul
title Cài Đặt Token Discord - Discord Quest Phantom
cls
echo ==================================================================
echo    👻 DISCORD QUEST PHANTOM - THIẾT LẬP TOKEN NHANH
echo ==================================================================
echo.
echo Hướng dẫn lấy token:
echo   1. Mở trình duyệt vào https://discord.com/app và đăng nhập.
echo   2. Nhấn F12, chọn tab Console.
echo   3. Dán lệnh trích xuất token và copy chuỗi token (không lấy dấu ngoặc kép).
echo.
echo ==================================================================
set /p TOKEN="Dán (Paste) Discord Token của bạn vào đây rồi ấn Enter: "

if "%TOKEN%"=="" (
    echo.
    echo [!] Token không được để trống!
    pause
    exit /b
)

:: Loại bỏ dấu ngoặc kép nếu người dùng vô tình dán kèm
set TOKEN=%TOKEN:"=%

echo %TOKEN%> .token

echo.
echo [OK] Đã lưu token thành công vào file .token!
echo.
set /p RUN_NOW="Bạn có muốn khởi động tool ngay bây giờ không? (y/n): "
if /i "%RUN_NOW%"=="y" (
    if exist discord-quest-phantom.exe (
        start discord-quest-phantom.exe
    ) else if exist discord-quest-completer.exe (
        start discord-quest-completer.exe
    ) else (
        python main.py
    )
)
