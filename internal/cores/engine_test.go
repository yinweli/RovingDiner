package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteEngine(t *testing.T) {
	suite.Run(t, new(SuiteEngine))
}

// SuiteEngine 驗證驅動引擎對 exprs.Resolver 的委派與 Lock 後綴路由(Engine.go);
// 詞條內容的逐項驗證見 attrRead_test.go / attrRefRead_test.go,此處只驗分派機制。
type SuiteEngine struct {
	suite.Suite
}

func (this *SuiteEngine) TestEngineAttr() {
	runtime := NewRuntime(0)
	runtime.Game.morale = NewValue(30, 2)
	eng := &Engine{runtime: runtime}

	// 值表命中
	value, ok := eng.Attr("morale", nil)
	this.True(ok)
	this.Equal(float64(30), value.Num())

	// Lock 後綴 → 剝後綴查鎖表
	value, ok = eng.Attr("moraleLock", nil)
	this.True(ok)
	this.Equal(float64(2), value.Num())

	// 未知名稱 → 失敗
	_, ok = eng.Attr("nope", nil)
	this.False(ok)

	// 非鎖屬性的 Lock 後綴 → 失敗(round 為「寫」無鎖,鎖表無此鍵)
	_, ok = eng.Attr("roundLock", nil)
	this.False(ok)
}

func (this *SuiteEngine) TestEngineAttrRef() {
	eng := &Engine{runtime: NewRuntime(0)}
	ref := NewRefGuest(&Guest{calm: NewValue(5, 1)})

	// 值表命中
	value, ok := eng.AttrRef(ref, "calm", nil)
	this.True(ok)
	this.Equal(float64(5), value.Num())

	// Lock 後綴 → 剝後綴查鎖表
	value, ok = eng.AttrRef(ref, "calmLock", nil)
	this.True(ok)
	this.Equal(float64(1), value.Num())

	// 未知名稱 → 失敗
	_, ok = eng.AttrRef(ref, "nope", nil)
	this.False(ok)
}

func (this *SuiteEngine) TestEngineExecAssignGlobal() {
	runtime := NewRuntime(0)
	runtime.Game.score = NewValue(10, 0)
	eng := NewEngine(runtime, nil, nil, nil, nil, nil)

	// 全域帶值賦值;RHS 經 exprs 求值(含內建函式 min,驗證 builtin 注入)
	this.True(eng.ExecAssign("score", "", false, AssignAdd, this.expr("min(5, 8)")))
	this.Equal(int32(15), runtime.Game.GetScore().GetValue())

	// @ 鎖定(不帶右值,value 為 nil、不求值)
	this.True(eng.ExecAssign("score", "", false, AssignLock, nil))
	this.Equal(int32(1), runtime.Game.GetScore().GetLock())

	// 未知全域屬性 → no-op
	this.False(eng.ExecAssign("nope", "", false, AssignSet, this.expr("1")))

	// RHS 評估失敗(除 0)→ no-op
	this.False(eng.ExecAssign("energy", "", false, AssignSet, this.expr("1 / 0")))

	// RHS 非數值(字串)→ no-op
	this.False(eng.ExecAssign("energy", "", false, AssignSet, this.expr("'text'")))
}

func (this *SuiteEngine) TestEngineExecAssignRef() {
	card := &Card{instanceID: 1}
	runtime := NewRuntime(0)
	runtime.Game.drawLast = card
	eng := NewEngine(runtime, nil, nil, nil, nil, nil)

	// 引用左值寫入(最後抽出卡牌的出牌費用設為 3)
	this.True(eng.ExecAssign("drawLast", "cost", true, AssignSet, this.expr("3")))
	this.Equal(int32(3), card.GetCost().GetValue())

	// 引用屬性型別不符(卡牌引用寫顧客屬性 calm)→ no-op
	this.False(eng.ExecAssign("drawLast", "calm", true, AssignSet, this.expr("3")))

	// 未知引用屬性 → no-op
	this.False(eng.ExecAssign("drawLast", "nope", true, AssignSet, this.expr("3")))

	// 引用解析為空物件(drawLast 為 nil)→ no-op
	runtime.Game.drawLast = nil
	this.False(eng.ExecAssign("drawLast", "cost", true, AssignSet, this.expr("3")))
}

func (this *SuiteEngine) TestEngineSelectObject() {
	runtime := NewRuntime(0)
	runtime.Game.drawLast = &Card{instanceID: 7}
	eng := NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{}, nil)

	result, ok := eng.selectObject("drawLast", nil) // 已登錄 → 派發至詞條
	this.True(ok)
	this.Equal([]InstanceID{7}, result)

	_, ok = eng.selectObject("nope", nil) // 未登錄命令對象 → ok=false
	this.False(ok)
}

// expr 解析算術式來源為 *exprs.Expr;解析失敗即測試失敗。
func (this *SuiteEngine) expr(source string) *exprs.Expr {
	result, err := exprs.Parse(source)
	this.Require().NoError(err)
	return result
}
