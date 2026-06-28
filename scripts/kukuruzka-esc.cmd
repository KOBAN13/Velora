@echo off
setlocal
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0kukuruzka-esc.ps1" %*
exit /b %ERRORLEVEL%
