package games

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteValidate(t *testing.T) {
	suite.Run(t, new(SuiteValidate))
}

// SuiteValidate 驗證命令語言詞彙表的 Validate(validate.go)逐名查全域表。
// 操作命令詞彙表 M8 / M9 前為空故 verb 一律未知;屬性修改命令查左值可寫性(M7 生效)。
type SuiteValidate struct {
	suite.Suite
}

func (this *SuiteValidate) TestValidateOperate() {
	// 操作命令:空詞彙表 → 未知命令(命令名稱先於命令對象被檢出)
	operate, err := Parse("handAdd(none, 10031, 1)")
	this.Require().NoError(err)
	this.Error(Validate(operate))
}

func (this *SuiteValidate) TestValidateAssign() {
	// 全域可寫屬性 → 通過
	this.NoError(Validate(this.parse("morale = 10")))
	// 全域唯讀屬性 → 報錯(nextPhase 僅 phaseJump 可設)
	this.Error(Validate(this.parse("nextPhase = '玩家行動'")))
	// 全域未知屬性 → 報錯
	this.Error(Validate(this.parse("nope = 1")))

	// 引用可寫屬性 → 通過
	this.NoError(Validate(this.parse("self.calm += 1")))
	// 引用唯讀屬性 → 報錯(cardID 唯讀)
	this.Error(Validate(this.parse("drawLast.cardID = 1")))
}

// parse 解析命令來源為 Command;解析失敗即測試失敗(供 Validate 斷言聚焦於語意檢查)。
func (this *SuiteValidate) parse(source string) Command {
	command, err := Parse(source)
	this.Require().NoError(err)
	return command
}
