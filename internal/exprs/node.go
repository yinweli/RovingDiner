package exprs

// node 是運算式抽象語法樹的節點;以型別開關於 evaluator 求值。
// M3 僅有字面值 / 一元 / 二元 / 三元四種純語言節點;條件對象(屬性 / 引用 / 函式)
// 屬於 Resolver 接縫,於 M4 再加(對齊【營業實作規格書 | 九、里程碑建議 | M4】)。
type node interface {
	isNode()
}

// nodeLiteral 字面值節點(數值 / 字串 / 布林 / 空物件)。
type nodeLiteral struct {
	value Value
}

func (nodeLiteral) isNode() {}

// nodeUnary 一元運算節點;op 為 tokenMinus(負號)或 tokenNot(否定)。
type nodeUnary struct {
	op      tokenKind
	operand node
}

func (nodeUnary) isNode() {}

// nodeBinary 二元運算節點;op 涵蓋算術(+ - * / %)、比較(< > <= >= == !=)
// 與邏輯(AND / OR)運算符。
type nodeBinary struct {
	op  tokenKind
	lhs node
	rhs node
}

func (nodeBinary) isNode() {}

// nodeTernary 三元運算節點(cond ? then : els);只評估被選中的分支
// (對齊【營業規格書 | 二十七、運算式 | 7】三元條件失敗規則)。
type nodeTernary struct {
	cond node
	then node
	els  node
}

func (nodeTernary) isNode() {}
