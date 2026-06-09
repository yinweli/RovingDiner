package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteEngine(t *testing.T) {
	suite.Run(t, new(SuiteEngine))
}

// SuiteEngine 驗證驅動引擎對 exprs.Resolver 的委派與 Lock 後綴路由(engine.go);
// 詞條內容的逐項驗證見 attrRead_test.go / attrRefRead_test.go,此處只驗分派機制。
type SuiteEngine struct {
	suite.Suite
}

func (this *SuiteEngine) TestEngineAttr() {
	runtime := NewRuntime(0)
	runtime.Game.Morale = Value{Value: 30, Lock: 2}
	eng := &engine{runtime: runtime}

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
	eng := &engine{runtime: NewRuntime(0)}
	ref := guestRef{guest: &Guest{Calm: Value{Value: 5, Lock: 1}}}

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
