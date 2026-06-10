package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteCommandFlow(t *testing.T) {
	suite.Run(t, new(SuiteCommandFlow))
}

// SuiteCommandFlow 驗證處理流程命令(commandFlow.go):cardRun / *Morph / cardify / restore / guestExit / guestReturn / guestRoam / guestSeat。
type SuiteCommandFlow struct {
	suite.Suite
}

func (this *SuiteCommandFlow) TestCardRun() {
	game := NewGame(0, nil, nil, nil, nil)
	game.energy = NewValue(5, 0)
	card := &Card{instanceID: 1, cardID: 103, cost: NewValue(2, 0)}
	game.Hand = []*Card{card}
	injectPort(game)

	commandCardRun(game, []InstanceID{1}, this.flag(true, true)) // 消耗點數、進棄牌堆
	this.Equal(int32(3), game.GetEnergy().GetValue())            // 5 - 2
	this.Equal(card, game.GetPlayLast())
	this.Equal(int32(1), game.GetPlayCount())
	this.Equal(int32(1), game.GetPlayTotal().Get(1)) // 卡 103 群組 1
	this.Empty(game.Hand)
	this.Equal(CardList{card}, game.Drop)
}

func (this *SuiteCommandFlow) TestCardRunNoop() {
	game := NewGame(0, nil, nil, nil, nil)
	game.energy = NewValue(5, 0)
	injectPort(game)

	sealed := &Card{instanceID: 1, seal: NewValue(0, 1)}
	game.Hand = []*Card{sealed}
	commandCardRun(game, []InstanceID{1}, this.flag(true, false)) // 封印 → no-op
	this.Len(game.Hand, 1)
	this.Equal(int32(5), game.GetEnergy().GetValue())

	poor := &Card{instanceID: 2, cost: NewValue(9, 0)}
	game.Hand = []*Card{poor}
	commandCardRun(game, []InstanceID{2}, this.flag(true, false)) // 點數不足 → no-op
	this.Equal(int32(5), game.GetEnergy().GetValue())

	noEnergy := &Card{instanceID: 3, cardID: 103}
	game.Hand = []*Card{noEnergy}
	commandCardRun(game, []InstanceID{3}, this.flag(false, false)) // 不耗點數、不進棄牌堆 → 留手牌
	this.Equal(noEnergy, game.GetPlayLast())
	this.Len(game.Hand, 1)

	inDrop := &Card{instanceID: 4, cardID: 103}
	game.Drop = []*Card{inDrop}
	commandCardRun(game, []InstanceID{4}, this.flag(false, true)) // 已在棄牌堆 → 不重複移動
	this.Len(game.Drop, 1)

	game.Exile = []*Card{{instanceID: 5}}
	commandCardRun(game, []InstanceID{5}, nil) // 流放牌堆(位置不符)→ no-op
	this.Len(game.Exile, 1)

	commandCardRun(game, []InstanceID{99}, nil) // 不存在 → no-op
}

func (this *SuiteCommandFlow) TestCardRunEnergyLock() {
	game := NewGame(0, nil, nil, nil, nil)
	game.energy = NewValue(5, 1) // 點數鎖定
	card := &Card{instanceID: 1, cardID: 103, cost: NewValue(2, 0)}
	game.Hand = []*Card{card}
	injectPort(game)

	commandCardRun(game, []InstanceID{1}, this.flag(true, true)) // 鎖定 → 照出牌、耗能不扣
	this.Equal(int32(5), game.GetEnergy().GetValue())
	this.Equal(CardList{card}, game.Drop)
}

