package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// execSelector 命令對象詞條的解析行為:以 Engine 為 context、arg 為 [...] 內參數,產出作用對象集合。
type execSelector func(eng *Engine, arg []exprs.Value) (result []InstanceID)

// tableSelector 命令對象詞彙表(名稱 → 解析行為);由本檔以 map 字面值填入,M5 為空。
// TODO(M8): 填入 none / self / deckTop … 等命令對象的解析行為(含 N 邊界 / filter / auto-shuffle)。
var tableSelector = map[string]execSelector{}

// HasSelector 回報命令對象詞彙表是否登錄 name;供 games.Validate 檢查命令對象。M5 表空恆 false。
func HasSelector(name string) bool {
	_, ok := tableSelector[name]
	return ok
}
