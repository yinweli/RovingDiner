package games

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// Validate 走訪命令 AST、逐名查 cores 命令詞彙表是否登錄。M5 為空表故任何名稱皆視為未知;詞條於 M6+ 填入後生效。
// 屬性可寫性(cores.HasAttr / cores.HasAttrRef)留待 M7、[N] 個數留待 M8、命令參數數量留待 M9。
func Validate(command Command) error {
	switch c := command.(type) {
	case commandOperate:
		if cores.HasCommand(c.verb) == false {
			return &exprs.SyntaxError{Pos: c.verbPos, Msg: "未知的操作命令：" + c.verb}
		} // if

		if cores.HasSelector(c.selector.selector) == false {
			return &exprs.SyntaxError{Pos: c.selector.selectorPos, Msg: "未知的命令對象：" + c.selector.selector}
		} // if

	case commandAssign:
		// 屬性可寫性(全域 cores.HasAttr / 引用 cores.HasAttrRef)留待 M7
	} // switch

	return nil
}
