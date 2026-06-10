package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteEffectAdvance(t *testing.T) {
	suite.Run(t, new(SuiteEffectAdvance))
}

// SuiteEffectAdvance 驗證推進效果流程（effectAdvance.go）:到期篩選 / 排序 / 退場（結束命令 × 層數 → 出佇列）/ 凍結跳過。
type SuiteEffectAdvance struct {
	suite.Suite
}

func (this *SuiteEffectAdvance) TestAdvanceEffect() {
	ended := []int32{}
	record := func(id int32) cores.EffectExec {
		return func(game *cores.Game) { ended = append(ended, id) }
	}

	data := tester.BuildData()
	game := newGameData(data)
	guest := cores.NewGuest(game, 501)
	frozen := cores.NewGuest(game, 501)
	game.Cardify.Push(frozen)

	data.SetEffect(901, cores.EffectData{Kind: cores.EffectPersist, RunRound: 2, RunOrder: 10, End: func(game *cores.Game) {
		this.Equal(guest, game.GetSelf().GetGuest()) // 退場結束命令以該效果 self 綁定執行
		ended = append(ended, 901)
	}})
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectPersist, RunRound: 2, RunOrder: 20, End: record(902)})
	data.SetEffect(903, cores.EffectData{Kind: cores.EffectPersist, RunRound: 9, End: record(903)})
	data.SetEffect(904, cores.EffectData{Kind: cores.EffectPersist, End: record(904)})
	data.SetEffect(905, cores.EffectData{Kind: cores.EffectPersist, RunRound: 1, End: record(905)})

	game.GetRound().Set(1)
	game.Effect.Push(cores.NewEffect(game, 901, cores.NewRefGuest(guest), 2))  // 結束回合 2
	game.Effect.Push(cores.NewEffect(game, 902, cores.Ref{}, 1))               // 結束回合 2
	game.Effect.Push(cores.NewEffect(game, 903, cores.Ref{}, 1))               // 結束回合 9 → 未到期
	game.Effect.Push(cores.NewEffect(game, 904, cores.Ref{}, 1))               // 作用回合 0 → 整場保留
	game.Effect.Push(cores.NewEffect(game, 905, cores.NewRefGuest(frozen), 1)) // 結束回合 1,凍結 → 不推進

	strayData := tester.BuildData() // 以另一份資料建構效果 999 → 本場查無編譯資料 → 略過
	strayData.SetEffect(999, cores.EffectData{Kind: cores.EffectPersist, RunRound: 1, End: record(999)})
	strayGame := cores.NewGame(0, 0, strayData, nil, nil)
	strayGame.GetRound().Set(1)
	game.Effect.Push(cores.NewEffect(strayGame, 999, cores.Ref{}, 1))

	game.GetRound().Set(2)
	advanceEffect(game)
	this.Equal([]int32{902, 901, 901}, ended) // 作用順序 20 先於 10;901 層數 2 → 結束命令 × 2
	this.Len(game.Effect, 4)                  // 903 / 904 / 905（凍結）/ 999（防禦）留佇列
	this.Nil(game.GetSelf())                  // self 還原（原 nil）
}
