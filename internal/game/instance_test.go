package game

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteInstance(t *testing.T) {
	suite.Run(t, new(SuiteInstance))
}

// SuiteInstance 驗證 instance.go 各實例型別的方法行為。
type SuiteInstance struct {
	suite.Suite
}

// TestValueLocked 驗證 Value.Locked 僅在鎖定計數 > 0 時回報鎖定。
func (this *SuiteInstance) TestValueLocked() {
	this.False(Value{Value: 5, Lock: 0}.Locked())  // 未鎖定
	this.True(Value{Value: 0, Lock: 1}.Locked())   // 鎖定一層
	this.True(Value{Value: 9, Lock: 3}.Locked())   // 鎖定多層；Value 不影響鎖定判定
	this.False(Value{Value: 0, Lock: -1}.Locked()) // 計數非正不算鎖定
}

// TestSelfIsNone 驗證 Self.IsNone 僅在卡牌與顧客皆為 nil 時回報空物件。
func (this *SuiteInstance) TestSelfIsNone() {
	this.True(Self{}.IsNone())                                // 空物件
	this.False(Self{Card: &Card{}}.IsNone())                  // self 為卡牌
	this.False(Self{Guest: &Guest{}}.IsNone())                // self 為顧客
	this.False(Self{Card: &Card{}, Guest: &Guest{}}.IsNone()) // 兩欄皆非 nil（防禦性，正常不應發生）
}
