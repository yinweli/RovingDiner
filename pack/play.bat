@echo off
REM ASCII only in this file: cmd parses .bat files with the console codepage, so
REM UTF-8 CJK comments break parsing on CP950 consoles (DBCS swallows ASCII bytes).
REM chcp 65001: rodi prints UTF-8 (seed line, errors); switch so it displays right.
chcp 65001 >nul

REM Start the game at stage 1; pause after exit so the seed stays visible for replay.
cd /d %~dp0
rodi.exe --stage 1
pause
