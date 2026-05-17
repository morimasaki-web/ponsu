@echo off
setlocal

REM Wrapper for environments where PowerShell script execution is restricted.
REM Uses ExecutionPolicy Bypass for this invocation only.

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0verify.ps1" %*
if errorlevel 1 exit /b %errorlevel%

endlocal
