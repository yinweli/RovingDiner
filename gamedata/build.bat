@echo off
REM 編譯表單程序
set output=output
set targetCode=../sheet
set targetData=../sheetdata

REM 產生程式碼與資料檔案
echo # Generate code and json
sheeter build --tag BS --config sheeter.yaml

REM 複製程式碼
echo # Copy code
rm -rf %targetCode%
cp -r %output%/codeGo/ %targetCode%

REM 複製資料檔案
echo # Copy json

for %%i in (%targetData%) do (
    rm -rf %%i
    cp -r %output%/json/ %%i
)

REM 清理暫存檔案
echo # Clean up
rm -rf %output%