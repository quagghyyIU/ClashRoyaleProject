@echo off
echo Building Go applications...
cd tcr

echo Building server...
go build -o server.exe ./cmd/server
if errorlevel 1 (
    echo Failed to build server!
    pause
    exit /b %errorlevel%
)

echo Building client...
go build -o client.exe ./cmd/client
if errorlevel 1 (
    echo Failed to build client!
    pause
    exit /b %errorlevel%
)

echo Launching server and clients...

REM Launch Server
start "TCR Server" /D "%~dp0tcr" server.exe

REM Wait a couple of seconds for the server to initialize
timeout /t 2 /nobreak > nul

REM Launch Client 1
start "TCR Client 1 (Player A)" /D "%~dp0tcr" client.exe

REM Launch Client 2
start "TCR Client 2 (Player B)" /D "%~dp0tcr" client.exe

echo All components launched.
cd .. 