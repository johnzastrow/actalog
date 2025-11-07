@echo off
REM ActaLog Build Script for Windows
REM This script provides an alternative to Makefile for Windows users

setlocal

REM Set project-local cache directories to avoid Windows permission issues
set PROJECT_DIR=%~dp0
set CACHE_DIR=%PROJECT_DIR%.cache
set GOCACHE=%CACHE_DIR%\go-build
set GOMODCACHE=%CACHE_DIR%\go-mod
set GOTMPDIR=%CACHE_DIR%\tmp

REM Create directories if they don't exist
if not exist bin mkdir bin
if not exist %CACHE_DIR% mkdir %CACHE_DIR%
if not exist %GOCACHE% mkdir %GOCACHE%
if not exist %GOMODCACHE% mkdir %GOMODCACHE%
if not exist %GOTMPDIR% mkdir %GOTMPDIR%

if "%1"=="" goto help
if "%1"=="build" goto build
if "%1"=="run" goto run
if "%1"=="test" goto test
if "%1"=="clean" goto clean
if "%1"=="fmt" goto fmt
if "%1"=="help" goto help
goto help

:build
echo Building ActaLog...
go build -o bin\actalog.exe cmd\actalog\main.go
if %errorlevel% equ 0 (
    echo Build complete: bin\actalog.exe
) else (
    echo Build failed!
    exit /b 1
)
goto end

:run
echo Running ActaLog...
go run cmd\actalog\main.go
goto end

:test
echo Running tests...
go test -v -race -coverprofile=coverage.out ./...
if %errorlevel% equ 0 (
    go tool cover -html=coverage.out -o coverage.html
    echo Coverage report generated: coverage.html
)
goto end

:clean
echo Cleaning build artifacts...
if exist bin rmdir /s /q bin
if exist .cache rmdir /s /q .cache
if exist coverage.out del coverage.out
if exist coverage.html del coverage.html
if exist *.db del *.db
if exist *.sqlite del *.sqlite
if exist *.sqlite3 del *.sqlite3
echo Clean complete
goto end

:fmt
echo Formatting code...
go fmt ./...
goto end

:help
echo ActaLog - Windows Build Script
echo.
echo Usage: build.bat [command]
echo.
echo Available commands:
echo   build      Build the application
echo   run        Run the application
echo   test       Run tests with coverage
echo   clean      Clean build artifacts
echo   fmt        Format Go code
echo   help       Show this help message
echo.
goto end

:end
endlocal
