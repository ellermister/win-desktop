@echo off
setlocal

set "EXE=win-desktop.exe"

if not exist "upx.exe" (
  echo Error: upx.exe not found in current directory.
  exit /b 1
)

if not exist "%EXE%" (
  echo Error: %EXE% not found.
  exit /b 1
)

upx.exe --best --lzma "%EXE%"
exit /b %ERRORLEVEL%
