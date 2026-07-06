@echo off
REM Optional Authenticode signing for MusicVisualizerPro.exe
REM Requires Windows SDK signtool and a valid .pfx certificate.

set "TARGET=dist\MusicVisualizerPro.exe"
if not exist "%TARGET%" (
  echo Build first: build-windows.bat or scripts\package-windows.sh
  exit /b 1
)

if "%SIGN_CERT%"=="" (
  echo.
  echo Code signing skipped — set environment variables:
  echo   SIGN_CERT     = path to your .pfx certificate
  echo   SIGN_PASSWORD = certificate password
  echo   SIGN_TSURL    = optional timestamp server ^(default: http://timestamp.digicert.com^)
  echo.
  echo Unsigned builds still run via SmartScreen: More info -^> Run anyway
  exit /b 0
)

if "%SIGN_TSURL%"=="" set "SIGN_TSURL=http://timestamp.digicert.com"

echo Signing %TARGET% ...
signtool sign /fd SHA256 /tr "%SIGN_TSURL%" /td SHA256 /f "%SIGN_CERT%" /p "%SIGN_PASSWORD%" "%TARGET%"
if errorlevel 1 (
  echo Signing failed.
  exit /b 1
)

signtool verify /pa "%TARGET%"
if errorlevel 1 (
  echo Verification failed.
  exit /b 1
)

echo Signed and verified: %TARGET%
