package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuitePhaseRoundEnd(t *testing.T) {
	suite.Run(t, new(SuitePhaseRoundEnd))
}

// SuitePhaseRoundEnd 驗證回合結束階段（phaseRoundEnd.go）:roundEnd 觸發 / 耐心遞減 / 點數歸零與保留 / 推進效果接線。
type SuitePhaseRoundEnd struct {
	suite.Suite
}

func (this *SuitePhaseRoundEnd) TestPhaseRoundEnd() {
	fired := 0
	ended := 0
	data := tester.BuildData()
	game := newGameData(data)
	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerRoundEnd, Trigger: func(game *cores.Game) { fired++ }})
	data.SetEffect(905, cores.EffectData{Kind: cores.EffectPersist, RunRound: 1, End: func(game *cores.Game) { ended++ }})
	game.GetRoundMax().Set(99) // 結算已接線:佈置回合上限與士氣,避免終止判定命中
	game.GetMorale().Set(30)
	game.GetRound().Set(1)
	game.Effect.Push(cores.NewEffect(game, 901, cores.Ref{}, 1))
	game.Effect.Push(cores.NewEffect(game, 905, cores.Ref{}, 1)) // 結束回合 1 → 本站推進退場

	seated := cores.NewGuest(game, 501) // Calm 3
	game.Seat.Place(1, seated)
	locked := cores.NewGuest(game, 501)
	locked.GetCalm().Lock()
	game.Seat.Place(2, locked)
	roam := cores.NewGuest(game, 501)
	game.Roam.Push(roam)
	game.GetEnergy().Set(5)

	this.Equal(cores.PhaseRoundStart, phaseRoundEnd(game))
	this.Equal(1, fired)
	this.Equal(int32(2), seated.GetCalm().GetValue()) // 3 - 1
	this.Equal(int32(3), locked.GetCalm().GetValue()) // 鎖定 → 不扣
	this.Equal(int32(2), roam.GetCalm().GetValue())   // 遊蕩列表照扣
	this.Equal(int32(0), game.GetEnergy().GetValue()) // 點數歸零
	this.Equal(1, ended)                              // 推進效果接線:到期退場
	this.Len(game.Effect, 1)                          // 901（整場保留）留佇列

	// 出牌點數保留鎖定 → 點數保留
	game.GetEnergy().Set(4)
	game.GetEnergyKeep().Lock()
	this.Equal(cores.PhaseRoundStart, phaseRoundEnd(game))
	this.Equal(int32(4), game.GetEnergy().GetValue())
}

// TestCalmDrop 驗證回合結束的耐心 -1 屬性事件（流程寫入白名單）:帶對象編號與前後值;鎖定不扣以 Before == After 表達。
func (this *SuitePhaseRoundEnd) TestCalmDrop() {
	game, record := newGameRecord()
	guest := cores.NewGuest(game, 501) // Calm 3

	calmDrop(game, guest)
	this.Equal(int32(2), guest.GetCalm().GetValue())
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventProperty, DataID: 501, InstanceID: guest.GetInstanceID(), Attr: "calm", Op: cores.AssignSub, Operand: 1, Before: 3, After: 2}, record.Event[0])

	guest.GetCalm().Lock()
	calmDrop(game, guest) // 鎖定不扣 → Before == After
	this.Require().Len(record.Event, 2)
	this.Equal(float64(2), record.Event[1].Before)
	this.Equal(float64(2), record.Event[1].After)
}
