package exprs

import (
	"math"
)

// evaluator 走訪 AST 求值;M3 無遊戲接縫,M4 起注入 Resolver 與內建函式。
type evaluator struct {
}

// eval 依節點型別分派求值;回傳值與是否成功(ok == false 為評估失敗)。
func (this *evaluator) eval(n node) (result Value, ok bool) {
	switch n := n.(type) {
	case nodeLiteral:
		return n.value, true

	case nodeUnary:
		return this.evalUnary(n)

	case nodeBinary:
		return this.evalBinary(n)

	case nodeTernary:
		return this.evalTernary(n)

	default:
		return Value{}, false
	} // switch
}

// evalUnary 求值一元運算:負號僅適用數值;否定先取真假判定再反轉。
func (this *evaluator) evalUnary(n nodeUnary) (result Value, ok bool) {
	operand, okOperand := this.eval(n.operand)

	if okOperand == false {
		return Value{}, false
	} // if

	switch n.op {
	case tokenMinus:
		if operand.IsNum() == false {
			return Value{}, false
		} // if

		return NewNum(-operand.Num()), true

	case tokenNot:
		flag, okFlag := operand.Truthy()

		if okFlag == false {
			return Value{}, false
		} // if

		return NewBool(flag == false), true

	default:
		return Value{}, false
	} // switch
}

// evalBinary 求值二元運算;AND / OR 走短路求值,其餘先評估兩運算元再分派。
func (this *evaluator) evalBinary(n nodeBinary) (result Value, ok bool) {
	if n.op == tokenAnd || n.op == tokenOr {
		return this.evalLogical(n)
	} // if

	lhs, okLhs := this.eval(n.lhs)

	if okLhs == false {
		return Value{}, false
	} // if

	rhs, okRhs := this.eval(n.rhs)

	if okRhs == false {
		return Value{}, false
	} // if

	switch n.op {
	case tokenPlus, tokenMinus, tokenStar, tokenSlash, tokenPercent:
		return evalArith(n.op, lhs, rhs)

	case tokenLT, tokenGT, tokenLE, tokenGE:
		return evalOrder(n.op, lhs, rhs)

	case tokenEQ, tokenNE:
		return evalEqual(n.op, lhs, rhs)

	default:
		return Value{}, false
	} // switch
}

// evalLogical 求值邏輯且 / 或,採短路:AND 左假即假、OR 左真即真,皆不評估右運算元
// (對齊【營業實作規格書 | 九、里程碑建議 | M3】短路設計;真假判定見 Value.Truthy)。
func (this *evaluator) evalLogical(n nodeBinary) (result Value, ok bool) {
	lhs, okLhs := this.eval(n.lhs)

	if okLhs == false {
		return Value{}, false
	} // if

	left, okLeft := lhs.Truthy()

	if okLeft == false {
		return Value{}, false
	} // if

	if n.op == tokenAnd && left == false {
		return NewBool(false), true // AND 左假短路
	} // if

	if n.op == tokenOr && left {
		return NewBool(true), true // OR 左真短路
	} // if

	rhs, okRhs := this.eval(n.rhs)

	if okRhs == false {
		return Value{}, false
	} // if

	right, okRight := rhs.Truthy()

	if okRight == false {
		return Value{}, false
	} // if

	return NewBool(right), true
}

// evalTernary 求值三元:條件失敗則整個三元失敗、兩分支皆不評估;條件成功只評估被選分支
// (對齊【營業規格書 | 二十七、運算式 | 7】三元條件失敗)。
func (this *evaluator) evalTernary(n nodeTernary) (result Value, ok bool) {
	cond, okCond := this.eval(n.cond)

	if okCond == false {
		return Value{}, false
	} // if

	flag, okFlag := cond.Truthy()

	if okFlag == false {
		return Value{}, false
	} // if

	if flag {
		return this.eval(n.then)
	} // if

	return this.eval(n.els)
}

// evalArith 求值算術運算;非數值運算元或 除 0 / 取餘 0 皆評估失敗
// (對齊【營業規格書 | 二十七、運算式 | 7】)。
func evalArith(op tokenKind, lhs, rhs Value) (result Value, ok bool) {
	if lhs.IsNum() == false || rhs.IsNum() == false {
		return Value{}, false
	} // if

	a := lhs.Num()
	b := rhs.Num()

	switch op {
	case tokenPlus:
		return NewNum(a + b), true

	case tokenMinus:
		return NewNum(a - b), true

	case tokenStar:
		return NewNum(a * b), true

	case tokenSlash:
		if b == 0 {
			return Value{}, false
		} // if

		return NewNum(a / b), true

	case tokenPercent:
		if b == 0 {
			return Value{}, false
		} // if

		return NewNum(math.Mod(a, b)), true

	default:
		return Value{}, false
	} // switch
}

// evalOrder 求值大小比較(< > <= >=);僅數值對數值合法,其餘評估失敗
// (字串 / 布林 / 空物件無大小順序)。
func evalOrder(op tokenKind, lhs, rhs Value) (result Value, ok bool) {
	if lhs.IsNum() == false || rhs.IsNum() == false {
		return Value{}, false
	} // if

	a := lhs.Num()
	b := rhs.Num()

	switch op {
	case tokenLT:
		return NewBool(a < b), true

	case tokenGT:
		return NewBool(a > b), true

	case tokenLE:
		return NewBool(a <= b), true

	case tokenGE:
		return NewBool(a >= b), true

	default:
		return Value{}, false
	} // switch
}

// evalEqual 求值相等比較(== !=);依 valueEqual 判定,跨型別評估失敗。
func evalEqual(op tokenKind, lhs, rhs Value) (result Value, ok bool) {
	equal, okEqual := valueEqual(lhs, rhs)

	if okEqual == false {
		return Value{}, false
	} // if

	if op == tokenNE {
		return NewBool(equal == false), true
	} // if

	return NewBool(equal), true
}

// valueEqual 依【營業規格書 | 二十七、運算式 | 2、7】判定相等:同型別才可比,
// 跨型別評估失敗;空物件只與空物件相等(M4 起物件引用比實例編號)。
func valueEqual(lhs, rhs Value) (result, ok bool) {
	switch {
	case lhs.IsNum() && rhs.IsNum():
		return lhs.Num() == rhs.Num(), true

	case lhs.IsText() && rhs.IsText():
		return lhs.Text() == rhs.Text(), true

	case lhs.IsBool() && rhs.IsBool():
		return lhs.Bool() == rhs.Bool(), true

	case lhs.IsNone() && rhs.IsNone():
		return true, true

	default:
		return false, false
	} // switch
}
