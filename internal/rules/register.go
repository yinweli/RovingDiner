package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// Register 對營業實例裝備全部詞彙(屬性讀寫 / 操作命令 / 命令對象 / 內建函式);
// games 於建構 Game 後呼叫一次。詞彙字面值由各概念檔自持(attrRead.go 等), 本入口僅逐詞條餵給 Game.Register*。
func Register(game *cores.Game) {
	for k, v := range attrRead {
		game.RegisterAttrRead(k, v)
	} // for

	for k, v := range attrWrite {
		game.RegisterAttrWrite(k, v)
	} // for

	for k, v := range attrRefRead {
		game.RegisterAttrRefRead(k, v)
	} // for

	for k, v := range attrRefWrite {
		game.RegisterAttrRefWrite(k, v)
	} // for

	for k, v := range command {
		game.RegisterCommand(k, v)
	} // for

	for k, v := range selector {
		game.RegisterSelector(k, v.resolve) // 引擎只需解析行為; arity 屬 Validate 期靜態知識, 不入 Game
	} // for

	for k, v := range builtin {
		game.RegisterBuiltin(k, v)
	} // for
}
