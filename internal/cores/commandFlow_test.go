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
	runtime := NewRuntime(0)
	runtime.Game.Energy = Value{Value: 5}
	card := &Card{InstanceID: 1, CardID: 103, Cost: Value{Value: 2}}
	runtime.Hand = []*Card{card}
	eng := this.engine(runtime)

	commandCardRun(eng, []InstanceID{1}, this.flag(true, true)) // 消耗點數、進棄牌堆
	this.Equal(int32(3), runtime.Game.Energy.Value)             // 5 - 2
	this.Equal(card, runtime.Game.PlayLast)
	this.Equal(int32(1), runtime.Game.PlayCount)
	this.Equal(int32(1), runtime.Game.PlayTotal[1]) // 卡 103 群組 1
	this.Empty(runtime.Hand)
	this.Equal([]*Card{card}, runtime.Drop)
}

func (this *SuiteCommandFlow) TestCardRunNoop() {
	runtime := NewRuntime(0)
	runtime.Game.Energy = Value{Value: 5}
	eng := this.engine(runtime)

	sealed := &Card{InstanceID: 1, Seal: Value{Lock: 1}}
	runtime.Hand = []*Card{sealed}
	commandCardRun(eng, []InstanceID{1}, this.flag(true, false)) // 封印 → no-op
	this.Len(runtime.Hand, 1)
	this.Equal(int32(5), runtime.Game.Energy.Value)

	poor := &Card{InstanceID: 2, Cost: Value{Value: 9}}
	runtime.Hand = []*Card{poor}
	commandCardRun(eng, []InstanceID{2}, this.flag(true, false)) // 點數不足 → no-op
	this.Equal(int32(5), runtime.Game.Energy.Value)

	noEnergy := &Card{InstanceID: 3, CardID: 103}
	runtime.Hand = []*Card{noEnergy}
	commandCardRun(eng, []InstanceID{3}, this.flag(false, false)) // 不耗點數、不進棄牌堆 → 留手牌
	this.Equal(noEnergy, runtime.Game.PlayLast)
	this.Len(runtime.Hand, 1)

	inDrop := &Card{InstanceID: 4, CardID: 103}
	runtime.Drop = []*Card{inDrop}
	commandCardRun(eng, []InstanceID{4}, this.flag(false, true)) // 已在棄牌堆 → 不重複移動
	this.Len(runtime.Drop, 1)

	runtime.Exile = []*Card{{InstanceID: 5}}
	commandCardRun(eng, []InstanceID{5}, nil) // 流放牌堆(位置不符)→ no-op
	this.Len(runtime.Exile, 1)

	commandCardRun(eng, []InstanceID{99}, nil) // 不存在 → no-op
}

func (this *SuiteCommandFlow) TestMorph() {
	runtime := NewRuntime(0)
	card := &Card{InstanceID: runtime.NextID(), CardID: 102, Cost: Value{Value: 9}}
	id := card.InstanceID
	runtime.Hand = []*Card{card}
	eng := this.engine(runtime)

	commandHandMorph(eng, []InstanceID{id}, nums(7)) // 群組 7 → 抽中卡 101
	this.Equal(int32(101), card.CardID)              // cardID 換成抽中值
	this.NotEqual(id, card.InstanceID)               // 重分配實例編號
	this.Equal(int32(0), card.Cost.Value)            // 載卡 101 資料(無 Cost → 0)
	this.Equal(int32(102), runtime.Game.MorphOldID)
	this.Equal(int32(101), runtime.Game.MorphNewID)
	this.Equal(card, runtime.Game.MorphLast)
	this.Equal(int32(1), runtime.Game.MorphCount)
	this.Len(runtime.Hand, 1) // 留原牌堆
}

