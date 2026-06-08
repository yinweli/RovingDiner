package exprs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteEval(t *testing.T) {
	suite.Run(t, new(SuiteEval))
}

// SuiteEval 驗證求值語意:字面值、否定、邏輯短路、算術(含除 0 / 取餘 0)、大小 / 相等比較(含跨型別失敗)、
// 三元惰性求值、優先序,以及評估失敗以 ok == false 表達。測試順序對齊 eval.go 的 dispatch(節點型別序)。
type SuiteEval struct {
	suite.Suite
}

func (this *SuiteEval) TestEvalLiteral() {
	this.Equal(5.0, this.num("5"))
	this.Equal(1.5, this.num("1.5"))
	this.True(this.boolean("true"))

	value, ok := this.eval("'hi'")
	this.True(ok)
	this.True(value.IsText())
	this.Equal("hi", value.Text())

	value, ok = this.eval("none")
	this.True(ok)
	this.True(value.IsNone())
}

func (this *SuiteEval) TestEvalNot() {
	this.False(this.boolean("!true"))
	this.True(this.boolean("!false"))
	this.True(this.boolean("!0"))  // 0 為假 → 否定為真
	this.False(this.boolean("!5")) // 非 0 為真 → 否定為假
	this.fail("!'a'")
	this.fail("!none")
}

func (this *SuiteEval) TestEvalLogical() {
	this.False(this.boolean("true AND false"))
	this.True(this.boolean("true AND true"))
	this.True(this.boolean("false OR true"))
	this.True(this.boolean("1 AND 2")) // 非 0 皆為真
	this.False(this.boolean("0 OR 0"))
	this.True(this.boolean("5 OR 0"))
}

func (this *SuiteEval) TestEvalLogicalTruthyFail() {
	this.fail("'a' AND true") // 左側無真假判定
	this.fail("true AND 'a'") // 右側無真假判定,左真才評估右側
	this.fail("none OR false")
}

func (this *SuiteEval) TestEvalLogicalShortCircuit() {
	this.False(this.boolean("false AND (1 / 0)")) // AND 左假,右不評估,無除 0 失敗
	this.True(this.boolean("true OR (1 / 0)"))    // OR 左真,右不評估
	this.fail("true AND (1 / 0)")                 // 左真,右被評估 → 除 0 失敗
	this.fail("false OR (1 / 0)")                 // 左假,右被評估 → 除 0 失敗
	this.fail("(1 / 0) AND true")                 // 左運算元評估失敗 → 整體失敗
}

func (this *SuiteEval) TestEvalArith() {
	this.Equal(3.0, this.num("1 + 2"))
	this.Equal(7.0, this.num("2 * 3 + 1"))
	this.Equal(2.5, this.num("10 / 4")) // 中間值可為小數,不四捨五入
	this.Equal(1.0, this.num("10 % 3"))
	this.Equal(-3.0, this.num("2 - 5"))
	this.Equal(-5.0, this.num("-(2 + 3)"))
}

func (this *SuiteEval) TestEvalArithFail() {
	this.fail("1 / 0")   // 除 0
	this.fail("5 % 0")   // 取餘 0
	this.fail("'a' + 1") // 非數值參與算術
	this.fail("true + 1")
	this.fail("none + 1")
	this.fail("-'a'")        // 負號僅適用數值
	this.fail("(1 / 0) + 1") // 左運算元評估失敗 → 整體失敗
	this.fail("1 + (1 / 0)") // 右運算元評估失敗 → 整體失敗
}

func (this *SuiteEval) TestEvalOrder() {
	this.True(this.boolean("3 > 2"))
	this.True(this.boolean("2 >= 2"))
	this.False(this.boolean("3 < 2"))
	this.False(this.boolean("1 < 0"))
	this.True(this.boolean("2 <= 3"))
	this.False(this.boolean("3 <= 2"))
	this.fail("'a' < 'b'") // 字串無大小順序
	this.fail("true < false")
	this.fail("none < none")
}

func (this *SuiteEval) TestEvalEqual() {
	this.True(this.boolean("2 == 2"))
	this.True(this.boolean("2 != 3"))
	this.True(this.boolean("'a' == 'a'"))
	this.False(this.boolean("'a' == 'b'"))
	this.True(this.boolean("'a' != 'b'"))
	this.True(this.boolean("true == true"))
	this.False(this.boolean("true == false"))
	this.True(this.boolean("none == none")) // 空物件只與空物件相等
	this.False(this.boolean("none != none"))
}

func (this *SuiteEval) TestEvalEqualCrossType() {
	this.fail("1 == 'a'")
	this.fail("1 == true")
	this.fail("1 == none")
	this.fail("'a' == true")
	this.fail("none == 1")
}

func (this *SuiteEval) TestEvalTernary() {
	this.Equal(1.0, this.num("true ? 1 : 2"))
	this.Equal(2.0, this.num("false ? 1 : 2"))
	this.Equal(10.0, this.num("1 ? 10 : 20")) // 非 0 條件為真
	this.Equal(20.0, this.num("0 ? 10 : 20"))
}

func (this *SuiteEval) TestEvalTernaryLazy() {
	this.Equal(1.0, this.num("true ? 1 : (1 / 0)")) // 未選分支不評估
	this.Equal(2.0, this.num("false ? (1 / 0) : 2"))
	this.fail("'a' ? 1 : 2") // 條件失敗 → 整個三元失敗
	this.fail("(1 / 0) ? 1 : 2")
}

func (this *SuiteEval) TestEvalPrecedence() {
	this.Equal(-4.0, this.num("1 - 2 - 3"))            // 左結合
	this.Equal(14.0, this.num("2 + 3 * 4"))            // 乘緊於加
	this.Equal(20.0, this.num("(2 + 3) * 4"))          // 括號最高
	this.True(this.boolean("true OR false AND false")) // AND 緊於 OR → OR(true, false)
}

// eval 解析並求值 source(解析失敗即 fail 測試);回傳求值結果與是否成功。
func (this *SuiteEval) eval(source string) (result Value, ok bool) {
	expr, err := Parse(source)
	this.Require().NoError(err, source)
	return expr.Eval(Env{})
}

// num 求值 source 並斷言為成功的數值,回傳其數值。
func (this *SuiteEval) num(source string) float64 {
	value, ok := this.eval(source)
	this.Require().True(ok, source)
	this.Require().True(value.IsNum(), source)
	return value.Num()
}

// boolean 求值 source 並斷言為成功的布林值,回傳其布林值。
func (this *SuiteEval) boolean(source string) bool {
	value, ok := this.eval(source)
	this.Require().True(ok, source)
	this.Require().True(value.IsBool(), source)
	return value.Bool()
}

// fail 求值 source 並斷言評估失敗(ok == false)。
func (this *SuiteEval) fail(source string) {
	_, ok := this.eval(source)
	this.Require().False(ok, source)
}
