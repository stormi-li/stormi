@echo off
setlocal enabledelayedexpansion

REM 获取所有包含 "nsqd" 的进程的PID
for /f "tokens=2" %%a in ('tasklist ^| findstr nsqd') do (
    set /a "PID=%%a"
    echo Terminating nsqd process with PID !PID!...
    taskkill /PID !PID! /F > nul
)

echo All nsqd processes terminated.