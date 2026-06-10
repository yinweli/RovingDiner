package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteEffectStack(t *testing.T) {
	suite.Run(t, new(SuiteEffectStack))
}

// SuiteEffectStack 驗證堆疊處理（effectStack.go）:免疫閘門 / 新建 / 既有疊層 / 上限夾 / 堆疊時間 / 同份比對。
type SuiteEffectStack struct {
	suite.Suite
}

func (this *SuiteEffectStack) TestEffectStack() {
	guest := &Guest{instanceID: 11, effectImmune: Immune{count: map[int32]int32{}}}
	runtime := NewRuntime(0)
	runtime.Game.round = NewValue(5, 0)
	eng := this.engine(runtime, map[int32]effectData{
		801: {Stack: 2, StackMax: 3, RunRound: 2}, // 增量 2、上限 3、作用回合 2
		802: {Stack: 0, StackMax: 0, RunRound: 0}, // 0 → 1、無上限、整場
	})

	// 新效果:不存在 → 建立、層數 = min(增量, 上限)、回層數
	this.Equal(int32(2), effectStack(eng, NewRefGuest(guest), 801, 0)) // 增量 2 ≤ 上限 3
	this.Require().Len(runtime.Effect, 1)
	this.Equal(int32(2), runtime.Effect[0].GetStack())
	this.Equal(int32(6), runtime.Effect[0].GetExpire()) // 5 + 2 − 1

	// 既有同份 → 疊層、min(2+2, 3) = 3、回實際增加 1
	this.Equal(int32(1), effectStack(eng, NewRefGuest(guest), 801, 0)) // 2 → 3（達上限,只 +1）
	this.Len(runtime.Effect, 1)                                        // 仍一筆
	this.Equal(int32(3), runtime.Effect[0].GetStack())

	// 已達上限再疊 → 回 0
	this.Equal(int32(0), effectStack(eng, NewRefGuest(guest), 801, 0))

	// override N:802 無上限、0/1→1 被 override 取代
	this.Equal(int32(5), effectStack(eng, NewRefGuest(guest), 802, 5)) // override 5
	this.Require().Len(runtime.Effect, 2)
	this.Equal(int32(5), runtime.Effect[1].GetStack())
	this.Equal(int32(0), runtime.Effect[1].GetExpire()) // RunRound 0 → 整場
}

func (this *SuiteEffectStack) TestEffectStackImmune() {
	guest := &Guest{instanceID: 11, effectImmune: Immune{count: map[int32]int32{5: 1}}} // 免疫群組 5
	runtime := NewRuntime(0)
	eng := this.engine(runtime, map[int32]effectData{
		801: {Group: 5, Stack: 1},
	})

	this.Equal(int32(0), effectStack(eng, NewRefGuest(guest), 801, 0)) // 顧客免疫 → 0、不入列
	this.Empty(runtime.Effect)

	// 卡牌 self 不受免疫閘門（免疫為顧客限定）
	this.Equal(int32(1), effectStack(eng, NewRefCard(&Card{instanceID: 12}), 801, 0))
	this.Len(runtime.Effect, 1)

	this.Equal(int32(0), effectStack(eng, NewRefGuest(guest), 999, 0)) // 查無編譯資料 → no-op
}

func (this *SuiteEffectStack) TestEffectStackTime() {
	guest := &Guest{instanceID: 11, effectImmune: Immune{count: map[int32]int32{}}}

	// 刷新:再疊時重算結束回合
	refresh := NewRuntime(0)
	refresh.Game.round = NewValue(5, 0)
	engRefresh := this.engine(refresh, map[int32]effectData{
		801: {StackTime: StackTimeRefresh, RunRound: 3},
	})
	effectStack(engRefresh, NewRefGuest(guest), 801, 0) // 建立、Expire = 5 + 3 − 1 = 7
	this.Equal(int32(7), refresh.Effect[0].GetExpire())
	refresh.Game.round = NewValue(8, 0)
	effectStack(engRefresh, NewRefGuest(guest), 801, 0) // 刷新、Expire = 8 + 3 − 1 = 10
	this.Equal(int32(10), refresh.Effect[0].GetExpire())

	// 不變:再疊時維持原結束回合
	stay := NewRuntime(0)
	stay.Game.round = NewValue(5, 0)
	engStay := this.engine(stay, map[int32]effectData{
		801: {StackTime: StackTimeStay, RunRound: 3},
	})
	effectStack(engStay, NewRefGuest(guest), 801, 0) // Expire = 7
	stay.Game.round = NewValue(8, 0)
	effectStack(engStay, NewRefGuest(guest), 801, 0) // 不變、Expire 維持 7
	this.Equal(int32(7), stay.Effect[0].GetExpire())
}

// === 測試輔助（置尾） ===

func (this *SuiteEffectStack) engine(runtime *Runtime, effect map[int32]effectData) *Engine {
	eng := NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{}, nil)
	eng.effect = effect // 白箱覆寫:測試用自訂編譯效果

	return eng
}
