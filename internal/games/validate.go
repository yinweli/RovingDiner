package games

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// Validate 走訪命令 AST、逐名查 cores 命令詞彙表是否登錄。
// 操作命令查 verb / 命令對象(M6+ 詞條到位後生效);屬性修改命令查左值可寫性(M7)——
// 全域左值查 cores.HasAttrWrite、引用左值的屬性查 cores.HasAttrRefWrite(唯讀 / 未知屬性即報錯)。
// 引用左值的引用基底是否為合法物件引用、[N] 個數(M8)、命令參數數量(M9)留待後續。
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
		if c.isRef {
			if cores.HasAttrRefWrite(c.refAttr) == false {
				return &exprs.SyntaxError{Pos: c.refAttrPos, Msg: "不可寫入的引用屬性：" + c.refAttr}
			} // if
		} else {
			if cores.HasAttrWrite(c.base) == false {
				return &exprs.SyntaxError{Pos: c.basePos, Msg: "不可寫入的屬性：" + c.base}
			} // if
		} // if
	} // switch

	return nil
}
