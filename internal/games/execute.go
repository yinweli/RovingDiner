package games

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// execute 派發單一命令至引擎執行(走法 X:games 持 AST 型別 switch、engine 做行為分派)。
// commandAssign → engine.ExecAssign(屬性修改命令,M7);commandOperate → 操作命令執行(M9 接入)。
// 對應【營業實作規格書 | 二、套件結構】games 對外執行面;名稱合法性應先經 Validate。
func execute(game *cores.Game, command Command) {
	switch c := command.(type) {
	case commandAssign:
		game.ExecAssign(c.base, c.refAttr, c.isRef, c.op, c.value)

	case commandOperate:
		game.ExecOperate(c.verb, c.selector.selector, c.selector.param, c.arg)
	} // switch
}
