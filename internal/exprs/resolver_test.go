package exprs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteResolver(t *testing.T) {
	suite.Run(t, new(SuiteResolver))
}

// SuiteResolver 驗證 exprs↔games 接縫:條件對象經 Resolver(Attr / AttrRef)解析、
// 內建函式經 Builtin 註冊表分派(優先於查詢函式),以及各種解析失敗與 self / none 守衛。
type SuiteResolver struct {
	suite.Suite
}

func (this *SuiteResolver) TestEvalIdentAttr() {
	env := Env{Resolver: &stubResolver{attr: map[string]Value{"morale": NewNum(5)}}}

	value, ok := this.eval("morale", env)
	this.True(ok)
	this.Equal(5.0, value.Num())

	value, ok = this.eval("morale + 1", env)
	this.True(ok)
	this.Equal(6.0, value.Num()) // 屬性值參與算術
}

func (this *SuiteResolver) TestEvalIdentRef() {
	env := Env{Resolver: &stubResolver{attr: map[string]Value{"self": NewRef(stubRef{id: 1})}}}

	value, ok := this.eval("self", env)
	this.True(ok)
	this.True(value.IsRef())
}

func (this *SuiteResolver) TestEvalIdentFail() {
	env := Env{Resolver: &stubResolver{attr: map[string]Value{}}}
	this.False(this.ok("unknown", env))  // 屬性不存在
	this.False(this.ok("morale", Env{})) // 無 Resolver → 失敗
}

func (this *SuiteResolver) TestEvalCallBuiltin() {
	env := Env{Builtin: stubBuiltin()}

	value, ok := this.eval("min(3, 1, 2)", env)
	this.True(ok)
	this.Equal(1.0, value.Num())

	value, ok = this.eval("max(3, 1, 2)", env)
	this.True(ok)
	this.Equal(3.0, value.Num())
}

func (this *SuiteResolver) TestEvalCallBuiltinFail() {
	env := Env{Builtin: stubBuiltin()}
	this.False(this.ok("min(1)", env))        // 少於 2 個參數
	this.False(this.ok("min(1, 'a')", env))   // 非數值參數
	this.False(this.ok("min(1, 1 / 0)", env)) // 參數評估失敗 → 整體失敗
	this.False(this.ok("foo(1, 2)", Env{}))   // 既非內建函式、亦無 Resolver → 失敗
}

func (this *SuiteResolver) TestEvalCallQuery() {
	// argCount 為測試用查詢函式,回傳參數個數,藉此驗證參數有確實傳入 Resolver
	env := Env{Resolver: &stubResolver{}}

	value, ok := this.eval("argCount('>=', 2)", env)
	this.True(ok)
	this.Equal(2.0, value.Num())
}

func (this *SuiteResolver) TestEvalCallPrecedence() {
	// 同名時內建函式優先於查詢函式:Resolver 不認得 min(Attr 會失敗),仍應走內建而成功
	env := Env{Resolver: &stubResolver{attr: map[string]Value{}}, Builtin: stubBuiltin()}

	value, ok := this.eval("min(2, 5)", env)
	this.True(ok)
	this.Equal(2.0, value.Num())
}

func (this *SuiteResolver) TestEvalRefAttr() {
	env := Env{Resolver: &stubResolver{
		attr:    map[string]Value{"self": NewRef(stubRef{id: 1})},
		refAttr: map[int]map[string]Value{1: {"calm": NewNum(7)}},
	}}

	value, ok := this.eval("self.calm", env)
	this.True(ok)
	this.Equal(7.0, value.Num())
	this.True(this.boolean("self.calm > 5", env))
}

func (this *SuiteResolver) TestEvalRefQuery() {
	env := Env{Resolver: &stubResolver{attr: map[string]Value{"self": NewRef(stubRef{id: 1})}}}

	value, ok := this.eval("self.argCount(101, 202)", env)
	this.True(ok)
	this.Equal(2.0, value.Num()) // 參數有傳入 AttrRef
}

func (this *SuiteResolver) TestEvalRefFail() {
	env := Env{Resolver: &stubResolver{
		attr:    map[string]Value{"self": NewNone(), "round": NewNum(3)},
		refAttr: map[int]map[string]Value{1: {"calm": NewNum(7)}},
	}}
	this.False(this.ok("self.calm", env))    // 主體為空物件 → 失敗
	this.False(this.ok("round.calm", env))   // 主體非引用(數值) → 失敗
	this.False(this.ok("unknown.calm", env)) // 主體名稱無法解析 → 失敗
	this.False(this.ok("self.calm", Env{}))  // 無 Resolver → 失敗

	env2 := Env{Resolver: &stubResolver{attr: map[string]Value{"self": NewRef(stubRef{id: 1})}}}
	this.False(this.ok("self.unknown", env2))         // 引用屬性不存在 → 失敗
	this.False(this.ok("self.argCount(1 / 0)", env2)) // 引用查詢函式參數評估失敗 → 失敗
}

