@echo off
setlocal enabledelayedexpansion
if "%1"=="upload" (
    git add .
    git commit -m v1.1.7
    git push origin master
    exit
)


if "%1"=="nsqdkill" (
    for /f "tokens=2" %%a in ('tasklist ^| findstr nsqd') do (
        set /a "PID=%%a"
        echo Terminating nsqd process with PID !PID!...
        taskkill /PID !PID! /F > nul
    )
    exit
)