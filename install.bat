@echo off
setlocal enabledelayedexpansion

echo Harvest CLI Utility Installation Script for Windows
echo ===================================================
echo.

:: Check if Go is installed
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Error: Go is not installed or not in your PATH.
    echo Please install Go from https://golang.org/dl/ and try again.
    exit /b 1
)

:: Check Go version
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
set GO_VERSION=%GO_VERSION:go=%
for /f "tokens=1,2 delims=." %%a in ("%GO_VERSION%") do (
    set GO_MAJOR=%%a
    set GO_MINOR=%%b
)

if %GO_MAJOR% LSS 1 (
    echo Warning: Go version %GO_VERSION% detected.
    echo This project recommends Go 1.18 or higher.
    set /p CONTINUE="Continue anyway? (y/n): "
    if /i "!CONTINUE!" neq "y" exit /b 1
) else (
    if %GO_MAJOR% EQU 1 (
        if %GO_MINOR% LSS 18 (
            echo Warning: Go version %GO_VERSION% detected.
            echo This project recommends Go 1.18 or higher.
            set /p CONTINUE="Continue anyway? (y/n): "
            if /i "!CONTINUE!" neq "y" exit /b 1
        )
    )
)

echo Go version %GO_VERSION% detected. ✓
echo.

:: Build the project
echo Building Harvest CLI utility...
go build -o h.exe

:: Determine installation path
set "INSTALL_DIR=%USERPROFILE%\bin"
set "INSTALL_PATH=%INSTALL_DIR%\h.exe"

:: Create the bin directory if it doesn't exist
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

:: Copy the binary to the installation path
echo Installing to %INSTALL_PATH%...
copy /Y h.exe "%INSTALL_PATH%" >nul

:: Check if the directory is in PATH
echo Checking if %INSTALL_DIR% is in your PATH...
echo %PATH% | findstr /C:"%INSTALL_DIR%" >nul
if %ERRORLEVEL% neq 0 (
    echo Adding %INSTALL_DIR% to your PATH...
    
    :: Add to user PATH permanently
    setx PATH "%PATH%;%INSTALL_DIR%"
    
    echo Please restart your command prompt to use the 'h' command.
) else (
    echo %INSTALL_DIR% is already in your PATH. ✓
)

:: Check if config.json exists
if not exist "config.json" (
    echo Creating sample config.json...
    (
        echo {
        echo   "projects": [
        echo     {
        echo       "id": 123,
        echo       "name": "Project A",
        echo       "tasks": [
        echo         {
        echo           "id": 456,
        echo           "name": "Software Development"
        echo         },
        echo         {
        echo           "id": 678,
        echo           "name": "Non-Billable"
        echo         }
        echo       ]
        echo     }
        echo   ],
        echo   "default_project": "Project A",
        echo   "default_task": "Software Development",
        echo   "harvest_api": {
        echo     "account_id": "YOUR_HARVEST_ACCOUNT_ID",
        echo     "token": "YOUR_HARVEST_API_TOKEN"
        echo   }
        echo }
    ) > config.json
    
    echo A sample config.json has been created.
    echo Please edit it with your Harvest API credentials before using the tool.
)

echo.
echo Installation complete!
echo You can now use the Harvest CLI utility by running 'h' in your command prompt.
echo For help, run 'h --help'

:: Provide instructions for getting Harvest API credentials
echo.
echo Don't forget to update your config.json with your Harvest API credentials:
echo 1. Log in to your Harvest account
echo 2. Go to Settings ^> Developer
echo 3. Create a new personal access token
echo 4. Note your Account ID and Token
echo 5. Update the config.json file with these values

:: Provide instructions for first use
echo.
echo Quick start:
echo - List today's entries: h list
echo - Create a new entry: h create
echo - Create with defaults: h create -D
echo - Update an entry: h update
echo - Delete entries: h delete

pause 