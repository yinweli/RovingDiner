package expr

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteEval(t *testing.T) {
	suite.Run(t, new(SuiteEval))
}

// SuiteEval 驗證求值語意:條件對象解析、比較 / 算術失敗、ref 比較、短路、三元惰性、內建函式。
type SuiteEval struct {
	suite.Suite
}

// eval 解析並以指定 Resolver 求值,回傳結果與成功旗標。
func (this *SuiteEval) eval(source string, resolver Resolver) (value Value, ok bool) {
	expr, err := Parse(source)
	this.Require().NoError(err)
	return expr.Eval(resolver)
}

func (this *SuiteEval) TestProperty() {
	resolver := &mockResolver{property: map[string]Value{"morale": NewNumber(50)}}

	value, ok := this.eval("morale", resolver)
	this.Require().True(ok)
	this.Equal(50.0, value.Number())

	value, ok = this.eval("morale > 0", resolver)
	this.Require().True(ok)
	this.True(value.Bool())

	_, ok = this.eval("missing", resolver) // 未登記屬性 → 解析失敗
	this.False(ok)
}

func (this *SuiteEval) TestMember() {
	resolver := &mockResolver{
		property: map[string]Value{"self": NewRef(1)},
		member:   map[string]Value{"calm": NewNumber(7)},
	}

	value, ok := this.eval("self.calm", resolver)
	this.Require().True(ok)
	this.Equal(7.0, value.Number())

	value, ok = this.eval("self.calm > 5", resolver)
	this.Require().True(ok)
	this.True(value.Bool())

	_, ok = this.eval("self.unknown", resolver) // 未登記引用屬性 → 失敗
	this.False(ok)
}

func (this *SuiteEval) TestMemberBaseFail() {
	resolver := &mockResolver{member: map[string]Value{"calm": NewNumber(7)}} // self 未登記(未固定)
	_, ok := this.eval("self.calm", resolver)
	this.False(ok)
}

func (this *SuiteEval) TestMemberOnNoneFail() {
	resolver := &mockResolver{
		property: map[string]Value{"self": NewNone()}, // self 為空物件
		member:   map[string]Value{"calm": NewNumber(7)},
	}
	_, ok := this.eval("self.calm", resolver)
	this.False(ok) // 空物件存取屬性 → 失敗
}

func (this *SuiteEval) TestQueryFunction() {
	// 查詢函式引數應原樣傳入:第一參數字串 '>='、第二參數數值 2
	resolver := &mockResolver{
		onProperty: func(name string, arg []Value) (value Value, ok bool) {
			match := name == "tableCount" && len(arg) == 2 &&
				arg[0].IsString() && arg[0].Str() == ">=" &&
				arg[1].IsNumber() && arg[1].Number() == 2
			return NewBool(match), true
		},
	}

	value, ok := this.eval("tableCount('>=', 2)", resolver)
	this.Require().True(ok)
	this.True(value.Bool())
}

func (this *SuiteEval) TestBuiltinMaxMin() {
	value, ok := this.eval("max(1, 5, 3)", nil)
	this.Require().True(ok)
	this.Equal(5.0, value.Number())

	value, ok = this.eval("min(1, 5, 3)", nil)
	this.Require().True(ok)
	this.Equal(1.0, value.Number())

	_, ok = this.eval("max(1)", nil) // 至少 2 引數
	this.False(ok)

	_, ok = this.eval("max(1, 'x')", nil) // 非數值參與
	this.False(ok)
}

func (this *SuiteEval) TestRefEquality() {
	resolver := &mockResolver{property: map[string]Value{
		"self":     NewRef(1),
		"seatLast": NewRef(1),
		"exitLast": NewRef(2),
		"empty":    NewNone(),
	}}

	this.evalTrue("self == seatLast", resolver)  // 同實例編號
	this.evalFalse("self == exitLast", resolver) // 不同實例編號
	this.evalTrue("self != exitLast", resolver)
	this.evalTrue("empty == empty", resolver) // 空物件只與空物件相等
	this.evalFalse("self == empty", resolver) // 非空 vs 空
}

