package games

import (
	"fmt"

	"github.com/yinweli/RovingDiner/internal/exprs"
	"github.com/yinweli/RovingDiner/internal/rules"
)

// Validate 走訪命令 AST、逐名查 rules 命令詞彙表是否登錄(詞彙 SSOT 在 rules、查名不需建構 Game)。
// 操作命令查 verb / 命令對象 / 命令對象 [...] 參數數量(依詞彙表 arity, 【營業規格書 | 二十四、命令對象清單 | 參數規則】
// 裸寫 / 數量不符視為語法錯誤; 執行期解析仍寬鬆 no-op); 屬性修改命令查左值可寫性——
// 全域左值查 rules.HasAttrWrite、引用左值的屬性查 rules.HasAttrRefWrite(唯讀 / 未知屬性即報錯)。
// 引用左值的引用基底是否為合法物件引用(HasObjectRef)、命令(verb)參數數量(M9)留待後續。
func Validate(command Command) error {
	switch c := command.(type) {
	case commandOperate:
		if rules.HasCommand(c.verb) == false {
			return &exprs.SyntaxError{Pos: c.verbPos, Msg: "未知的操作命令:" + c.verb}
		} // if

		if rules.HasSelector(c.selector.selector) == false {
			return &exprs.SyntaxError{Pos: c.selector.selectorPos, Msg: "未知的命令對象:" + c.selector.selector}
		} // if

		arity, _ := rules.SelectorArity(c.selector.selector)

		if len(c.selector.param) != arity {
			return &exprs.SyntaxError{Pos: c.selector.selectorPos, Msg: fmt.Sprintf(
				"命令對象 %v 的參數數量不符: 需要 %v 個、實得 %v 個", c.selector.selector, arity, len(c.selector.param))}
		} // if

	case commandAssign:
		if c.isRef {
			if rules.HasAttrRefWrite(c.refAttr) == false {
				return &exprs.SyntaxError{Pos: c.refAttrPos, Msg: "不可寫入的引用屬性:" + c.refAttr}
			} // if
		} else {
			if rules.HasAttrWrite(c.base) == false {
				return &exprs.SyntaxError{Pos: c.basePos, Msg: "不可寫入的屬性:" + c.base}
			} // if
		} // if
	} // switch

	return nil
}