func (this *SuiteCommandFlow) TestCardRunEffect() {
	card := &Card{instanceID: 10, cardID: 101, effectID: NewIDList(801)}
	game := NewGame(0, nil, nil, nil, nil)
	game.Hand = []*Card{card}
	injectPort(game)
	game.effectData = map[int32]effectData{
		801: {Kind: EffectTrigger, TargetKind: TargetNone, Stack: 1},
	}

	commandCardRun(game, []InstanceID{10}, this.flag(false, false)) // 不耗點、不進棄牌
	this.Require().Len(game.Effect, 1)                              // 實例效果列表啟動 → 效果入佇列
	this.Equal(int32(801), game.Effect[0].GetEffectID())
}

func (this *SuiteCommandFlow) TestCardRunEffectMove() {
	card := &Card{instanceID: 10, cardID: 101, effectID: NewIDList(802)}
	game := NewGame(0, nil, nil, nil, nil)
	game.Hand = []*Card{card}
	injectPort(game)
	game.effectData = map[int32]effectData{
		802: {Kind: EffectImmed, TargetKind: TargetNone, Immed: func(e *Game) {
			removeCard(e, ContainerHand, card)
			placeCard(e, ContainerExile, card) // 立即命令在出牌途中把本卡移至流放牌堆
		}},
	}

	commandCardRun(game, []InstanceID{10}, this.flag(false, true)) // 不耗點、進棄牌堆
	this.Empty(game.Exile)                                         // 進棄牌堆前重新定位 → 自當前容器(流放)移出、不殘留
	this.Equal(CardList{card}, game.Drop)                          // 僅進棄牌堆一次、不重複
}

func (this *SuiteCommandFlow) TestMorph() {
	game := NewGame(0, nil, nil, nil, nil)
	card := &Card{instanceID: game.NextID(), cardID: 102, cost: NewValue(9, 0)}
	id := card.GetInstanceID()
	game.Hand = []*Card{card}
	injectPort(game)

	commandHandMorph(game, []InstanceID{id}, nums(7)) // 群組 7 → 抽中卡 101
	this.Equal(int32(101), card.GetCardID())          // cardID 換成抽中值
	this.NotEqual(id, card.GetInstanceID())           // 重分配實例編號
	this.Equal(int32(0), card.GetCost().GetValue())   // 載卡 101 資料(無 Cost → 0)
	this.Equal(int32(102), game.GetMorphOldID())
	this.Equal(int32(101), game.GetMorphNewID())
	this.Equal(card, game.GetMorphLast())
	this.Equal(int32(1), game.GetMorphCount())
	this.Len(game.Hand, 1) // 留原牌堆
}

func (this *SuiteCommandFlow) TestMorphNoop() {
	game := NewGame(0, nil, nil, nil, nil)
	card := &Card{instanceID: 1, cardID: 102}
	game.Hand = []*Card{card}
	injectPort(game)

	commandDeckMorph(game, []InstanceID{1}, nums(7)) // 位置不符(卡在手牌)→ no-op
	this.Equal(int32(102), card.GetCardID())

	commandHandMorph(game, []InstanceID{1}, nums(8)) // 群組總權重 0 → no-op
	commandHandMorph(game, []InstanceID{1}, nums(9)) // 抽中 999 無資料 → 跳過
	commandHandMorph(game, []InstanceID{1}, nil)     // 缺參數 → no-op
	this.Equal(int32(102), card.GetCardID())

	game.Drop = []*Card{{instanceID: 2, cardID: 102}}
	commandDropMorph(game, []InstanceID{2}, nums(7))
	this.Equal(int32(101), game.Drop[0].GetCardID())
	game.Exile = []*Card{{instanceID: 3, cardID: 102}}
	commandExileMorph(game, []InstanceID{3}, nums(7))
	this.Equal(int32(101), game.Exile[0].GetCardID())
}