func (this *SuiteCommandFlow) TestMorphNoop() {
	runtime := NewRuntime(0)
	card := &Card{InstanceID: 1, CardID: 102}
	runtime.Hand = []*Card{card}
	eng := this.engine(runtime)

	commandDeckMorph(eng, []InstanceID{1}, nums(7)) // 位置不符(卡在手牌)→ no-op
	this.Equal(int32(102), card.CardID)

	commandHandMorph(eng, []InstanceID{1}, nums(8)) // 群組總權重 0 → no-op
	commandHandMorph(eng, []InstanceID{1}, nums(9)) // 抽中 999 無資料 → 跳過
	commandHandMorph(eng, []InstanceID{1}, nil)     // 缺參數 → no-op
	this.Equal(int32(102), card.CardID)

	runtime.Drop = []*Card{{InstanceID: 2, CardID: 102}}
	commandDropMorph(eng, []InstanceID{2}, nums(7))
	this.Equal(int32(101), runtime.Drop[0].CardID)
	runtime.Exile = []*Card{{InstanceID: 3, CardID: 102}}
	commandExileMorph(eng, []InstanceID{3}, nums(7))
	this.Equal(int32(101), runtime.Exile[0].CardID)
}

func (this *SuiteCommandFlow) TestCardify() {
	runtime := NewRuntime(0)
	runtime.Game.Round = 7
	guest := &Guest{InstanceID: 11, SeatID: 2}
	runtime.Seat[2] = guest
	eng := this.engine(runtime)

	commandCardify(eng, []InstanceID{11}, nums(101)) // 卡牌化為卡 101
	this.Nil(runtime.Seat[2])                        // 1. 自座位移除
	this.Require().Len(runtime.Cardify, 1)           // 2. 入卡牌化列表
	this.Equal(int32(0), guest.SeatID)
	this.Equal(int32(7), guest.Freeze) // 3. 凍結
	this.Require().Len(runtime.Hand, 1)
	card := runtime.Hand[0]
	this.Equal(int32(101), card.CardID)
	this.Equal(guest, card.Cardify)      // 5. 綁來源
	this.Equal(int32(1), card.Keep.Lock) // 6. 不棄 + 1

	runtime.Roam = []*Guest{{InstanceID: 12}}
	commandCardify(eng, []InstanceID{12}, nums(101)) // 不在座位列表 → no-op
	this.Len(runtime.Cardify, 1)

	runtime.Seat[3] = &Guest{InstanceID: 13, SeatID: 3}
	commandCardify(eng, []InstanceID{13}, nums(999)) // 卡牌資料不存在 → no-op
	commandCardify(eng, []InstanceID{13}, nil)       // 缺卡牌編號 → no-op
	this.NotNil(runtime.Seat[3])
}

func (this *SuiteCommandFlow) TestRestore() {
	runtime := NewRuntime(0)
	runtime.Game.Round = 10
	guest := &Guest{InstanceID: 11, Freeze: 4}
	card := &Card{InstanceID: 1, CardID: 101, Cardify: guest, Keep: Value{Lock: 1}}
	runtime.Hand = []*Card{card}
	runtime.Cardify = []*Guest{guest}
	runtime.Effect = []*Effect{{Self: Self{Guest: guest}, Expire: 5}}
	eng := this.engine(runtime) // 座位 1,2,3 全空 → randomEmptySeat → seat 1

	commandRestore(eng, []InstanceID{1}, nil)
	this.Empty(runtime.Cardify)        // 2. 移出卡牌化列表
	this.Equal(guest, runtime.Seat[1]) // 3. 入隨機空座位
	this.Equal(int32(1), guest.SeatID)
	this.Equal(int32(11), runtime.Effect[0].Expire) // 4. 5 + (10 - 4)
	this.Equal(int32(0), guest.Freeze)              // 5. 解凍
	this.Nil(card.Cardify)                          // 6. 解綁
	this.Equal(int32(0), card.Keep.Lock)            // 7. 不棄 - 1
}

