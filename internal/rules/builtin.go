package rules

import (
	"slices"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

// builtinEntry 內建函式詞條: 運算行為 + 最少參數數量。行為與 metadata 同詞條單一定義點
// (比照 attrReadEntry; M28 R4A), 供 games.ValidateExpr 靜態校驗(內建函式為 varargs, 故為下限而非定數)。
type builtinEntry struct {
	run    exprs.Builtin // 運算行為(builtin* 具名函式)
	minArg int           // 最少參數數量(【營業規格書 | 二十六、內建函式清單】參數數量欄)
}

// builtin 內建函式註冊表(min / max); 對應【營業規格書 | 二十六、內建函式清單】。
// 內建函式為純運算、不讀系統狀態, 與 attrRead / attrWrite 等同屬套件層全域詞彙表(建後唯讀、等同常數);
// 由 Game 於求值期經 Env() 帶入 exprs.Env, 保持 exprs 自身 game-agnostic。
// 每一詞條對應一個獨立的 builtin* 具名函式(比照讀寫詞彙表), 參數校驗共用 numArg。
var builtin = map[string]builtinEntry{
	"min": {run: builtinMin, minArg: 2},
	"max": {run: builtinMax, minArg: 2},
}

// BuiltinMinArg 回報內建函式詞條的最少參數數量(查無回 ok=false); 供 games.ValidateExpr 靜態校驗
// (M28 R4A; 內建函式優先於查詢函式, 與求值路徑同序)。
func BuiltinMinArg(name string) (minArg int, ok bool) {
	entry, okEntry := builtin[name]

	if okEntry == false {
		return 0, false
	} // if

	return entry.minArg, true
}

// builtinMin 取所有參數的最小值; 少於 2 個參數或任一非數值即失敗(對齊【二十六】評估失敗)。
func builtinMin(arg []exprs.Value) (result exprs.Value, ok bool) {
	num, valid := numArg(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	return exprs.NewNum(slices.Min(num)), true
}

// builtinMax 取所有參數的最大值; 少於 2 個參數或任一非數值即失敗(對齊【二十六】評估失敗)。
func builtinMax(arg []exprs.Value) (result exprs.Value, ok bool) {
	num, valid := numArg(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	return exprs.NewNum(slices.Max(num)), true
}
