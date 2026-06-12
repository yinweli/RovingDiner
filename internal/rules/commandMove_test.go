package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteCommandMove(t *testing.T) {
	suite.Run(t, new(SuiteCommandMove))
}

// SuiteCommandMove 驗證容器搬移命令(commandMove.go): 12 個 *To* 移動、位置不符 / 不存在 no-op、洗牌參數、事件(*Last / Count / Total)。
type SuiteCommandMove struct {
	suite.Suite
}

func (this *SuiteCommandMove) TestMoveToHand() {
	for _, tc := range []struct {
		name string
		verb cores.CommandFunc
		put  func(game *cores.Game, card *cores.Card)
	}{
		{"deckToHand", commandDeckToHand, func(game *cores.Game, card *cores.Card) { game.Deck.Push(card) }},
		{"dropToHand", commandDropToHand, func(game *cores.Game, card *cores.Card) { game.Drop.Push(card) }},
		{"exileToHand", commandExileToHand, func(game *cores.Game, card *cores.Card) { game.Exile.Push(card) }},
	} {
		game := newGame()
		card := cores.NewCard(game, 101)
		tc.put(game, card)

		tc.verb(game, []cores.InstanceID{card.GetInstanceID()}, nil)

		this.Equal(cores.CardList{card}, game.Hand, tc.name)      // 移入手牌頂端
		this.Equal(card, game.GetDrawLast(), tc.name)             // 最後抽出
		this.Equal(int32(1), game.GetDrawCount(), tc.name)        // 回合抽牌計數
		this.Equal(int32(1), game.GetDrawTotal().Get(1), tc.name) // 群組 1 累積
	} // for
}

func (this *SuiteCommandMove) TestMoveToDeck() {
	for _, tc := range []struct {
		name string
		verb cores.CommandFunc
		put  func(game *cores.Game, card *cores.Card)
	}{
		{"handToDeck", commandHandToDeck, func(game *cores.Game, card *cores.Card) { game.Hand.Push(card) }},
		{"dropToDeck", commandDropToDeck, func(game *cores.Game, card *cores.Card) { game.Drop.Push(card) }},
		{"exileToDeck", commandExileToDeck, func(game *cores.Game, card *cores.Card) { game.Exile.Push(card) }},
	} {
		game := newGame()
		card := cores.NewCard(game, 101)
		tc.put(game, card)

		tc.verb(game, []cores.InstanceID{card.GetInstanceID()}, nil) // 洗牌參數缺(nil)→ 不洗牌

		this.Equal(cores.CardList{card}, game.Deck, tc.name) // 移入抽牌牌堆
		this.Nil(game.GetDrawLast(), tc.name)                // 進抽牌牌堆無事件
	} // for
}

func (this *SuiteCommandMove) TestMoveToDrop() {
	for _, tc := range []struct {
		name string
		verb cores.CommandFunc
		put  func(game *cores.Game, card *cores.Card)
	}{
		{"handToDrop", commandHandToDrop, func(game *cores.Game, card *cores.Card) { game.Hand.Push(card) }},
		{"deckToDrop", commandDeckToDrop, func(game *cores.Game, card *cores.Card) { game.Deck.Push(card) }},
		{"exileToDrop", commandExileToDrop, func(game *cores.Game, card *cores.Card) { game.Exile.Push(card) }},
	} {
		game := newGame()
		card := cores.NewCard(game, 101)
		tc.put(game, card)

		tc.verb(game, []cores.InstanceID{card.GetInstanceID()}, nil)

		this.Equal(cores.CardList{card}, game.Drop, tc.name)
		this.Equal(card, game.GetDropLast(), tc.name)
		this.Equal(int32(1), game.GetDropTotal().Get(1), tc.name)
	} // for
}

func (this *SuiteCommandMove) TestMoveToExile() {
	for _, tc := range []struct {
		name string
		verb cores.CommandFunc
		put  func(game *cores.Game, card *cores.Card)
	}{
		{"handToExile", commandHandToExile, func(game *cores.Game, card *cores.Card) { game.Hand.Push(card) }},
		{"deckToExile", commandDeckToExile, func(game *cores.Game, card *cores.Card) { game.Deck.Push(card) }},
		{"dropToExile", commandDropToExile, func(game *cores.Game, card *cores.Card) { game.Drop.Push(card) }},
	} {
		game := newGame()
		card := cores.NewCard(game, 101)
		tc.put(game, card)

		tc.verb(game, []cores.InstanceID{card.GetInstanceID()}, nil)

		this.Equal(cores.CardList{card}, game.Exile, tc.name)
		this.Equal(card, game.GetExileLast(), tc.name)
		this.Equal(int32(1), game.GetExileTotal().Get(1), tc.name)
	} // for
}

func (this *SuiteCommandMove) TestMovePositionMismatch() {
	game := newGame()
	card := cores.NewCard(game, 101)
	game.Hand.Push(card) // 卡在手牌, 但 deckToHand 期望源為抽牌牌堆

	commandDeckToHand(game, []cores.InstanceID{card.GetInstanceID()}, nil)

	this.Equal(cores.CardList{card}, game.Hand) // 位置不符 → no-op, 仍在手牌
	this.Nil(game.GetDrawLast())
}

func (this *SuiteCommandMove) TestMoveNotFound() {
	game := newGame()
	game.Deck.Push(cores.NewCard(game, 101))

	commandDeckToHand(game, []cores.InstanceID{99}, nil) // 編號不存在 → no-op

	this.Empty(game.Hand)
	this.Len(game.Deck, 1)
}

func (this *SuiteCommandMove) TestMoveShuffle() {
	game := newGame()
	move := cores.NewCard(game, 101)
	game.Deck.Push(move)
	game.Deck.Push(cores.NewCard(game, 101))

	commandDeckToHand(game, []cores.InstanceID{move.GetInstanceID()}, []exprs.Value{exprs.NewBool(true)}) // 洗牌=true: 移後洗抽牌牌堆

	this.Len(game.Hand, 1)
	this.Len(game.Deck, 1) // 另一張留抽牌牌堆(恆等替身: 張數保留)
}

func (this *SuiteCommandMove) TestMoveCardGroup() {
	game := newGame()
	known := cores.NewCard(game, 101) // 101 → 群組 1
	stray := strayCard(999)           // 999 → 資料缺失 → 群組 0
	game.Hand = cores.CardList{known, stray}

	commandHandToDrop(game, []cores.InstanceID{known.GetInstanceID(), stray.GetInstanceID()}, nil)

	this.Equal(int32(1), game.GetDropTotal().Get(1)) // 群組 1
	this.Equal(int32(1), game.GetDropTotal().Get(0)) // 群組 0(靜態資料缺失)
	this.Equal(int32(2), game.GetDropCount())
}