func (this *SuiteCommandFlow) TestRestoreNoop() {
	runtime := NewRuntime(0)
	plain := &Card{InstanceID: 2, CardID: 101}
	runtime.Hand = []*Card{plain}
	eng := this.engine(runtime)

	commandRestore(eng, []InstanceID{2}, nil)  // self.cardify = none → no-op
	commandRestore(eng, []InstanceID{99}, nil) // 非卡牌 → no-op
	this.Nil(plain.Cardify)

	full := NewRuntime(0)
	guest := &Guest{InstanceID: 11}
	card := &Card{InstanceID: 1, CardID: 101, Cardify: guest, Keep: Value{Lock: 1}}
	full.Hand = []*Card{card}
	full.Cardify = []*Guest{guest}
	full.Seat[1], full.Seat[2], full.Seat[3] = &Guest{}, &Guest{}, &Guest{} // 座位全占用
	engFull := this.engine(full)

	commandRestore(engFull, []InstanceID{1}, nil) // 剩餘座位 = 0 → no-op
	this.NotNil(card.Cardify)
	this.Len(full.Cardify, 1)
}

func (this *SuiteCommandFlow) TestGuestExit() {
	runtime := NewRuntime(0)
	runtime.Game.Score = Value{Value: 10}
	runtime.Game.Morale = Value{Value: 20}
	guest := &Guest{InstanceID: 11, SeatID: 2, Score: Value{Value: 3}, Morale: Value{Value: 5}}
	runtime.Seat[2] = guest
	eng := this.engine(runtime)

	commandGuestExit(eng, []InstanceID{11}, this.flag(true, true)) // 給滿意、扣士氣
	this.Equal(guest, runtime.Game.ExitLast)
	this.Equal(int32(2), runtime.Game.ExitLastSeat)
	this.Equal(int32(1), runtime.Game.ExitCount)
	this.Nil(runtime.Seat[2])                        // 自座位移除
	this.Equal(int32(13), runtime.Game.Score.Value)  // 10 + 3
	this.Equal(int32(15), runtime.Game.Morale.Value) // 20 - 5(無格擋 / 護盾)
	this.Equal(guest, runtime.Game.DamageGuest)      // morale 特例來源 = 離場顧客
	this.Equal(int32(5), runtime.Game.DamageValue)
}

func (this *SuiteCommandFlow) TestGuestExitNoScoreMorale() {
	runtime := NewRuntime(0)
	runtime.Game.Score = Value{Value: 10}
	runtime.Game.Morale = Value{Value: 20}
	guest := &Guest{InstanceID: 11, SeatID: 2, Score: Value{Value: 3}, Morale: Value{Value: 5}}
	runtime.Seat[2] = guest
	eng := this.engine(runtime)

	commandGuestExit(eng, []InstanceID{11}, this.flag(false, false)) // 不給滿意、不扣士氣
	this.Equal(int32(10), runtime.Game.Score.Value)
	this.Equal(int32(20), runtime.Game.Morale.Value)
	this.Nil(runtime.Seat[2]) // 仍移除

	commandGuestExit(eng, []InstanceID{99}, nil) // 非顧客實例 → no-op
}

func (this *SuiteCommandFlow) TestGuestReturn() {
	runtime := NewRuntime(0)
	guest := &Guest{InstanceID: 11, Sate: Value{Lock: 1}, SateSeal: Value{Lock: 1}, CalmSeal: Value{Lock: 1}}
	runtime.Roam = []*Guest{guest}
	eng := this.engine(runtime) // 座位全空 → seat 1

	commandGuestReturn(eng, []InstanceID{11}, nil)
	this.Empty(runtime.Roam)
	this.Equal(guest, runtime.Seat[1])
	this.Equal(int32(1), guest.SeatID)
	this.Equal(int32(0), guest.Sate.Lock) // 解入列自動鎖
	this.Equal(int32(0), guest.SateSeal.Lock)
	this.Equal(int32(0), guest.CalmSeal.Lock)
}

func (this *SuiteCommandFlow) TestGuestReturnNoop() {
	full := NewRuntime(0)
	full.Game.Morale = Value{Value: 20}
	guest := &Guest{InstanceID: 11, Morale: Value{Value: 5}}
	full.Roam = []*Guest{guest}
	full.Seat[1], full.Seat[2], full.Seat[3] = &Guest{}, &Guest{}, &Guest{} // 全占用
	engFull := this.engine(full)

	commandGuestReturn(engFull, []InstanceID{11}, nil) // 剩餘座位 = 0 → 走 guestExit(false, true)
	this.Empty(full.Roam)
	this.Equal(int32(1), full.Game.ExitCount)
	this.Equal(int32(15), full.Game.Morale.Value) // 扣士氣 20 - 5

	seated := NewRuntime(0)
	seated.Seat[1] = &Guest{InstanceID: 11, SeatID: 1}
	engSeated := this.engine(seated)
	commandGuestReturn(engSeated, []InstanceID{11}, nil) // 非遊蕩列表 → no-op
	this.NotNil(seated.Seat[1])
}

