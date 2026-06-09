package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteCommandMove(t *testing.T) {
	suite.Run(t, new(SuiteCommandMove))
}

// SuiteCommandMove 驗證容器搬移命令(commandMove.go):12 個 *To* 移動、位置不符 / 不存在 no-op、洗牌參數、事件(*Last / Count / Total)。
type SuiteCommandMove struct {
	suite.Suite
}

func (this *SuiteCommandMove) TestMoveToHand() {
	for _, tc := range []struct {
		name string
		verb commandFunc
		put  func(runtime *Runtime, card *Card)
	}{
		{"deckToHand", commandDeckToHand, func(runtime *Runtime, card *Card) { runtime.Deck = []*Card{card} }},
		{"dropToHand", commandDropToHand, func(runtime *Runtime, card *Card) { runtime.Drop = []*Card{card} }},
		{"exileToHand", commandExileToHand, func(runtime *Runtime, card *Card) { runtime.Exile = []*Card{card} }},
	} {
		runtime := NewRuntime(0)
		card := &Card{InstanceID: 1, CardID: 101}
		tc.put(runtime, card)
		eng := this.engine(runtime)

		tc.verb(eng, []InstanceID{1}, nil)

		this.Equal([]*Card{card}, runtime.Hand, tc.name)         // 移入手牌頂端
		this.Equal(card, runtime.Game.DrawLast, tc.name)         // 最後抽出
		this.Equal(int32(1), runtime.Game.DrawCount, tc.name)    // 回合抽牌計數
		this.Equal(int32(1), runtime.Game.DrawTotal[1], tc.name) // 群組 1 累積
	} // for
}

func (this *SuiteCommandMove) TestMoveToDeck() {
	for _, tc := range []struct {
		name string
		verb commandFunc
		put  func(runtime *Runtime, card *Card)
	}{
		{"handToDeck", commandHandToDeck, func(runtime *Runtime, card *Card) { runtime.Hand = []*Card{card} }},
		{"dropToDeck", commandDropToDeck, func(runtime *Runtime, card *Card) { runtime.Drop = []*Card{card} }},
		{"exileToDeck", commandExileToDeck, func(runtime *Runtime, card *Card) { runtime.Exile = []*Card{card} }},
	} {
		runtime := NewRuntime(0)
		card := &Card{InstanceID: 1, CardID: 101}
		tc.put(runtime, card)
		eng := this.engine(runtime)

		tc.verb(eng, []InstanceID{1}, nil) // 洗牌參數缺(nil)→ 不洗牌

		this.Equal([]*Card{card}, runtime.Deck, tc.name) // 移入抽牌牌堆
		this.Nil(runtime.Game.DrawLast, tc.name)         // 進抽牌牌堆無事件
	} // for
}

func (this *SuiteCommandMove) TestMoveToDrop() {
	for _, tc := range []struct {
		name string
		verb commandFunc
		put  func(runtime *Runtime, card *Card)
	}{
		{"handToDrop", commandHandToDrop, func(runtime *Runtime, card *Card) { runtime.Hand = []*Card{card} }},
		{"deckToDrop", commandDeckToDrop, func(runtime *Runtime, card *Card) { runtime.Deck = []*Card{card} }},
		{"exileToDrop", commandExileToDrop, func(runtime *Runtime, card *Card) { runtime.Exile = []*Card{card} }},
	} {
		runtime := NewRuntime(0)
		card := &Card{InstanceID: 1, CardID: 101}
		tc.put(runtime, card)
		eng := this.engine(runtime)

		tc.verb(eng, []InstanceID{1}, nil)

		this.Equal([]*Card{card}, runtime.Drop, tc.name)
		this.Equal(card, runtime.Game.DropLast, tc.name)
		this.Equal(int32(1), runtime.Game.DropTotal[1], tc.name)
	} // for
}

func (this *SuiteCommandMove) TestMoveToExile() {
	for _, tc := range []struct {
		name string
		verb commandFunc
		put  func(runtime *Runtime, card *Card)
	}{
		{"handToExile", commandHandToExile, func(runtime *Runtime, card *Card) { runtime.Hand = []*Card{card} }},
		{"deckToExile", commandDeckToExile, func(runtime *Runtime, card *Card) { runtime.Deck = []*Card{card} }},
		{"dropToExile", commandDropToExile, func(runtime *Runtime, card *Card) { runtime.Drop = []*Card{card} }},
	} {
		runtime := NewRuntime(0)
		card := &Card{InstanceID: 1, CardID: 101}
		tc.put(runtime, card)
		eng := this.engine(runtime)

		tc.verb(eng, []InstanceID{1}, nil)

		this.Equal([]*Card{card}, runtime.Exile, tc.name)
		this.Equal(card, runtime.Game.ExileLast, tc.name)
		this.Equal(int32(1), runtime.Game.ExileTotal[1], tc.name)
	} // for
}

