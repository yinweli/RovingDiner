package exprs

// parser 是遞迴下降解析器,逐一消費 lexer 產出的 token(尾端必為 tokenEOF)。
type parser struct {
	token []token
	pos   int
}

// peek 取得目前 token(不前進);停在尾端 tokenEOF 時持續回傳 EOF。
func (this *parser) peek() token {
	return this.token[this.pos]
}

// next 取得目前 token 並前進;已在尾端 tokenEOF 時不再前進。
func (this *parser) next() (result token) {
	result = this.token[this.pos]

	if this.pos < len(this.token)-1 {
		this.pos++
	} // if

	return result
}

// parseExpr 解析完整運算式:邏輯或之後可接三元 ? :(右結合,只評估被選分支)。
// 對齊【營業規格書 | 二十七、運算式 | 1】<運算式>。
func (this *parser) parseExpr() (result node, err error) {
	cond, errCond := this.parseOr()

	if errCond != nil {
		return nil, errCond
	} // if

	if this.peek().kind != tokenQuestion {
		return cond, nil
	} // if

	this.next() // 吃 ?
	then, errThen := this.parseExpr()

	if errThen != nil {
		return nil, errThen
	} // if

	if this.peek().kind != tokenColon {
		return nil, newError(this.peek().pos, "三元運算式缺少 ':'")
	} // if

	this.next() // 吃 :
	els, errElse := this.parseExpr()

	if errElse != nil {
		return nil, errElse
	} // if

	return nodeTernary{cond: cond, then: then, els: els}, nil
}

// parseOr 解析邏輯或(OR;左結合)。對齊【營業規格書 | 二十七、運算式 | 1】<邏輯或>。
func (this *parser) parseOr() (result node, err error) {
	result, err = this.parseAnd()

	if err != nil {
		return nil, err
	} // if

	for this.peek().kind == tokenOr {
		this.next()
		rhs, errRhs := this.parseAnd()

		if errRhs != nil {
			return nil, errRhs
		} // if

		result = nodeBinary{op: tokenOr, lhs: result, rhs: rhs}
	} // for

	return result, nil
}

// parseAnd 解析邏輯且(AND;左結合,緊於 OR)。對齊【營業規格書 | 二十七、運算式 | 1】<邏輯且>。
func (this *parser) parseAnd() (result node, err error) {
	result, err = this.parseNot()

	if err != nil {
		return nil, err
	} // if

	for this.peek().kind == tokenAnd {
		this.next()
		rhs, errRhs := this.parseNot()

		if errRhs != nil {
			return nil, errRhs
		} // if

		result = nodeBinary{op: tokenAnd, lhs: result, rhs: rhs}
	} // for

	return result, nil
}

// parseNot 解析一元否定(!;右結合,緊於 AND)。對齊【營業規格書 | 二十七、運算式 | 1】<否定>。
func (this *parser) parseNot() (result node, err error) {
	if this.peek().kind == tokenNot {
		this.next()
		operand, errOperand := this.parseNot()

		if errOperand != nil {
			return nil, errOperand
		} // if

		return nodeUnary{op: tokenNot, operand: operand}, nil
	} // if

	return this.parseCompare()
}

// parseCompare 解析比較(< > <= >= == !=;單一、不串接)。對齊【營業規格書 | 二十七、運算式 | 1】<比較>。
func (this *parser) parseCompare() (result node, err error) {
	result, err = this.parseAdd()

	if err != nil {
		return nil, err
	} // if

	switch this.peek().kind {
	case tokenLT, tokenGT, tokenLE, tokenGE, tokenEQ, tokenNE:
		op := this.next().kind
		rhs, errRhs := this.parseAdd()

		if errRhs != nil {
			return nil, errRhs
		} // if

		return nodeBinary{op: op, lhs: result, rhs: rhs}, nil

	default:
		return result, nil // 無比較符:單一算術式
	} // switch
}

// parseAdd 解析加減(+ -;左結合)。對齊【營業規格書 | 二十七、運算式 | 1】<算術式>。
func (this *parser) parseAdd() (result node, err error) {
	result, err = this.parseMul()

	if err != nil {
		return nil, err
	} // if

	for this.peek().kind == tokenPlus || this.peek().kind == tokenMinus {
		op := this.next().kind
		rhs, errRhs := this.parseMul()

		if errRhs != nil {
			return nil, errRhs
		} // if

		result = nodeBinary{op: op, lhs: result, rhs: rhs}
	} // for

	return result, nil
}

// parseMul 解析乘除取餘(* / %;左結合)。對齊【營業規格書 | 二十七、運算式 | 1】<項>。
func (this *parser) parseMul() (result node, err error) {
	result, err = this.parseFactor()

	if err != nil {
		return nil, err
	} // if

	for this.peek().kind == tokenStar || this.peek().kind == tokenSlash || this.peek().kind == tokenPercent {
		op := this.next().kind
		rhs, errRhs := this.parseFactor()

		if errRhs != nil {
			return nil, errRhs
		} // if

		result = nodeBinary{op: op, lhs: result, rhs: rhs}
	} // for

	return result, nil
}

// parseFactor 解析因子:開頭一元負號(右結合)或基本運算元。
// 對齊【營業規格書 | 二十七、運算式 | 1】<因子>。
func (this *parser) parseFactor() (result node, err error) {
	if this.peek().kind == tokenMinus {
		this.next()
		operand, errOperand := this.parseFactor()

		if errOperand != nil {
			return nil, errOperand
		} // if

		return nodeUnary{op: tokenMinus, operand: operand}, nil
	} // if

	return this.parsePrimary()
}

// parsePrimary 解析基本運算元:字面值(數值 / 字串 / 布林 / none)或括號分組。
// 條件對象(屬性 / 引用 / 函式)為 Resolver 接縫,於 M4 再加。
func (this *parser) parsePrimary() (result node, err error) {
	token := this.peek()

	switch token.kind {
	case tokenNum:
		this.next()
		return nodeLiteral{value: NewNum(token.num)}, nil

	case tokenText:
		this.next()
		return nodeLiteral{value: NewText(token.text)}, nil

	case tokenBool:
		this.next()
		return nodeLiteral{value: NewBool(token.flag)}, nil

	case tokenNone:
		this.next()
		return nodeLiteral{value: NewNone()}, nil

	case tokenLParen:
		this.next()
		inner, errInner := this.parseExpr()

		if errInner != nil {
			return nil, errInner
		} // if

		if this.peek().kind != tokenRParen {
			return nil, newError(this.peek().pos, "括號未閉合,缺少 ')'")
		} // if

		this.next() // 吃 )
		return inner, nil

	case tokenEOF:
		return nil, newError(token.pos, "運算式不完整,缺少運算元")

	default:
		return nil, newError(token.pos, "預期運算元(數值/字串/布林/none/括號),但看到:"+token.text)
	} // switch
}
