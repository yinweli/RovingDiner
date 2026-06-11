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

// SuiteEffectStack 驗證堆疊處理(effectStack.go): 免疫閘門 / 新建 / 既有疊層 / 上限夾 / 堆疊時間 / 同份比對。
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

	// 新效果: 不存在 → 建立、層數 = min(增量, 上限)、回佇列實例與層數
	effect, added := effectStack(game, cores.NewRefGuest(guest), 801, 0) // 增量 2 ≤ 上限 3
	this.Equal(int32(2), added)
	this.Require().Len(game.Effect, 1)
	this.Same(game.Effect[0], effect) // 回佇列實例(常駐 啟動命令憑此取實例編號)
	this.Equal(int32(2), game.Effect[0].GetStack())
	this.Equal(int32(6), game.Effect[0].GetExpire()) // 5 + 2 − 1

	// 既有同份 → 疊層、min(2+2, 3) = 3、回實際增加 1
	this.Equal(int32(1), stackAdded(game, cores.NewRefGuest(guest), 801, 0)) // 2 → 3(達上限, 只 +1)
	this.Len(game.Effect, 1)                                                 // 仍一筆
	this.Equal(int32(3), game.Effect[0].GetStack())

	// 已達上限再疊 → 回 0
	this.Equal(int32(0), stackAdded(game, cores.NewRefGuest(guest), 801, 0))

	// override N: 802 無上限、0/1→1 被 override 取代
	this.Equal(int32(5), stackAdded(game, cores.NewRefGuest(guest), 802, 5)) // override 5
	this.Require().Len(game.Effect, 2)
	this.Equal(int32(5), game.Effect[1].GetStack())
	this.Equal(int32(0), game.Effect[1].GetExpire()) // RunRound 0 → 整場
}

// TestEffectStackEmit 驗證堆疊處理的效果事件: 新建入列 / 既有實際增層 / 刷新改變結束回合發 加入(載層數 / 結束回合快照);
// 堆疊滿且結束回合不變 → 不發。
func (this *SuiteEffectStack) TestEffectStackEmit() {
	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{Stack: 2, StackMax: 3})
	data.SetEffect(803, cores.EffectData{Stack: 3, StackMax: 3, StackTime: cores.StackTimeRefresh, RunRound: 2})
	game, record := newGameDataRecord(data)
	guest := &cores.Guest{}

	effect, _ := effectStack(game, cores.NewRefGuest(guest), 801, 0) // 新建入列 → 加入(快照: 層數 2、整場保留)
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventEffect, EffectID: 801, EffectInstanceID: effect.GetInstanceID(), Stage: cores.EffectStageJoin, Stack: 2, Alive: true}, record.Event[0])

	effectStack(game, cores.NewRefGuest(guest), 801, 0) // 疊層(實際 +1) → 加入(快照: 層數 3)
	this.Require().Len(record.Event, 2)
	this.Equal(int32(3), record.Event[1].Stack)

	effectStack(game, cores.NewRefGuest(guest), 801, 0) // 堆疊滿(增層 0)且結束回合不變 → 不發
	this.Len(record.Event, 2)

	effectStack(game, cores.NewRefGuest(guest), 803, 0) // 新建入列(滿層 3、結束回合 = 0 + 2 - 1)
	this.Require().Len(record.Event, 3)
	this.Equal(int32(1), record.Event[2].Expire)

	game.GetRound().Set(5)
	effectStack(game, cores.NewRefGuest(guest), 803, 0) // 堆疊滿但刷新改變結束回合 → 照發(快照載刷後值)
	this.Require().Len(record.Event, 4)
	this.Equal(cores.EventData{Kind: cores.EventEffect, Round: 5, EffectID: 803, EffectInstanceID: record.Event[2].EffectInstanceID, Stage: cores.EffectStageJoin, Stack: 3, Expire: 6, Alive: true}, record.Event[3]) // Round 為 Emit 座標蓋章
}

func (this *SuiteEffectStack) TestEffectStackImmune() {
	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{Group: 5, Stack: 1})
	game := newGameData(data)
	guest := &cores.Guest{}
	guest.GetEffectImmune().Add(5) // 免疫群組 5

	effect, added := effectStack(game, cores.NewRefGuest(guest), 801, 0) // 顧客免疫 → nil、0、不入列
	this.Nil(effect)
	this.Equal(int32(0), added)
	this.Empty(game.Effect)

	// 卡牌 self 不受免疫閘門(免疫為顧客限定)
	this.Equal(int32(1), stackAdded(game, cores.NewRefCard(&cores.Card{}), 801, 0))
	this.Len(game.Effect, 1)

	this.Equal(int32(0), stackAdded(game, cores.NewRefGuest(guest), 999, 0)) // 查無編譯資料 → no-op
}

func (this *SuiteEffectStack) TestEffectStackTime() {
	guest := &cores.Guest{}

	// 刷新: 再疊時重算結束回合
	refreshData := tester.BuildData()
	refreshData.SetEffect(801, cores.EffectData{StackTime: cores.StackTimeRefresh, RunRound: 3})
	refresh := newGameData(refreshData)
	refresh.GetRound().Set(5)
	effectStack(refresh, cores.NewRefGuest(guest), 801, 0) // 建立、Expire = 5 + 3 − 1 = 7
	this.Equal(int32(7), refresh.Effect[0].GetExpire())
	refresh.GetRound().Set(8)
	effectStack(refresh, cores.NewRefGuest(guest), 801, 0) // 刷新、Expire = 8 + 3 − 1 = 10
	this.Equal(int32(10), refresh.Effect[0].GetExpire())

	// 不變: 再疊時維持原結束回合
	stayData := tester.BuildData()
	stayData.SetEffect(801, cores.EffectData{StackTime: cores.StackTimeStay, RunRound: 3})
	stay := newGameData(stayData)
	stay.GetRound().Set(5)
	effectStack(stay, cores.NewRefGuest(guest), 801, 0) // Expire = 7
	stay.GetRound().Set(8)
	effectStack(stay, cores.NewRefGuest(guest), 801, 0) // 不變、Expire 維持 7
	this.Equal(int32(7), stay.Effect[0].GetExpire())
}

// === 測試輔助(置尾) ===

// stackAdded 包裝 effectStack 只取實際增加層數(M18 簽章改回傳實例 + 層數, 多數斷言只關心層數)。
func stackAdded(game *cores.Game, self cores.Ref, effectID, override int32) int32 {
	_, added := effectStack(game, self, effectID, override)
	return added
}
