package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteEffectStack(t *testing.T) {
	suite.Run(t, new(SuiteEffectStack))
}

// SuiteEffectStack 驗證堆疊處理（effectStack.go）:免疫閘門 / 新建 / 既有疊層 / 上限夾 / 堆疊時間 / 同份比對。
type SuiteEffectStack struct {
	suite.Suite
}

func (this *SuiteEffectStack) TestEffectStack() {
	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{Stack: 2, StackMax: 3, RunRound: 2}) // 增量 2、上限 3、作用回合 2
	data.SetEffect(802, cores.EffectData{Stack: 0, StackMax: 0, RunRound: 0}) // 0 → 1、無上限、整場
	game := newGameData(data)
	game.GetRound().Set(5)
	guest := &cores.Guest{}

	// 新效果:不存在 → 建立、層數 = min(增量, 上限)、回層數
	this.Equal(int32(2), effectStack(game, cores.NewRefGuest(guest), 801, 0)) // 增量 2 ≤ 上限 3
	this.Require().Len(game.Effect, 1)
	this.Equal(int32(2), game.Effect[0].GetStack())
	this.Equal(int32(6), game.Effect[0].GetExpire()) // 5 + 2 − 1

	// 既有同份 → 疊層、min(2+2, 3) = 3、回實際增加 1
	this.Equal(int32(1), effectStack(game, cores.NewRefGuest(guest), 801, 0)) // 2 → 3（達上限,只 +1）
	this.Len(game.Effect, 1)                                                  // 仍一筆
	this.Equal(int32(3), game.Effect[0].GetStack())

	// 已達上限再疊 → 回 0
	this.Equal(int32(0), effectStack(game, cores.NewRefGuest(guest), 801, 0))

	// override N:802 無上限、0/1→1 被 override 取代
	this.Equal(int32(5), effectStack(game, cores.NewRefGuest(guest), 802, 5)) // override 5
	this.Require().Len(game.Effect, 2)
	this.Equal(int32(5), game.Effect[1].GetStack())
	this.Equal(int32(0), game.Effect[1].GetExpire()) // RunRound 0 → 整場
}

func (this *SuiteEffectStack) TestEffectStackImmune() {
	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{Group: 5, Stack: 1})
	game := newGameData(data)
	guest := &cores.Guest{}
	guest.GetEffectImmune().Add(5) // 免疫群組 5

	this.Equal(int32(0), effectStack(game, cores.NewRefGuest(guest), 801, 0)) // 顧客免疫 → 0、不入列
	this.Empty(game.Effect)

	// 卡牌 self 不受免疫閘門（免疫為顧客限定）
	this.Equal(int32(1), effectStack(game, cores.NewRefCard(&cores.Card{}), 801, 0))
	this.Len(game.Effect, 1)

	this.Equal(int32(0), effectStack(game, cores.NewRefGuest(guest), 999, 0)) // 查無編譯資料 → no-op
}

func (this *SuiteEffectStack) TestEffectStackTime() {
	guest := &cores.Guest{}

	// 刷新:再疊時重算結束回合
	refreshData := tester.BuildData()
	refreshData.SetEffect(801, cores.EffectData{StackTime: cores.StackTimeRefresh, RunRound: 3})
	refresh := newGameData(refreshData)
	refresh.GetRound().Set(5)
	effectStack(refresh, cores.NewRefGuest(guest), 801, 0) // 建立、Expire = 5 + 3 − 1 = 7
	this.Equal(int32(7), refresh.Effect[0].GetExpire())
	refresh.GetRound().Set(8)
	effectStack(refresh, cores.NewRefGuest(guest), 801, 0) // 刷新、Expire = 8 + 3 − 1 = 10
	this.Equal(int32(10), refresh.Effect[0].GetExpire())

	// 不變:再疊時維持原結束回合
	stayData := tester.BuildData()
	stayData.SetEffect(801, cores.EffectData{StackTime: cores.StackTimeStay, RunRound: 3})
	stay := newGameData(stayData)
	stay.GetRound().Set(5)
	effectStack(stay, cores.NewRefGuest(guest), 801, 0) // Expire = 7
	stay.GetRound().Set(8)
	effectStack(stay, cores.NewRefGuest(guest), 801, 0) // 不變、Expire 維持 7
	this.Equal(int32(7), stay.Effect[0].GetExpire())
}
