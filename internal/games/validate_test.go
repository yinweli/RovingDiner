package games

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteValidate(t *testing.T) {
	suite.Run(t, new(SuiteValidate))
}

// SuiteValidate 驗證命令語言詞彙表的 Validate(validate.go)逐名查全域表。
// M5 詞彙表為空,故操作命令名稱一律未知(詞條於 M6+ 填入後生效);屬性修改命令的可寫性驗證留待 M7。
type SuiteValidate struct {
	suite.Suite
}

func (this *SuiteValidate) TestValidate() {
	// 操作命令:空詞彙表 → 未知命令(命令名稱先於命令對象被檢出)
	operate, err := Parse("handAdd(none, 10031, 1)")
	this.Require().NoError(err)
	this.Error(Validate(operate))

	// 屬性修改命令:M5 的 Validate 尚不檢查可寫性(留待 M7)→ 不報錯
	assign, err := Parse("morale = 10")
	this.Require().NoError(err)
	this.NoError(Validate(assign))
}