func (this *SuiteCommandFlow) TestGuestRoam() {
	runtime := NewRuntime(0)
	guest := &Guest{InstanceID: 11, SeatID: 2}
	runtime.Seat[2] = guest
	eng := this.engine(runtime)

	commandGuestRoam(eng, []InstanceID{11}, nil)
	this.Nil(runtime.Seat[2])
	this.Require().Len(runtime.Roam, 1)
	this.Equal(int32(0), guest.SeatID)
	this.Equal(int32(1), guest.Sate.Lock) // 自動鎖
	this.Equal(int32(1), guest.SateSeal.Lock)
	this.Equal(int32(1), guest.CalmSeal.Lock)

	commandGuestRoam(eng, []InstanceID{11}, nil) // 已在遊蕩(非座位)→ no-op
	this.Len(runtime.Roam, 1)
}

func (this *SuiteCommandFlow) TestGuestSeat() {
	runtime := NewRuntime(0)
	guest := &Guest{InstanceID: 11}
	runtime.Wait = []*Guest{guest}
	eng := this.engine(runtime) // 座位全空 → seat 1

	commandGuestSeat(eng, nil, nil)
	this.Empty(runtime.Wait)
	this.Equal(guest, runtime.Seat[1])
	this.Equal(int32(1), guest.SeatID)
	this.Equal(guest, runtime.Game.SeatLast)
	this.Equal(int32(1), runtime.Game.SeatCount)

	commandGuestSeat(eng, nil, nil) // 排隊佇列空 → no-op
	this.Equal(int32(1), runtime.Game.SeatCount)

	full := NewRuntime(0)
	full.Wait = []*Guest{{InstanceID: 12}}
	full.Seat[1], full.Seat[2], full.Seat[3] = &Guest{}, &Guest{}, &Guest{}
	engFull := this.engine(full)
	commandGuestSeat(engFull, nil, nil) // 無空座位 → no-op
	this.Len(full.Wait, 1)
}

func (this *SuiteCommandFlow) TestRemoveGuestContainer() {
	runtime := NewRuntime(0)
	g1 := &Guest{InstanceID: 1, SeatID: 1}
	g2 := &Guest{InstanceID: 2}
	g3 := &Guest{InstanceID: 3}
	g4 := &Guest{InstanceID: 4}
	keep := &Guest{InstanceID: 5}
	runtime.Seat[1] = g1
	runtime.Wait = []*Guest{g2, keep} // 兩位:移除 g2、保留 keep(removeGuest 保留分支)
	runtime.Roam = []*Guest{g3}
	runtime.Cardify = []*Guest{g4}
	eng := this.engine(runtime)

	removeGuestContainer(eng, ContainerSeat, g1)
	removeGuestContainer(eng, ContainerWait, g2)
	removeGuestContainer(eng, ContainerRoam, g3)
	removeGuestContainer(eng, ContainerCardify, g4)
	removeGuestContainer(eng, ContainerHand, g1) // 非顧客容器 → default no-op

	this.Nil(runtime.Seat[1])
	this.Equal([]*Guest{keep}, runtime.Wait) // g2 移除、keep 保留
	this.Empty(runtime.Roam)
	this.Empty(runtime.Cardify)
}

// === 測試輔助(置尾) ===

func (this *SuiteCommandFlow) engine(runtime *Runtime) *Engine {
	return NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{}, nil)
}

// flag 把兩個布林包成參數列表(供帶兩個布林旗標的流程命令:cardRun 消耗點數 / 進棄牌堆、guestExit 給滿意 / 扣士氣)。
func (this *SuiteCommandFlow) flag(a, b bool) []exprs.Value {
	return []exprs.Value{exprs.NewBool(a), exprs.NewBool(b)}
}