func (this *SuiteEval) TestCrossTypeFail() {
	resolver := &mockResolver{property: map[string]Value{"morale": NewNumber(1), "self": NewRef(1)}}

	_, ok := this.eval("morale == 'x'", resolver) // 數值 vs 字串
	this.False(ok)

	_, ok = this.eval("self == morale", resolver) // 物件引用 vs 數值
	this.False(ok)

	_, ok = this.eval("1 < 'a'", nil) // 跨型別大小比較
	this.False(ok)
}

func (this *SuiteEval) TestOrderingNonNumberFail() {
	resolver := &mockResolver{property: map[string]Value{"self": NewRef(1), "seatLast": NewRef(2)}}

	_, ok := this.eval("self < seatLast", resolver) // 物件引用不可比大小
	this.False(ok)

	_, ok = this.eval("'a' < 'b'", nil) // 字串不可比大小
	this.False(ok)
}

func (this *SuiteEval) TestArithFail() {
	resolver := &mockResolver{property: map[string]Value{"self": NewRef(1)}}

	_, ok := this.eval("1 / 0", nil) // 除 0
	this.False(ok)

	_, ok = this.eval("1 % 0", nil) // 取餘 0
	this.False(ok)

	_, ok = this.eval("1 + 'a'", nil) // 非數值參與算術
	this.False(ok)

	_, ok = this.eval("self + 1", resolver) // 物件引用參與算術
	this.False(ok)
}

func (this *SuiteEval) TestShortCircuit() {
	resolver := &mockResolver{} // 所有屬性皆未登記 → missing 會失敗

	this.evalEqualBool("false AND missing", resolver, false) // 左假 → 不評右側
	this.evalEqualBool("true OR missing", resolver, true)    // 左真 → 不評右側

	_, ok := this.eval("true AND missing", resolver) // 右側被評估 → 失敗
	this.False(ok)

	_, ok = this.eval("false OR missing", resolver) // 右側被評估 → 失敗
	this.False(ok)
}

func (this *SuiteEval) TestGuardIdiom() {
	// 守衛式:self 為空物件時,因短路而不存取 self.calm
	resolver := &mockResolver{property: map[string]Value{"self": NewNone(), "empty": NewNone()}}

	value, ok := this.eval("self != empty AND self.calm > 5", resolver)
	this.Require().True(ok)
	this.False(value.Bool()) // self == empty → 守衛為假 → 整體為假
}

func (this *SuiteEval) TestTernaryLazy() {
	resolver := &mockResolver{} // missing 未登記

	value, ok := this.eval("1 ? 10 : missing", resolver)
	this.Require().True(ok)
	this.Equal(10.0, value.Number()) // 未選中分支不評估

	value, ok = this.eval("0 ? missing : 20", resolver)
	this.Require().True(ok)
	this.Equal(20.0, value.Number())

	_, ok = this.eval("missing ? 1 : 2", resolver) // 條件失敗 → 整體失敗
	this.False(ok)
}

func (this *SuiteEval) TestBareArithAsBool() {
	resolver := &mockResolver{property: map[string]Value{"morale": NewNumber(5)}}

	this.evalTrue("morale AND 1", resolver) // 非 0 數值協調為真

	_, ok := this.eval("'x' AND 1", nil) // 字串無法協調為布林 → 失敗
	this.False(ok)
}

// evalTrue 斷言運算式求值成功且為布林真。
func (this *SuiteEval) evalTrue(source string, resolver Resolver) {
	value, ok := this.eval(source, resolver)
	this.Require().Truef(ok, "應求值成功: %s", source)
	this.Truef(value.Bool(), "應為真: %s", source)
}

// evalFalse 斷言運算式求值成功且為布林假。
func (this *SuiteEval) evalFalse(source string, resolver Resolver) {
	value, ok := this.eval(source, resolver)
	this.Require().Truef(ok, "應求值成功: %s", source)
	this.Falsef(value.Bool(), "應為假: %s", source)
}

// evalEqualBool 斷言運算式求值成功且布林結果等於 expect。
func (this *SuiteEval) evalEqualBool(source string, resolver Resolver, expect bool) {
	value, ok := this.eval(source, resolver)
	this.Require().Truef(ok, "應求值成功: %s", source)
	this.Equalf(expect, value.Bool(), "布林結果不符: %s", source)
}
