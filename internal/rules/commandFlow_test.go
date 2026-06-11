package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteCommandFlow(t *testing.T) {
	suite.Run(t, new(SuiteCommandFlow))
}

// SuiteCommandFlow 驗證處理流程命令(commandFlow.go): cardRun / *Morph / cardify / restore / guestExit / guestReturn / guestRoam / guestSeat。
type SuiteCommandFlow struct {
	suite.Suite
}

func (this *SuiteCommandFlow) TestCardRun() {
	game := newGame()
	game.GetEnergy().Set(5)
	card := cores.NewCard(game, 101) // 群組 1
	card.GetCost().Set(2)
	game.Hand = cores.CardList{card}

	commandCardRun(game, []cores.InstanceID{card.GetInstanceID()}, this.flag(true, true)) // 消耗點數、進棄牌堆
	this.Equal(int32(3), game.GetEnergy().GetValue())                                     // 5 - 2
	this.Equal(card, game.GetPlayLast())
	this.Equal(int32(1), game.GetPlayCount())
	this.Equal(int32(1), game.GetPlayTotal().Get(1)) // 卡 101 群組 1
	this.Empty(game.Hand)
	this.Equal(cores.CardList{card}, game.Drop)
}

func (this *SuiteCommandFlow) TestCardRunNoop() {
	game := newGame()
	game.GetEnergy().Set(5)

	sealed := cores.NewCard(game, 101)
	sealed.GetSeal().Lock()
	game.Hand = cores.CardList{sealed}
	commandCardRun(game, []cores.InstanceID{sealed.GetInstanceID()}, this.flag(true, false)) // 封印 → no-op
	this.Len(game.Hand, 1)
	this.Equal(int32(5), game.GetEnergy().GetValue())

	poor := cores.NewCard(game, 101)
	poor.GetCost().Set(9)
	game.Hand = cores.CardList{poor}
	commandCardRun(game, []cores.InstanceID{poor.GetInstanceID()}, this.flag(true, false)) // 點數不足 → no-op
	this.Equal(int32(5), game.GetEnergy().GetValue())

	noEnergy := cores.NewCard(game, 101)
	game.Hand = cores.CardList{noEnergy}
	commandCardRun(game, []cores.InstanceID{noEnergy.GetInstanceID()}, this.flag(false, false)) // 不耗點數、不進棄牌堆 → 留手牌
	this.Equal(noEnergy, game.GetPlayLast())
	this.Len(game.Hand, 1)

	inDrop := cores.NewCard(game, 101)
	game.Drop = cores.CardList{inDrop}
	commandCardRun(game, []cores.InstanceID{inDrop.GetInstanceID()}, this.flag(false, true)) // 已在棄牌堆 → 不重複移動
	this.Len(game.Drop, 1)

	exiled := cores.NewCard(game, 101)
	game.Exile = cores.CardList{exiled}
	commandCardRun(game, []cores.InstanceID{exiled.GetInstanceID()}, nil) // 流放牌堆(位置不符)→ no-op
	this.Len(game.Exile, 1)

	commandCardRun(game, []cores.InstanceID{99}, nil) // 不存在 → no-op
}

func (this *SuiteCommandFlow) TestCardRunEnergyLock() {
	game := newGame()
	game.GetEnergy().Set(5)
	game.GetEnergy().Lock() // 點數鎖定
	card := cores.NewCard(game, 101)
	card.GetCost().Set(2)
	game.Hand = cores.CardList{card}

	commandCardRun(game, []cores.InstanceID{card.GetInstanceID()}, this.flag(true, true)) // 鎖定 → 照出牌、耗能不扣
	this.Equal(int32(5), game.GetEnergy().GetValue())
	this.Equal(cores.CardList{card}, game.Drop)
}

func (this *SuiteCommandFlow) TestCardRunEffect() {
	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{Kind: cores.EffectTrigger, TargetKind: cores.TargetNone, Stack: 1})
	game := newGameData(data)
	card := cores.NewCard(game, 101)
	card.GetEffectID().Add(801)
	game.Hand = cores.CardList{card}

	commandCardRun(game, []cores.InstanceID{card.GetInstanceID()}, this.flag(false, false)) // 不耗點、不進棄牌
	this.Require().Len(game.Effect, 1)                                                      // 實例效果列表啟動 → 效果入佇列
	this.Equal(int32(801), game.Effect[0].GetEffectID())
}

