package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuitePhase(t *testing.T) {
	suite.Run(t, new(SuitePhase))
}

// SuitePhase 驗證核心流程分派（phase.go）:RunPhase 七站分派、哨兵 recover 與停機、跨站共用 helper。
type SuitePhase struct {
	suite.Suite
}

// TestRunPhase 驗證 RunPhase 分派七站並以空盤面走完最小一局:回合結束的結算判定「全場清空」→ 哨兵跳出 →
// recover 轉營業成功站 → 停機;非哨兵 panic 原樣重拋。
func (this *SuitePhase) TestRunPhase() {
	game := newGame()
	this.Equal(cores.PhaseRoundStart, RunPhase(game, cores.PhaseGameStart))
	this.Equal(cores.PhasePlayerAction, RunPhase(game, cores.PhaseRoundStart))
	this.Equal(cores.PhaseGuestAction, RunPhase(game, cores.PhasePlayerAction)) // FakeOperator 恆回玩家結束
	this.Equal(cores.PhaseRoundEnd, RunPhase(game, cores.PhaseGuestAction))     // 行動佇列空 → 收尾
	this.Equal(cores.PhaseGameSucc, RunPhase(game, cores.PhaseRoundEnd))        // 空盤面結算 → 全場清空 → 哨兵 → 營業成功
	this.Equal(cores.PhaseNone, RunPhase(game, cores.PhaseGameSucc))
	this.Equal(cores.PhaseNone, RunPhase(game, cores.PhaseGameFail))
	this.Equal(cores.PhaseNone, RunPhase(game, cores.PhaseNone)) // 停機

	this.PanicsWithValue("boom", func() { // 非哨兵 panic → 原樣重拋
		data := tester.BuildData()
		bad := newGameData(data)
		data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerRoundEnd, Trigger: func(game *cores.Game) { panic("boom") }})
		bad.Effect.Push(cores.NewEffect(bad, 901, cores.Ref{}, 1))
		RunPhase(bad, cores.PhaseRoundEnd)
	})

	// 踏站接線:已知階段先 SetPhase 再發 phase 切換事件（座標即新階段）;PhaseNone 不踏站不發（M18）
	wired, record := newGameRecord()
	RunPhase(wired, cores.PhaseGameStart)
	this.Equal(cores.PhaseGameStart, wired.GetPhase())
	this.Require().NotEmpty(record.Event)
	this.Equal(cores.EventPhase, record.Event[0].Kind)
	this.Equal(cores.PhaseGameStart, record.Event[0].Phase)

	count := len(record.Event)
	RunPhase(wired, cores.PhaseNone)
	this.Len(record.Event, count) // 停機 → 無事件
}

// TestEnergyFill 驗證 energyFill 點數補滿:低於上限補至上限、高於上限保留、鎖定不補。
func (this *SuitePhase) TestEnergyFill() {
	game := newGame()
	game.GetEnergyMax().Set(3)

	energyFill(game)
	this.Equal(int32(3), game.GetEnergy().GetValue()) // 低於上限 → 補滿

	game.GetEnergy().Set(5)
	energyFill(game)
	this.Equal(int32(5), game.GetEnergy().GetValue()) // 高於上限 → 保留

	game.GetEnergy().Set(1)
	game.GetEnergy().Lock()
	energyFill(game)
	this.Equal(int32(1), game.GetEnergy().GetValue()) // 鎖定 → 不補
}

// TestSkillGroup 驗證 skillGroup 取技能群組編號;技能資料缺失回 0。
func (this *SuitePhase) TestSkillGroup() {
	game := newGame()
	this.Equal(int32(3), skillGroup(game, 301))
	this.Equal(int32(0), skillGroup(game, 999)) // 技能資料缺失 → 0
}

// TestCardSkill 驗證 cardSkill 取卡牌技能編號（玩家出牌範圍事件的技能操作元）;卡牌資料缺失回 0。
func (this *SuitePhase) TestCardSkill() {
	game := newGame()
	this.Equal(int32(301), cardSkill(game, 103))
	this.Equal(int32(0), cardSkill(game, 101)) // 卡 101 無技能 → 0
	this.Equal(int32(0), cardSkill(game, 999)) // 卡牌資料缺失 → 0
}
