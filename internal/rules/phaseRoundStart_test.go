package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuitePhaseRoundStart(t *testing.T) {
	suite.Run(t, new(SuitePhaseRoundStart))
}

// SuitePhaseRoundStart 驗證回合開始階段(phaseRoundStart.go): 回合遞增 / 計數歸零 / 兩觸發 / 點數補滿 / 入座迴圈 / 下一階段分派。
type SuitePhaseRoundStart struct {
	suite.Suite
}

func (this *SuitePhaseRoundStart) TestPhaseRoundStart() {
	ready := 0
	start := 0
	seat := 0
	data := tester.BuildData()
	game := newGameData(data)
	game.GetEnergyMax().Set(3)
	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerRoundReady, Trigger: func(game *cores.Game) { ready++ }})
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerRoundStart, Trigger: func(game *cores.Game) { start++ }})
	data.SetEffect(903, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerGuestSeat, Trigger: func(game *cores.Game) { seat++ }})

	for _, itor := range []int32{901, 902, 903} {
		game.Effect.Push(cores.NewEffect(game, itor, cores.Ref{}, 1))
	} // for

	game.EventSeat(cores.NewGuest(game, 501)) // 髒回合計數 → 驗證 RoundReset 歸零後重計
	game.Wait = cores.WaitList{cores.NewGuest(game, 501), cores.NewGuest(game, 501), cores.NewGuest(game, 501), cores.NewGuest(game, 501)}

	this.Equal(cores.PhasePlayerAction, phaseRoundStart(game)) // 無跳轉 → 預設玩家行動
	this.Equal(int32(1), game.GetRound().GetValue())           // 回合 +1
	this.Equal(1, ready)
	this.Equal(1, start)
	this.Equal(int32(3), game.GetEnergy().GetValue()) // 點數補滿
	this.Equal(3, seat)                               // 三空位 → 三位入座(每位觸發一次)
	this.Equal(int32(3), game.GetSeatCount())         // RoundReset 歸零後重計 3
	this.Len(game.Wait, 1)                            // 第四位仍排隊
	this.NotNil(game.Seat[1])
	this.NotNil(game.Seat[2])
	this.NotNil(game.Seat[3])

	// 下一階段分派: 顧客行動 / 回合結束, 皆清除跳轉
	game.SetNextPhase(cores.PhaseGuestAction)
	this.Equal(cores.PhaseGuestAction, phaseRoundStart(game))
	this.Equal(cores.PhaseNone, game.GetNextPhase())

	game.SetNextPhase(cores.PhaseRoundEnd)
	this.Equal(cores.PhaseRoundEnd, phaseRoundStart(game))
	this.Equal(cores.PhaseNone, game.GetNextPhase())
}
