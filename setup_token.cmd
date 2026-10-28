@echo off
rem ===========================================================================
rem IMPORTANT: This file MUST stay CRLF-encoded and ASCII-only.
rem cmd.exe parses batch files byte-by-byte: LF-only line endings corrupt
rem commands inside ( ) blocks and the window closes instantly.
rem Delayed expansion is intentionally NOT enabled (it strips literal "!").
rem ===========================================================================
setlocal EnableExtensions
cd /d "%~dp0"
title Setup Discord Token - Discord Quest Phantom
cls

rem NOTE: chcp 65001 is deliberately NOT set before reading the token.
rem Switching to the UTF-8 codepage makes `set /p` drop pasted input on some
rem Windows builds. The codepage is restored just before launching the app.

echo ==================================================================
echo    DISCORD QUEST PHANTOM - TOKEN SETUP / THIET LAP TOKEN
echo ==================================================================
echo.
echo [EN] How to get your Discord token:
echo   1. Open https://discord.com/app in your browser and log in.
echo   2. Press F12, then open the Console tab.
echo      If the browser blocks pasting, type: allow pasting
echo   3. Paste the token extraction snippet from README.md and press Enter.
echo   4. Copy the printed token string.
echo.
echo [VI] Huong dan lay token:
echo   1. Mo trinh duyet vao https://discord.com/app va dang nhap.
echo   2. Nhan F12, chuyen sang tab Console.
echo      Neu trinh duyet chan dan, go: allow pasting
echo   3. Dan doan ma trich xuat token trong README.md roi nhan Enter.
echo   4. Sao chep chuoi token vua hien ra.
echo.
echo ==================================================================
echo.

set "TOKEN="
set /p "TOKEN=Paste Discord Token here / Dan Token vao day roi nhan Enter: "

if not defined TOKEN goto :empty

rem Strip surrounding quotes and whitespace.
for /f "tokens=* delims=" %%A in ("%TOKEN%") do set "TOKEN=%%~A"
set "TOKEN=%TOKEN: =%"

if not defined TOKEN goto :empty

rem Discord tokens are ~70 chars; reject obvious junk before saving.
set "TMPT=%TOKEN%"
set "TLEN=0"
:count
if not defined TMPT goto :counted
set "TMPT=%TMPT:~1%"
set /a TLEN+=1
goto :count

:counted
if %TLEN% LSS 20 (
    echo.
    echo  [!] That does not look like a Discord token - only %TLEN% chars.
    echo  [!] Chuoi vua nhap qua ngan, co ve khong phai token Discord.
    echo.
    pause
    exit /b 1
)

rem Write the token with no trailing newline. The parenthesised block avoids
rem cmd.exe's numeric-filename redirection bug ("1234> .token").
> ".token" (
    <nul set /p "=%TOKEN%"
)

if not exist ".token" goto :write_failed

echo.
echo  [OK] Token saved to .token / Da luu token vao file .token
echo.

rem Switch to UTF-8 only now, so the app's emoji/box output renders correctly.
chcp 65001 >nul 2>&1

rem Offer to start the app right away.
echo  [*] Starting Discord Quest Phantom...
echo.

if exist "discord-quest-phantom.exe" (
    "discord-quest-phantom.exe"
    goto :done
)
if exist "discord-quest-completer.exe" (
    "discord-quest-completer.exe"
    goto :done
)

echo  [i] Token saved. Now double-click run.cmd to start.
echo  [i] Da luu token. Hay nhap dup vao run.cmd de chay.
echo.
pause
goto :done

:empty
echo.
echo  [!] Token cannot be empty / Token khong duoc de trong!
echo.
pause
exit /b 1

:write_failed
echo.
echo  [!] Could not write .token / Khong the ghi file .token.
echo  [!] Check folder permissions / Kiem tra quyen ghi thu muc.
echo.
pause
exit /b 1

:done
endlocal
exit /b 0
