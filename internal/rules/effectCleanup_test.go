package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteEffectCleanup(t *testing.T) {
	suite.Run(t, new(SuiteEffectCleanup))
}

// SuiteEffectCleanup 驗證清理效果流程（effectCleanup.go）:失效對象篩選 / 無目標排除 / 排序 / 退場 / 凍結整體 no-op。
type SuiteEffectCleanup struct {
	suite.Suite
}

func (this *SuiteEffectCleanup) TestCleanupEffect() {
	ended := []int32{}
	record := func(id int32) cores.EffectExec {
		return func(game *cores.Game) { ended = append(ended, id) }
	}

	data := tester.BuildData()
	game := newGameData(data)
	dead := cores.NewGuest(game, 501)
	alive := cores.NewGuest(game, 501)

	data.SetEffect(901, cores.EffectData{Kind: cores.EffectPersist, TargetKind: cores.TargetGuestRand, RunOrder: 10, End: record(901)})
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectPersist, TargetKind: cores.TargetGuestRand, RunOrder: 20, End: record(902)})
	data.SetEffect(903, cores.EffectData{Kind: cores.EffectTrigger, TargetKind: cores.TargetNone, End: record(903)})
	data.SetEffect(904, cores.EffectData{Kind: cores.EffectPersist, TargetKind: cores.TargetGuestRand, End: record(904)})

	game.Effect.Push(cores.NewEffect(game, 901, cores.NewRefGuest(dead), 2))
	game.Effect.Push(cores.NewEffect(game, 902, cores.NewRefGuest(dead), 1))
	game.Effect.Push(cores.NewEffect(game, 903, cores.NewRefGuest(dead), 1))  // 無目標效果 → 不在清理範圍（即使 self 同實例）
	game.Effect.Push(cores.NewEffect(game, 904, cores.NewRefGuest(alive), 1)) // self 他人 → 不清

	strayData := tester.BuildData() // 以另一份資料建構效果 999 → 本場查無編譯資料 → 略過
	strayData.SetEffect(999, cores.EffectData{Kind: cores.EffectPersist, TargetKind: cores.TargetGuestRand, End: record(999)})
	game.Effect.Push(cores.NewEffect(cores.NewGame(0, strayData, nil, nil), 999, cores.NewRefGuest(dead), 1))

	cleanupEffect(game, cores.NewRefGuest(dead))
	this.Equal([]int32{902, 901, 901}, ended) // 作用順序 20 先於 10;901 層數 2 → 結束命令 × 2
	this.Len(game.Effect, 3)                  // 903（無目標）/ 904（他人）/ 999（防禦）留佇列
	this.Nil(game.GetSelf())                  // self 還原（原 nil）

	// 凍結中顧客不視為失效對象 → 整體 no-op
	frozen := cores.NewGuest(game, 501)
	game.Cardify.Push(frozen)
	data.SetEffect(905, cores.EffectData{Kind: cores.EffectPersist, TargetKind: cores.TargetGuestRand, End: record(905)})
	game.Effect.Push(cores.NewEffect(game, 905, cores.NewRefGuest(frozen), 1))

	ended = nil
	cleanupEffect(game, cores.NewRefGuest(frozen))
	this.Empty(ended)
	this.Len(game.Effect, 4)
}
