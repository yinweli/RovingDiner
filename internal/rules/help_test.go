package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteHelp(t *testing.T) {
	suite.Run(t, new(SuiteHelp))
}

// SuiteHelp 驗證 help.go 的無狀態輔助:參數取值 / 分組計數與比較 / 引用鎖定取值。
type SuiteHelp struct {
	suite.Suite
}

func (this *SuiteHelp) TestOneInt() {
	num, ok := oneInt([]exprs.Value{exprs.NewNum(5)})
	this.True(ok)
	this.Equal(int32(5), num)

	num, ok = oneInt([]exprs.Value{exprs.NewNum(3.9)}) // 浮點截斷為整數
	this.True(ok)
	this.Equal(int32(3), num)

	_, ok = oneInt(nil) // 數量不符(0 個)
	this.False(ok)

	_, ok = oneInt([]exprs.Value{exprs.NewNum(1), exprs.NewNum(2)}) // 數量不符(2 個)
	this.False(ok)

	_, ok = oneInt([]exprs.Value{exprs.NewText("x")}) // 型別不符
	this.False(ok)
}

func (this *SuiteHelp) TestTwoInt() {
	a, b, ok := twoInt([]exprs.Value{exprs.NewNum(2), exprs.NewNum(5)})
	this.True(ok)
	this.Equal(int32(2), a)
	this.Equal(int32(5), b)

	_, _, ok = twoInt([]exprs.Value{exprs.NewNum(1)}) // 數量不符(1 個)
	this.False(ok)

	_, _, ok = twoInt([]exprs.Value{exprs.NewText("x"), exprs.NewNum(1)}) // 第一參型別不符
	this.False(ok)

	_, _, ok = twoInt([]exprs.Value{exprs.NewNum(1), exprs.NewText("x")}) // 第二參型別不符
	this.False(ok)
}

func (this *SuiteHelp) TestNumArg() {
	num, ok := numArg([]exprs.Value{exprs.NewNum(1), exprs.NewNum(2.5)})
	this.True(ok)
	this.Equal([]float64{1, 2.5}, num)

	_, ok = numArg([]exprs.Value{exprs.NewNum(1)}) // 數量不符(少於 2 個)
	this.False(ok)

	_, ok = numArg([]exprs.Value{exprs.NewNum(1), exprs.NewText("x")}) // 型別不符
	this.False(ok)
}

func (this *SuiteHelp) TestArgInt() {
	n, ok := argInt([]exprs.Value{exprs.NewNum(3.9)}) // 浮點截斷
	this.True(ok)
	this.Equal(int32(3), n)

	_, ok = argInt(nil) // 缺漏
	this.False(ok)

	_, ok = argInt([]exprs.Value{exprs.NewText("x")}) // 非數值
	this.False(ok)
}

func (this *SuiteHelp) TestArgNum() {
	n, ok := argNum([]exprs.Value{exprs.NewNum(2.5)})
	this.True(ok)
	this.Equal(float64(2.5), n)

	_, ok = argNum(nil) // 缺漏
	this.False(ok)

	_, ok = argNum([]exprs.Value{exprs.NewBool(true)}) // 非數值
	this.False(ok)
}

func (this *SuiteHelp) TestArgBool() {
	this.True(argBool([]exprs.Value{exprs.NewBool(true)}))
	this.False(argBool([]exprs.Value{exprs.NewBool(false)}))
	this.False(argBool(nil))                            // 缺漏
	this.False(argBool([]exprs.Value{exprs.NewNum(1)})) // 非布林
	// 讀第 k 個參數由呼叫端傳 arg[k:]:此處讀 index 1(copy / clone 的洗牌位於 index 1)
	this.True(argBool([]exprs.Value{exprs.NewNum(1), exprs.NewBool(true)}[1:]))
}

func (this *SuiteHelp) TestIntList() {
	this.Equal([]int32{1, 2}, intList([]exprs.Value{exprs.NewNum(1), exprs.NewText("x"), exprs.NewNum(2)})) // 略過非數值
	this.Nil(intList(nil))
}

func (this *SuiteHelp) TestArgTail() {
	full := []exprs.Value{exprs.NewNum(1), exprs.NewNum(2)}
	this.Len(argTail(full, 1), 1) // 自 index 1
	this.Nil(argTail(full, 2))    // from == len → nil
	this.Nil(argTail(full, 5))    // from > len → nil
}

func (this *SuiteHelp) TestGroupSize() {
	game := cores.NewGame(0, 0, tester.BuildData(), nil, nil)
	card := cores.CardList{cores.NewCard(game, 101), cores.NewCard(game, 101), cores.NewCard(game, 102)}

	result, ok := groupSize(card, game.GetSheet(), []exprs.Value{exprs.NewNum(0)}) // N==0 全量
	this.True(ok)
	this.Equal(float64(3), result.Num())

	result, ok = groupSize(card, game.GetSheet(), []exprs.Value{exprs.NewNum(1)}) // 群組 1 = 卡 101 兩張
	this.True(ok)
	this.Equal(float64(2), result.Num())

	_, ok = groupSize(card, game.GetSheet(), nil) // 參數不符 → 失敗
	this.False(ok)
}

