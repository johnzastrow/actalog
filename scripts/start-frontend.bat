@echo off
REM start-frontend.bat - small wrapper to call the bash guided script if available
echo ActaLog frontend starter — Windows wrapper

where bash >nul 2>&1
if errorlevel 1 (
  echo "bash.exe not found in PATH. Run the script using Git Bash, WSL, or run scripts\start-frontend.sh manually."
  exit /b 1
)

REM Invoke the bash guided script
bash "%~dp0start-frontend.sh"
