package expr

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteParser(t *testing.T) {
	suite.Run(t, new(SuiteParser))
}

// SuiteParser 驗證解析:優先級 / 結合性(透過純字面值求值觀察)、合法語法成功、非法語法回錯。
type SuiteParser struct {
	suite.Suite
}

// num 解析並對「純字面值」運算式求值(無需 Resolver),回傳數值結果。
func (this *SuiteParser) num(source string) float64 {
	expr, err := Parse(source)
	this.Require().NoError(err)

	value, ok := expr.Eval(nil)
	this.Require().True(ok)
	this.Require().True(value.IsNumber())

	return value.Number()
}

// truth 解析並對「純字面值」運算式求值,回傳布林結果。
func (this *SuiteParser) truth(source string) bool {
	expr, err := Parse(source)
	this.Require().NoError(err)

	value, ok := expr.Eval(nil)
	this.Require().True(ok)
	this.Require().True(value.IsBool())

	return value.Bool()
}

func (this *SuiteParser) TestArithPrecedence() {
	this.Equal(7.0, this.num("1 + 2 * 3"))   // 乘除先於加減
	this.Equal(9.0, this.num("(1 + 2) * 3")) // 括號最先
	this.Equal(10.0, this.num("2 * 3 + 4"))
	this.Equal(1.0, this.num("10 % 3"))
}

func (this *SuiteParser) TestLeftAssociative() {
	this.Equal(5.0, this.num("10 - 2 - 3")) // (10-2)-3
	this.Equal(2.0, this.num("8 / 2 / 2"))  // (8/2)/2
}

func (this *SuiteParser) TestUnaryMinus() {
	this.Equal(1.0, this.num("-2 + 3"))
	this.Equal(-5.0, this.num("-(2 + 3)"))
	this.Equal(-6.0, this.num("2 * -3"))
}

func (this *SuiteParser) TestComparisonLooserThanArith() {
	this.True(this.truth("1 + 1 == 2")) // 算術先於比較
	this.True(this.truth("2 * 2 > 3"))
}

func (this *SuiteParser) TestNotPrecedence() {
	this.True(this.truth("!0"))        // 0 → 假 → 否定 → 真
	this.False(this.truth("!5"))       // 非 0 → 真 → 否定 → 假
	this.True(this.truth("!(1 == 2)")) // 否定整個比較
	this.True(this.truth("!1 == 0"))   // 否定低於比較:!(1 == 0) → !假 → 真
}

func (this *SuiteParser) TestAndOrPrecedence() {
	this.True(this.truth("true OR false AND false"))  // AND 緊於 OR:true OR (false AND false)
	this.False(this.truth("true AND false OR false")) // (true AND false) OR false
	this.True(this.truth("false AND true OR true"))   // (false AND true) OR true
}

func (this *SuiteParser) TestTernary() {
	this.Equal(10.0, this.num("1 ? 10 : 20"))
	this.Equal(20.0, this.num("0 ? 10 : 20"))
	this.Equal(2.0, this.num("1 ? 2 : 0 ? 3 : 4")) // 右結合:1 ? 2 : (0 ? 3 : 4)
	this.Equal(3.0, this.num("0 ? 1 : 1 ? 3 : 4")) // 否則分支續解析三元
}

func (this *SuiteParser) TestValidSyntax() {
	source := []string{
		"morale > 0 AND self.calm >= 5",
		"tableCount('>=', 2)",
		"max(1, 2, 3)",
		"self != exitLast",
		"round == roundMax ? 1 : 0",
		"self.effectStack(101) > 0",
	}

	for _, itor := range source {
		_, err := Parse(itor)
		this.NoErrorf(err, "應可解析: %s", itor)
	} // for
}

func (this *SuiteParser) TestSyntaxError() {
	source := []string{
		"",          // 空輸入
		"1 +",       // 缺右運算元
		"(1",        // 括號未閉合
		"1 ? 2",     // 三元缺 :
		"a.",        // '.' 後缺屬性名
		"1 < 2 < 3", // 鏈式比較(多餘 token)
		"1 2",       // 多餘 token
		")",         // 未預期 token
		"foo(1,",    // 引數列未閉合
	}

	for _, itor := range source {
		_, err := Parse(itor)
		this.Errorf(err, "應回語法錯誤: %q", itor)
	} // for
}
