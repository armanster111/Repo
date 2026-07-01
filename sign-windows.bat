@echo off
REM Optional code signing template.
REM You need a valid Authenticode certificate (.pfx) to stop SmartScreen warnings.

if "%SIGN_CERT%"=="" (
  echo Set SIGN_CERT to your .pfx path and SIGN_PASSWORD to sign the executable.
  exit /b 1
)

signtool sign /fd SHA256 /f "%SIGN_CERT%" /p "%SIGN_PASSWORD%" dist\music-visualizer.exe
signtool verify /pa dist\music-visualizer.exe
