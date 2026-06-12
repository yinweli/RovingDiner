package exprs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteExpr(t *testing.T) {
	suite.Run(t, new(SuiteExpr))
}

// SuiteExpr 驗證對外入口 Parse / Expr / Eval: 成功解析後可求值、詞法錯誤上拋、
// 結尾多餘內容報錯, 以及 parse-once / eval-many(同一 Expr 多次求值結果一致)。
type SuiteExpr struct {
	suite.Suite
}

func (this *SuiteExpr) TestParse() {
	expr, err := Parse("1 + 2 * 3")
	this.Require().NoError(err)
	this.Require().NotNil(expr)

	value, ok := expr.Eval(Env{})
	this.True(ok)
	this.Equal(7.0, value.Num())
}

func (this *SuiteExpr) TestParseLexError() {
	// 詞法錯誤(未知字元 / 字串未結束 / 裸 =)由 Parse 直接上拋, 型別仍為 SyntaxError
	source := []string{
		"1 & 2", // 未知字元
		"'oops", // 字串未結束
		"a = 1", // 裸 = 不屬於運算式
	}

	for _, itor := range source {
		_, err := Parse(itor)
		this.Require().Error(err, itor)
		_, ok := err.(*SyntaxError)
		this.True(ok, itor)
	} // for
}

func (this *SuiteExpr) TestParseTrailing() {
	_, err := Parse("1 2") // 解析出 1 之後仍有多餘 token
	this.Require().Error(err)
	syntaxError, ok := err.(*SyntaxError)
	this.Require().True(ok)
	this.Equal(2, syntaxError.Pos) // 指向多餘的第二個 token
}

// TestExprNames 驗證名稱使用走訪: 識別子 / 函式 / 引用各型帶位置與參數個數、依來源出現順序;
// 純字面值回空、nil 防禦。
func (this *SuiteExpr) TestExprNames() {
	expr, err := Parse("morale + min(1, handSize(2)) > self.effectGroup(5)")
	this.Require().NoError(err)
	this.Equal([]NameUse{
		{Name: "morale", Pos: 0},
		{Name: "min", Pos: 9, Argc: 2},
		{Name: "handSize", Pos: 16, Argc: 1},
		{Ref: true, Name: "self", Pos: 31, Attr: "effectGroup", AttrPos: 36, Argc: 1},
	}, expr.Names())

	literal, errLiteral := Parse("-(1 + 2) > 0 ? 'a' : 'b'") // 純字面值(含一元 / 二元 / 三元節點)→ 空
	this.Require().NoError(errLiteral)
	this.Empty(literal.Names())

	var none *Expr
	this.Empty(none.Names()) // nil 防禦
}

func (this *SuiteExpr) TestExprReuse() {
	// parse 一次、求值多次, 結果穩定一致(對齊 parse-once / eval-many 設計)
	expr, err := Parse("(2 + 3) * 4")
	this.Require().NoError(err)

	for range 3 {
		value, ok := expr.Eval(Env{})
		this.True(ok)
		this.Equal(20.0, value.Num())
	} // for
}
