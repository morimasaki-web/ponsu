@echo off
setlocal enabledelayedexpansion

REM Wrapper to avoid PowerShell ExecutionPolicy issues.
REM Usage examples:
REM   ponsu\scripts\migrate.cmd up
REM   ponsu\scripts\migrate.cmd down
REM   ponsu\scripts\migrate.cmd up 1

set "CMD=%~1"
if "%CMD%"=="" set "CMD=up"

set "STEPS=%~2"
if "%STEPS%"=="" set "STEPS=0"

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0migrate.ps1" -Command "%CMD%" -Steps %STEPS%
exit /b %ERRORLEVEL%
