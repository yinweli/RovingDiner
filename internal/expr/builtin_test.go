package expr

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteBuiltin(t *testing.T) {
	suite.Run(t, new(SuiteBuiltin))
}

// SuiteBuiltin 直接驗證內建函式實作:註冊表內容、max / min 語意、引數校驗,不經 parser。
type SuiteBuiltin struct {
	suite.Suite
}

// TestBuiltinRegistry 確認註冊表只登記規格書【二十六、內建函式清單】的 max / min,且皆可呼叫。
func (this *SuiteBuiltin) TestBuiltinRegistry() {
	this.Len(builtin, 2)
	this.NotNil(builtin["max"])
	this.NotNil(builtin["min"])
}

func (this *SuiteBuiltin) TestBuiltinMax() {
	value, ok := builtinMax([]Value{NewNum(1), NewNum(5), NewNum(3)})
	this.Require().True(ok)
	this.Equal(5.0, value.Num())

	value, ok = builtinMax([]Value{NewNum(-2.5), NewNum(-1)}) // 負數 / 小數
	this.Require().True(ok)
	this.Equal(-1.0, value.Num())

	value, ok = builtinMax([]Value{NewNum(4), NewNum(4)}) // 相等取其一
	this.Require().True(ok)
	this.Equal(4.0, value.Num())

	_, ok = builtinMax([]Value{NewNum(1)}) // 少於 2 引數 → 失敗
	this.False(ok)

	_, ok = builtinMax(nil) // 零引數 → 失敗
	this.False(ok)

	_, ok = builtinMax([]Value{NewNum(1), NewText("x")}) // 含字串 → 失敗
	this.False(ok)

	_, ok = builtinMax([]Value{NewNum(1), NewBool(true)}) // 含布林 → 失敗
	this.False(ok)

	_, ok = builtinMax([]Value{NewNum(1), NewRef(2)}) // 含物件引用 → 失敗
	this.False(ok)
}

func (this *SuiteBuiltin) TestBuiltinMin() {
	value, ok := builtinMin([]Value{NewNum(1), NewNum(5), NewNum(3)})
	this.Require().True(ok)
	this.Equal(1.0, value.Num())

	value, ok = builtinMin([]Value{NewNum(-2.5), NewNum(-1)}) // 負數 / 小數
	this.Require().True(ok)
	this.Equal(-2.5, value.Num())

	_, ok = builtinMin([]Value{NewNum(1)}) // 少於 2 引數 → 失敗
	this.False(ok)

	_, ok = builtinMin([]Value{NewNum(1), NewText("x")}) // 含非數值 → 失敗
	this.False(ok)
}

func (this *SuiteBuiltin) TestNumericArg() {
	num, ok := numericArg([]Value{NewNum(1), NewNum(2), NewNum(3)}, 2)
	this.Require().True(ok)
	this.Equal([]float64{1, 2, 3}, num)

	num, ok = numericArg([]Value{NewNum(7)}, 1) // 剛好達到下限
	this.Require().True(ok)
	this.Equal([]float64{7}, num)

	_, ok = numericArg([]Value{NewNum(1)}, 2) // 不足下限 → 失敗
	this.False(ok)

	_, ok = numericArg([]Value{NewNum(1), NewText("x")}, 2) // 含非數值 → 失敗
	this.False(ok)
}
