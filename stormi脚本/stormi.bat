@echo off

if "%1"=="upload" (
    git add .
    git commit -m v1.1.7
    git push origin master
    exit
)


