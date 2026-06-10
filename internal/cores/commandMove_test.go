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
		put  func(game *Game, card *Card)
	}{
		{"deckToHand", commandDeckToHand, func(game *Game, card *Card) { game.Deck = []*Card{card} }},
		{"dropToHand", commandDropToHand, func(game *Game, card *Card) { game.Drop = []*Card{card} }},
		{"exileToHand", commandExileToHand, func(game *Game, card *Card) { game.Exile = []*Card{card} }},
	} {
		game := NewGame(0, nil, nil, nil, nil)
		card := &Card{instanceID: 1, cardID: 101}
		tc.put(game, card)
		injectPort(game)

		tc.verb(game, []InstanceID{1}, nil)

		this.Equal(CardList{card}, game.Hand, tc.name)            // 移入手牌頂端
		this.Equal(card, game.GetDrawLast(), tc.name)             // 最後抽出
		this.Equal(int32(1), game.GetDrawCount(), tc.name)        // 回合抽牌計數
		this.Equal(int32(1), game.GetDrawTotal().Get(1), tc.name) // 群組 1 累積
	} // for
}

func (this *SuiteCommandMove) TestMoveToDeck() {
	for _, tc := range []struct {
		name string
		verb commandFunc
		put  func(game *Game, card *Card)
	}{
		{"handToDeck", commandHandToDeck, func(game *Game, card *Card) { game.Hand = []*Card{card} }},
		{"dropToDeck", commandDropToDeck, func(game *Game, card *Card) { game.Drop = []*Card{card} }},
		{"exileToDeck", commandExileToDeck, func(game *Game, card *Card) { game.Exile = []*Card{card} }},
	} {
		game := NewGame(0, nil, nil, nil, nil)
		card := &Card{instanceID: 1, cardID: 101}
		tc.put(game, card)
		injectPort(game)

		tc.verb(game, []InstanceID{1}, nil) // 洗牌參數缺(nil)→ 不洗牌

		this.Equal(CardList{card}, game.Deck, tc.name) // 移入抽牌牌堆
		this.Nil(game.GetDrawLast(), tc.name)          // 進抽牌牌堆無事件
	} // for
}

func (this *SuiteCommandMove) TestMoveToDrop() {
	for _, tc := range []struct {
		name string
		verb commandFunc
		put  func(game *Game, card *Card)
	}{
		{"handToDrop", commandHandToDrop, func(game *Game, card *Card) { game.Hand = []*Card{card} }},
		{"deckToDrop", commandDeckToDrop, func(game *Game, card *Card) { game.Deck = []*Card{card} }},
		{"exileToDrop", commandExileToDrop, func(game *Game, card *Card) { game.Exile = []*Card{card} }},
	} {
		game := NewGame(0, nil, nil, nil, nil)
		card := &Card{instanceID: 1, cardID: 101}
		tc.put(game, card)
		injectPort(game)

		tc.verb(game, []InstanceID{1}, nil)

		this.Equal(CardList{card}, game.Drop, tc.name)
		this.Equal(card, game.GetDropLast(), tc.name)
		this.Equal(int32(1), game.GetDropTotal().Get(1), tc.name)
	} // for
}

func (this *SuiteCommandMove) TestMoveToExile() {
	for _, tc := range []struct {
		name string
		verb commandFunc
		put  func(game *Game, card *Card)
	}{
		{"handToExile", commandHandToExile, func(game *Game, card *Card) { game.Hand = []*Card{card} }},
		{"deckToExile", commandDeckToExile, func(game *Game, card *Card) { game.Deck = []*Card{card} }},
		{"dropToExile", commandDropToExile, func(game *Game, card *Card) { game.Drop = []*Card{card} }},
	} {
		game := NewGame(0, nil, nil, nil, nil)
		card := &Card{instanceID: 1, cardID: 101}
		tc.put(game, card)
		injectPort(game)

		tc.verb(game, []InstanceID{1}, nil)

		this.Equal(CardList{card}, game.Exile, tc.name)
		this.Equal(card, game.GetExileLast(), tc.name)
		this.Equal(int32(1), game.GetExileTotal().Get(1), tc.name)
	} // for
}

