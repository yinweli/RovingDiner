package games

import (
	"fmt"

	"github.com/yinweli/RovingDiner/internal/exprs"
	"github.com/yinweli/RovingDiner/internal/rules"
)

// Validate 走訪命令 AST、逐名查 rules 命令詞彙表是否登錄(詞彙 SSOT 在 rules、查名不需建構 Game)。
// 操作命令查 verb / 命令對象 / 命令對象 [...] 參數數量(依詞彙表 arity, 【營業規格書 | 二十四、命令對象清單 | 參數規則】
// 裸寫 / 數量不符視為語法錯誤; 執行期解析仍寬鬆 no-op); 屬性修改命令查左值——
// 全域左值查 rules.HasAttrWrite、引用左值查基底合法性(rules.HasObjectRef, 非物件引用 / 未知名稱即報錯)
// 與屬性可寫性(rules.HasAttrRefWrite, 唯讀 / 未知屬性即報錯)。內嵌算術式(右值 / 命令對象參數 / 其餘參數)
// 一律經 ValidateExpr 查運算式詞彙、錯誤位置回算至命令座標(M28 R4A)。命令(verb)參數數量(M9)留待後續。
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

		for index, itor := range c.selector.param {
			if err := validateExprAt(itor, c.selector.paramPos[index]); err != nil {
				return err
			} // if
		} // for

		for index, itor := range c.arg {
			if err := validateExprAt(itor, c.argPos[index]); err != nil {
				return err
			} // if
		} // for

	case commandAssign:
		if c.isRef {
			if rules.HasObjectRef(c.base) == false {
				return &exprs.SyntaxError{Pos: c.basePos, Msg: "不是合法的物件引用:" + c.base}
			} // if

			if rules.HasAttrRefWrite(c.refAttr) == false {
				return &exprs.SyntaxError{Pos: c.refAttrPos, Msg: "不可寫入的引用屬性:" + c.refAttr}
			} // if
		} else if rules.HasAttrWrite(c.base) == false {
			return &exprs.SyntaxError{Pos: c.basePos, Msg: "不可寫入的屬性:" + c.base}
		} // if

		if c.value != nil { // @ # 無右值
			return validateExprAt(c.value, c.valuePos)
		} // if
	} // switch

	return nil
}

// ValidateExpr 走訪運算式名稱使用、逐名查 rules 讀側詞彙表(M28 R4A; 比照 Validate 之於命令, 查名不需建構 Game):
// 識別子 / 查詢函式查全域讀表(名稱 + 參數數量, 無括號與零參數呼叫等價)、內建函式查最少參數數量(求值同序優先),
// 引用屬性查基底合法性(rules.HasObjectRef)與引用讀表(名稱 + 參數數量)。執行期求值仍寬鬆評估失敗。
func ValidateExpr(expr *exprs.Expr) error {
	for _, itor := range expr.Names() {
		if itor.Ref {
			if rules.HasObjectRef(itor.Name) == false {
				return &exprs.SyntaxError{Pos: itor.Pos, Msg: "不是合法的物件引用:" + itor.Name}
			} // if

			arity, ok := rules.AttrRefReadArity(itor.Attr)

			if ok == false {
				return &exprs.SyntaxError{Pos: itor.AttrPos, Msg: "未知的引用屬性:" + itor.Attr}
			} // if

			if itor.Argc != arity {
				return &exprs.SyntaxError{Pos: itor.AttrPos, Msg: fmt.Sprintf(
					"引用屬性 %v 的參數數量不符: 需要 %v 個、實得 %v 個", itor.Attr, arity, itor.Argc)}
			} // if

			continue
		} // if

		if minArg, ok := rules.BuiltinMinArg(itor.Name); ok {
			if itor.Argc < minArg {
				return &exprs.SyntaxError{Pos: itor.Pos, Msg: fmt.Sprintf(
					"內建函式 %v 至少需要 %v 個參數、實得 %v 個", itor.Name, minArg, itor.Argc)}
			} // if

			continue
		} // if

		arity, ok := rules.AttrReadArity(itor.Name)

		if ok == false {
			return &exprs.SyntaxError{Pos: itor.Pos, Msg: "未知的屬性名稱:" + itor.Name}
		} // if

		if itor.Argc != arity {
			return &exprs.SyntaxError{Pos: itor.Pos, Msg: fmt.Sprintf(
				"屬性 %v 的參數數量不符: 需要 %v 個、實得 %v 個", itor.Name, arity, itor.Argc)}
		} // if
	} // for

	return nil
}

// validateExprAt 查內嵌算術式詞彙並把錯誤位置加上 offset 回算至命令來源座標(重用 offsetError)。
func validateExprAt(expr *exprs.Expr, offset int) error {
	if err := ValidateExpr(expr); err != nil {
		return offsetError(err, offset)
	} // if

	return nil
}
