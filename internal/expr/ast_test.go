package expr

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteAst(t *testing.T) {
	suite.Run(t, new(SuiteAst))
}

// SuiteAst 驗證求值語意:條件對象解析、比較 / 算術失敗、ref 比較、短路、三元惰性、內建分派。
type SuiteAst struct {
	suite.Suite
}

// eval 解析並以指定 Resolver 求值,回傳結果與成功旗標。
func (this *SuiteAst) eval(source string, resolver Resolver) (value Value, ok bool) {
	expr, err := Parse(source)
	this.Require().NoError(err)
	return expr.Eval(resolver)
}

func (this *SuiteAst) TestNodePropertyEval() {
	resolver := &mockResolver{attr: map[string]Value{"morale": NewNum(50)}}

	value, ok := this.eval("morale", resolver)
	this.Require().True(ok)
	this.Equal(50.0, value.Num())

	value, ok = this.eval("morale > 0", resolver)
	this.Require().True(ok)
	this.True(value.Bool())

	_, ok = this.eval("missing", resolver) // 未登記屬性 → 解析失敗
	this.False(ok)
}

func (this *SuiteAst) TestNodeCallEval() {
	// 查詢函式引數應原樣傳入:第一參數字串 '>='、第二參數數值 2
	resolver := &mockResolver{
		onAttr: func(name string, arg []Value) (value Value, ok bool) {
			match := name == "tableCount" && len(arg) == 2 &&
				arg[0].IsText() && arg[0].Text() == ">=" &&
				arg[1].IsNum() && arg[1].Num() == 2
			return NewBool(match), true
		},
	}

	value, ok := this.eval("tableCount('>=', 2)", resolver)
	this.Require().True(ok)
	this.True(value.Bool())
}

func (this *SuiteAst) TestNodeCallBuiltin() {
	// nodeCall 先查內建函式註冊表、再交給 Resolver:內建名優先,即使 resolver 對該名會失敗仍走內建實作。
	// (max / min 的語意與引數校驗詳見 builtin_test.go。)
	resolver := &mockResolver{
		onAttr: func(name string, arg []Value) (value Value, ok bool) {
			return Value{}, false // 任何查詢函式都失敗
		},
	}

	value, ok := this.eval("max(1, 5, 3)", resolver)
	this.Require().True(ok)
	this.Equal(5.0, value.Num()) // 未落到 resolver → 證明走內建

	_, ok = this.eval("tableCount(1)", resolver) // 非內建名 → 落到 resolver(此處設計為失敗)
	this.False(ok)
}

func (this *SuiteAst) TestNodeMemberEval() {
	resolver := &mockResolver{
		attr:    map[string]Value{"self": NewRef(1)},
		attrRef: map[string]Value{"calm": NewNum(7)},
	}

	value, ok := this.eval("self.calm", resolver)
	this.Require().True(ok)
	this.Equal(7.0, value.Num())

	value, ok = this.eval("self.calm > 5", resolver)
	this.Require().True(ok)
	this.True(value.Bool())

	_, ok = this.eval("self.unknown", resolver) // 未登記引用屬性 → 失敗
	this.False(ok)
}

func (this *SuiteAst) TestNodeMemberBaseFail() {
	resolver := &mockResolver{attrRef: map[string]Value{"calm": NewNum(7)}} // self 未登記(未固定)
	_, ok := this.eval("self.calm", resolver)
	this.False(ok)
}

func (this *SuiteAst) TestNodeMemberNoneFail() {
	resolver := &mockResolver{
		attr:    map[string]Value{"self": NewNone()}, // self 為空物件
		attrRef: map[string]Value{"calm": NewNum(7)},
	}
	_, ok := this.eval("self.calm", resolver)
	this.False(ok) // 空物件存取屬性 → 失敗
}

func (this *SuiteAst) TestNodeUnaryEval() {
	value, ok := this.eval("-5", nil) // 一元負號:數值取負
	this.Require().True(ok)
	this.Equal(-5.0, value.Num())

	_, ok = this.eval("-'x'", nil) // 負號施於非數值 → 失敗
	this.False(ok)

	value, ok = this.eval("!true", nil) // 否定:布林取反
	this.Require().True(ok)
	this.False(value.Bool())

	value, ok = this.eval("!0", nil) // 否定:數值經 AsBool 協調(0 → 假 → 取反為真)
	this.Require().True(ok)
	this.True(value.Bool())

	_, ok = this.eval("!'x'", nil) // 否定施於字串 → 無法協調為布林 → 失敗
	this.False(ok)
}