func (this *SuiteCommandFlow) TestCardRunEffectMove() {
	data := tester.BuildData()
	game := newGameData(data)
	card := cores.NewCard(game, 101)
	card.GetEffectID().Add(802)
	game.Hand = cores.CardList{card}
	data.SetEffect(802, cores.EffectData{Kind: cores.EffectImmed, TargetKind: cores.TargetNone, Immed: func(e *cores.Game) {
		removeCard(e, cores.ContainerHand, card)
		placeCard(e, cores.ContainerHand, cores.ContainerExile, card) // 立即命令在出牌途中把本卡移至流放牌堆
	}})

	commandCardRun(game, []cores.InstanceID{card.GetInstanceID()}, this.flag(false, true)) // 不耗點、進棄牌堆
	this.Empty(game.Exile)                                                                 // 進棄牌堆前重新定位 → 自當前容器(流放)移出、不殘留
	this.Equal(cores.CardList{card}, game.Drop)                                            // 僅進棄牌堆一次、不重複
}

func (this *SuiteCommandFlow) TestMorph() {
	ended := 0
	data := tester.BuildData()
	game := newGameData(data)
	card := cores.NewCard(game, 102)
	card.GetCost().Set(9)
	id := card.GetInstanceID()
	game.Hand = cores.CardList{card}
	other := cores.NewCard(game, 102)
	game.Deck = cores.CardList{other}

	data.SetEffect(801, cores.EffectData{Kind: cores.EffectPersist, TargetKind: cores.TargetCardRand, End: func(game *cores.Game) { ended++ }})
	game.Effect.Push(cores.NewEffect(game, 801, cores.NewRefCard(card), 2))  // 變身卡的殘留效果 → 步驟 5 清理
	game.Effect.Push(cores.NewEffect(game, 801, cores.NewRefCard(other), 1)) // 他卡效果 → 留佇列

	commandHandMorph(game, []cores.InstanceID{id}, nums(7)) // 群組 7 → 抽中卡 101
	this.Equal(int32(101), card.GetCardID())                // cardID 換成抽中值
	this.NotEqual(id, card.GetInstanceID())                 // 重分配實例編號
	this.Equal(int32(0), card.GetCost().GetValue())         // 載卡 101 資料(無 Cost → 0)
	this.Equal(int32(102), game.GetMorphOldID())
	this.Equal(int32(101), game.GetMorphNewID())
	this.Equal(card, game.GetMorphLast())
	this.Equal(int32(1), game.GetMorphCount())
	this.Len(game.Hand, 1) // 留原牌堆

	this.Equal(2, ended) // 5. 清理變身前綁定本卡的效果: 結束命令 × 層數 2
	this.Require().Len(game.Effect, 1)
	this.Equal(other, game.Effect[0].GetSelf().GetCard()) // 他卡效果不受影響
}

func (this *SuiteCommandFlow) TestMorphNoop() {
	game := newGame()
	card := cores.NewCard(game, 102)
	id := []cores.InstanceID{card.GetInstanceID()}
	game.Hand = cores.CardList{card}

	commandDeckMorph(game, id, nums(7)) // 位置不符(卡在手牌)→ no-op
	this.Equal(int32(102), card.GetCardID())

	commandHandMorph(game, id, nums(8)) // 群組總權重 0 → no-op
	commandHandMorph(game, id, nums(9)) // 抽中 999 無資料 → 跳過
	commandHandMorph(game, id, nil)     // 缺參數 → no-op
	this.Equal(int32(102), card.GetCardID())

	drop := cores.NewCard(game, 102)
	game.Drop = cores.CardList{drop}
	commandDropMorph(game, []cores.InstanceID{drop.GetInstanceID()}, nums(7))
	this.Equal(int32(101), game.Drop[0].GetCardID())

	exiled := cores.NewCard(game, 102)
	game.Exile = cores.CardList{exiled}
	commandExileMorph(game, []cores.InstanceID{exiled.GetInstanceID()}, nums(7))
	this.Equal(int32(101), game.Exile[0].GetCardID())
}

