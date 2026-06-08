package exprs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteValue(t *testing.T) {
	suite.Run(t, new(SuiteValue))
}

// SuiteValue 驗證求值結果型別 Value 的建構 / 取值、真假判定(Truthy)與四捨五入(Round)。
type SuiteValue struct {
	suite.Suite
}

func (this *SuiteValue) TestValueNum() {
	value := NewNum(1.5)
	this.True(value.IsNum())
	this.Equal(1.5, value.Num())
	this.False(value.IsBool())
	this.False(value.IsText())
	this.False(value.IsNone())
}

func (this *SuiteValue) TestValueBool() {
	value := NewBool(true)
	this.True(value.IsBool())
	this.True(value.Bool())
	this.False(value.IsNum())
}

func (this *SuiteValue) TestValueText() {
	value := NewText("玩家行動")
	this.True(value.IsText())
	this.Equal("玩家行動", value.Text())
	this.False(value.IsNum())
}

func (this *SuiteValue) TestValueNone() {
	value := NewNone()
	this.True(value.IsNone())
	this.False(value.IsNum())
	this.False(value.IsBool())
	this.False(value.IsText())
}

func (this *SuiteValue) TestValueRef() {
	ref := stubRef{id: 7}
	value := NewRef(ref)
	this.True(value.IsRef())
	this.Equal(Ref(ref), value.Ref())
	this.False(value.IsNum())
	this.False(value.IsNone())
}

func (this *SuiteValue) TestValueTruthy() {
	result, ok := NewBool(true).Truthy()
	this.True(ok)
	this.True(result)

	result, ok = NewBool(false).Truthy()
	this.True(ok)
	this.False(result)

	result, ok = NewNum(5).Truthy()
	this.True(ok)
	this.True(result) // 非 0 為真

	result, ok = NewNum(0).Truthy()
	this.True(ok)
	this.False(result) // 0 為假

	_, ok = NewText("x").Truthy()
	this.False(ok) // 字串無真假判定

	_, ok = NewNone().Truthy()
	this.False(ok) // 空物件無真假判定

	_, ok = NewRef(stubRef{id: 1}).Truthy()
	this.False(ok) // 物件引用無真假判定
}

func (this *SuiteValue) TestRound() {
	this.Equal(int32(2), Round(2.4))
	this.Equal(int32(3), Round(2.5)) // half-away-from-zero
	this.Equal(int32(-2), Round(-2.4))
	this.Equal(int32(-3), Round(-2.5))
	this.Equal(int32(0), Round(0))
	this.Equal(int32(2), Round(1.5))
}
