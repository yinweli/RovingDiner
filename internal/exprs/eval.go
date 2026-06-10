package exprs

import (
	"math"
)

// evaluator 走訪 AST 求值;條件對象 / 函式經 env(Resolver + 內建函式註冊表)解析,
// env 於 Eval 時帶入、無全域可變狀態。
type evaluator struct {
	env Env
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

	case nodeIdent:
		return this.evalIdent(n)

	case nodeCall:
		return this.evalCall(n)

	case nodeRef:
		return this.evalRef(n)

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

// evalIdent 求值無括號條件對象(全域屬性 / 物件引用 / self),交由 Resolver.Attr 解析。
func (this *evaluator) evalIdent(n nodeIdent) (result Value, ok bool) {
	if this.env.Resolver == nil {
		return Value{}, false
	} // if

	return this.env.Resolver.Attr(n.name, nil)
}

// evalCall 求值函式呼叫:先評估參數,再以名稱查內建函式註冊表;未命中則視為 Resolver 查詢函式。
// 內建函式優先於查詢函式(對齊【營業規格書 | 二十六、內建函式清單】與【二十三、屬性清單】查詢函式之區隔)。
func (this *evaluator) evalCall(n nodeCall) (result Value, ok bool) {
	arg, okArg := this.evalArgs(n.arg)

	if okArg == false {
		return Value{}, false
	} // if

	if builtin, exist := this.env.Builtin[n.name]; exist {
		return builtin(arg)
	} // if

	if this.env.Resolver == nil {
		return Value{}, false
	} // if

	return this.env.Resolver.Attr(n.name, arg)
}

// evalRef 求值引用屬性 / 引用查詢函式:先以 Resolver.Attr 解析引用主體,主體須為物件引用
// (空物件 / 型別不符 → 失敗,對齊【營業規格書 | 二十七、運算式 | 5】),再以 Resolver.AttrRef 取子屬性。
func (this *evaluator) evalRef(n nodeRef) (result Value, ok bool) {
	if this.env.Resolver == nil {
		return Value{}, false
	} // if

	base, okBase := this.env.Resolver.Attr(n.name, nil)

	if okBase == false {
		return Value{}, false
	} // if

	if base.IsRef() == false {
		return Value{}, false
	} // if

	arg, okArg := this.evalArgs(n.arg)

	if okArg == false {
		return Value{}, false
	} // if

	return this.env.Resolver.AttrRef(base.Ref(), n.attr, arg)
}

// evalArgs 逐一求值參數;任一參數評估失敗則整體失敗。
func (this *evaluator) evalArgs(arg []node) (result []Value, ok bool) {
	result = make([]Value, 0, len(arg))

	for _, itor := range arg {
		value, okValue := this.eval(itor)

		if okValue == false {
			return nil, false
		} // if

		result = append(result, value)
	} // for

	return result, true
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

	case lhs.isObject() && rhs.isObject():
		return objectEqual(lhs, rhs), true

	default:
		return false, false
	} // switch
}

// objectEqual 判定兩個物件值(空物件 / 物件引用)是否相等:空物件只與空物件相等、
// 兩引用比實例編號、一空一非空為不等(對齊【營業規格書 | 二十七、運算式 | 2】)。
func objectEqual(lhs, rhs Value) (result bool) {
	if lhs.IsNone() && rhs.IsNone() {
		return true
	} // if

	if lhs.IsRef() && rhs.IsRef() {
		return lhs.Ref().IsSame(rhs.Ref())
	} // if

	return false // 一空一非空
}
