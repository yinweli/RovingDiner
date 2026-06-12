package roditool

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
	"github.com/yinweli/RovingDiner/internal/games"
)

// 單筆檢查引擎(【營業實作規格書 | 附錄：企劃驗證器】): 三種文法各一個入口, 依欄位種類分派;
// 與表單檢查(checkEffect 的文法欄)共用同一驗證路徑——兩個工具 = 同一引擎的兩種輸入模式。
// 回 nil 即通過; 錯誤為帶位置中文(exprs.SyntaxError 系)。參照類檢查需整份表, 屬表單檢查。

// CheckExpr 單筆檢查運算式欄位(觸發條件 / 觸發次數): 文法(exprs.Parse)+ 詞彙(games.ValidateExpr)。
func CheckExpr(source string) error {
	expr, err := exprs.Parse(source)

	if err != nil {
		return err
	} // if

	return games.ValidateExpr(expr)
}

// CheckCommand 單筆檢查命令欄位(立即 / 觸發 / 啟動 / 結束命令): 文法(games.Parse)+ 詞彙與參數(games.Validate)。
func CheckCommand(source string) error {
	command, err := games.Parse(source)

	if err != nil {
		return err
	} // if

	return games.Validate(command)
}

// CheckThreshold 單筆檢查門檻配對欄位(門檻值^技能編號): 格式(cores.ParseThreshold 單一來源;
// 技能編號參照需技能表, 由表單檢查把關)。
func CheckThreshold(source string) error {
	_, err := cores.ParseThreshold(source)
	return err
}
