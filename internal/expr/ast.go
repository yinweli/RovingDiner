package expr

import "math"

const (
	builtinMax = "max" // 內建函式:取最大值
	builtinMin = "min" // 內建函式:取最小值
)

// node 是 AST 節點;eval 在給定 Resolver 下求值,ok=false 即【二十七〇8】評估失敗狀態,沿樹上傳。
type node interface {
	eval(resolver Resolver) (value Value, ok bool)
}

// literalNode 字面值節點(數字 / 字串 / 布林);求值即回傳預存 Value。
type literalNode struct {
	value Value
}

func (this *literalNode) eval(_ Resolver) (value Value, ok bool) {
	return this.value, true
}

// propertyNode 裸條件對象節點:全域屬性 / 物件引用(如 morale、self、drawLast),走 Resolver.Property。
type propertyNode struct {
	name string
}

func (this *propertyNode) eval(resolver Resolver) (value Value, ok bool) {
	return resolver.Property(this.name, nil)
}

// callNode 函式呼叫節點:內建函式(max / min)或全域查詢函式(如 tableCount('>=', 2))。
type callNode struct {
	name string
	arg  []node
}

func (this *callNode) eval(resolver Resolver) (value Value, ok bool) {
	arg, valid := evalArgs(resolver, this.arg)
	if valid == false {
		return Value{}, false
	} // if

	if this.name == builtinMax || this.name == builtinMin {
		return evalBuiltin(this.name, arg)
	} // if

	return resolver.Property(this.name, arg)
}

// memberNode 引用屬性節點:base.name 或 base.name(arg)(如 self.calm、self.effectStack(101));
// 先以 Resolver.Property 解析 base 引用,再走 Resolver.Member。
type memberNode struct {
	base string
	name string
	arg  []node
	call bool // 是否帶括號(查詢函式);false 為裸屬性存取(傳 nil 引數)
}

func (this *memberNode) eval(resolver Resolver) (value Value, ok bool) {
	target, valid := resolver.Property(this.base, nil)
	if valid == false {
		return Value{}, false
	} // if

	var arg []Value

	if this.call == true {
		arg, valid = evalArgs(resolver, this.arg)
		if valid == false {
			return Value{}, false
		} // if
	} // if

	return resolver.Member(target, this.name, arg)
}

// unaryNode 一元運算節點:tokenMinus(數值取負)或 tokenNot(布林否定)。
type unaryNode struct {
	op      tokenKind
	operand node
}

func (this *unaryNode) eval(resolver Resolver) (value Value, ok bool) {
	operand, valid := this.operand.eval(resolver)
	if valid == false {
		return Value{}, false
	} // if

	if this.op == tokenMinus {
		if operand.kind != kindNumber {
			return Value{}, false
		} // if

		return NewNumber(-operand.number), true
	} // if

	boolean, asBool := AsBool(operand) // tokenNot
	if asBool == false {
		return Value{}, false
	} // if

	return NewBool(boolean == false), true
}

// binaryNode 二元運算節點:依運算符分派至邏輯 / 比較 / 算術。
type binaryNode struct {
	op    tokenKind
	left  node
	right node
}

func (this *binaryNode) eval(resolver Resolver) (value Value, ok bool) {
	if this.op == tokenAnd || this.op == tokenOr {
		return this.evalLogic(resolver)
	} // if

	if isCompareToken(this.op) == true {
		return this.evalCompare(resolver)
	} // if

	return this.evalArith(resolver)
}

// evalLogic 求值邏輯 AND / OR,採短路:AND 左假即停、OR 左真即停;左側本身失敗仍整體失敗。
func (this *binaryNode) evalLogic(resolver Resolver) (value Value, ok bool) {
	left, valid := this.left.eval(resolver)
	if valid == false {
		return Value{}, false
	} // if

	leftBool, asBool := AsBool(left)
	if asBool == false {
		return Value{}, false
	} // if

	if this.op == tokenAnd && leftBool == false {
		return NewBool(false), true // AND 左側為假 → 短路
	} // if

	if this.op == tokenOr && leftBool == true {
		return NewBool(true), true // OR 左側為真 → 短路
	} // if

	right, validRight := this.right.eval(resolver)
	if validRight == false {
		return Value{}, false
	} // if

	rightBool, asBoolRight := AsBool(right)
	if asBoolRight == false {
		return Value{}, false
	} // if

	return NewBool(rightBool), true
}

// evalCompare 求值比較運算;== / != 走相等比較(同型別),其餘大小比較僅限數值。
func (this *binaryNode) evalCompare(resolver Resolver) (value Value, ok bool) {
	left, valid := this.left.eval(resolver)
	if valid == false {
		return Value{}, false
	} // if

	right, validRight := this.right.eval(resolver)
	if validRight == false {
		return Value{}, false
	} // if

	if this.op == tokenEQ || this.op == tokenNE {
		return this.evalEqual(left, right)
	} // if

	return evalOrder(this.op, left, right)
}