func (this *SuiteAst) TestNodeBinaryLogic() {
	resolver := &mockResolver{} // 所有屬性皆未登記 → missing 會失敗

	this.evalEqualBool("false AND missing", resolver, false) // 左假 → 不評右側
	this.evalEqualBool("true OR missing", resolver, true)    // 左真 → 不評右側

	_, ok := this.eval("true AND missing", resolver) // 右側被評估 → 失敗
	this.False(ok)

	_, ok = this.eval("false OR missing", resolver) // 右側被評估 → 失敗
	this.False(ok)
}

func (this *SuiteAst) TestNodeBinaryLogicGuard() {
	// 守衛式:self 為空物件時,因短路而不存取 self.calm
	resolver := &mockResolver{attr: map[string]Value{"self": NewNone(), "empty": NewNone()}}

	value, ok := this.eval("self != empty AND self.calm > 5", resolver)
	this.Require().True(ok)
	this.False(value.Bool()) // self == empty → 守衛為假 → 整體為假
}

func (this *SuiteAst) TestNodeBinaryLogicAsBool() {
	resolver := &mockResolver{attr: map[string]Value{"morale": NewNum(5)}}

	this.evalTrue("morale AND 1", resolver) // 非 0 數值協調為真

	_, ok := this.eval("'x' AND 1", nil) // 字串無法協調為布林 → 失敗
	this.False(ok)
}

func (this *SuiteAst) TestNodeBinaryEqual() {
	resolver := &mockResolver{attr: map[string]Value{
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

func (this *SuiteAst) TestNodeBinaryOrder() {
	resolver := &mockResolver{attr: map[string]Value{"self": NewRef(1), "seatLast": NewRef(2)}}

	_, ok := this.eval("self < seatLast", resolver) // 物件引用不可比大小
	this.False(ok)

	_, ok = this.eval("'a' < 'b'", nil) // 字串不可比大小
	this.False(ok)
}

func (this *SuiteAst) TestNodeBinaryArith() {
	resolver := &mockResolver{attr: map[string]Value{"self": NewRef(1)}}

	_, ok := this.eval("1 / 0", nil) // 除 0
	this.False(ok)

	_, ok = this.eval("1 % 0", nil) // 取餘 0
	this.False(ok)

	_, ok = this.eval("1 + 'a'", nil) // 非數值參與算術
	this.False(ok)

	_, ok = this.eval("self + 1", resolver) // 物件引用參與算術
	this.False(ok)
}

func (this *SuiteAst) TestCrossTypeCompare() {
	resolver := &mockResolver{attr: map[string]Value{"morale": NewNum(1), "self": NewRef(1)}}

	_, ok := this.eval("morale == 'x'", resolver) // 數值 vs 字串(evalEqual)
	this.False(ok)

	_, ok = this.eval("self == morale", resolver) // 物件引用 vs 數值(evalEqual)
	this.False(ok)

	_, ok = this.eval("1 < 'a'", nil) // 數值 vs 字串大小比較(evalOrder)
	this.False(ok)
}

func (this *SuiteAst) TestNodeTernaryEval() {
	resolver := &mockResolver{} // missing 未登記

	value, ok := this.eval("1 ? 10 : missing", resolver)
	this.Require().True(ok)
	this.Equal(10.0, value.Num()) // 未選中分支不評估

	value, ok = this.eval("0 ? missing : 20", resolver)
	this.Require().True(ok)
	this.Equal(20.0, value.Num())

	_, ok = this.eval("missing ? 1 : 2", resolver) // 條件失敗 → 整體失敗
	this.False(ok)
}

// evalTrue 斷言運算式求值成功且為布林真。
func (this *SuiteAst) evalTrue(source string, resolver Resolver) {
	value, ok := this.eval(source, resolver)
	this.Require().Truef(ok, "應求值成功: %s", source)
	this.Truef(value.Bool(), "應為真: %s", source)
}

// evalFalse 斷言運算式求值成功且為布林假。
func (this *SuiteAst) evalFalse(source string, resolver Resolver) {
	value, ok := this.eval(source, resolver)
	this.Require().Truef(ok, "應求值成功: %s", source)
	this.Falsef(value.Bool(), "應為假: %s", source)
}

// evalEqualBool 斷言運算式求值成功且布林結果等於 expect。
func (this *SuiteAst) evalEqualBool(source string, resolver Resolver, expect bool) {
	value, ok := this.eval(source, resolver)
	this.Require().Truef(ok, "應求值成功: %s", source)
	this.Equalf(expect, value.Bool(), "布林結果不符: %s", source)
}