// TestMorphEmit 驗證變身的實例事件: 銷毀舊(舊編號)+ 建立新(新編號)兩發, 位置不變無容器事件(EventInstance 唯一真身; M18 拍板)。
func (this *SuiteCommandFlow) TestMorphEmit() {
	game, record := newGameRecord()
	card := cores.NewCard(game, 102)
	oldInstance := card.GetInstanceID()
	game.Hand = cores.CardList{card}

	morph(game, []cores.InstanceID{oldInstance}, nums(7), cores.ContainerHand) // 群組 7 → 抽中卡 101
	this.Require().Len(record.Event, 2)
	this.Equal(cores.EventData{Kind: cores.EventInstance, DataID: 102, InstanceID: oldInstance, Alive: false}, record.Event[0])
	this.Equal(cores.EventData{Kind: cores.EventInstance, DataID: 101, InstanceID: card.GetInstanceID(), Alive: true}, record.Event[1])
}

func (this *SuiteCommandFlow) TestCardify() {
	game := newGame()
	game.GetRound().Set(7)
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(2, guest)

	commandCardify(game, []cores.InstanceID{guest.GetInstanceID()}, nums(101)) // 卡牌化為卡 101
	this.Nil(game.Seat[2])                                                     // 1. 自座位移除
	this.Require().Len(game.Cardify, 1)                                        // 2. 入卡牌化列表
	this.Equal(int32(0), guest.GetSeatID())
	this.Equal(int32(7), guest.GetFreeze()) // 3. 凍結
	this.Require().Len(game.Hand, 1)
	card := game.Hand[0]
	this.Equal(int32(101), card.GetCardID())
	this.Equal(guest, card.GetCardify())           // 5. 綁來源
	this.Equal(int32(1), card.GetKeep().GetLock()) // 6. 不棄 + 1

	roam := cores.NewGuest(game, 501)
	game.Roam.Push(roam)
	commandCardify(game, []cores.InstanceID{roam.GetInstanceID()}, nums(101)) // 不在座位列表 → no-op
	this.Len(game.Cardify, 1)

	seated := cores.NewGuest(game, 501)
	game.Seat.Place(3, seated)
	commandCardify(game, []cores.InstanceID{seated.GetInstanceID()}, nums(999)) // 卡牌資料不存在 → no-op
	commandCardify(game, []cores.InstanceID{seated.GetInstanceID()}, nil)       // 缺卡牌編號 → no-op
	this.NotNil(game.Seat[3])
}

func (this *SuiteCommandFlow) TestRestore() {
	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{}) // 自訂效果(整場保留), 測試以 SetExpire 佈置結束回合
	game := newGameData(data)
	game.GetRound().Set(10)
	guest := cores.NewGuest(game, 501)
	guest.SetFreeze(4)
	card := cores.NewCard(game, 101)
	card.CardifyBind(guest)
	game.Hand = cores.CardList{card}
	game.Cardify.Push(guest)
	effect := cores.NewEffect(game, 801, cores.NewRefGuest(guest), 1)
	effect.SetExpire(5)
	game.Effect.Push(effect)

	commandRestore(game, []cores.InstanceID{card.GetInstanceID()}, nil)
	this.Empty(game.Cardify)        // 2. 移出卡牌化列表
	this.Equal(guest, game.Seat[1]) // 3. 入隨機空座位
	this.Equal(int32(1), guest.GetSeatID())
	this.Equal(int32(11), game.Effect[0].GetExpire()) // 4. 5 + (10 - 4)
	this.Equal(int32(0), guest.GetFreeze())           // 5. 解凍
	this.Nil(card.GetCardify())                       // 6. 解綁
	this.Equal(int32(0), card.GetKeep().GetLock())    // 7. 不棄 - 1
}

func (this *SuiteCommandFlow) TestRestoreNoop() {
	game := newGame()
	plain := cores.NewCard(game, 101)
	game.Hand = cores.CardList{plain}

	commandRestore(game, []cores.InstanceID{plain.GetInstanceID()}, nil) // self.cardify = none → no-op
	commandRestore(game, []cores.InstanceID{99}, nil)                    // 非卡牌 → no-op
	this.Nil(plain.GetCardify())

	full := newGame()
	guest := cores.NewGuest(full, 501)
	card := cores.NewCard(full, 101)
	card.CardifyBind(guest)
	full.Hand = cores.CardList{card}
	full.Cardify.Push(guest)
	full.Seat[1], full.Seat[2], full.Seat[3] = &cores.Guest{}, &cores.Guest{}, &cores.Guest{} // 座位全占用

	commandRestore(full, []cores.InstanceID{card.GetInstanceID()}, nil) // 剩餘座位 = 0 → no-op
	this.NotNil(card.GetCardify())
	this.Len(full.Cardify, 1)
}

