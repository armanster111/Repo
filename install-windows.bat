@echo off
setlocal
title Install Music Visualizer Pro

set "TARGET=%LOCALAPPDATA%\MusicVisualizerPro"
if not exist "%TARGET%" mkdir "%TARGET%"

echo Installing to %TARGET% ...
copy /Y "dist\music-visualizer.exe" "%TARGET%\music-visualizer.exe" >nul
if errorlevel 1 (
  echo Build the app first with build-windows.bat
  pause
  exit /b 1
)

powershell -NoProfile -ExecutionPolicy Bypass -Command "$s=(New-Object -COM WScript.Shell).CreateShortcut([Environment]::GetFolderPath('Desktop') + '\Music Visualizer Pro.lnk'); $s.TargetPath='%TARGET%\music-visualizer.exe'; $s.WorkingDirectory='%TARGET%'; $s.Save()"
powershell -NoProfile -ExecutionPolicy Bypass -Command "$s=(New-Object -COM WScript.Shell).CreateShortcut([Environment]::GetFolderPath('Startup') + '\Music Visualizer Pro.lnk'); $s.TargetPath='%TARGET%\music-visualizer.exe'; $s.WorkingDirectory='%TARGET%'; $s.Save()"

echo Installed. Desktop shortcut created.
pause
