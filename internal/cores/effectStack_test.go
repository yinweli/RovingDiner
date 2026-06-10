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
	game := NewGame(0, nil, nil, nil, nil)
	game.round = NewValue(5, 0)
	this.inject(game, map[int32]effectData{
		801: {Stack: 2, StackMax: 3, RunRound: 2}, // 增量 2、上限 3、作用回合 2
		802: {Stack: 0, StackMax: 0, RunRound: 0}, // 0 → 1、無上限、整場
	})

	// 新效果:不存在 → 建立、層數 = min(增量, 上限)、回層數
	this.Equal(int32(2), effectStack(game, NewRefGuest(guest), 801, 0)) // 增量 2 ≤ 上限 3
	this.Require().Len(game.Effect, 1)
	this.Equal(int32(2), game.Effect[0].GetStack())
	this.Equal(int32(6), game.Effect[0].GetExpire()) // 5 + 2 − 1

	// 既有同份 → 疊層、min(2+2, 3) = 3、回實際增加 1
	this.Equal(int32(1), effectStack(game, NewRefGuest(guest), 801, 0)) // 2 → 3（達上限,只 +1）
	this.Len(game.Effect, 1)                                            // 仍一筆
	this.Equal(int32(3), game.Effect[0].GetStack())

	// 已達上限再疊 → 回 0
	this.Equal(int32(0), effectStack(game, NewRefGuest(guest), 801, 0))

	// override N:802 無上限、0/1→1 被 override 取代
	this.Equal(int32(5), effectStack(game, NewRefGuest(guest), 802, 5)) // override 5
	this.Require().Len(game.Effect, 2)
	this.Equal(int32(5), game.Effect[1].GetStack())
	this.Equal(int32(0), game.Effect[1].GetExpire()) // RunRound 0 → 整場
}

func (this *SuiteEffectStack) TestEffectStackImmune() {
	guest := &Guest{instanceID: 11, effectImmune: Immune{count: map[int32]int32{5: 1}}} // 免疫群組 5
	game := NewGame(0, nil, nil, nil, nil)
	this.inject(game, map[int32]effectData{
		801: {Group: 5, Stack: 1},
	})

	this.Equal(int32(0), effectStack(game, NewRefGuest(guest), 801, 0)) // 顧客免疫 → 0、不入列
	this.Empty(game.Effect)

	// 卡牌 self 不受免疫閘門（免疫為顧客限定）
	this.Equal(int32(1), effectStack(game, NewRefCard(&Card{instanceID: 12}), 801, 0))
	this.Len(game.Effect, 1)

	this.Equal(int32(0), effectStack(game, NewRefGuest(guest), 999, 0)) // 查無編譯資料 → no-op
}

func (this *SuiteEffectStack) TestEffectStackTime() {
	guest := &Guest{instanceID: 11, effectImmune: Immune{count: map[int32]int32{}}}

	// 刷新:再疊時重算結束回合
	refresh := NewGame(0, nil, nil, nil, nil)
	refresh.round = NewValue(5, 0)
	this.inject(refresh, map[int32]effectData{
		801: {StackTime: StackTimeRefresh, RunRound: 3},
	})
	effectStack(refresh, NewRefGuest(guest), 801, 0) // 建立、Expire = 5 + 3 − 1 = 7
	this.Equal(int32(7), refresh.Effect[0].GetExpire())
	refresh.round = NewValue(8, 0)
	effectStack(refresh, NewRefGuest(guest), 801, 0) // 刷新、Expire = 8 + 3 − 1 = 10
	this.Equal(int32(10), refresh.Effect[0].GetExpire())

	// 不變:再疊時維持原結束回合
	stay := NewGame(0, nil, nil, nil, nil)
	stay.round = NewValue(5, 0)
	this.inject(stay, map[int32]effectData{
		801: {StackTime: StackTimeStay, RunRound: 3},
	})
	effectStack(stay, NewRefGuest(guest), 801, 0) // Expire = 7
	stay.round = NewValue(8, 0)
	effectStack(stay, NewRefGuest(guest), 801, 0) // 不變、Expire 維持 7
	this.Equal(int32(7), stay.Effect[0].GetExpire())
}

// === 測試輔助（置尾） ===

// inject 注入測試替身(injectPort)並以自訂編譯效果覆寫衍生索引(不經 prepareEffect)。
func (this *SuiteEffectStack) inject(game *Game, effect map[int32]effectData) {
	injectPort(game)
	game.effectData = effect
}