func (this *SuiteCommandFlow) TestGuestExit() {
	ended := 0
	data := tester.BuildData()
	game := newGameData(data)
	game.GetScore().Set(10)
	game.GetMorale().Set(20)
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(2, guest)
	guest.GetScore().Set(3)
	guest.GetMorale().Set(5)

	data.SetEffect(801, cores.EffectData{Kind: cores.EffectPersist, TargetKind: cores.TargetGuestRand, End: func(game *cores.Game) { ended++ }})
	game.Effect.Push(cores.NewEffect(game, 801, cores.NewRefGuest(guest), 1)) // 離場顧客殘留效果 → 步驟 7 清理

	commandGuestExit(game, []cores.InstanceID{guest.GetInstanceID()}, this.flag(true, true)) // 給滿意、扣士氣
	this.Equal(guest, game.GetExitLast())
	this.Equal(int32(2), game.GetExitLastSeat())
	this.Equal(int32(1), game.GetExitCount())
	this.Nil(game.Seat[2])                             // 自座位移除
	this.Equal(int32(13), game.GetScore().GetValue())  // 10 + 3
	this.Equal(int32(15), game.GetMorale().GetValue()) // 20 - 5(無格擋 / 護盾)
	this.Equal(guest, game.GetDamageGuest())           // morale 特例來源 = 離場顧客
	this.Equal(int32(5), game.GetDamageValue())

	this.Equal(1, ended) // 7. 清理離場顧客殘留效果
	this.Empty(game.Effect)
}

func (this *SuiteCommandFlow) TestGuestExitNoScoreMorale() {
	game := newGame()
	game.GetScore().Set(10)
	game.GetMorale().Set(20)
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(2, guest)
	guest.GetScore().Set(3)
	guest.GetMorale().Set(5)

	commandGuestExit(game, []cores.InstanceID{guest.GetInstanceID()}, this.flag(false, false)) // 不給滿意、不扣士氣
	this.Equal(int32(10), game.GetScore().GetValue())
	this.Equal(int32(20), game.GetMorale().GetValue())
	this.Nil(game.Seat[2]) // 仍移除

	commandGuestExit(game, []cores.InstanceID{99}, nil) // 非顧客實例 → no-op

	frozen := cores.NewGuest(game, 501)
	game.Cardify.Push(frozen)
	commandGuestExit(game, []cores.InstanceID{frozen.GetInstanceID()}, this.flag(false, false)) // 卡牌化(凍結)中 → 該項 no-op(指令隔離防禦)
	this.Len(game.Cardify, 1)
}

func (this *SuiteCommandFlow) TestGuestExitScoreLock() {
	game := newGame()
	game.GetScore().Set(10)
	game.GetScore().Lock() // 滿意值鎖定
	game.GetMorale().Set(20)
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(2, guest)
	guest.GetScore().Set(3)
	guest.GetMorale().Set(5)

	commandGuestExit(game, []cores.InstanceID{guest.GetInstanceID()}, this.flag(true, true)) // 鎖定 → 照離場、滿意值不給
	this.Equal(int32(10), game.GetScore().GetValue())
	this.Nil(game.Seat[2])
}

// TestGuestExitEmit 驗證離場的事件接線: 容器事件(所在 → None, 銷毀離開)+ 給滿意值 / 扣士氣屬性事件(流程寫入白名單, 呼叫點包前後值)。
func (this *SuiteCommandFlow) TestGuestExitEmit() {
	game, record := newGameRecord()
	game.GetMorale().Set(30)
	guest := cores.NewGuest(game, 502) // Score 2、Morale 5
	game.Seat.Place(1, guest)

	guestExitOne(game, guest, cores.ContainerSeat, true, true)
	this.Require().Len(record.Event, 3)
	this.Equal(cores.EventData{Kind: cores.EventContainer, DataID: 502, InstanceID: guest.GetInstanceID(), From: cores.ContainerSeat, To: cores.ContainerNone}, record.Event[0])
	this.Equal(cores.EventData{Kind: cores.EventProperty, Attr: "score", Op: cores.AssignAdd, Operand: 2, Before: 0, After: 2}, record.Event[1])
	this.Equal(cores.EventData{Kind: cores.EventProperty, Attr: "morale", Op: cores.AssignSub, Operand: 5, Before: 30, After: 25}, record.Event[2])
}