// evalEqual 求值 == / !=:同型別才可比較,跨型別失敗;物件引用比實例編號(空物件只與空物件相等)。
func (this *binaryNode) evalEqual(left, right Value) (value Value, ok bool) {
	equal, valid := valueEqual(left, right)
	if valid == false {
		return Value{}, false // 跨型別比較 → 失敗(非 false)
	} // if

	if this.op == tokenNE {
		return NewBool(equal == false), true
	} // if

	return NewBool(equal), true
}

// evalOrder 求值 < > <= >=;僅數值可比大小,其餘型別失敗。
func evalOrder(op tokenKind, left, right Value) (value Value, ok bool) {
	if left.kind != kindNumber || right.kind != kindNumber {
		return Value{}, false
	} // if

	if op == tokenLT {
		return NewBool(left.number < right.number), true
	} // if

	if op == tokenGT {
		return NewBool(left.number > right.number), true
	} // if

	if op == tokenLE {
		return NewBool(left.number <= right.number), true
	} // if

	if op == tokenGE {
		return NewBool(left.number >= right.number), true
	} // if

	return Value{}, false
}

// evalArith 求值算術運算;運算元須皆為數值,/ 與 % 的除數為 0 時失敗。
func (this *binaryNode) evalArith(resolver Resolver) (value Value, ok bool) {
	left, valid := this.left.eval(resolver)
	if valid == false {
		return Value{}, false
	} // if

	right, validRight := this.right.eval(resolver)
	if validRight == false {
		return Value{}, false
	} // if

	if left.kind != kindNumber || right.kind != kindNumber {
		return Value{}, false
	} // if

	if this.op == tokenPlus {
		return NewNumber(left.number + right.number), true
	} // if

	if this.op == tokenMinus {
		return NewNumber(left.number - right.number), true
	} // if

	if this.op == tokenStar {
		return NewNumber(left.number * right.number), true
	} // if

	if this.op == tokenSlash {
		if right.number == 0 {
			return Value{}, false
		} // if

		return NewNumber(left.number / right.number), true
	} // if

	if this.op == tokenPercent {
		if right.number == 0 {
			return Value{}, false
		} // if

		return NewNumber(math.Mod(left.number, right.number)), true
	} // if

	return Value{}, false
}

// ternaryNode 三元運算節點:條件成立評真值分支、否則評假值分支;只評估被選中的分支。
type ternaryNode struct {
	cond node
	yes  node
	no   node
}

func (this *ternaryNode) eval(resolver Resolver) (value Value, ok bool) {
	cond, valid := this.cond.eval(resolver)
	if valid == false {
		return Value{}, false // 條件失敗 → 整個三元失敗(兩分支皆不評估)
	} // if

	boolean, asBool := AsBool(cond)
	if asBool == false {
		return Value{}, false
	} // if

	if boolean == true {
		return this.yes.eval(resolver)
	} // if

	return this.no.eval(resolver)
}

// isCompareToken 回傳 token 是否為比較運算符。
func isCompareToken(kind tokenKind) bool {
	return kind == tokenEQ || kind == tokenNE ||
		kind == tokenLT || kind == tokenGT ||
		kind == tokenLE || kind == tokenGE
}

// valueEqual 判斷兩 Value 是否相等;型別不符回傳 ok=false(跨型別比較失敗)。
func valueEqual(left, right Value) (equal, ok bool) {
	if left.kind != right.kind {
		return false, false
	} // if

	switch left.kind {
	case kindNumber:
		return left.number == right.number, true
	case kindString:
		return left.str == right.str, true
	case kindBool:
		return left.boolean == right.boolean, true
	case kindRef:
		return refEqual(left.ref, right.ref), true
	} // switch

	return false, false
}

// refEqual 比較兩物件引用:空物件只與空物件相等,否則比實例編號。
func refEqual(left, right ref) bool {
	if left.none == true || right.none == true {
		return left.none == right.none
	} // if

	return left.id == right.id
}

// evalArgs 依序評估引數;任一失敗即整體失敗。
func evalArgs(resolver Resolver, arg []node) (result []Value, ok bool) {
	result = make([]Value, 0, len(arg))

	for _, itor := range arg {
		value, valid := itor.eval(resolver)
		if valid == false {
			return nil, false
		} // if

		result = append(result, value)
	} // for

	return result, true
}

// evalBuiltin 計算內建函式 max / min(見【二十六】):至少 2 引數、全為數值,否則失敗。
func evalBuiltin(name string, arg []Value) (value Value, ok bool) {
	if len(arg) < 2 {
		return Value{}, false
	} // if

	for _, itor := range arg {
		if itor.kind != kindNumber {
			return Value{}, false
		} // if
	} // for

	result := arg[0].number

	for _, itor := range arg[1:] {
		if name == builtinMax && itor.number > result {
			result = itor.number
		} // if

		if name == builtinMin && itor.number < result {
			result = itor.number
		} // if
	} // for

	return NewNumber(result), true
}
