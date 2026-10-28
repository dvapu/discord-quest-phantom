@echo off
rem ===========================================================================
rem IMPORTANT: This file MUST stay CRLF-encoded and ASCII-only.
rem cmd.exe parses batch files byte-by-byte: LF-only line endings corrupt
rem commands inside ( ) blocks ("echo" -> "ho", "if exist" -> "xist") and the
rem window closes instantly before any `pause`. .gitattributes + CI enforce this.
rem Delayed expansion is intentionally NOT enabled, because it strips literal
rem "!" characters from echo lines (e.g. "[!]" would print as "[").
rem ===========================================================================
setlocal EnableExtensions
cd /d "%~dp0"
title Discord Quest Phantom

rem NOTE: no `chcp 65001` here. This launcher only prints ASCII, and setting the
rem UTF-8 codepage breaks `set /p` inside setup_token.cmd (pasted token is lost).
rem The codepage is switched to UTF-8 only right before launching the app.

echo ==================================================================
echo    Discord Quest Phantom - Windows Launcher
echo ==================================================================
echo.

rem ---- 1. Token check ------------------------------------------------------
if not exist ".token" (
    echo  [!] Chua tim thay file .token / .token file not found!
    echo  [*] Dang khoi dong trinh thiet lap token...
    echo.
    if not exist "setup_token.cmd" goto :missing_setup
    call "setup_token.cmd"
)

if not exist ".token" goto :no_token

rem ---- 2. Locate the executable --------------------------------------------
if exist "discord-quest-phantom.exe" (
    set "APP=discord-quest-phantom.exe"
    goto :run
)
if exist "discord-quest-completer.exe" (
    set "APP=discord-quest-completer.exe"
    goto :run
)

rem ---- 3. Source checkout fallback: build with Go if available -------------
where go >nul 2>&1
if not errorlevel 1 (
    if exist "cmd\completer\main.go" (
        echo  [*] Source tree detected, compiling discord-quest-phantom.exe...
        go build -ldflags="-s -w" -o "discord-quest-phantom.exe" ".\cmd\completer"
        if exist "discord-quest-phantom.exe" (
            echo  [+] Build succeeded.
            set "APP=discord-quest-phantom.exe"
            goto :run
        )
        echo  [-] Build failed.
    )
)

rem ---- 4. Python fallback --------------------------------------------------
where python >nul 2>&1
if not errorlevel 1 (
    if exist "main.py" (
        echo  [*] Falling back to Python runner...
        python main.py
        goto :end
    )
)

rem ---- 5. Nothing runnable -------------------------------------------------
echo.
echo ==================================================================
echo  [!] EXECUTABLE NOT FOUND / KHONG TIM THAY FILE THUC THI
echo ==================================================================
echo.
echo  This folder does not contain discord-quest-phantom.exe.
echo  Thu muc nay khong chua file discord-quest-phantom.exe.
echo.
echo  Please download the pre-built Windows ZIP (no Go/Python needed):
echo  Vui long tai ban dong goi san (khong can cai Go hay Python):
echo.
echo     https://github.com/dvapu/discord-quest-phantom/releases
echo.
echo  Then unzip it and double-click run.cmd.
echo  Giai nen file zip roi nhap dup vao run.cmd.
echo ==================================================================
echo.
pause
goto :end

:missing_setup
echo  [-] setup_token.cmd is missing from this folder.
echo  [-] Thieu file setup_token.cmd trong thu muc nay.
echo.
pause
goto :end

:no_token
echo.
echo  [-] Chua co token / No token provided.
echo  [-] Vui long chay lai sau khi da co token.
echo.
pause
goto :end

:run
rem UTF-8 only for the app itself, so emoji/box-drawing render correctly.
chcp 65001 >nul 2>&1
echo.
echo  [*] Starting %APP% ...
echo.
rem -keep-open=false: this launcher already pauses, so avoid a double prompt.
rem Any user-supplied flag comes afterwards and therefore wins.
"%APP%" -keep-open=false %*
set "EXITCODE=%errorlevel%"
echo.
echo ==================================================================
if not "%EXITCODE%"=="0" (
    echo  [!] Program exited with code %EXITCODE%.
    echo  [!] Chuong trinh ket thuc voi ma loi %EXITCODE%.
    echo.
    echo  Common causes / Nguyen nhan thuong gap:
    echo    - Invalid or expired Discord token - re-run setup_token.cmd
    echo    - No internet connection
    echo    - Discord changed its API - check for a newer release
) else (
    echo  [+] Finished successfully / Hoan tat thanh cong.
)
echo ==================================================================
echo.
pause
goto :end

:end
endlocal
exit /b 0
