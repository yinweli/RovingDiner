package exprs

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteParser(t *testing.T) {
	suite.Run(t, new(SuiteParser))
}

// SuiteParser 驗證遞迴下降解析: 三元、邏輯、比較、加減 / 乘除、一元負號、字面值 / 括號 / 識別子 / 函式 / 引用,
// 以及優先序 / 結合性與語法錯誤(含出錯位置)。測試順序對齊 parser.go 由外而內的遞迴下降宣告序; 樹形以 S-運算式字串比對。
type SuiteParser struct {
	suite.Suite
}

func (this *SuiteParser) TestParseTernary() {
	this.Equal("(?: true 1 2)", this.ast("true ? 1 : 2"))
	this.Equal("(?: true 1 (?: false 2 3))", this.ast("true ? 1 : false ? 2 : 3")) // 右結合
}

func (this *SuiteParser) TestParseLogical() {
	this.Equal("(or true (and false true))", this.ast("true OR false AND true")) // AND 緊於 OR
	this.Equal("(and (not true) false)", this.ast("!true AND false"))            // NOT 緊於 AND
	this.Equal("(not (not true))", this.ast("!!true"))
}

func (this *SuiteParser) TestParseComparison() {
	this.Equal("(< (+ 1 2) 3)", this.ast("1 + 2 < 3")) // 算術緊於比較
	this.Equal("(>= 2 2)", this.ast("2 >= 2"))
}

func (this *SuiteParser) TestParseAdditive() {
	this.Equal("(- (- 1 2) 3)", this.ast("1 - 2 - 3")) // 左結合
}

func (this *SuiteParser) TestParseMultiplicative() {
	this.Equal("(% (* 2 3) 4)", this.ast("2 * 3 % 4")) // 左結合
}

func (this *SuiteParser) TestParseUnaryMinus() {
	this.Equal("(neg 3)", this.ast("-3"))
	this.Equal("(neg (+ 2 3))", this.ast("-(2 + 3)"))
	this.Equal("(neg (neg 3))", this.ast("- -3")) // 一元負號右結合
}

func (this *SuiteParser) TestParseLiteral() {
	this.Equal("5", this.ast("5"))
	this.Equal("1.5", this.ast("1.5"))
	this.Equal("true", this.ast("true"))
	this.Equal("'hi'", this.ast("'hi'"))
	this.Equal("none", this.ast("none"))
}

func (this *SuiteParser) TestParseGrouping() {
	this.Equal("(* (+ 1 2) 3)", this.ast("(1 + 2) * 3"))
}

func (this *SuiteParser) TestParseIdent() {
	this.Equal("morale", this.ast("morale")) // 全域屬性
	this.Equal("self", this.ast("self"))     // 物件引用
	this.Equal("(> (+ morale 5) 10)", this.ast("morale + 5 > 10"))
	this.Equal("(!= self none)", this.ast("self != none"))
}

func (this *SuiteParser) TestParseCall() {
	this.Equal("(call min 1 2)", this.ast("min(1, 2)"))
	this.Equal("(call max 3 1 2)", this.ast("max(3, 1, 2)"))
	this.Equal("(call tableCount '>=' 2)", this.ast("tableCount('>=', 2)"))  // 字串參數
	this.Equal("(call max (call min 1 2) 3)", this.ast("max(min(1, 2), 3)")) // 巢狀函式
	this.Equal("(call now)", this.ast("now()"))                              // 無參數
}

func (this *SuiteParser) TestParseRef() {
	this.Equal("(ref self calm)", this.ast("self.calm"))
	this.Equal("(ref drawLast cardID)", this.ast("drawLast.cardID"))
	this.Equal("(ref self effectStack 101)", this.ast("self.effectStack(101)")) // 引用查詢函式
}

func (this *SuiteParser) TestParsePrecedence() {
	this.Equal("(+ 1 (* 2 3))", this.ast("1 + 2 * 3")) // 乘除緊於加減
	this.Equal("(+ (* 2 3) 1)", this.ast("2 * 3 + 1"))
}