func (this *SuiteHelp) TestGroupTotal() {
	total := cores.NewTally()
	total.Add(1)
	total.Add(1)
	total.Add(1)
	total.Add(2)
	total.Add(2)

	result, ok := groupTotal(&total, []exprs.Value{exprs.NewNum(0)}) // N==0 全加總
	this.True(ok)
	this.Equal(float64(5), result.Num())

	result, ok = groupTotal(&total, []exprs.Value{exprs.NewNum(1)})
	this.True(ok)
	this.Equal(float64(3), result.Num())

	_, ok = groupTotal(&total, nil) // 參數不符 → 失敗
	this.False(ok)
}

func (this *SuiteHelp) TestCompareOp() {
	for _, itor := range []struct {
		op   string
		a, b int32
		want bool
	}{
		{"<", 1, 2, true}, {"<", 2, 1, false},
		{">", 2, 1, true}, {">", 1, 2, false},
		{"<=", 2, 2, true}, {"<=", 3, 2, false},
		{">=", 2, 2, true}, {">=", 1, 2, false},
		{"==", 2, 2, true}, {"==", 1, 2, false},
		{"!=", 1, 2, true}, {"!=", 2, 2, false},
	} {
		result, ok := compareOp(itor.op, itor.a, itor.b)
		this.True(ok, itor.op)
		this.Equal(itor.want, result, itor.op)
	} // for

	_, ok := compareOp("~=", 1, 1) // 未知運算符 → 失敗
	this.False(ok)
}

func (this *SuiteHelp) TestCardLock() {
	game := cores.NewGame(0, 0, tester.BuildData(), nil, nil)
	card := cores.NewCard(game, 101)
	card.GetCost().Lock()
	pick := func(c *cores.Card) int32 { return c.GetCost().GetLock() }

	result, ok := cardLock(cores.NewRefCard(card), pick)
	this.True(ok)
	this.Equal(float64(1), result.Num())

	_, ok = cardLock(cores.NewRefGuest(cores.NewGuest(game, 501)), pick) // 非卡牌引用 → 失敗
	this.False(ok)
}

func (this *SuiteHelp) TestGuestLock() {
	game := cores.NewGame(0, 0, tester.BuildData(), nil, nil)
	guest := cores.NewGuest(game, 501)
	guest.GetCalm().Lock()
	pick := func(g *cores.Guest) int32 { return g.GetCalm().GetLock() }

	result, ok := guestLock(cores.NewRefGuest(guest), pick)
	this.True(ok)
	this.Equal(float64(1), result.Num())

	_, ok = guestLock(cores.NewRefCard(cores.NewCard(game, 101)), pick) // 非顧客引用 → 失敗
	this.False(ok)
}

func (this *SuiteHelp) TestCardGroup() {
	game := newGame()
	this.Equal(int32(1), cardGroup(game, cores.NewCard(game, 101))) // 卡 101 → 群組 1
	this.Equal(int32(0), cardGroup(game, strayCard(999)))           // 靜態資料缺失 → 0
}

func (this *SuiteHelp) TestRemoveCard() {
	game := newGame()
	card := cores.NewCard(game, 101)
	game.Hand.Push(card)

	removeCard(game, cores.ContainerNone, card) // 非四牌堆 → default no-op、不 panic
	this.Len(game.Hand, 1)

	removeCard(game, cores.ContainerHand, card)
	this.Empty(game.Hand)
}

func (this *SuiteHelp) TestPlaceCard() {
	game := newGame()
	card := cores.NewCard(game, 101)

	placeCard(game, cores.ContainerNone, card) // 非四牌堆 → default no-op、不改任何牌堆
	this.Empty(game.Hand)
	this.Empty(game.Deck)
	this.Empty(game.Drop)
	this.Empty(game.Exile)

	placeCard(game, cores.ContainerDeck, card) // 進抽牌牌堆:無事件
	this.Equal(cores.CardList{card}, game.Deck)
	this.Nil(game.GetDrawLast())
}

func (this *SuiteHelp) TestPlaceCardTrigger() {
	fired := map[cores.TriggerKind]bool{}
	record := func(timing cores.TriggerKind) cores.EffectExec {
		return func(*cores.Game) { fired[timing] = true }
	}

	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardDraw, Trigger: record(cores.TriggerCardDraw)})
	data.SetEffect(802, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardDrop, Trigger: record(cores.TriggerCardDrop)})
	data.SetEffect(803, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardExile, Trigger: record(cores.TriggerCardExile)})
	game := newGameData(data)

	for _, itor := range []int32{801, 802, 803} {
		game.Effect.Push(cores.NewEffect(game, itor, cores.Ref{}, 1))
	} // for

	placeCard(game, cores.ContainerHand, cores.NewCard(game, 101))  // 進手牌 → cardDraw
	placeCard(game, cores.ContainerDrop, cores.NewCard(game, 101))  // 進棄牌牌堆 → cardDrop
	placeCard(game, cores.ContainerExile, cores.NewCard(game, 101)) // 進流放牌堆 → cardExile

	this.True(fired[cores.TriggerCardDraw])
	this.True(fired[cores.TriggerCardDrop])
	this.True(fired[cores.TriggerCardExile])
}

func (this *SuiteHelp) TestDamageSource() {
	guest := &cores.Guest{}
	guestRef := cores.NewRefGuest(guest)
	cardRef := cores.NewRefCard(&cores.Card{})

	this.Same(guest, damageSource(&guestRef)) // 顧客 self → 該顧客
	this.Nil(damageSource(&cardRef))          // 卡牌 self → 空物件
	this.Nil(damageSource(nil))               // self 未綁定 → 空物件
}
