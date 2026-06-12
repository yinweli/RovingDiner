package games

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteValidate(t *testing.T) {
	suite.Run(t, new(SuiteValidate))
}

// SuiteValidate 驗證命令語言詞彙表的 Validate(validate.go)逐名查全域表。
// 操作命令查 verb / 命令對象 / 命令對象 [...] 參數數量(M28); 屬性修改命令查左值可寫性(M7 生效)。
type SuiteValidate struct {
	suite.Suite
}

func (this *SuiteValidate) TestValidateOperate() {
	// 已登錄命令 + 合法命令對象 → 通過(M9 詞條到位後生效)
	this.NoError(Validate(this.parse("handAdd(none, 10031, 1)")))
	// 未登錄命令 → 報錯
	this.Error(Validate(this.parse("noSuchCommand(none)")))
	// 未登錄命令對象 → 報錯
	this.Error(Validate(this.parse("handAdd(noSuchSelector)")))
}

// TestValidateArity 驗證命令對象 [...] 參數數量校驗(M28; 依【二十四、命令對象清單 | 參數規則】裸寫 / 數量不符即報錯)。
func (this *SuiteValidate) TestValidateArity() {
	// 數量相符 → 通過(參數為算術式, 數量看 token 個數不看值)
	this.NoError(Validate(this.parse("guestExit(guestPick[1], 1, 1)")))
	this.NoError(Validate(this.parse("handToDrop(handAll[0])")))
	this.NoError(Validate(this.parse("handToDrop(handPick[1 + 1, 0])")))
	// 裸寫缺參數 → 報錯
	this.Error(Validate(this.parse("guestExit(guestPick, 1, 1)")))
	this.Error(Validate(this.parse("handToDrop(handAll)")))
	// 數量不足 / 過多 → 報錯
	this.Error(Validate(this.parse("handToDrop(handPick[2])")))
	this.Error(Validate(this.parse("handAdd(none[1], 10031, 1)")))
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
	this.NoError(Validate(this.parse("drawLast.cost = 1")))
	// 引用唯讀屬性 → 報錯(cardID 唯讀)
	this.Error(Validate(this.parse("drawLast.cardID = 1")))
	// 引用基底非物件引用 → 報錯(M28; morale 為數值屬性、不可作 <基底>.<屬性> 基底)
	this.Error(Validate(this.parse("morale.cost = 1")))
	// 引用基底未登錄 → 報錯(M28; 先前僅查屬性可寫性會漏過)
	this.Error(Validate(this.parse("nope.cost = 1")))
}

// parse 解析命令來源為 Command; 解析失敗即測試失敗(供 Validate 斷言聚焦於語意檢查)。
func (this *SuiteValidate) parse(source string) Command {
	command, err := Parse(source)
	this.Require().NoError(err)
	return command
}
