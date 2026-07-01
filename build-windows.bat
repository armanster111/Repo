@echo off
setlocal

title Build Music Visualizer Pro

where go >nul 2>nul
if errorlevel 1 (
  echo Go is not installed.
  echo.
  echo This script can install Go from the official Windows package using winget.
  echo If Windows asks for permission, choose Yes.
  echo.
  choice /C YN /M "Install Go now"
  if errorlevel 2 (
    echo Install Go from https://go.dev/dl/ then run this file again.
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
    echo Go installation failed. Install Go from https://go.dev/dl/ then run this file again.
    pause
    exit /b 1
  )

  set "PATH=%PATH%;C:\Program Files\Go\bin;%USERPROFILE%\go\bin"
  where go >nul 2>nul
  if errorlevel 1 (
    echo Go was installed, but Windows has not refreshed PATH yet.
    echo Close this window, open the extracted folder again, and double-click build-windows.bat once more.
    pause
    exit /b 1
  )
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
echo.
echo Creating a Desktop shortcut...
powershell -NoProfile -ExecutionPolicy Bypass -Command "$ErrorActionPreference='Stop'; $target=(Resolve-Path 'dist\music-visualizer.exe').Path; $workdir=(Resolve-Path 'dist').Path; $shortcut=[Environment]::GetFolderPath('Desktop') + '\Music Visualizer Pro.lnk'; $s=(New-Object -COM WScript.Shell).CreateShortcut($shortcut); $s.TargetPath=$target; $s.WorkingDirectory=$workdir; $s.Save()"
if errorlevel 1 (
  echo Shortcut creation failed, but the app was built successfully.
  echo Open dist\music-visualizer.exe from this folder.
) else (
  echo Desktop shortcut created.
)
echo.
choice /C YN /M "Open Music Visualizer Pro now"
if errorlevel 2 (
  echo Done. You can open Music Visualizer Pro later from dist\music-visualizer.exe.
) else (
  start "" "dist\music-visualizer.exe"
)
pause
