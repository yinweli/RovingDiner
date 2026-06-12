package exprs

// node 是運算式抽象語法樹的節點; 以型別開關於 evaluator 求值。
// M3 僅有字面值 / 一元 / 二元 / 三元四種純語言節點; 條件對象(屬性 / 引用 / 函式)
// 屬於 Resolver 接縫, 於 M4 再加(對齊【營業實作規格書 | 九、里程碑建議 | M4】)。
type node interface {
	isNode()
}

// nodeLiteral 字面值節點(數值 / 字串 / 布林 / 空物件)。
type nodeLiteral struct {
	value Value
}

func (nodeLiteral) isNode() {}

// nodeUnary 一元運算節點; op 為 tokenMinus(負號)或 tokenNot(否定)。
type nodeUnary struct {
	op      tokenKind
	operand node
}

func (nodeUnary) isNode() {}

// nodeBinary 二元運算節點; op 涵蓋算術(+ - * / %)、比較(< > <= >= == !=)
// 與邏輯(AND / OR)運算符。
type nodeBinary struct {
	op  tokenKind
	lhs node
	rhs node
}

func (nodeBinary) isNode() {}

// nodeTernary 三元運算節點(cond ? then : els); 只評估被選中的分支
// (對齊【營業規格書 | 二十七、運算式 | 7】三元條件失敗規則)。
type nodeTernary struct {
	cond node
	then node
	els  node
}

func (nodeTernary) isNode() {}

// nodeIdent 條件對象識別子節點: 無括號的全域屬性或物件引用(如 morale、drawLast、self);
// 求值時交由 Resolver.Attr 解析(對齊【營業規格書 | 二十七、運算式 | 5】屬性 / self)。
type nodeIdent struct {
	name string
	pos  int // name 於來源的 rune 位置(供 Names 走訪的詞彙錯誤定位; M28 R4A)
}

func (nodeIdent) isNode() {}

// nodeCall 函式呼叫節點: name(args)。求值時先查內建函式註冊表(min / max…), 未命中則
// 視為 Resolver 的查詢函式(對齊【營業規格書 | 二十六、內建函式清單】與【二十七、運算式 | 9】)。
type nodeCall struct {
	name string
	pos  int // name 於來源的 rune 位置(同 nodeIdent.pos)
	arg  []node
}

func (nodeCall) isNode() {}

// nodeRef 引用屬性 / 引用查詢函式節點: name.attr 或 name.attr(args)。求值時先以 Resolver.Attr
// 解析引用主體、再以 Resolver.AttrRef 取其子屬性(對齊【營業規格書 | 二十七、運算式 | 5】引用屬性)。
type nodeRef struct {
	name    string
	pos     int // name 於來源的 rune 位置(同 nodeIdent.pos)
	attr    string
	attrPos int // attr 於來源的 rune 位置
	arg     []node
}

func (nodeRef) isNode() {}
