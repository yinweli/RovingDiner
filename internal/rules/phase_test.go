package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuitePhase(t *testing.T) {
	suite.Run(t, new(SuitePhase))
}

// SuitePhase 驗證核心流程分派（phase.go）:runPhase 七站分派與停機、跨站共用 helper。
type SuitePhase struct {
	suite.Suite
}

// TestRunPhase 驗證 runPhase 分派七站並以空盤面走完最小一輪;PhaseNone / 未知 → 停機。
func (this *SuitePhase) TestRunPhase() {
	game := newGame()
	this.Equal(cores.PhaseRoundStart, runPhase(game, cores.PhaseGameStart))
	this.Equal(cores.PhasePlayerAction, runPhase(game, cores.PhaseRoundStart))
	this.Equal(cores.PhaseGuestAction, runPhase(game, cores.PhasePlayerAction)) // FakeOperator 恆回玩家結束
	this.Equal(cores.PhaseRoundEnd, runPhase(game, cores.PhaseGuestAction))     // 行動佇列空 → 收尾
	this.Equal(cores.PhaseRoundStart, runPhase(game, cores.PhaseRoundEnd))
	this.Equal(cores.PhaseNone, runPhase(game, cores.PhaseGameSucc))
	this.Equal(cores.PhaseNone, runPhase(game, cores.PhaseGameFail))
	this.Equal(cores.PhaseNone, runPhase(game, cores.PhaseNone)) // 停機
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
