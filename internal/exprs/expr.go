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

// Eval 對已解析的運算式求值; 對齊【營業規格書 | 二十七、運算式】的求值語意。
// env 帶入條件對象解析器 Resolver 與內建函式註冊表(由 games 提供); 純語言運算式可傳 Env{}。
// 評估失敗不拋錯, 以 ok == false 表達(失敗條件詳見【營業規格書 | 二十七、運算式 | 7】;
// 後續行為由呼叫方各自定義, 詳見【營業規格書 | 二十七、運算式 | 8】)。
// 中間值為 float64、過程不四捨五入(捨入由呼叫方以 Round 處理)。
func (this *Expr) Eval(env Env) (result Value, ok bool) {
	evaluator := &evaluator{env: env}
	return evaluator.eval(this.root)
}