func (this *SuiteCommandFlow) TestGuestReturn() {
	game := newGame()
	guest := &cores.Guest{}
	guest.RoamLock() // 入列自動鎖(sate / sateSeal / calmSeal 各 1)
	game.Roam.Push(guest)

	commandGuestReturn(game, []cores.InstanceID{guest.GetInstanceID()}, nil)
	this.Empty(game.Roam)
	this.Equal(guest, game.Seat[1])
	this.Equal(int32(1), guest.GetSeatID())
	this.Equal(int32(0), guest.GetSate().GetLock()) // 解入列自動鎖
	this.Equal(int32(0), guest.GetSateSeal().GetLock())
	this.Equal(int32(0), guest.GetCalmSeal().GetLock())
}

func (this *SuiteCommandFlow) TestGuestReturnNoop() {
	full := newGame()
	full.GetMorale().Set(20)
	guest := cores.NewGuest(full, 501)
	guest.GetMorale().Set(5)
	full.Roam.Push(guest)
	full.Seat[1], full.Seat[2], full.Seat[3] = &cores.Guest{}, &cores.Guest{}, &cores.Guest{} // 全占用

	commandGuestReturn(full, []cores.InstanceID{guest.GetInstanceID()}, nil) // 剩餘座位 = 0 → 走 guestExit(false, true)
	this.Empty(full.Roam)
	this.Equal(int32(1), full.GetExitCount())
	this.Equal(int32(15), full.GetMorale().GetValue()) // 扣士氣 20 - 5

	seated := newGame()
	stay := cores.NewGuest(seated, 501)
	seated.Seat.Place(1, stay)
	commandGuestReturn(seated, []cores.InstanceID{stay.GetInstanceID()}, nil) // 非遊蕩列表 → no-op
	this.NotNil(seated.Seat[1])
}

func (this *SuiteCommandFlow) TestGuestRoam() {
	game := newGame()
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(2, guest)
	sealLock := guest.GetSateSeal().GetLock() // 501 封印飽食初始鎖 1

	commandGuestRoam(game, []cores.InstanceID{guest.GetInstanceID()}, nil)
	this.Nil(game.Seat[2])
	this.Require().Len(game.Roam, 1)
	this.Equal(int32(0), guest.GetSeatID())
	this.Equal(int32(1), guest.GetSate().GetLock()) // 自動鎖
	this.Equal(sealLock+1, guest.GetSateSeal().GetLock())
	this.Equal(int32(1), guest.GetCalmSeal().GetLock())

	commandGuestRoam(game, []cores.InstanceID{guest.GetInstanceID()}, nil) // 已在遊蕩(非座位)→ no-op
	this.Len(game.Roam, 1)
}

func (this *SuiteCommandFlow) TestGuestSeat() {
	game := newGame()
	guest := cores.NewGuest(game, 501)
	game.Wait.Insert(guest)

	commandGuestSeat(game, nil, nil)
	this.Empty(game.Wait)
	this.Equal(guest, game.Seat[1])
	this.Equal(int32(1), guest.GetSeatID())
	this.Equal(guest, game.GetSeatLast())
	this.Equal(int32(1), game.GetSeatCount())

	commandGuestSeat(game, nil, nil) // 排隊佇列空 → no-op
	this.Equal(int32(1), game.GetSeatCount())

	full := newGame()
	full.Wait.Insert(cores.NewGuest(full, 501))
	full.Seat[1], full.Seat[2], full.Seat[3] = &cores.Guest{}, &cores.Guest{}, &cores.Guest{}
	commandGuestSeat(full, nil, nil) // 無空座位 → no-op
	this.Len(full.Wait, 1)
}

// TestGuestSeatOneEmit 驗證入座的容器事件: 排隊 → 座位、To == Seat 帶座位編號。
func (this *SuiteCommandFlow) TestGuestSeatOneEmit() {
	game, record := newGameRecord()
	guest := cores.NewGuest(game, 501)
	game.Wait.Insert(guest)

	this.True(guestSeatOne(game))
	this.Require().NotEmpty(record.Event)
	this.Equal(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: guest.GetInstanceID(), From: cores.ContainerWait, To: cores.ContainerSeat, SeatID: 1}, record.Event[0])
}

