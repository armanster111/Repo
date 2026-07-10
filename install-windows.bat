@echo off
setlocal
title Install Muse

set "SOURCE=%~dp0Muse.exe"
if not exist "%SOURCE%" set "SOURCE=%~dp0dist\Muse.exe"
if not exist "%SOURCE%" set "SOURCE=%~dp0dist\muse.exe"
if not exist "%SOURCE%" (
  echo Could not find Muse.exe next to this installer.
  echo Download Muse-Windows.zip, extract it, then run this file again.
  pause
  exit /b 1
)

set "TARGET=%LOCALAPPDATA%\Muse"
if not exist "%TARGET%" mkdir "%TARGET%"

echo Installing to %TARGET% ...
copy /Y "%SOURCE%" "%TARGET%\Muse.exe" >nul

powershell -NoProfile -ExecutionPolicy Bypass -Command "$s=(New-Object -COM WScript.Shell).CreateShortcut([Environment]::GetFolderPath('Desktop') + '\Muse.lnk'); $s.TargetPath='%TARGET%\Muse.exe'; $s.WorkingDirectory='%TARGET%'; $s.Save()"

echo Installed. Desktop shortcut created.
echo You can also run Muse.exe directly without installing.
choice /C YN /M "Open Muse now"
if not errorlevel 2 start "" "%TARGET%\Muse.exe"
pause
