@echo off
setlocal

where go >nul 2>nul
if errorlevel 1 (
  echo Go is not installed. Install it from https://go.dev/dl/ and run this file again.
  pause
  exit /b 1
)

if not exist dist mkdir dist
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-H windowsgui" -o dist\music-visualizer.exe .\cmd\music-visualizer
if errorlevel 1 (
  echo Build failed.
  pause
  exit /b 1
)

echo Built dist\music-visualizer.exe
pause