func (this *SuiteCommandFlow) TestRemoveGuestContainer() {
	game := newGame()
	g1 := cores.NewGuest(game, 501)
	g2 := cores.NewGuest(game, 501)
	g3 := cores.NewGuest(game, 501)
	g4 := cores.NewGuest(game, 501)
	keep := cores.NewGuest(game, 501)
	game.Seat.Place(1, g1)
	game.Wait = cores.WaitList{g2, keep} // 兩位: 移除 g2、保留 keep
	game.Roam.Push(g3)
	game.Cardify.Push(g4)

	removeGuestContainer(game, cores.ContainerSeat, g1)
	removeGuestContainer(game, cores.ContainerWait, g2)
	removeGuestContainer(game, cores.ContainerRoam, g3)
	removeGuestContainer(game, cores.ContainerCardify, g4)
	removeGuestContainer(game, cores.ContainerHand, g1) // 非顧客容器 → default no-op

	this.Nil(game.Seat[1])
	this.Equal(cores.WaitList{keep}, game.Wait) // g2 移除、keep 保留
	this.Empty(game.Roam)
	this.Empty(game.Cardify)
}

func (this *SuiteCommandFlow) TestCardRunTrigger() {
	fired := false
	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardPlay, Trigger: func(*cores.Game) { fired = true }})
	game := newGameData(data)
	card := cores.NewCard(game, 101)
	game.Hand = cores.CardList{card}
	game.Effect.Push(cores.NewEffect(game, 801, cores.Ref{}, 1))

	commandCardRun(game, []cores.InstanceID{card.GetInstanceID()}, this.flag(false, false)) // 不耗點、不進棄牌
	this.True(fired)                                                                        // 玩家出牌觸發
}

func (this *SuiteCommandFlow) TestMorphTrigger() {
	fired := false
	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardMorph, Trigger: func(*cores.Game) { fired = true }})
	game := newGameData(data)
	card := cores.NewCard(game, 102)
	game.Hand = cores.CardList{card}
	game.Effect.Push(cores.NewEffect(game, 801, cores.Ref{}, 1))

	morph(game, []cores.InstanceID{card.GetInstanceID()}, []exprs.Value{exprs.NewNum(7)}, cores.ContainerHand) // 抽獎群組 7 → 變身
	this.True(fired)                                                                                           // 卡牌變身觸發
}

func (this *SuiteCommandFlow) TestGuestSeatTrigger() {
	fired := false
	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerGuestSeat, Trigger: func(*cores.Game) { fired = true }})
	game := newGameData(data)
	game.Wait.Insert(cores.NewGuest(game, 501))
	game.Effect.Push(cores.NewEffect(game, 801, cores.Ref{}, 1))

	commandGuestSeat(game, nil, nil)
	this.True(fired) // 顧客入座觸發
}

func (this *SuiteCommandFlow) TestGuestExitTrigger() {
	fired := []cores.TriggerKind{}
	record := func(timing cores.TriggerKind) cores.EffectExec {
		return func(*cores.Game) { fired = append(fired, timing) }
	}

	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerExitAny, Trigger: record(cores.TriggerExitAny)})
	data.SetEffect(802, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerExitSate, Trigger: record(cores.TriggerExitSate)})
	data.SetEffect(803, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerExitCalm, Trigger: record(cores.TriggerExitCalm)})
	data.SetEffect(804, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerExitDone, Trigger: record(cores.TriggerExitDone)})
	data.SetEffect(805, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerDamage, Trigger: record(cores.TriggerDamage)})
	game := newGameData(data)
	game.GetMorale().Set(20)
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(2, guest)
	guest.GetScore().Set(3)
	guest.GetMorale().Set(5)

	for _, itor := range []int32{801, 802, 803, 804, 805} {
		game.Effect.Push(cores.NewEffect(game, itor, cores.Ref{}, 1))
	} // for

	guestExitOne(game, guest, cores.ContainerSeat, true, true) // 給滿意 + 扣士氣
	// 依序: 離場(2) → 飽食(3) → 生氣(4) → 離場後(6) → 士氣受損(9, 扣士氣經 moraleDamage)
	this.Equal([]cores.TriggerKind{cores.TriggerExitAny, cores.TriggerExitSate, cores.TriggerExitCalm, cores.TriggerExitDone, cores.TriggerDamage}, fired)
}

// === 測試輔助(置尾) ===

// flag 把兩個布林包成參數列表(供帶兩個布林旗標的流程命令: cardRun 消耗點數 / 進棄牌堆、guestExit 給滿意 / 扣士氣)。
func (this *SuiteCommandFlow) flag(a, b bool) []exprs.Value {
	return []exprs.Value{exprs.NewBool(a), exprs.NewBool(b)}
}
