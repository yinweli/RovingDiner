package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteEffectTrigger(t *testing.T) {
	suite.Run(t, new(SuiteEffectTrigger))
}

// SuiteEffectTrigger 驗證觸發時機流程（effectTrigger.go）:篩選 / 排序 / 觸發條件 / 觸發次數 / 觸發後行為 / self 綁定。
type SuiteEffectTrigger struct {
	suite.Suite
}

func (this *SuiteEffectTrigger) TestFireTrigger() {
	fired := []int32{}
	record := func(id int32) EffectCommand {
		return func(game *Game) { fired = append(fired, id) }
	}

	game := NewGame(0, nil, nil, nil, nil)
	game.Effect = EffectList{
		{instanceID: 1, effectID: 402, stack: 1}, // 觸發、時機符、RunOrder 10
		{instanceID: 2, effectID: 403, stack: 1}, // 觸發、時機符、RunOrder 20（最先）
		{instanceID: 3, effectID: 401, stack: 1}, // 觸發但時機不符 → 不入列
		{instanceID: 4, effectID: 201, stack: 1}, // 立即類型 → 不入列
		{instanceID: 5, effectID: 999, stack: 1}, // 查無編譯資料 → 略過
	}
	effect := map[int32]effectData{
		402: {Kind: EffectTrigger, TriggerKind: TriggerCardPlay, RunOrder: 10, Trigger: record(402)},
		403: {Kind: EffectTrigger, TriggerKind: TriggerCardPlay, RunOrder: 20, Trigger: record(403)},
		401: {Kind: EffectTrigger, TriggerKind: TriggerCardDraw, Trigger: record(401)},
		201: {Kind: EffectImmed, TriggerKind: TriggerCardPlay, Trigger: record(201)},
	}
	injectPort(game)
	game.effectData = effect // 白箱覆寫:測試用自訂編譯效果

	fireTrigger(game, TriggerCardPlay)
	this.Equal([]int32{403, 402}, fired) // 作用順序 20 先於 10;時機 / 類型不符與查無資料皆未觸發
	this.Len(game.Effect, 5)             // 觸發後行為預設保留 → 佇列不變
}

func (this *SuiteEffectTrigger) TestFireOne() {
	count := 0
	endRun := 0
	guest := &Guest{instanceID: 7}

	game := NewGame(0, nil, nil, nil, nil)
	target := &Effect{instanceID: 1, effectID: 402, stack: 2, self: NewRefGuest(guest)}
	game.Effect = EffectList{target}
	effect := map[int32]effectData{
		402: {
			Kind:         EffectTrigger,
			TriggerKind:  TriggerCardPlay,
			TriggerAfter: TriggerAfterRemove,
			Count:        this.expr("3"),
			Trigger: func(game *Game) {
				this.Equal(guest, game.self.GetGuest()) // self 綁定為該效果 self
				count++
			},
			End: func(game *Game) { endRun++ },
		},
	}
	this.inject(game, effect)

	fireOne(game, target)
	this.Equal(6, count)    // 堆疊層數 2 × 觸發次數 3
	this.Equal(2, endRun)   // 結束命令 重複 堆疊層數 2 次
	this.Empty(game.Effect) // 觸發後移除 → 出佇列
	this.Nil(game.self)     // self 還原（原 nil）

	// 觸發條件不成立 → 觸發命令不執行、效果不移除
	blocked := 0
	gameCond := NewGame(0, nil, nil, nil, nil)
	effectCond := &Effect{instanceID: 1, effectID: 402, stack: 1}
	gameCond.Effect = EffectList{effectCond}
	this.inject(gameCond, map[int32]effectData{
		402: {Kind: EffectTrigger, TriggerKind: TriggerCardPlay, Cond: this.expr("morale > 5"), Trigger: func(game *Game) { blocked++ }},
	}) // morale 0 → 條件假
	fireOne(gameCond, effectCond)
	this.Equal(0, blocked)
	this.Len(gameCond.Effect, 1)

	// 觸發次數 = 0 → 該效果中止:即使觸發後行為為移除,也不執行、不移除
	gameStop := NewGame(0, nil, nil, nil, nil)
	effectStop := &Effect{instanceID: 1, effectID: 402, stack: 1}
	gameStop.Effect = EffectList{effectStop}
	this.inject(gameStop, map[int32]effectData{
		402: {Kind: EffectTrigger, TriggerKind: TriggerCardPlay, TriggerAfter: TriggerAfterRemove, Count: this.expr("0"), Trigger: func(game *Game) { blocked++ }},
	})
	fireOne(gameStop, effectStop)
	this.Equal(0, blocked)
	this.Len(gameStop.Effect, 1) // 中止先於移除 → 留佇列
}

func (this *SuiteEffectTrigger) TestCondPass() {
	game := this.inject(NewGame(0, nil, nil, nil, nil), nil)

	this.True(condPass(game, nil)) // 空欄 → 恆成立

	game.morale = NewValue(10, 0)
	this.True(condPass(game, this.expr("morale > 5"))) // 真

	game.morale = NewValue(3, 0)
	this.False(condPass(game, this.expr("morale > 5"))) // 假

	this.False(condPass(game, this.expr("self.calm"))) // self 未綁 → 評估失敗 → 不成立
	this.False(condPass(game, this.expr("'x'")))       // 非真值（text，Truthy ok=false）→ 不成立
}

func (this *SuiteEffectTrigger) TestTriggerCount() {
	game := this.inject(NewGame(0, nil, nil, nil, nil), nil)

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

func (this *SuiteEffectTrigger) TestRunEffectCommand() {
	game := this.inject(NewGame(0, nil, nil, nil, nil), nil)
	count := 0
	command := func(game *Game) { count++ }

	runEffectCommand(game, command, 3)
	this.Equal(3, count)

	count = 0
	runEffectCommand(game, nil, 5) // nil 命令 → 整體略過
	this.Equal(0, count)

	runEffectCommand(game, command, 0) // times <= 0 → 不執行
	this.Equal(0, count)
}

// === 測試輔助（置尾） ===

// inject 注入測試替身(injectPort)並以自訂編譯效果覆寫衍生索引(不經 prepareEffect)。
func (this *SuiteEffectTrigger) inject(game *Game, effect map[int32]effectData) *Game {
	injectPort(game)
	game.effectData = effect
	return game
}

// expr 解析運算式字串為 *exprs.Expr;語法錯即測試失敗。供觸發條件 / 觸發次數構造。
func (this *SuiteEffectTrigger) expr(source string) *exprs.Expr {
	result, err := exprs.Parse(source)
	this.Require().NoError(err)

	return result
}
