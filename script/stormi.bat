if "%1"=="upload" (
    git add .
    git commit -m %2
    git push origin master
    exit
)