func (this *SuiteCommandFlow) TestCardify() {
	game := NewGame(0, nil, nil, nil, nil)
	game.round = NewValue(7, 0)
	guest := &Guest{instanceID: 11, seatID: 2}
	game.Seat[2] = guest
	injectPort(game)

	commandCardify(game, []InstanceID{11}, nums(101)) // 卡牌化為卡 101
	this.Nil(game.Seat[2])                            // 1. 自座位移除
	this.Require().Len(game.Cardify, 1)               // 2. 入卡牌化列表
	this.Equal(int32(0), guest.GetSeatID())
	this.Equal(int32(7), guest.GetFreeze()) // 3. 凍結
	this.Require().Len(game.Hand, 1)
	card := game.Hand[0]
	this.Equal(int32(101), card.GetCardID())
	this.Equal(guest, card.GetCardify())           // 5. 綁來源
	this.Equal(int32(1), card.GetKeep().GetLock()) // 6. 不棄 + 1

	game.Roam = []*Guest{{instanceID: 12}}
	commandCardify(game, []InstanceID{12}, nums(101)) // 不在座位列表 → no-op
	this.Len(game.Cardify, 1)

	game.Seat[3] = &Guest{instanceID: 13, seatID: 3}
	commandCardify(game, []InstanceID{13}, nums(999)) // 卡牌資料不存在 → no-op
	commandCardify(game, []InstanceID{13}, nil)       // 缺卡牌編號 → no-op
	this.NotNil(game.Seat[3])
}

func (this *SuiteCommandFlow) TestRestore() {
	game := NewGame(0, nil, nil, nil, nil)
	game.round = NewValue(10, 0)
	guest := &Guest{instanceID: 11, freeze: 4}
	card := &Card{instanceID: 1, cardID: 101, cardify: guest, keep: NewValue(0, 1)}
	game.Hand = []*Card{card}
	game.Cardify = []*Guest{guest}
	game.Effect = EffectList{{self: NewRefGuest(guest), expire: 5}}
	injectPort(game) // 座位 1,2,3 全空 → randomEmptySeat → seat 1

	commandRestore(game, []InstanceID{1}, nil)
	this.Empty(game.Cardify)        // 2. 移出卡牌化列表
	this.Equal(guest, game.Seat[1]) // 3. 入隨機空座位
	this.Equal(int32(1), guest.GetSeatID())
	this.Equal(int32(11), game.Effect[0].GetExpire()) // 4. 5 + (10 - 4)
	this.Equal(int32(0), guest.GetFreeze())           // 5. 解凍
	this.Nil(card.GetCardify())                       // 6. 解綁
	this.Equal(int32(0), card.GetKeep().GetLock())    // 7. 不棄 - 1
}

func (this *SuiteCommandFlow) TestRestoreNoop() {
	game := NewGame(0, nil, nil, nil, nil)
	plain := &Card{instanceID: 2, cardID: 101}
	game.Hand = []*Card{plain}
	injectPort(game)

	commandRestore(game, []InstanceID{2}, nil)  // self.cardify = none → no-op
	commandRestore(game, []InstanceID{99}, nil) // 非卡牌 → no-op
	this.Nil(plain.GetCardify())

	full := NewGame(0, nil, nil, nil, nil)
	guest := &Guest{instanceID: 11}
	card := &Card{instanceID: 1, cardID: 101, cardify: guest, keep: NewValue(0, 1)}
	full.Hand = []*Card{card}
	full.Cardify = []*Guest{guest}
	full.Seat[1], full.Seat[2], full.Seat[3] = &Guest{}, &Guest{}, &Guest{} // 座位全占用
	injectPort(full)

	commandRestore(full, []InstanceID{1}, nil) // 剩餘座位 = 0 → no-op
	this.NotNil(card.GetCardify())
	this.Len(full.Cardify, 1)
}

