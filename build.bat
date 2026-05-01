@echo off
setlocal

echo [1/3] Generating wire...
cd /d "%~dp0wire"
go run -mod=mod github.com/google/wire/cmd/wire
if %errorlevel% neq 0 (
    echo wire generation failed
    exit /b %errorlevel%
)

echo [2/3] Building...
cd /d "%~dp0"
go build -o bin/house-manager.exe ./cmd/server
if %errorlevel% neq 0 (
    echo build failed
    exit /b %errorlevel%
)

echo [3/3] Done: bin\house-manager.exe
