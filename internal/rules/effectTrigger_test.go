package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteEffectTrigger(t *testing.T) {
	suite.Run(t, new(SuiteEffectTrigger))
}

// SuiteEffectTrigger 驗證觸發時機流程（effectTrigger.go）:篩選 / 排序 / 凍結跳過 / 觸發條件 / 觸發次數 / 觸發後行為 / self 綁定。
type SuiteEffectTrigger struct {
	suite.Suite
}

func (this *SuiteEffectTrigger) TestFireTrigger() {
	fired := []int32{}
	record := func(id int32) cores.EffectExec {
		return func(game *cores.Game) { fired = append(fired, id) }
	}

	data := tester.BuildData()
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardPlay, RunOrder: 10, Trigger: record(902)})
	data.SetEffect(903, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardPlay, RunOrder: 20, Trigger: record(903)})
	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardDraw, Trigger: record(901)})
	data.SetEffect(701, cores.EffectData{Kind: cores.EffectImmed, TriggerKind: cores.TriggerCardPlay, Trigger: record(701)})
	game := newGameData(data)

	for _, itor := range []int32{902, 903, 901, 701} {
		game.Effect.Push(cores.NewEffect(game, itor, cores.Ref{}, 1))
	} // for

	strayData := tester.BuildData() // 以另一份資料建構效果 999 → 本場查無編譯資料 → 略過
	strayData.SetEffect(999, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardPlay, Trigger: record(999)})
	game.Effect.Push(cores.NewEffect(cores.NewGame(0, 0, strayData, nil, nil, nil), 999, cores.Ref{}, 1))

	frozen := cores.NewGuest(game, 501) // 凍結中顧客的效果 → 效果凍結跳過(【二十一｜凍結語意】)
	game.Cardify.Push(frozen)
	data.SetEffect(904, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardPlay, Trigger: record(904)})
	game.Effect.Push(cores.NewEffect(game, 904, cores.NewRefGuest(frozen), 1))

	fireTrigger(game, cores.TriggerCardPlay)
	this.Equal([]int32{903, 902}, fired) // 作用順序 20 先於 10;時機 / 類型不符、凍結中未觸發
	this.Len(game.Effect, 6)             // 觸發後行為預設保留 → 佇列不變
}

func (this *SuiteEffectTrigger) TestFireOne() {
	count := 0
	endRun := 0
	guest := &cores.Guest{}

	data := tester.BuildData()
	data.SetEffect(902, cores.EffectData{
		Kind:         cores.EffectTrigger,
		TriggerKind:  cores.TriggerCardPlay,
		TriggerAfter: cores.TriggerAfterRemove,
		Count:        this.expr("3"),
		Trigger: func(game *cores.Game) {
			this.Equal(guest, game.GetSelf().GetGuest()) // self 綁定為該效果 self
			count++
		},
		End: func(game *cores.Game) { endRun++ },
	})
	game := newGameData(data)
	target := cores.NewEffect(game, 902, cores.NewRefGuest(guest), 2)
	game.Effect.Push(target)

	fireOne(game, target)
	this.Equal(6, count)     // 堆疊層數 2 × 觸發次數 3
	this.Equal(2, endRun)    // 結束命令 重複 堆疊層數 2 次
	this.Empty(game.Effect)  // 觸發後移除 → 出佇列
	this.Nil(game.GetSelf()) // self 還原（原 nil）

	// 觸發條件不成立 → 觸發命令不執行、效果不移除
	blocked := 0
	condData := tester.BuildData()
	condData.SetEffect(902, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardPlay, Cond: this.expr("morale > 5"), Trigger: func(game *cores.Game) { blocked++ }})
	gameCond := newGameData(condData) // morale 0 → 條件假
	effectCond := cores.NewEffect(gameCond, 902, cores.Ref{}, 1)
	gameCond.Effect.Push(effectCond)
	fireOne(gameCond, effectCond)
	this.Equal(0, blocked)
	this.Len(gameCond.Effect, 1)

	// 觸發次數 = 0 → 該效果中止:即使觸發後行為為移除,也不執行、不移除
	stopData := tester.BuildData()
	stopData.SetEffect(902, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardPlay, TriggerAfter: cores.TriggerAfterRemove, Count: this.expr("0"), Trigger: func(game *cores.Game) { blocked++ }})
	gameStop := newGameData(stopData)
	effectStop := cores.NewEffect(gameStop, 902, cores.Ref{}, 1)
	gameStop.Effect.Push(effectStop)
	fireOne(gameStop, effectStop)
	this.Equal(0, blocked)
	this.Len(gameStop.Effect, 1) // 中止先於移除 → 留佇列
}

func (this *SuiteEffectTrigger) TestCondPass() {
	game := newGame()

	this.True(condPass(game, nil)) // 空欄 → 恆成立

	game.GetMorale().Set(10)
	this.True(condPass(game, this.expr("morale > 5"))) // 真

	game.GetMorale().Set(3)
	this.False(condPass(game, this.expr("morale > 5"))) // 假

	this.False(condPass(game, this.expr("self.calm"))) // self 未綁 → 評估失敗 → 不成立
	this.False(condPass(game, this.expr("'x'")))       // 非真值（text，Truthy ok=false）→ 不成立
}

func (this *SuiteEffectTrigger) TestTriggerCount() {
	game := newGame()

	n, ok := triggerCount(game, nil) // 留空 → 1
	this.True(ok)
	this.Equal(int32(1), n)

	n, ok = triggerCount(game, this.expr("2 + 1"))
	this.True(ok)
	this.Equal(int32(3), n)

	_, ok = triggerCount(game, this.expr("0")) // = 0 → 中止
	this.False(ok)

	_, ok = triggerCount(game, this.expr("3 - 5")) // < 0 → 中止
	this.False(ok)

	_, ok = triggerCount(game, this.expr("self.sate")) // 評估失敗 → 中止
	this.False(ok)

	_, ok = triggerCount(game, this.expr("'x'")) // 非數值 → 中止
	this.False(ok)
}

// TestRunEffectExec 驗證 runEffectExec 連續執行命令 times 次;nil 命令整體略過、times <= 0 不執行。
func (this *SuiteEffectTrigger) TestRunEffectExec() {
	game := newGame()
	count := 0
	command := func(game *cores.Game) { count++ }

	runEffectExec(game, command, 3)
	this.Equal(3, count)

	count = 0
	runEffectExec(game, nil, 5) // nil 命令 → 整體略過
	this.Equal(0, count)

	runEffectExec(game, command, 0) // times <= 0 → 不執行
	this.Equal(0, count)
}

// === 測試輔助（置尾） ===

// expr 解析運算式字串為 *exprs.Expr;語法錯即測試失敗。供觸發條件 / 觸發次數構造。
func (this *SuiteEffectTrigger) expr(source string) *exprs.Expr {
	result, err := exprs.Parse(source)
	this.Require().NoError(err)
	return result
}
