@echo off
setlocal
title Install Music Visualizer Pro

set "SOURCE=%~dp0MusicVisualizerPro.exe"
if not exist "%SOURCE%" set "SOURCE=%~dp0dist\MusicVisualizerPro.exe"
if not exist "%SOURCE%" set "SOURCE=%~dp0dist\music-visualizer.exe"
if not exist "%SOURCE%" (
  echo Could not find MusicVisualizerPro.exe next to this installer.
  echo Download MusicVisualizerPro-Windows.zip, extract it, then run this file again.
  pause
  exit /b 1
)

set "TARGET=%LOCALAPPDATA%\MusicVisualizerPro"
if not exist "%TARGET%" mkdir "%TARGET%"

echo Installing to %TARGET% ...
copy /Y "%SOURCE%" "%TARGET%\MusicVisualizerPro.exe" >nul

powershell -NoProfile -ExecutionPolicy Bypass -Command "$s=(New-Object -COM WScript.Shell).CreateShortcut([Environment]::GetFolderPath('Desktop') + '\Music Visualizer Pro.lnk'); $s.TargetPath='%TARGET%\MusicVisualizerPro.exe'; $s.WorkingDirectory='%TARGET%'; $s.Save()"

echo Installed. Desktop shortcut created.
echo You can also run MusicVisualizerPro.exe directly without installing.
choice /C YN /M "Open Music Visualizer Pro now"
if not errorlevel 2 start "" "%TARGET%\MusicVisualizerPro.exe"
pause
