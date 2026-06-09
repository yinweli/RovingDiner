package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteBuiltin(t *testing.T) {
	suite.Run(t, new(SuiteBuiltin))
}

// SuiteBuiltin 驗證內建函式註冊表(builtin.go):min / max 摺疊、參數不足 / 非數值失敗。
type SuiteBuiltin struct {
	suite.Suite
}

func (this *SuiteBuiltin) TestBuiltinMinMax() {
	maxFunc := builtin["max"]
	minFunc := builtin["min"]
	this.Require().NotNil(maxFunc)
	this.Require().NotNil(minFunc)

	value, ok := maxFunc([]exprs.Value{exprs.NewNum(3), exprs.NewNum(7), exprs.NewNum(5)})
	this.True(ok)
	this.Equal(float64(7), value.Num())

	value, ok = minFunc([]exprs.Value{exprs.NewNum(3), exprs.NewNum(7), exprs.NewNum(5)})
	this.True(ok)
	this.Equal(float64(3), value.Num())

	_, ok = maxFunc([]exprs.Value{exprs.NewNum(1)}) // 少於 2 參數 → 失敗
	this.False(ok)

	_, ok = maxFunc([]exprs.Value{exprs.NewText("a"), exprs.NewNum(1)}) // 首參非數值 → 失敗
	this.False(ok)

	_, ok = minFunc([]exprs.Value{exprs.NewNum(1), exprs.NewBool(true)}) // 後續參數非數值 → 失敗
	this.False(ok)
}