func (this *SuiteCommandFlow) TestGuestExit() {
	game := NewGame(0, nil, nil, nil, nil)
	game.score = NewValue(10, 0)
	game.morale = NewValue(20, 0)
	guest := &Guest{instanceID: 11, seatID: 2, score: NewValue(3, 0), morale: NewValue(5, 0)}
	game.Seat[2] = guest
	injectPort(game)

	commandGuestExit(game, []InstanceID{11}, this.flag(true, true)) // 給滿意、扣士氣
	this.Equal(guest, game.GetExitLast())
	this.Equal(int32(2), game.GetExitLastSeat())
	this.Equal(int32(1), game.GetExitCount())
	this.Nil(game.Seat[2])                             // 自座位移除
	this.Equal(int32(13), game.GetScore().GetValue())  // 10 + 3
	this.Equal(int32(15), game.GetMorale().GetValue()) // 20 - 5(無格擋 / 護盾)
	this.Equal(guest, game.GetDamageGuest())           // morale 特例來源 = 離場顧客
	this.Equal(int32(5), game.GetDamageValue())
}

func (this *SuiteCommandFlow) TestGuestExitNoScoreMorale() {
	game := NewGame(0, nil, nil, nil, nil)
	game.score = NewValue(10, 0)
	game.morale = NewValue(20, 0)
	guest := &Guest{instanceID: 11, seatID: 2, score: NewValue(3, 0), morale: NewValue(5, 0)}
	game.Seat[2] = guest
	injectPort(game)

	commandGuestExit(game, []InstanceID{11}, this.flag(false, false)) // 不給滿意、不扣士氣
	this.Equal(int32(10), game.GetScore().GetValue())
	this.Equal(int32(20), game.GetMorale().GetValue())
	this.Nil(game.Seat[2]) // 仍移除

	commandGuestExit(game, []InstanceID{99}, nil) // 非顧客實例 → no-op
}

func (this *SuiteCommandFlow) TestGuestExitScoreLock() {
	game := NewGame(0, nil, nil, nil, nil)
	game.score = NewValue(10, 1) // 滿意值鎖定
	game.morale = NewValue(20, 0)
	guest := &Guest{instanceID: 11, seatID: 2, score: NewValue(3, 0), morale: NewValue(5, 0)}
	game.Seat[2] = guest
	injectPort(game)

	commandGuestExit(game, []InstanceID{11}, this.flag(true, true)) // 鎖定 → 照離場、滿意值不給
	this.Equal(int32(10), game.GetScore().GetValue())
	this.Nil(game.Seat[2])
}

func (this *SuiteCommandFlow) TestGuestReturn() {
	game := NewGame(0, nil, nil, nil, nil)
	guest := &Guest{instanceID: 11, sate: NewValue(0, 1), sateSeal: NewValue(0, 1), calmSeal: NewValue(0, 1)}
	game.Roam = []*Guest{guest}
	injectPort(game) // 座位全空 → seat 1

	commandGuestReturn(game, []InstanceID{11}, nil)
	this.Empty(game.Roam)
	this.Equal(guest, game.Seat[1])
	this.Equal(int32(1), guest.GetSeatID())
	this.Equal(int32(0), guest.GetSate().GetLock()) // 解入列自動鎖
	this.Equal(int32(0), guest.GetSateSeal().GetLock())
	this.Equal(int32(0), guest.GetCalmSeal().GetLock())
}

func (this *SuiteCommandFlow) TestGuestReturnNoop() {
	full := NewGame(0, nil, nil, nil, nil)
	full.morale = NewValue(20, 0)
	guest := &Guest{instanceID: 11, morale: NewValue(5, 0)}
	full.Roam = []*Guest{guest}
	full.Seat[1], full.Seat[2], full.Seat[3] = &Guest{}, &Guest{}, &Guest{} // 全占用
	injectPort(full)

	commandGuestReturn(full, []InstanceID{11}, nil) // 剩餘座位 = 0 → 走 guestExit(false, true)
	this.Empty(full.Roam)
	this.Equal(int32(1), full.GetExitCount())
	this.Equal(int32(15), full.GetMorale().GetValue()) // 扣士氣 20 - 5

	seated := NewGame(0, nil, nil, nil, nil)
	seated.Seat[1] = &Guest{instanceID: 11, seatID: 1}
	injectPort(seated)
	commandGuestReturn(seated, []InstanceID{11}, nil) // 非遊蕩列表 → no-op
	this.NotNil(seated.Seat[1])
}

