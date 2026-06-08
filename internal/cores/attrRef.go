package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// execAttrRef 引用屬性詞條的讀取行為:自 ref(卡牌 / 顧客)以 engine 為 context 取子屬性;arg 供引用查詢函式。
type execAttrRef func(eng *engine, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool)

// tableAttrRef 引用屬性詞彙表(名稱 → 讀取行為);由本檔以 map 字面值填入,M5 為空。
// TODO(M6): 填入 calm / sate / cost … 等引用屬性的讀取行為。
var tableAttrRef = map[string]execAttrRef{}

// HasAttrRef 回報引用屬性詞彙表是否登錄 name;供 games.Validate 於 M7 檢查引用屬性可寫性。M5 表空恆 false。
func HasAttrRef(name string) bool {
	_, ok := tableAttrRef[name]
	return ok
}
