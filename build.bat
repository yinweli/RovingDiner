@echo off
REM ASCII only in this file: cmd parses .bat files with the console codepage, so
REM UTF-8 CJK comments break parsing on CP950 consoles (DBCS swallows ASCII bytes).
REM chcp 65001: sheeter/roditool print UTF-8; switch so their output displays right.
chcp 65001 >nul

REM Build sheet code and data, then run the sheet check (roditool).
REM This file sits at the repo/package root and cds into gamedata, where
REM sheeter.yaml and the xlsx live; all relative paths below assume that cwd.
REM Shared by the dev repo and the game install package: tools are called bare;
REM inside the package the PATH line below hits the exes at the package root,
REM in the dev repo the repo root has no exes so lookup falls through to the
REM versions installed by task install.
cd /d %~dp0gamedata
set PATH=%~dp0;%PATH%
set output=output
set targetCode=..\sheet
set targetData=..\sheetdata

echo # Generate code and json
sheeter build --tag BS --config sheeter.yaml
if errorlevel 1 exit /b 1

REM Copy generated Go code (dev repo only; the package ships no go.mod and the
REM sheet readers are already compiled into rodi.exe).
if exist ..\go.mod (
    echo # Copy code
    rmdir /s /q %targetCode% 2>nul
    xcopy %output%\codeGo %targetCode% /e /i /q /y
    if errorlevel 1 exit /b 1
)

echo # Copy json
rmdir /s /q %targetData% 2>nul
xcopy %output%\json %targetData% /e /i /q /y
if errorlevel 1 exit /b 1

echo # Clean up
rmdir /s /q %output%

echo # Sheet check
roditool sheet --data %targetData%
if errorlevel 1 exit /b 1
