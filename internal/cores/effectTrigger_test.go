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
		return func(eng *Engine) { fired = append(fired, id) }
	}

	runtime := NewRuntime(0)
	runtime.Effect = []*Effect{
		{InstanceID: 1, EffectID: 402, Stack: 1}, // 觸發、時機符、RunOrder 10
		{InstanceID: 2, EffectID: 403, Stack: 1}, // 觸發、時機符、RunOrder 20（最先）
		{InstanceID: 3, EffectID: 401, Stack: 1}, // 觸發但時機不符 → 不入列
		{InstanceID: 4, EffectID: 201, Stack: 1}, // 立即類型 → 不入列
		{InstanceID: 5, EffectID: 999, Stack: 1}, // 查無編譯資料 → 略過
	}
	effect := map[int32]effectData{
		402: {Kind: EffectTrigger, TriggerKind: TriggerCardPlay, Trigger: record(402)},
		403: {Kind: EffectTrigger, TriggerKind: TriggerCardPlay, Trigger: record(403)},
		401: {Kind: EffectTrigger, TriggerKind: TriggerCardDraw, Trigger: record(401)},
		201: {Kind: EffectImmed, TriggerKind: TriggerCardPlay, Trigger: record(201)},
	}
	eng := this.engine(runtime, effect)

	fireTrigger(eng, TriggerCardPlay)
	this.Equal([]int32{403, 402}, fired) // 作用順序 20 先於 10;時機 / 類型不符與查無資料皆未觸發
	this.Len(runtime.Effect, 5)          // 觸發後行為預設保留 → 佇列不變
}

func (this *SuiteEffectTrigger) TestFireOne() {
	count := 0
	endRun := 0
	guest := &Guest{InstanceID: 7}

	runtime := NewRuntime(0)
	target := &Effect{InstanceID: 1, EffectID: 402, Stack: 2, Self: Self{Guest: guest}}
	runtime.Effect = []*Effect{target}
	effect := map[int32]effectData{
		402: {
			Kind:         EffectTrigger,
			TriggerKind:  TriggerCardPlay,
			TriggerAfter: TriggerAfterRemove,
			Count:        this.expr("3"),
			Trigger: func(eng *Engine) {
				this.Equal(guest, eng.self.Guest) // self 綁定為該效果 self
				count++
			},
			End: func(eng *Engine) { endRun++ },
		},
	}
	eng := this.engine(runtime, effect)

	fireOne(eng, target)
	this.Equal(6, count)       // 堆疊層數 2 × 觸發次數 3
	this.Equal(2, endRun)      // 結束命令 重複 堆疊層數 2 次
	this.Empty(runtime.Effect) // 觸發後移除 → 出佇列
	this.Nil(eng.self)         // self 還原（原 nil）

	// 觸發條件不成立 → 觸發命令不執行、效果不移除
	blocked := 0
	runtimeCond := NewRuntime(0)
	effectCond := &Effect{InstanceID: 1, EffectID: 402, Stack: 1}
	runtimeCond.Effect = []*Effect{effectCond}
	engCond := this.engine(runtimeCond, map[int32]effectData{
		402: {Kind: EffectTrigger, TriggerKind: TriggerCardPlay, Cond: this.expr("morale > 5"), Trigger: func(eng *Engine) { blocked++ }},
	}) // morale 0 → 條件假
	fireOne(engCond, effectCond)
	this.Equal(0, blocked)
	this.Len(runtimeCond.Effect, 1)

	// 觸發次數 = 0 → 該效果中止:即使觸發後行為為移除,也不執行、不移除
	runtimeStop := NewRuntime(0)
	effectStop := &Effect{InstanceID: 1, EffectID: 402, Stack: 1}
	runtimeStop.Effect = []*Effect{effectStop}
	engStop := this.engine(runtimeStop, map[int32]effectData{
		402: {Kind: EffectTrigger, TriggerKind: TriggerCardPlay, TriggerAfter: TriggerAfterRemove, Count: this.expr("0"), Trigger: func(eng *Engine) { blocked++ }},
	})
	fireOne(engStop, effectStop)
	this.Equal(0, blocked)
	this.Len(runtimeStop.Effect, 1) // 中止先於移除 → 留佇列
}

func (this *SuiteEffectTrigger) TestCondPass() {
	eng := this.engine(NewRuntime(0), nil)

	this.True(condPass(eng, nil)) // 空欄 → 恆成立

	eng.runtime.Game.Morale.Value = 10
	this.True(condPass(eng, this.expr("morale > 5"))) // 真

	eng.runtime.Game.Morale.Value = 3
	this.False(condPass(eng, this.expr("morale > 5"))) // 假

	this.False(condPass(eng, this.expr("self.calm"))) // self 未綁 → 評估失敗 → 不成立
	this.False(condPass(eng, this.expr("'x'")))       // 非真值（text，Truthy ok=false）→ 不成立
}

func (this *SuiteEffectTrigger) TestTriggerCount() {
	eng := this.engine(NewRuntime(0), nil)

	n, ok := triggerCount(eng, nil) // 留空 → 1
	this.True(ok)
	this.Equal(int32(1), n)

	n, ok = triggerCount(eng, this.expr("2 + 1"))
	this.True(ok)
	this.Equal(int32(3), n)

	_, ok = triggerCount(eng, this.expr("0")) // = 0 → 中止
	this.False(ok)

	_, ok = triggerCount(eng, this.expr("3 - 5")) // < 0 → 中止
	this.False(ok)

	_, ok = triggerCount(eng, this.expr("self.sate")) // 評估失敗 → 中止
	this.False(ok)

	_, ok = triggerCount(eng, this.expr("'x'")) // 非數值 → 中止
	this.False(ok)
}

func (this *SuiteEffectTrigger) TestRunEffectCommand() {
	eng := this.engine(NewRuntime(0), nil)
	count := 0
	command := func(eng *Engine) { count++ }

	runEffectCommand(eng, command, 3)
	this.Equal(3, count)

	count = 0
	runEffectCommand(eng, nil, 5) // nil 命令 → 整體略過
	this.Equal(0, count)

	runEffectCommand(eng, command, 0) // times <= 0 → 不執行
	this.Equal(0, count)
}

// === 測試輔助（置尾） ===

func (this *SuiteEffectTrigger) engine(runtime *Runtime, effect map[int32]effectData) *Engine {
	eng := NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{}, nil)
	eng.effect = effect // 白箱覆寫:測試用自訂編譯效果(不經 prepareEffect)

	return eng
}

// expr 解析運算式字串為 *exprs.Expr;語法錯即測試失敗。供觸發條件 / 觸發次數構造。
func (this *SuiteEffectTrigger) expr(source string) *exprs.Expr {
	result, err := exprs.Parse(source)
	this.Require().NoError(err)

	return result
}