func (this *SuiteCommandFlow) TestGuestRoam() {
	game := NewGame(0, nil, nil, nil, nil)
	guest := &Guest{instanceID: 11, seatID: 2}
	game.Seat[2] = guest
	injectPort(game)

	commandGuestRoam(game, []InstanceID{11}, nil)
	this.Nil(game.Seat[2])
	this.Require().Len(game.Roam, 1)
	this.Equal(int32(0), guest.GetSeatID())
	this.Equal(int32(1), guest.GetSate().GetLock()) // 自動鎖
	this.Equal(int32(1), guest.GetSateSeal().GetLock())
	this.Equal(int32(1), guest.GetCalmSeal().GetLock())

	commandGuestRoam(game, []InstanceID{11}, nil) // 已在遊蕩(非座位)→ no-op
	this.Len(game.Roam, 1)
}

func (this *SuiteCommandFlow) TestGuestSeat() {
	game := NewGame(0, nil, nil, nil, nil)
	guest := &Guest{instanceID: 11}
	game.Wait = []*Guest{guest}
	injectPort(game) // 座位全空 → seat 1

	commandGuestSeat(game, nil, nil)
	this.Empty(game.Wait)
	this.Equal(guest, game.Seat[1])
	this.Equal(int32(1), guest.GetSeatID())
	this.Equal(guest, game.GetSeatLast())
	this.Equal(int32(1), game.GetSeatCount())

	commandGuestSeat(game, nil, nil) // 排隊佇列空 → no-op
	this.Equal(int32(1), game.GetSeatCount())

	full := NewGame(0, nil, nil, nil, nil)
	full.Wait = []*Guest{{instanceID: 12}}
	full.Seat[1], full.Seat[2], full.Seat[3] = &Guest{}, &Guest{}, &Guest{}
	injectPort(full)
	commandGuestSeat(full, nil, nil) // 無空座位 → no-op
	this.Len(full.Wait, 1)
}

func (this *SuiteCommandFlow) TestRemoveGuestContainer() {
	game := NewGame(0, nil, nil, nil, nil)
	g1 := &Guest{instanceID: 1, seatID: 1}
	g2 := &Guest{instanceID: 2}
	g3 := &Guest{instanceID: 3}
	g4 := &Guest{instanceID: 4}
	keep := &Guest{instanceID: 5}
	game.Seat[1] = g1
	game.Wait = []*Guest{g2, keep} // 兩位:移除 g2、保留 keep(removeGuest 保留分支)
	game.Roam = []*Guest{g3}
	game.Cardify = []*Guest{g4}
	injectPort(game)

	removeGuestContainer(game, ContainerSeat, g1)
	removeGuestContainer(game, ContainerWait, g2)
	removeGuestContainer(game, ContainerRoam, g3)
	removeGuestContainer(game, ContainerCardify, g4)
	removeGuestContainer(game, ContainerHand, g1) // 非顧客容器 → default no-op

	this.Nil(game.Seat[1])
	this.Equal(WaitList{keep}, game.Wait) // g2 移除、keep 保留
	this.Empty(game.Roam)
	this.Empty(game.Cardify)
}

func (this *SuiteCommandFlow) TestCardRunTrigger() {
	fired := false
	game := NewGame(0, nil, nil, nil, nil)
	game.Hand = []*Card{{instanceID: 10, cardID: 101}}
	game.Effect = EffectList{{instanceID: 1, effectID: 801, stack: 1}}
	injectPort(game)
	game.effectData = map[int32]effectData{
		801: {Kind: EffectTrigger, TriggerKind: TriggerCardPlay, Trigger: func(*Game) { fired = true }},
	}

	commandCardRun(game, []InstanceID{10}, this.flag(false, false)) // 不耗點、不進棄牌
	this.True(fired)                                                // 玩家出牌觸發
}