func (this *SuiteResolver) TestEvalObjectCompare() {
	env := Env{Resolver: &stubResolver{attr: map[string]Value{
		"self":  NewRef(stubRef{id: 1}),
		"alias": NewRef(stubRef{id: 1}), // 同一實例編號
		"other": NewRef(stubRef{id: 2}),
		"empty": NewNone(),
	}}}
	this.True(this.boolean("self == alias", env))  // 比實例編號:同
	this.False(this.boolean("self == other", env)) // 不同實例
	this.True(this.boolean("self != other", env))
	this.True(this.boolean("self != empty", env)) // 引用 != 空物件
	this.False(this.boolean("self == empty", env))
	this.True(this.boolean("empty == none", env)) // 解析到的空物件 == none 字面值
	this.False(this.ok("self == 5", env))         // 跨型別 → 失敗
}

func (this *SuiteResolver) TestEvalSelfGuard() {
	// self 為空物件:守衛 self != none 為假 → AND 短路、不評估 self.calm,整體為假(不失敗)
	envNone := Env{Resolver: &stubResolver{attr: map[string]Value{"self": NewNone()}}}
	this.False(this.boolean("self != none AND self.calm > 5", envNone))

	// self 為引用且 calm == 7:守衛成立 → 評估 self.calm > 5 → 真
	envRef := Env{Resolver: &stubResolver{
		attr:    map[string]Value{"self": NewRef(stubRef{id: 1})},
		refAttr: map[int]map[string]Value{1: {"calm": NewNum(7)}},
	}}
	this.True(this.boolean("self != none AND self.calm > 5", envRef))
}

// eval 以給定 env 解析並求值 source(解析失敗即 fail 測試)。
func (this *SuiteResolver) eval(source string, env Env) (result Value, ok bool) {
	expr, err := Parse(source)
	this.Require().NoError(err, source)
	return expr.Eval(env)
}

// boolean 求值並斷言為成功的布林值,回傳其布林值。
func (this *SuiteResolver) boolean(source string, env Env) bool {
	value, ok := this.eval(source, env)
	this.Require().True(ok, source)
	this.Require().True(value.IsBool(), source)
	return value.Bool()
}

// ok 求值並回傳是否成功(供失敗案例斷言)。
func (this *SuiteResolver) ok(source string, env Env) bool {
	_, ok := this.eval(source, env)
	return ok
}

// stubRef 測試用物件引用,以 id 代表實例編號(IsSame 比 id)。
type stubRef struct {
	id int
}

func (this stubRef) IsSame(other Ref) bool {
	stub, ok := other.(stubRef)
	return ok && this.id == stub.id
}

// stubResolver 測試用 Resolver:attr 以名稱查主表、refAttr 以「引用 id + 屬性名」查子表;
// 名稱 argCount 視為查詢函式,回傳傳入的參數個數(便於驗證參數確實送達)。
type stubResolver struct {
	attr    map[string]Value
	refAttr map[int]map[string]Value
}

func (this *stubResolver) Attr(name string, arg []Value) (result Value, ok bool) {
	if name == "argCount" {
		return NewNum(float64(len(arg))), true
	} // if

	value, exist := this.attr[name]
	return value, exist
}

func (this *stubResolver) AttrRef(ref Ref, name string, arg []Value) (result Value, ok bool) {
	stub, okRef := ref.(stubRef)

	if okRef == false {
		return Value{}, false
	} // if

	if name == "argCount" {
		return NewNum(float64(len(arg))), true
	} // if

	attr, exist := this.refAttr[stub.id]

	if exist == false {
		return Value{}, false
	} // if

	value, exist := attr[name]
	return value, exist
}

// stubBuiltin 提供與【營業規格書 | 二十六、內建函式清單】一致的 min / max(≥2 個數值參數)。
func stubBuiltin() map[string]Builtin {
	return map[string]Builtin{
		"min": stubFold(func(a, b float64) float64 { return min(a, b) }),
		"max": stubFold(func(a, b float64) float64 { return max(a, b) }),
	}
}

// stubFold 以 pick 摺疊參數;少於 2 個或任一非數值即失敗。
func stubFold(pick func(a, b float64) float64) Builtin {
	return func(arg []Value) (result Value, ok bool) {
		if len(arg) < 2 {
			return Value{}, false
		} // if

		if arg[0].IsNum() == false {
			return Value{}, false
		} // if

		acc := arg[0].Num()

		for _, itor := range arg[1:] {
			if itor.IsNum() == false {
				return Value{}, false
			} // if

			acc = pick(acc, itor.Num())
		} // for

		return NewNum(acc), true
	}
}
