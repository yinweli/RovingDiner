package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuitePhaseGameEnd(t *testing.T) {
	suite.Run(t, new(SuitePhaseGameEnd))
}

// SuitePhaseGameEnd 驗證終止階段(phaseGameEnd.go): gameSucc / gameFail 觸發與 PhaseNone 停機訊號。
type SuitePhaseGameEnd struct {
	suite.Suite
}

func (this *SuitePhaseGameEnd) TestPhaseGameSucc() {
	fired := 0
	data := tester.BuildData()
	game := newGameData(data)
	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerGameSucc, Trigger: func(game *cores.Game) { fired++ }})
	game.Effect.Push(cores.NewEffect(game, 901, cores.Ref{}, 1))

	this.Equal(cores.PhaseNone, phaseGameSucc(game))
	this.Equal(1, fired)
}

func (this *SuitePhaseGameEnd) TestPhaseGameFail() {
	fired := 0
	data := tester.BuildData()
	game := newGameData(data)
	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerGameFail, Trigger: func(game *cores.Game) { fired++ }})
	game.Effect.Push(cores.NewEffect(game, 901, cores.Ref{}, 1))

	this.Equal(cores.PhaseNone, phaseGameFail(game))
	this.Equal(1, fired)
}
