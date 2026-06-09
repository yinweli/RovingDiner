package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// commandFunc 操作命令詞條的執行行為:以 Engine 為 context、target 為已解析命令對象、arg 為其餘參數。
type commandFunc func(eng *Engine, target []InstanceID, arg []exprs.Value) (err error)

// command 操作命令詞彙表(名稱 → 執行行為);由本檔以 map 字面值填入,M5 為空。
// TODO(M9): 填入 handAdd / deckShuffle … 等操作命令的執行行為。
var command = map[string]commandFunc{}

// HasCommand 回報操作命令詞彙表是否登錄 name;供 games.Validate 檢查命令 verb。M5 表空恆 false。
func HasCommand(name string) bool {
	_, ok := command[name]
	return ok
}
