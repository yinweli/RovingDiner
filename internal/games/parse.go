package games

import (
	"strings"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// Parse 對外入口: 純文法解析單一命令來源為 AST, 不碰詞彙表(成員合法性見 Validate)。
func Parse(source string) (result Command, err error) {
	parser := &commandParser{char: []rune(source)}
	parser.skipSpace()

	command, errParse := parser.parse()

	if errParse != nil {
		return nil, errParse
	} // if

	parser.skipSpace()

	if parser.eof() == false {
		return nil, parser.errorAt(parser.pos, "命令結尾有多餘的內容:"+parser.rest())
	} // if

	return command, nil
}

// Command 命令 AST 的封閉介面; parse-once、與 engine 無關。名稱以原始字串擷取,
// 成員合法性留待 Validate(M6+); 執行於 M7 / M9 補。
type Command interface {
	isCommand()
}

// commandParser 以 rune 切片掃描命令來源; pos 為當前 rune 索引(0 起算, 供錯誤定位與內嵌運算式偏移)。
type commandParser struct {
	char []rune
	pos  int
}

// parse 解析單一命令: 讀首個識別子, 後接 '(' 為操作命令, 否則為屬性修改命令。
func (this *commandParser) parse() (result Command, err error) {
	name, start, ok := this.readIdent()

	if ok == false {
		return nil, this.errorAt(this.pos, "命令開頭需要是屬性名稱或操作命令名稱")
	} // if

	this.skipSpace()

	if this.peek() == '(' {
		return this.parseOperate(name, start)
	} // if

	return this.parseAssign(name, start)
}

// parseAssign 解析屬性修改命令: 可選的 .<引用屬性>、賦值符, 以及(帶值賦值時)至來源結尾的算術式。
// 名稱合法性不在此驗, 留待 Validate(M6+)。
func (this *commandParser) parseAssign(base string, baseStart int) (result Command, err error) {
	command := commandAssign{base: base, basePos: baseStart}

	if this.peek() == '.' {
		this.pos++ // 吃掉 '.'
		this.skipSpace()
		attrName, attrStart, okAttr := this.readIdent()

		if okAttr == false {
			return nil, this.errorAt(this.pos, "『.』後需要引用屬性名稱")
		} // if

		command.refAttr = attrName
		command.refAttrPos = attrStart
		command.isRef = true
	} // if

	this.skipSpace()
	op, okOp := this.readAssignOp()

	if okOp == false {
		return nil, this.errorAt(this.pos, "缺少賦值符(= += -= *= /= %= @ #)")
	} // if

	command.op = op

	if op == cores.AssignLock || op == cores.AssignUnlock {
		return command, nil // @ / # 不帶算術式
	} // if

	this.skipSpace()
	value, errValue := this.parseExprRest(this.pos)

	if errValue != nil {
		return nil, errValue
	} // if

	command.value = value
	return command, nil
}

// parseOperate 解析操作命令:'(' 之後先解析命令對象(第一參數), 再以 , 分隔解析其餘參數至 ')'。
func (this *commandParser) parseOperate(verb string, verbStart int) (result Command, err error) {
	this.pos++ // 吃掉 '('

	selector, errSelector := this.parseSelector()

	if errSelector != nil {
		return nil, errSelector
	} // if

	arg := []*exprs.Expr{}
	this.skipSpace()

	for this.peek() == ',' {
		this.pos++ // 吃掉 ','
		value, errValue := this.parseArg()

		if errValue != nil {
			return nil, errValue
		} // if

		arg = append(arg, value)
		this.skipSpace()
	} // for

	if this.peek() != ')' {
		return nil, this.errorAt(this.pos, "操作命令的參數需以 , 分隔、並以 ) 結束")
	} // if

	this.pos++ // 吃掉 ')'
	return commandOperate{verb: verb, verbPos: verbStart, selector: selector, arg: arg}, nil
}

// parseSelector 解析命令對象: 命令對象名稱, 後接可選的 [<參數>, ...]。
func (this *commandParser) parseSelector() (result selectorArg, err error) {
	this.skipSpace()
	name, start, ok := this.readIdent()

	if ok == false {
		return selectorArg{}, this.errorAt(this.pos, "命令對象需要是名稱(無命令對象時填 none)")
	} // if

	result.selector = name
	result.selectorPos = start
	this.skipSpace()

	if this.peek() == '[' {
		this.pos++ // 吃掉 '['
		param, errParam := this.parseSelectorParam()

		if errParam != nil {
			return selectorArg{}, errParam
		} // if

		result.param = param
	} // if

	return result, nil
}

// parseSelectorParam 解析命令對象 [...] 內以 , 分隔的參數算術式列表('[' 已消耗、消耗對應的 ']')。
func (this *commandParser) parseSelectorParam() (result []*exprs.Expr, err error) {
	result = []*exprs.Expr{}
	this.skipSpace()

	if this.peek() == ']' {
		this.pos++ // 空參數列表(數量是否合法留待命令對象 M8 驗證)
		return result, nil
	} // if

	for {
		value, errValue := this.parseArg()

		if errValue != nil {
			return nil, errValue
		} // if

		result = append(result, value)
		this.skipSpace()

		switch this.peek() {
		case ',':
			this.pos++

		case ']':
			this.pos++
			return result, nil

		default:
			return nil, this.errorAt(this.pos, "命令對象參數需以 , 分隔、並以 ] 結束")
		} // switch
	} // for
}

// parseArg 解析單一參數算術式: 掃描至深度 0 的 , ) ] 或來源結尾為界, 委由 exprs.Parse 解析。
func (this *commandParser) parseArg() (result *exprs.Expr, err error) {
	this.skipSpace()
	start := this.pos
	end := this.scanArg(start)
	sub := string(this.char[start:end])

	if strings.TrimSpace(sub) == "" {
		return nil, this.errorAt(start, "缺少參數內容")
	} // if

	expr, errParse := exprs.Parse(sub)

	if errParse != nil {
		return nil, offsetError(errParse, start)
	} // if

	this.pos = end
	return expr, nil
}

// parseExprRest 把 start 起至來源結尾整段視為一段算術式(屬性修改命令的右值), 委由 exprs.Parse 解析。
func (this *commandParser) parseExprRest(start int) (result *exprs.Expr, err error) {
	sub := string(this.char[start:])

	if strings.TrimSpace(sub) == "" {
		return nil, this.errorAt(start, "賦值符後缺少算術式")
	} // if

	expr, errParse := exprs.Parse(sub)

	if errParse != nil {
		return nil, offsetError(errParse, start)
	} // if

	this.pos = len(this.char)
	return expr, nil
}

// scanArg 自 start 掃描一段參數算術式的結尾界線: 回傳深度 0 的 , ) ] 索引, 或來源結尾索引。
// 以 ( [ 計深度、) ] 減深度; 單引號字串內的標點不計入, 使字串參數與巢狀函式參數的逗號不被誤判。
func (this *commandParser) scanArg(start int) (end int) {
	char := this.char
	size := len(char)
	depth := 0
	i := start

	for i < size {
		switch char[i] {
		case '\'':
			i++

			for i < size && char[i] != '\'' {
				i++
			} // for

			if i < size {
				i++ // 吃掉結尾單引號
			} // if

		case '(', '[':
			depth++
			i++

		case ')', ']':
			if depth == 0 {
				return i
			} // if

			depth--
			i++

		case ',':
			if depth == 0 {
				return i
			} // if

			i++

		default:
			i++
		} // switch
	} // for

	return size
}

// readAssignOp 讀取賦值符(= += -= *= /= %= @ #); 成功時推進 pos。
func (this *commandParser) readAssignOp() (op cores.AssignKind, ok bool) {
	switch this.peek() {
	case '=':
		this.pos++
		return cores.AssignSet, true

	case '@':
		this.pos++
		return cores.AssignLock, true

	case '#':
		this.pos++
		return cores.AssignUnlock, true

	case '+':
		if this.peekNext() == '=' {
			this.pos += 2
			return cores.AssignAdd, true
		} // if

	case '-':
		if this.peekNext() == '=' {
			this.pos += 2
			return cores.AssignSub, true
		} // if

	case '*':
		if this.peekNext() == '=' {
			this.pos += 2
			return cores.AssignMul, true
		} // if

	case '/':
		if this.peekNext() == '=' {
			this.pos += 2
			return cores.AssignDiv, true
		} // if

	case '%':
		if this.peekNext() == '=' {
			this.pos += 2
			return cores.AssignMod, true
		} // if
	} // switch

	return cores.AssignSet, false
}

// readIdent 讀取識別子(ASCII 英文字母 / 數字 / 底線, 首字非數字); 成功回傳文字與起始 rune 索引。
func (this *commandParser) readIdent() (text string, start int, ok bool) {
	start = this.pos

	if this.pos >= len(this.char) || isIdentStart(this.char[this.pos]) == false {
		return "", start, false
	} // if

	for this.pos < len(this.char) && isIdentPart(this.char[this.pos]) {
		this.pos++
	} // for

	return string(this.char[start:this.pos]), start, true
}

// skipSpace 略過空白字元。
func (this *commandParser) skipSpace() {
	for this.pos < len(this.char) && isSpace(this.char[this.pos]) {
		this.pos++
	} // for
}

// eof 回傳是否已掃描至來源結尾。
func (this *commandParser) eof() bool {
	return this.pos >= len(this.char)
}

// peek 回傳當前 rune; 已至結尾回傳 0。
func (this *commandParser) peek() rune {
	if this.pos < len(this.char) {
		return this.char[this.pos]
	} // if

	return 0
}

// peekNext 回傳下一個 rune; 越界回傳 0。
func (this *commandParser) peekNext() rune {
	if this.pos+1 < len(this.char) {
		return this.char[this.pos+1]
	} // if

	return 0
}

// rest 回傳當前位置起的剩餘文字(供結尾多餘內容的錯誤訊息); 呼叫方已先確認非結尾。
func (this *commandParser) rest() string {
	return string(this.char[this.pos:])
}

// errorAt 建立帶位置的語法錯誤, 重用 exprs.SyntaxError 使命令與運算式錯誤格式一致。
func (this *commandParser) errorAt(pos int, msg string) error {
	return &exprs.SyntaxError{Pos: pos, Msg: msg}
}

// commandAssign 屬性修改命令: <左值> <賦值符> [<算術式>](【營業規格書 | 十七、命令 | 1】)。
type commandAssign struct {
	base       string           // 左值基底: 全域屬性, 或引用左值的引用基底(self / drawLast …)
	basePos    int              // base 於來源的 rune 位置(供 Validate 報位置)
	refAttr    string           // 引用屬性名; 非引用左值時為空字串
	refAttrPos int              // refAttr 於來源的 rune 位置
	isRef      bool             // 左值為 <基底>.<引用屬性> 形式時為真
	op         cores.AssignKind // 賦值符(下沉 cores、命令執行據此分派)
	value      *exprs.Expr      // 右值算術式; 鎖定 / 解鎖(@ #)時為 nil
}

// isCommand 標記 commandAssign 為 Command 封閉介面成員。
func (commandAssign) isCommand() {}

// commandOperate 操作命令: <命令>(命令對象, <參數>, ...)(【營業規格書 | 十七、命令 | 2】)。
type commandOperate struct {
	verb     string        // 命令動詞
	verbPos  int           // verb 於來源的 rune 位置
	selector selectorArg   // 命令對象(固定的第一參數)
	arg      []*exprs.Expr // 命令對象之後的其餘參數算術式(含 varargs)
}

// isCommand 標記 commandOperate 為 Command 封閉介面成員。
func (commandOperate) isCommand() {}

// selectorArg 命令對象(操作命令第一參數); param 為 [...] 內參數算術式, 無中括號時為 nil。
type selectorArg struct {
	selector    string
	selectorPos int
	param       []*exprs.Expr
}

// offsetError 把內嵌算術式的 SyntaxError 位置加上偏移, 回算至命令來源的 rune 索引;
// 非 SyntaxError 為防禦性處理(exprs.Parse 僅產出 SyntaxError)。
func offsetError(err error, offset int) error {
	syntaxError, ok := err.(*exprs.SyntaxError)

	if ok == false {
		return &exprs.SyntaxError{Pos: offset, Msg: err.Error()}
	} // if

	return &exprs.SyntaxError{Pos: offset + syntaxError.Pos, Msg: syntaxError.Msg}
}

// isSpace 回傳 rune 是否為空白字元。
func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// isIdentStart 回傳 rune 是否可作識別子起始字元(ASCII 英文字母或底線)。
func isIdentStart(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// isIdentPart 回傳 rune 是否可作識別子後續字元(起始字元或數字)。
func isIdentPart(c rune) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}
