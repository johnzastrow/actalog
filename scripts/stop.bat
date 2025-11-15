@echo off
setlocal enabledelayedexpansion

echo === Stopping ActaLog Services ===
echo.

REM Get the current directory
set "PROJECT_DIR=%~dp0"
set "PROJECT_NAME=actionlog"

REM Kill backend processes
echo Stopping backend...

REM Kill actalog.exe processes
for /f "tokens=2" %%a in ('tasklist /FI "IMAGENAME eq actalog.exe" /FO LIST ^| find "PID:"') do (
    echo   Killing actalog process: %%a
    taskkill /F /PID %%a 2>nul
)

REM Kill go.exe processes that might be running the server
for /f "tokens=1,2" %%a in ('wmic process where "name='go.exe'" get ProcessId^,CommandLine /format:list ^| findstr /i "ProcessId CommandLine"') do (
    set "line=%%a"
    if "!line:~0,10!"=="CommandLine" (
        REM Check for actionlog or actalog in command line
        if not "!line!"=="!line:actionlog=!" (
            set "next_is_pid=1"
        )
        if not "!line!"=="!line:actalog=!" (
            set "next_is_pid=1"
        )
        REM Check for cmd/actalog path
        if not "!line!"=="!line:cmd\actalog=!" (
            set "next_is_pid=1"
        )
    )
    if "!line:~0,9!"=="ProcessId" (
        if "!next_is_pid!"=="1" (
            for /f "tokens=2 delims==" %%p in ("%%a") do (
                echo   Killing go process: %%p
                taskkill /F /PID %%p 2>nul
            )
            set "next_is_pid="
        )
    )
)

REM Kill any remaining backend processes on any port (fallback sweep)
REM This catches instances running on non-standard ports
for /f "tokens=5" %%a in ('netstat -aon ^| find "LISTENING"') do (
    set "pid=%%a"
    if defined pid (
        REM Check if this PID belongs to actalog or go.exe with our project
        for /f "tokens=1" %%b in ('tasklist /FI "PID eq !pid!" /NH 2^>nul ^| findstr /i "actalog.exe go.exe"') do (
            echo   Killing backend process on port: !pid!
            taskkill /F /PID !pid! 2>nul
        )
    )
)

echo ✓ Backend stopped

REM Kill frontend processes
echo Stopping frontend...

REM Kill node.exe processes related to this project
for /f "tokens=1,2" %%a in ('wmic process where "name='node.exe'" get ProcessId^,CommandLine /format:list ^| findstr /i "ProcessId CommandLine"') do (
    set "line=%%a"
    if "!line:~0,10!"=="CommandLine" (
        REM Check for vite (dev server)
        if not "!line!"=="!line:vite=!" (
            set "next_is_pid=1"
        )
        REM Check for npm with dev command
        if not "!line!"=="!line:npm=!" (
            if not "!line!"=="!line:dev=!" (
                set "next_is_pid=1"
            )
        )
        REM Check for project directory paths
        if not "!line!"=="!line:actionlog=!" (
            set "next_is_pid=1"
        )
        if not "!line!"=="!line:\web=!" (
            if not "!line!"=="!line:node=!" (
                set "next_is_pid=1"
            )
        )
    )
    if "!line:~0,9!"=="ProcessId" (
        if "!next_is_pid!"=="1" (
            for /f "tokens=2 delims==" %%p in ("%%a") do (
                echo   Killing node process: %%p
                taskkill /F /PID %%p 2>nul
            )
            set "next_is_pid="
        )
    )
)

REM Kill any remaining frontend processes on any port (fallback sweep)
REM This catches dev servers running on non-standard ports
for /f "tokens=5" %%a in ('netstat -aon ^| find "LISTENING"') do (
    set "pid=%%a"
    if defined pid (
        REM Check if this PID belongs to node.exe
        for /f "tokens=1" %%b in ('tasklist /FI "PID eq !pid!" /NH 2^>nul ^| findstr /i "node.exe"') do (
            REM Additional check: see if it's a dev server port range (3000-6000)
            for /f "tokens=2 delims=:" %%c in ('netstat -ano ^| findstr "!pid!" ^| findstr "LISTENING"') do (
                set "port=%%c"
                set "port=!port: =!"
                if !port! geq 3000 if !port! leq 6000 (
                    echo   Killing node process on port !port!: !pid!
                    taskkill /F /PID !pid! 2>nul
                    goto :next_frontend_pid
                )
            )
            :next_frontend_pid
        )
    )
)

echo ✓ Frontend stopped

echo.
echo ✓ All services stopped
pause
