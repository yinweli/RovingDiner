package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuitePhaseGameStart(t *testing.T) {
	suite.Run(t, new(SuitePhaseGameStart))
}

// SuitePhaseGameStart 驗證營業開始階段（phaseGameStart.go）:設定載入 / 歸零設置 / 累積表清空 / 前置技能 / gameStart 觸發。
type SuitePhaseGameStart struct {
	suite.Suite
}

func (this *SuitePhaseGameStart) TestPhaseGameStart() {
	count := 0
	fired := 0
	data := tester.BuildData()
	game := newGameData(data)
	game.PrefixSkill = []int32{301, 999} // 999:查無技能 → 效果列表空、不啟動（防禦）
	data.SetEffect(401, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) { count++ }})
	data.SetEffect(402, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) { count++ }})
	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerGameStart, Trigger: func(game *cores.Game) { fired++ }})
	game.Effect.Push(cores.NewEffect(game, 901, cores.Ref{}, 1))

	game.GetRound().Set(7) // 佈置髒狀態 → 驗證歸零
	game.Settling = true
	game.GetDrawTotal().Add(1)
	game.GetDropTotal().Add(1)
	game.GetPlayTotal().Add(1)
	game.GetExileTotal().Add(1)

	this.Equal(cores.PhaseRoundStart, phaseGameStart(game))
	this.Equal(int32(12), game.GetRoundMax().GetValue()) // 設定六鍵載入
	this.Equal(int32(30), game.GetMorale().GetValue())
	this.Equal(int32(50), game.GetMoraleMax().GetValue())
	this.Equal(int32(3), game.GetEnergyMax().GetValue())
	this.Equal(int32(10), game.GetHandMax().GetValue())
	this.Equal(int32(5), game.GetDrawMax().GetValue())

	this.Equal(int32(0), game.GetRound().GetValue()) // 回合 / 結算旗標歸零
	this.False(game.Settling)
	this.Equal(int32(3), game.GetEnergy().GetValue()) // 點數補滿
	this.Equal(int32(0), game.GetDrawTotal().Sum())   // 四累積表清空
	this.Equal(int32(0), game.GetDropTotal().Sum())
	this.Equal(int32(0), game.GetPlayTotal().Sum())
	this.Equal(int32(0), game.GetExileTotal().Sum())

	this.Equal(2, count) // 前置技能 301 → 效果 401 / 402 各啟動一次
	this.Equal(1, fired) // 營業開始觸發
}

// TestSettingNum 驗證 settingNum 讀數字設定值;缺鍵 / 空值 / 非數字回 0。
func (this *SuitePhaseGameStart) TestSettingNum() {
	game := newGame()
	game.GetSheet().Setting.Data["Empty"] = &sheeter.Setting{ID: "Empty"}
	game.GetSheet().Setting.Data["Bad"] = &sheeter.Setting{ID: "Bad", Value: []string{"x"}}

	this.Equal(float64(12), settingNum(game, "RoundMax"))
	this.Equal(float64(0), settingNum(game, "nope"))  // 缺鍵 → 0
	this.Equal(float64(0), settingNum(game, "Empty")) // 空值 → 0
	this.Equal(float64(0), settingNum(game, "Bad"))   // 非數字 → 0
}