func (this *SuiteCommandFlow) TestMorphTrigger() {
	fired := false
	game := NewGame(0, nil, nil, nil, nil)
	game.Hand = []*Card{{instanceID: 10, cardID: 102}}
	game.Effect = EffectList{{instanceID: 1, effectID: 801, stack: 1}}
	injectPort(game)
	game.effectData = map[int32]effectData{
		801: {Kind: EffectTrigger, TriggerKind: TriggerCardMorph, Trigger: func(*Game) { fired = true }},
	}

	morph(game, []InstanceID{10}, []exprs.Value{exprs.NewNum(7)}, ContainerHand) // 抽獎群組 7 → 變身
	this.True(fired)                                                             // 卡牌變身觸發
}

func (this *SuiteCommandFlow) TestGuestSeatTrigger() {
	fired := false
	game := NewGame(0, nil, nil, nil, nil)
	game.Wait = []*Guest{{instanceID: 11}}
	game.Effect = EffectList{{instanceID: 1, effectID: 801, stack: 1}}
	injectPort(game)
	game.effectData = map[int32]effectData{
		801: {Kind: EffectTrigger, TriggerKind: TriggerGuestSeat, Trigger: func(*Game) { fired = true }},
	}

	commandGuestSeat(game, nil, nil)
	this.True(fired) // 顧客入座觸發
}

func (this *SuiteCommandFlow) TestGuestExitTrigger() {
	fired := []TriggerKind{}
	record := func(timing TriggerKind) EffectCommand { return func(*Game) { fired = append(fired, timing) } }

	game := NewGame(0, nil, nil, nil, nil)
	game.morale = NewValue(20, 0)
	guest := &Guest{instanceID: 11, seatID: 2, score: NewValue(3, 0), morale: NewValue(5, 0)}
	game.Seat[2] = guest
	game.Effect = EffectList{
		{instanceID: 1, effectID: 801, stack: 1},
		{instanceID: 2, effectID: 802, stack: 1},
		{instanceID: 3, effectID: 803, stack: 1},
		{instanceID: 4, effectID: 804, stack: 1},
		{instanceID: 5, effectID: 805, stack: 1},
	}
	injectPort(game)
	game.effectData = map[int32]effectData{
		801: {Kind: EffectTrigger, TriggerKind: TriggerExitAny, Trigger: record(TriggerExitAny)},
		802: {Kind: EffectTrigger, TriggerKind: TriggerExitSate, Trigger: record(TriggerExitSate)},
		803: {Kind: EffectTrigger, TriggerKind: TriggerExitCalm, Trigger: record(TriggerExitCalm)},
		804: {Kind: EffectTrigger, TriggerKind: TriggerExitDone, Trigger: record(TriggerExitDone)},
		805: {Kind: EffectTrigger, TriggerKind: TriggerDamage, Trigger: record(TriggerDamage)},
	}

	guestExitOne(game, guest, ContainerSeat, true, true) // 給滿意 + 扣士氣
	// 依序:離場(2) → 飽食(3) → 生氣(4) → 離場後(6) → 士氣受損(9,扣士氣經 moraleDamage)
	this.Equal([]TriggerKind{TriggerExitAny, TriggerExitSate, TriggerExitCalm, TriggerExitDone, TriggerDamage}, fired)
}

// === 測試輔助(置尾) ===

// flag 把兩個布林包成參數列表(供帶兩個布林旗標的流程命令:cardRun 消耗點數 / 進棄牌堆、guestExit 給滿意 / 扣士氣)。
func (this *SuiteCommandFlow) flag(a, b bool) []exprs.Value {
	return []exprs.Value{exprs.NewBool(a), exprs.NewBool(b)}
}