func (this *SuiteCommandMove) TestMovePositionMismatch() {
	game := NewGame(0, nil, nil, nil, nil)
	card := &Card{instanceID: 1, cardID: 101}
	game.Hand = []*Card{card} // 卡在手牌,但 deckToHand 期望源為抽牌牌堆
	injectPort(game)

	commandDeckToHand(game, []InstanceID{1}, nil)

	this.Equal(CardList{card}, game.Hand) // 位置不符 → no-op,仍在手牌
	this.Nil(game.GetDrawLast())
}

func (this *SuiteCommandMove) TestMoveNotFound() {
	game := NewGame(0, nil, nil, nil, nil)
	game.Deck = []*Card{{instanceID: 1, cardID: 101}}
	injectPort(game)

	commandDeckToHand(game, []InstanceID{99}, nil) // 編號不存在 → no-op

	this.Empty(game.Hand)
	this.Len(game.Deck, 1)
}

func (this *SuiteCommandMove) TestMoveShuffle() {
	game := NewGame(0, nil, nil, nil, nil)
	game.Deck = []*Card{{instanceID: 1, cardID: 101}, {instanceID: 2, cardID: 101}}
	injectPort(game)

	commandDeckToHand(game, []InstanceID{1}, []exprs.Value{exprs.NewBool(true)}) // 洗牌=true:移 id1 後洗抽牌牌堆

	this.Len(game.Hand, 1)
	this.Len(game.Deck, 1) // id2 留抽牌牌堆(恆等替身:張數保留)
}

func (this *SuiteCommandMove) TestMoveCardGroup() {
	game := NewGame(0, nil, nil, nil, nil)
	game.Hand = []*Card{{instanceID: 1, cardID: 101}, {instanceID: 2, cardID: 999}} // 101→群組 1、999→資料缺失→群組 0
	injectPort(game)

	commandHandToDrop(game, []InstanceID{1, 2}, nil)

	this.Equal(int32(1), game.GetDropTotal().Get(1)) // 群組 1
	this.Equal(int32(1), game.GetDropTotal().Get(0)) // 群組 0(靜態資料缺失)
	this.Equal(int32(2), game.GetDropCount())
}

func (this *SuiteCommandMove) TestMoveDefaults() {
	game := NewGame(0, nil, nil, nil, nil)
	card := &Card{instanceID: 1}
	injectPort(game)

	removeCard(game, ContainerNone, card) // 非四牌堆 → default no-op、不 panic
	placeCard(game, ContainerNone, card)  // 非四牌堆 → default no-op、不改任何牌堆

	this.Empty(game.Hand)
	this.Empty(game.Deck)
	this.Empty(game.Drop)
	this.Empty(game.Exile)
}

func (this *SuiteCommandMove) TestPlaceCardTrigger() {
	fired := map[TriggerKind]bool{}
	record := func(timing TriggerKind) EffectCommand { return func(*Game) { fired[timing] = true } }

	game := NewGame(0, nil, nil, nil, nil)
	game.Effect = EffectList{
		{instanceID: 1, effectID: 801, stack: 1},
		{instanceID: 2, effectID: 802, stack: 1},
		{instanceID: 3, effectID: 803, stack: 1},
	}
	injectPort(game)
	game.effectData = map[int32]effectData{
		801: {Kind: EffectTrigger, TriggerKind: TriggerCardDraw, Trigger: record(TriggerCardDraw)},
		802: {Kind: EffectTrigger, TriggerKind: TriggerCardDrop, Trigger: record(TriggerCardDrop)},
		803: {Kind: EffectTrigger, TriggerKind: TriggerCardExile, Trigger: record(TriggerCardExile)},
	}

	placeCard(game, ContainerHand, &Card{instanceID: 10, cardID: 101})  // 進手牌 → cardDraw
	placeCard(game, ContainerDrop, &Card{instanceID: 11, cardID: 101})  // 進棄牌牌堆 → cardDrop
	placeCard(game, ContainerExile, &Card{instanceID: 12, cardID: 101}) // 進流放牌堆 → cardExile

	this.True(fired[TriggerCardDraw])
	this.True(fired[TriggerCardDrop])
	this.True(fired[TriggerCardExile])
}

// === 測試輔助(置尾) ===

// engine 組裝測試引擎:注入 game、共用 buildSheet 靜態表、決定性 fake Operator / Rander。
