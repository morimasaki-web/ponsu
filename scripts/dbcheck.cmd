@echo off
setlocal enabledelayedexpansion

REM Wrapper to avoid PowerShell ExecutionPolicy issues.
REM Usage:
REM   ponsu\scripts\dbcheck.cmd

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0dbcheck.ps1"
exit /b %ERRORLEVEL%
