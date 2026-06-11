package games

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/tester"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(SuiteRun))
}

// SuiteRun 驗證營業執行入口（run.go）:跑通完整一局（成功 / 失敗）、決定性、Compiler 閉包（命令編譯 + 結算尾、語法錯跳過）。
type SuiteRun struct {
	suite.Suite
}

// TestRun 驗證跑通第一局:關卡 601 兩位顧客耐心耗盡生氣離場 → 全場清空 → 營業成功;
// 同 seed + 同關卡 + 同輸入重跑同結果（決定性）;高耐心關卡 603 耗到回合上限 → 失敗。
func (this *SuiteRun) TestRun() {
	this.True(Run(1, 601, tester.BuildSheet(), tester.FakeOperator{}, nil))
	this.True(Run(1, 601, tester.BuildSheet(), tester.FakeOperator{}, nil)) // 決定性:重跑同結果
	this.False(Run(1, 603, tester.BuildSheet(), tester.FakeOperator{}, nil))
}

// TestRunCompile 驗證 Compiler 閉包:前置技能掛真命令字串——編譯執行（含結算尾）則排隊顧客離場、首結算全場清空而獲勝;
// 語法錯命令 → 該效果整個跳過、顧客存活耗到回合上限而落敗。勝負翻轉證明編譯路徑真的執行。
func (this *SuiteRun) TestRunCompile() {
	sheet := tester.BuildSheet()
	sheet.Effect.Data[405] = &sheeter.Effect{ID: 405, CommandImmed: "guestExit(guestWait[1], false, false)"}
	sheet.Effect.Data[406] = &sheeter.Effect{ID: 406, CommandImmed: "guestExit(guestWait[1]"} // 語法錯 → prepareEffect 跳過
	sheet.Skill.Data[302] = &sheeter.Skill{ID: 302, EffectID: []int32{405}}
	sheet.Skill.Data[303] = &sheeter.Skill{ID: 303, EffectID: []int32{406}}
	sheet.Stage.Data[604] = &sheeter.Stage{ID: 604, Name: "編譯成功", WaitID: []int32{502}, PrefixSkillID: []int32{302}}
	sheet.Stage.Data[605] = &sheeter.Stage{ID: 605, Name: "編譯失敗", WaitID: []int32{502}, PrefixSkillID: []int32{303}}

	this.True(Run(1, 604, sheet, tester.FakeOperator{}, nil))  // 命令執行 → 顧客離場 → 結算尾判全場清空 → 成功
	this.False(Run(1, 605, sheet, tester.FakeOperator{}, nil)) // 語法錯效果被跳過 → 顧客存活 → 回合上限失敗
}
