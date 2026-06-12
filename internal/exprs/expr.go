package exprs

// Parse 把運算式字串解析為 Expr; 語法錯誤回傳帶位置的 SyntaxError。
// 文法與優先序對齊【營業規格書 | 二十七、運算式 | 1】
// (括號 > 乘除餘 > 加減 > 比較 > 否定 > AND > OR > 三元; 三元右結合、其餘左結合)。
func Parse(source string) (result *Expr, err error) {
	token, errLex := lex(source)

	if errLex != nil {
		return nil, errLex
	} // if

	parser := &parser{token: token}
	root, errParse := parser.parseExpr()

	if errParse != nil {
		return nil, errParse
	} // if

	if parser.peek().kind != tokenEOF {
		return nil, newError(parser.peek().pos, "運算式結尾有多餘的內容:"+parser.peek().text)
	} // if

	return &Expr{root: root}, nil
}

// Expr 是 parse 一次後可重複求值的運算式; 對齊【營業實作規格書 | 九、里程碑建議 | M3】
// 的 parse-once 設計(靜態載入時解析一次、執行期多次求值)。
type Expr struct {
	root node
}

// NameUse 條件對象名稱使用紀錄(Names 走訪輸出): 識別子 / 函式呼叫以 Name + Argc 表達
// (無括號識別子 Argc = 0, 與零參數呼叫同制——求值路徑等價), 引用型(name.attr)另帶 Attr 與其位置。
type NameUse struct {
	Ref     bool   // 引用型(name.attr)
	Name    string // 識別子 / 函式名 / 引用基底名
	Pos     int    // Name 於來源的 rune 位置
	Attr    string // 引用屬性名(僅引用型)
	AttrPos int    // Attr 於來源的 rune 位置(僅引用型)
	Argc    int    // 參數個數(引用型為 attr 的參數個數)
}

// Names 走訪 AST 收集全部條件對象名稱使用(識別子 / 函式 / 引用), 依來源出現順序;
// 供 games.ValidateExpr 靜態查詞彙(M28 R4A)。純字面值運算式回空。
func (this *Expr) Names() []NameUse {
	if this == nil {
		return nil
	} // if

	return collectName(this.root, nil)
}

// collectName 遞迴收集節點樹的名稱使用(深度優先、參數靠後, 即來源出現順序)。
func collectName(n node, result []NameUse) []NameUse {
	switch n := n.(type) {
	case nodeUnary:
		result = collectName(n.operand, result)

	case nodeBinary:
		result = collectName(n.lhs, result)
		result = collectName(n.rhs, result)

	case nodeTernary:
		result = collectName(n.cond, result)
		result = collectName(n.then, result)
		result = collectName(n.els, result)

	case nodeIdent:
		result = append(result, NameUse{Name: n.name, Pos: n.pos})

	case nodeCall:
		result = append(result, NameUse{Name: n.name, Pos: n.pos, Argc: len(n.arg)})

		for _, itor := range n.arg {
			result = collectName(itor, result)
		} // for

	case nodeRef:
		result = append(result, NameUse{Ref: true, Name: n.name, Pos: n.pos, Attr: n.attr, AttrPos: n.attrPos, Argc: len(n.arg)})

		for _, itor := range n.arg {
			result = collectName(itor, result)
		} // for
	} // switch

	return result
}

// Eval 對已解析的運算式求值; 對齊【營業規格書 | 二十七、運算式】的求值語意。
// env 帶入條件對象解析器 Resolver 與內建函式註冊表(由 games 提供); 純語言運算式可傳 Env{}。
// 評估失敗不拋錯, 以 ok == false 表達(失敗條件詳見【營業規格書 | 二十七、運算式 | 7】;
// 後續行為由呼叫方各自定義, 詳見【營業規格書 | 二十七、運算式 | 8】)。
// 中間值為 float64、過程不四捨五入(捨入由呼叫方以 Round 處理)。
func (this *Expr) Eval(env Env) (result Value, ok bool) {
	evaluator := &evaluator{env: env}
	return evaluator.eval(this.root)
}
