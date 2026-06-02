package expr

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteExpr(t *testing.T) {
	suite.Run(t, new(SuiteExpr))
}

// SuiteExpr 以代表性 spec 運算式做端到端驗證:Parse 後對真實感的 Resolver 求值,並驗一次編譯多次求值。
type SuiteExpr struct {
	suite.Suite
}

func (this *SuiteExpr) TestTriggerCondition() {
	// 觸發條件:餐廳士氣值為正且自身耐心值達標(守衛、引用屬性、邏輯運算)
	resolver := &mockResolver{
		property: map[string]Value{"morale": NewNumber(30), "self": NewRef(1)},
		member:   map[string]Value{"calm": NewNumber(6)},
	}

	expr, err := Parse("morale > 0 AND self.calm >= 5")
	this.Require().NoError(err)

	value, ok := expr.Eval(resolver)
	this.Require().True(ok)
	this.True(value.Bool())
}

func (this *SuiteExpr) TestTriggerCount() {
	// 觸發次數:三元決定次數(條件、三元、屬性比較)
	resolver := &mockResolver{property: map[string]Value{"round": NewNumber(5), "roundMax": NewNumber(5)}}

	expr, err := Parse("round == roundMax ? 1 : 0")
	this.Require().NoError(err)

	value, ok := expr.Eval(resolver)
	this.Require().True(ok)
	this.Equal(1.0, value.Number())
}

func (this *SuiteExpr) TestExitReasonByRef() {
	// 離場原因:以 self == exitLast 區分(物件引用比較)
	resolver := &mockResolver{property: map[string]Value{"self": NewRef(7), "exitLast": NewRef(7)}}

	expr, err := Parse("self == exitLast")
	this.Require().NoError(err)

	value, ok := expr.Eval(resolver)
	this.Require().True(ok)
	this.True(value.Bool())
}

func (this *SuiteExpr) TestArithWithBuiltin() {
	// 命令 RHS 算術式:內建函式夾住下限(屬性、內建函式、算術運算)
	resolver := &mockResolver{property: map[string]Value{"morale": NewNumber(-3)}}

	expr, err := Parse("max(morale, 0) + 1")
	this.Require().NoError(err)

	value, ok := expr.Eval(resolver)
	this.Require().True(ok)
	this.Equal(1.0, value.Number()) // 下限夾為零後再加一,結果為一
}

func (this *SuiteExpr) TestQueryFunction() {
	// 查詢函式:桌數查詢(字串運算子、數值門檻)
	resolver := &mockResolver{
		onProperty: func(name string, arg []Value) (value Value, ok bool) {
			if name == "tableCount" && len(arg) == 2 {
				return NewNumber(3), true // 假裝有 3 桌滿足條件
			} // if

			return Value{}, false
		},
	}

	expr, err := Parse("tableCount('>=', 2) > 0")
	this.Require().NoError(err)

	value, ok := expr.Eval(resolver)
	this.Require().True(ok)
	this.True(value.Bool())
}

func (this *SuiteExpr) TestCompileOnceEvalMany() {
	// 同一份編譯結果對不同 Runtime(以不同 Resolver 模擬)多次求值
	expr, err := Parse("morale <= 0")
	this.Require().NoError(err)

	alive := &mockResolver{property: map[string]Value{"morale": NewNumber(10)}}
	dead := &mockResolver{property: map[string]Value{"morale": NewNumber(0)}}

	value, ok := expr.Eval(alive)
	this.Require().True(ok)
	this.False(value.Bool())

	value, ok = expr.Eval(dead)
	this.Require().True(ok)
	this.True(value.Bool())
}
