@echo off
setlocal

title Build Music Visualizer Pro (developers)

where go >nul 2>nul
if errorlevel 1 (
  echo Go is not installed.
  echo.
  echo Most users should download the prebuilt MusicVisualizerPro.exe instead of building.
  echo Developers can install Go from https://go.dev/dl/ or let this script install it with winget.
  echo.
  choice /C YN /M "Install Go now"
  if errorlevel 2 (
    pause
    exit /b 1
  )

  where winget >nul 2>nul
  if errorlevel 1 (
    echo winget is not available on this PC.
    echo Install Go from https://go.dev/dl/ then run this file again.
    pause
    exit /b 1
  )

  winget install --id GoLang.Go --source winget --accept-package-agreements --accept-source-agreements
  if errorlevel 1 (
    echo Go installation failed.
    pause
    exit /b 1
  )

  set "PATH=%PATH%;C:\Program Files\Go\bin;%USERPROFILE%\go\bin"
  where go >nul 2>nul
  if errorlevel 1 (
    echo Go was installed, but Windows has not refreshed PATH yet.
    echo Close this window, open the folder again, and double-click build-windows.bat once more.
    pause
    exit /b 1
  )
)

if not exist dist mkdir dist
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags="-s -w -H windowsgui" -o dist\MusicVisualizerPro.exe .\cmd\music-visualizer
if errorlevel 1 (
  echo Build failed.
  pause
  exit /b 1
)

echo Built dist\MusicVisualizerPro.exe
echo.
choice /C YN /M "Open Music Visualizer Pro now"
if not errorlevel 2 start "" "dist\MusicVisualizerPro.exe"
pause