func (this *SuiteCommandMove) TestMovePositionMismatch() {
	runtime := NewRuntime(0)
	card := &Card{InstanceID: 1, CardID: 101}
	runtime.Hand = []*Card{card} // 卡在手牌,但 deckToHand 期望源為抽牌牌堆
	eng := this.engine(runtime)

	commandDeckToHand(eng, []InstanceID{1}, nil)

	this.Equal([]*Card{card}, runtime.Hand) // 位置不符 → no-op,仍在手牌
	this.Nil(runtime.Game.DrawLast)
}

func (this *SuiteCommandMove) TestMoveNotFound() {
	runtime := NewRuntime(0)
	runtime.Deck = []*Card{{InstanceID: 1, CardID: 101}}
	eng := this.engine(runtime)

	commandDeckToHand(eng, []InstanceID{99}, nil) // 編號不存在 → no-op

	this.Empty(runtime.Hand)
	this.Len(runtime.Deck, 1)
}

func (this *SuiteCommandMove) TestMoveShuffle() {
	runtime := NewRuntime(0)
	runtime.Deck = []*Card{{InstanceID: 1, CardID: 101}, {InstanceID: 2, CardID: 101}}
	eng := this.engine(runtime)

	commandDeckToHand(eng, []InstanceID{1}, []exprs.Value{exprs.NewBool(true)}) // 洗牌=true:移 id1 後洗抽牌牌堆

	this.Len(runtime.Hand, 1)
	this.Len(runtime.Deck, 1) // id2 留抽牌牌堆(恆等替身:張數保留)
}

func (this *SuiteCommandMove) TestMoveCardGroup() {
	runtime := NewRuntime(0)
	runtime.Hand = []*Card{{InstanceID: 1, CardID: 101}, {InstanceID: 2, CardID: 999}} // 101→群組 1、999→資料缺失→群組 0
	eng := this.engine(runtime)

	commandHandToDrop(eng, []InstanceID{1, 2}, nil)

	this.Equal(int32(1), runtime.Game.DropTotal[1]) // 群組 1
	this.Equal(int32(1), runtime.Game.DropTotal[0]) // 群組 0(靜態資料缺失)
	this.Equal(int32(2), runtime.Game.DropCount)
}

func (this *SuiteCommandMove) TestMoveDefaults() {
	runtime := NewRuntime(0)
	card := &Card{InstanceID: 1}
	eng := this.engine(runtime)

	removeCard(eng, ContainerNone, card) // 非四牌堆 → default no-op、不 panic
	placeCard(eng, ContainerNone, card)  // 非四牌堆 → default no-op、不改任何牌堆

	this.Empty(runtime.Hand)
	this.Empty(runtime.Deck)
	this.Empty(runtime.Drop)
	this.Empty(runtime.Exile)
}

func (this *SuiteCommandMove) TestPlaceCardTrigger() {
	fired := map[TriggerKind]bool{}
	record := func(timing TriggerKind) EffectCommand { return func(*Engine) { fired[timing] = true } }

	runtime := NewRuntime(0)
	runtime.Effect = []*Effect{
		{InstanceID: 1, EffectID: 801, Stack: 1},
		{InstanceID: 2, EffectID: 802, Stack: 1},
		{InstanceID: 3, EffectID: 803, Stack: 1},
	}
	eng := this.engine(runtime)
	eng.effect = map[int32]effectData{
		801: {Kind: EffectTrigger, TriggerKind: TriggerCardDraw, Trigger: record(TriggerCardDraw)},
		802: {Kind: EffectTrigger, TriggerKind: TriggerCardDrop, Trigger: record(TriggerCardDrop)},
		803: {Kind: EffectTrigger, TriggerKind: TriggerCardExile, Trigger: record(TriggerCardExile)},
	}

	placeCard(eng, ContainerHand, &Card{InstanceID: 10, CardID: 101})  // 進手牌 → cardDraw
	placeCard(eng, ContainerDrop, &Card{InstanceID: 11, CardID: 101})  // 進棄牌牌堆 → cardDrop
	placeCard(eng, ContainerExile, &Card{InstanceID: 12, CardID: 101}) // 進流放牌堆 → cardExile

	this.True(fired[TriggerCardDraw])
	this.True(fired[TriggerCardDrop])
	this.True(fired[TriggerCardExile])
}

// === 測試輔助(置尾) ===

// engine 組裝測試引擎:注入 runtime、共用 buildSheet 靜態表、決定性 fake Operator / Rander。
func (this *SuiteCommandMove) engine(runtime *Runtime) *Engine {
	return NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{}, nil)
}
