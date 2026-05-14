@echo off
setlocal enabledelayedexpansion

REM Wrapper to avoid PowerShell ExecutionPolicy issues.
REM Usage:
REM   ponsu\scripts\sqlc.cmd generate
REM   ponsu\scripts\sqlc.cmd version

set "CMD=%~1"
if "%CMD%"=="" set "CMD=generate"

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0sqlc.ps1" -Command "%CMD%"
exit /b %ERRORLEVEL%
