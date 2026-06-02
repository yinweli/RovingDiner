package expr

import "errors"

// parser 對 token 序列做遞迴下降解析。優先級階梯由低到高:
// 三元 → 邏輯或 → 邏輯且 → 否定 → 比較 → 加減 → 乘除 → 因子(對齊【營業規格書 | 二十七、運算式 | 3】優先順序表)。
type parser struct {
	token []token
	pos   int
}

// peek 回傳當前 token(越界回 EOF)。
func (this *parser) peek() token {
	if this.pos >= len(this.token) {
		return token{kind: tokenEOF, text: "<eof>"}
	} // if

	return this.token[this.pos]
}

// next 消耗並回傳當前 token(越界回 EOF 且不前進)。
func (this *parser) next() token {
	tok := this.peek()

	if this.pos < len(this.token) {
		this.pos++
	} // if

	return tok
}

// parseExpr 解析 <運算式>:邏輯式後可接三元 ? :(右結合)。
func (this *parser) parseExpr() (result node, err error) {
	cond, err := this.parseOr()
	if err != nil {
		return nil, err
	} // if

	if this.peek().kind != tokenQuestion {
		return cond, nil
	} // if

	this.next() // 吃 ?

	yes, err := this.parseExpr()
	if err != nil {
		return nil, err
	} // if

	if this.peek().kind != tokenColon {
		return nil, errors.New("expr: 三元運算式缺少 ':'")
	} // if

	this.next() // 吃 :

	no, err := this.parseExpr() // 右結合:假值分支續解析整層運算式
	if err != nil {
		return nil, err
	} // if

	return &ternaryNode{cond: cond, yes: yes, no: no}, nil
}

// parseOr 解析 <邏輯或>:邏輯且 {OR 邏輯且}(左結合)。
func (this *parser) parseOr() (result node, err error) {
	left, err := this.parseAnd()
	if err != nil {
		return nil, err
	} // if

	for this.peek().kind == tokenOr {
		this.next()

		right, errRight := this.parseAnd()
		if errRight != nil {
			return nil, errRight
		} // if

		left = &binaryNode{op: tokenOr, left: left, right: right}
	} // for

	return left, nil
}

// parseAnd 解析 <邏輯且>:否定 {AND 否定}(左結合)。
func (this *parser) parseAnd() (result node, err error) {
	left, err := this.parseNot()
	if err != nil {
		return nil, err
	} // if

	for this.peek().kind == tokenAnd {
		this.next()

		right, errRight := this.parseNot()
		if errRight != nil {
			return nil, errRight
		} // if

		left = &binaryNode{op: tokenAnd, left: left, right: right}
	} // for

	return left, nil
}

// parseNot 解析 <否定>:[!] 否定 | 比較(否定優先級低於比較,故 !a == b 解析為 !(a == b))。
func (this *parser) parseNot() (result node, err error) {
	if this.peek().kind == tokenNot {
		this.next()

		operand, errOperand := this.parseNot()
		if errOperand != nil {
			return nil, errOperand
		} // if

		return &unaryNode{op: tokenNot, operand: operand}, nil
	} // if

	return this.parseCompare()
}

// parseCompare 解析 <比較>:算術式 [比較符 算術式];至多一個比較符(不支援 a < b < c 鏈式)。
func (this *parser) parseCompare() (result node, err error) {
	left, err := this.parseAdd()
	if err != nil {
		return nil, err
	} // if

	if isCompareToken(this.peek().kind) == false {
		return left, nil
	} // if

	op := this.next().kind

	right, err := this.parseAdd()
	if err != nil {
		return nil, err
	} // if

	return &binaryNode{op: op, left: left, right: right}, nil
}

// parseAdd 解析 <算術式>:項 {(+ | -) 項}(左結合)。
func (this *parser) parseAdd() (result node, err error) {
	left, err := this.parseMul()
	if err != nil {
		return nil, err
	} // if

	for this.peek().kind == tokenPlus || this.peek().kind == tokenMinus {
		op := this.next().kind

		right, errRight := this.parseMul()
		if errRight != nil {
			return nil, errRight
		} // if

		left = &binaryNode{op: op, left: left, right: right}
	} // for

	return left, nil
}

