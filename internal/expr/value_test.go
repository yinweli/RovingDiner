package expr

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteValue(t *testing.T) {
	suite.Run(t, new(SuiteValue))
}

// SuiteValue 驗證 Value 標籤聯合的建構 / 取值、AsBool 布林協調、Round 四捨五入。
type SuiteValue struct {
	suite.Suite
}

func (this *SuiteValue) TestNumber() {
	value := NewNumber(1.5)
	this.True(value.IsNumber())
	this.False(value.IsString())
	this.False(value.IsBool())
	this.False(value.IsRef())
	this.False(value.IsNone())
	this.Equal(1.5, value.Number())
}

func (this *SuiteValue) TestString() {
	value := NewString("玩家行動")
	this.True(value.IsString())
	this.False(value.IsNumber())
	this.Equal("玩家行動", value.Str())
}

func (this *SuiteValue) TestBool() {
	value := NewBool(true)
	this.True(value.IsBool())
	this.False(value.IsNumber())
	this.True(value.Bool())
}

func (this *SuiteValue) TestRef() {
	value := NewRef(42)
	this.True(value.IsRef())
	this.False(value.IsNone())
	this.Equal(int64(42), value.RefID())
}

func (this *SuiteValue) TestNone() {
	value := NewNone()
	this.True(value.IsRef())
	this.True(value.IsNone())
	this.Equal(int64(0), value.RefID()) // 空物件無實例編號
}

func (this *SuiteValue) TestAsBool() {
	result, ok := AsBool(NewBool(true))
	this.True(ok)
	this.True(result)

	result, ok = AsBool(NewNumber(5)) // 非 0 → 真
	this.True(ok)
	this.True(result)

	result, ok = AsBool(NewNumber(0)) // 0 → 假
	this.True(ok)
	this.False(result)

	_, ok = AsBool(NewString("x")) // 字串 → 失敗
	this.False(ok)

	_, ok = AsBool(NewRef(1)) // 物件引用 → 失敗
	this.False(ok)
}

func (this *SuiteValue) TestRound() {
	this.Equal(int32(2), Round(2.4))
	this.Equal(int32(3), Round(2.5)) // half-away-from-zero
	this.Equal(int32(2), Round(2.0))
	this.Equal(int32(-2), Round(-2.4))
	this.Equal(int32(-3), Round(-2.5)) // half-away-from-zero(負向)
}
