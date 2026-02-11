@echo off
echo Building ClashDoc...

REM 1. Build frontend
echo Step 1: Building frontend...
cd web
call pnpm install
call pnpm run build
if %errorlevel% neq 0 (
    echo Frontend build failed!
    exit /b 1
)
cd ..

REM 2. Copy dist to cmd/server for embedding
echo Step 2: Copying dist for embedding...
xcopy /E /I /Y web\dist cmd\server\dist >nul 2>&1
if %errorlevel% neq 0 (
    echo Copy failed!
    exit /b 1
)

REM 3. Build Go application
echo Step 3: Building Go application...
go build -o clash-manager.exe ./cmd/server
if %errorlevel% neq 0 (
    echo Go build failed!
    exit /b 1
)

REM 4. Clean up
echo Step 4: Cleaning up...
rmdir /S /Q cmd\server\dist

echo.
echo Build completed successfully!
echo Output: clash-manager.exe
