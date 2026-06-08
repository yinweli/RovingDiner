package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// execAttr 全域屬性詞條的讀取行為:以 engine 為 context 求值;arg 供查詢函式型屬性(deckSize…),純屬性忽略。
type execAttr func(eng *engine, arg []exprs.Value) (result exprs.Value, ok bool)

// tableAttr 全域屬性詞彙表(名稱 → 讀取行為);由本檔以 map 字面值填入,M5 為空。
// TODO(M6): 填入 morale / score / deckSize … 等全域屬性的讀取行為。
var tableAttr = map[string]execAttr{}

// HasAttr 回報全域屬性詞彙表是否登錄 name;供 games.Validate 於 M7 檢查屬性可寫性。M5 表空恆 false。
func HasAttr(name string) bool {
	_, ok := tableAttr[name]
	return ok
}
