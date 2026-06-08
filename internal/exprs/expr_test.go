package exprs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteExpr(t *testing.T) {
	suite.Run(t, new(SuiteExpr))
}

// SuiteExpr 驗證對外入口 Parse / Expr / Eval:成功解析後可求值、詞法錯誤上拋、
// 結尾多餘內容報錯,以及 parse-once / eval-many(同一 Expr 多次求值結果一致)。
type SuiteExpr struct {
	suite.Suite
}

func (this *SuiteExpr) TestParse() {
	expr, err := Parse("1 + 2 * 3")
	this.Require().NoError(err)
	this.Require().NotNil(expr)

	value, ok := expr.Eval()
	this.True(ok)
	this.Equal(7.0, value.Num())
}

func (this *SuiteExpr) TestParseLexError() {
	// 詞法錯誤(未知字元 / 字串未結束 / 裸 =)由 Parse 直接上拋,型別仍為 SyntaxError
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

func (this *SuiteExpr) TestExprReuse() {
	// parse 一次、求值多次,結果穩定一致(對齊 parse-once / eval-many 設計)
	expr, err := Parse("(2 + 3) * 4")
	this.Require().NoError(err)

	for range 3 {
		value, ok := expr.Eval()
		this.True(ok)
		this.Equal(20.0, value.Num())
	} // for
}