// parseMul 解析 <項>:因子 {(* | / | %) 因子}(左結合)。
func (this *parser) parseMul() (result node, err error) {
	left, err := this.parseFactor()
	if err != nil {
		return nil, err
	} // if

	for this.peek().kind == tokenStar || this.peek().kind == tokenSlash || this.peek().kind == tokenPercent {
		op := this.next().kind

		right, errRight := this.parseFactor()
		if errRight != nil {
			return nil, errRight
		} // if

		left = &binaryNode{op: op, left: left, right: right}
	} // for

	return left, nil
}

// parseFactor 解析 <因子>:[-] 後接 字面值 / 括號運算式 / 識別子分支(unary minus 綁定最緊)。
func (this *parser) parseFactor() (result node, err error) {
	tok := this.peek()

	if tok.kind == tokenMinus {
		this.next()

		operand, errOperand := this.parseFactor()
		if errOperand != nil {
			return nil, errOperand
		} // if

		return &unaryNode{op: tokenMinus, operand: operand}, nil
	} // if

	if tok.kind == tokenNumber {
		this.next()
		return &literalNode{value: NewNumber(tok.number)}, nil
	} // if

	if tok.kind == tokenString {
		this.next()
		return &literalNode{value: NewString(tok.text)}, nil
	} // if

	if tok.kind == tokenBool {
		this.next()
		return &literalNode{value: NewBool(tok.boolean)}, nil
	} // if

	if tok.kind == tokenLParen {
		return this.parseParen()
	} // if

	if tok.kind == tokenIdent {
		return this.parseIdent()
	} // if

	return nil, errors.New("expr: 未預期的 token: " + tok.text)
}

// parseParen 解析括號:內為完整運算式(統一【營業規格書 | 二十七、運算式 | 1】的「算術式 / 運算式」分組,由型別於求值決定)。
func (this *parser) parseParen() (result node, err error) {
	this.next() // 吃 (

	inner, err := this.parseExpr()
	if err != nil {
		return nil, err
	} // if

	if this.peek().kind != tokenRParen {
		return nil, errors.New("expr: 括號未閉合")
	} // if

	this.next() // 吃 )
	return inner, nil
}

// parseIdent 解析識別子分支:name(arg) 函式呼叫 / name.attr 引用屬性 / 裸 name 條件對象。
func (this *parser) parseIdent() (result node, err error) {
	name := this.next().text

	if this.peek().kind == tokenLParen {
		arg, errArg := this.parseArgs()
		if errArg != nil {
			return nil, errArg
		} // if

		return &callNode{name: name, arg: arg}, nil
	} // if

	if this.peek().kind == tokenDot {
		return this.parseMember(name)
	} // if

	return &propertyNode{name: name}, nil
}

// parseMember 解析 name.attr 或 name.attr(arg)。
func (this *parser) parseMember(base string) (result node, err error) {
	this.next() // 吃 .

	if this.peek().kind != tokenIdent {
		return nil, errors.New("expr: '.' 後需要屬性名")
	} // if

	member := this.next().text

	if this.peek().kind == tokenLParen {
		arg, errArg := this.parseArgs()
		if errArg != nil {
			return nil, errArg
		} // if

		return &memberNode{base: base, name: member, arg: arg, call: true}, nil
	} // if

	return &memberNode{base: base, name: member, call: false}, nil
}

// parseArgs 解析引數列 ( arg {, arg} );每個引數為 <算術式> | <字串>,不含比較 / 邏輯 / 三元。
func (this *parser) parseArgs() (result []node, err error) {
	this.next() // 吃 (
	result = []node{}

	if this.peek().kind == tokenRParen {
		this.next()
		return result, nil
	} // if

	for {
		arg, errArg := this.parseAdd()
		if errArg != nil {
			return nil, errArg
		} // if

		result = append(result, arg)

		if this.peek().kind != tokenComma {
			break
		} // if

		this.next() // 吃 ,
	} // for

	if this.peek().kind != tokenRParen {
		return nil, errors.New("expr: 引數列未閉合")
	} // if

	this.next() // 吃 )
	return result, nil
}
