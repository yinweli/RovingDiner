package expr

import "errors"

// Parse 把運算式原始字串編譯成可重複求值的 Expr;語法錯誤回傳 error。
// 運算式為靜態資料(來自表格),建議於載入時 Parse 一次、之後對不同 Runtime 多次 Eval。
func Parse(source string) (expr *Expr, err error) {
	token, err := lex(source)
	if err != nil {
		return nil, err
	} // if

	parse := &parser{token: token}

	root, err := parse.parseExpr()
	if err != nil {
		return nil, err
	} // if

	if parse.peek().kind != tokenEOF {
		return nil, errors.New("expr: 多餘的 token: " + parse.peek().text)
	} // if

	return &Expr{root: root}, nil
}

// Expr 是編譯後的運算式(AST 根)。
type Expr struct {
	root node
}

// Eval 在給定 Resolver 下求值;ok=false 即【營業規格書 | 二十七、運算式 | 8】評估失敗狀態(不拋錯),
// 由呼叫方依場景定義後續行為(觸發條件 → 不成立、觸發次數 → 0、命令 → no-op)。
func (this *Expr) Eval(resolver Resolver) (value Value, ok bool) {
	return this.root.eval(resolver)
}