func (this *SuiteParser) TestParseError() {
	source := []string{
		"",            // 空輸入(EOF)
		"* 3",         // 缺左運算元
		"(1",          // 括號未閉合
		"()",          // 括號內缺運算元
		"1 +",         // 加法缺右運算元
		"2 *",         // 乘法缺右運算元
		"1 <",         // 比較缺右運算元
		"true OR",     // OR 缺右運算元
		"true AND",    // AND 缺右運算元
		"!",           // 否定缺運算元
		"-",           // 負號缺運算元
		"true ? 1",    // 三元缺 :
		"true ? : 2",  // 三元真值分支缺運算元
		"true ? 1 :",  // 三元假值分支缺運算元
		"self.",       // '.' 後缺引用屬性名
		"min(1",       // 函式參數未閉合
		"tableCount(", // 函式參數列表缺運算元
		"self.calm(",  // 引用查詢函式參數未閉合
	}

	for _, itor := range source {
		_, err := Parse(itor)
		this.Require().Error(err, itor)
	} // for
}

func (this *SuiteParser) TestParseErrorPosition() {
	this.Equal(0, this.errorPos(""))    // EOF 在位置 0
	this.Equal(2, this.errorPos("(1"))  // 缺 ) , 停在 EOF(長度 2)
	this.Equal(3, this.errorPos("1 +")) // 缺右運算元, 停在 EOF
}

// ast 解析 source 並回傳其 AST 的 S-運算式字串(解析失敗即 fail 測試)。
func (this *SuiteParser) ast(source string) string {
	expr, err := Parse(source)
	this.Require().NoError(err, source)
	return astString(expr.root)
}

// errorPos 解析 source(預期失敗)並回傳 SyntaxError 的位置。
func (this *SuiteParser) errorPos(source string) int {
	_, err := Parse(source)
	this.Require().Error(err, source)
	syntaxError, ok := err.(*SyntaxError)
	this.Require().True(ok)
	return syntaxError.Pos
}

// astString 把 AST 節點轉成 S-運算式字串, 便於樹形比對。
func astString(n node) string {
	switch n := n.(type) {
	case nodeLiteral:
		return literalString(n.value)

	case nodeUnary:
		return "(" + unaryOp(n.op) + " " + astString(n.operand) + ")"

	case nodeBinary:
		return "(" + binaryOp(n.op) + " " + astString(n.lhs) + " " + astString(n.rhs) + ")"

	case nodeTernary:
		return "(?: " + astString(n.cond) + " " + astString(n.then) + " " + astString(n.els) + ")"

	case nodeIdent:
		return n.name

	case nodeCall:
		return "(call " + n.name + argString(n.arg) + ")"

	case nodeRef:
		return "(ref " + n.name + " " + n.attr + argString(n.arg) + ")"

	default:
		return "?"
	} // switch
}

// argString 把參數節點列表渲染成空白前綴序列(無參數時為空字串)。
func argString(arg []node) string {
	result := ""

	for _, itor := range arg {
		result += " " + astString(itor)
	} // for

	return result
}

// literalString 把字面值轉成字串(字串值以單引號包夾、none 印 none)。
func literalString(value Value) string {
	switch {
	case value.IsNum():
		return strconv.FormatFloat(value.Num(), 'g', -1, 64)

	case value.IsBool():
		return strconv.FormatBool(value.Bool())

	case value.IsText():
		return "'" + value.Text() + "'"

	case value.IsNone():
		return "none"

	default:
		return "?"
	} // switch
}

// unaryOp 把一元運算符轉成顯示用標記。
func unaryOp(op tokenKind) string {
	switch op {
	case tokenMinus:
		return "neg"

	case tokenNot:
		return "not"

	default:
		return "?"
	} // switch
}

// binaryOp 把二元運算符轉成顯示用標記。
func binaryOp(op tokenKind) string {
	switch op {
	case tokenPlus:
		return "+"

	case tokenMinus:
		return "-"

	case tokenStar:
		return "*"

	case tokenSlash:
		return "/"

	case tokenPercent:
		return "%"

	case tokenLT:
		return "<"

	case tokenGT:
		return ">"

	case tokenLE:
		return "<="

	case tokenGE:
		return ">="

	case tokenEQ:
		return "=="

	case tokenNE:
		return "!="

	case tokenAnd:
		return "and"

	case tokenOr:
		return "or"

	default:
		return "?"
	} // switch
}
