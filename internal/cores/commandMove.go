package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 容器搬移命令（【營業規格書 | 二十五、操作命令清單】手牌 / 牌堆間移動）：
// 將命令對象中「位於來源容器」的卡牌移至目的容器（位置不符該項 no-op）；
// 進手牌 / 棄牌 / 流放各設 *Last / *Count / *Total（cardGroup 鍵）並觸發 cardDraw / cardDrop / cardExile；進抽牌牌堆無事件。
// 帶「洗牌」參數者（源或目的為抽牌牌堆）於操作後依參數洗抽牌牌堆。

func commandHandToDeck(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerHand, ContainerDeck, argBool(arg))
}

func commandHandToDrop(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerHand, ContainerDrop, false)
}

func commandHandToExile(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerHand, ContainerExile, false)
}

func commandDeckToDrop(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerDeck, ContainerDrop, argBool(arg))
}

func commandDeckToExile(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerDeck, ContainerExile, argBool(arg))
}

func commandDeckToHand(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerDeck, ContainerHand, argBool(arg))
}

func commandDropToDeck(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerDrop, ContainerDeck, argBool(arg))
}

func commandDropToExile(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerDrop, ContainerExile, false)
}

func commandDropToHand(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerDrop, ContainerHand, false)
}

func commandExileToDeck(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerExile, ContainerDeck, argBool(arg))
}

func commandExileToDrop(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerExile, ContainerDrop, false)
}

func commandExileToHand(game *Game, target []InstanceID, arg []exprs.Value) {
	moveCards(game, target, ContainerExile, ContainerHand, false)
}

// moveCards 把身分集中位於 source 容器的卡牌移至 dest；位置不符 / 不存在該項 no-op；shuffle 為真時操作後洗抽牌牌堆。
func moveCards(game *Game, target []InstanceID, source, dest ContainerKind, shuffle bool) {
	for _, itor := range target {
		card, where, ok := game.locateCard(itor)

		if ok == false || where != source {
			continue // 不存在 / 位置不符 → 該項 no-op
		} // if

		removeCard(game, source, card)
		placeCard(game, dest, card)
	} // for

	if shuffle {
		shuffleCard(game, game.Deck)
	} // if
}

// removeCard 自 source 牌堆移除指定卡牌（以實例編號比對）。
func removeCard(game *Game, source ContainerKind, card *Card) {
	switch source {
	case ContainerHand:
		game.Hand.Remove(card.GetInstanceID())

	case ContainerDeck:
		game.Deck.Remove(card.GetInstanceID())

	case ContainerDrop:
		game.Drop.Remove(card.GetInstanceID())

	case ContainerExile:
		game.Exile.Remove(card.GetInstanceID())

	default:
		// 不可達：removeCard 僅以四牌堆 source 呼叫
	} // switch
}

// placeCard 把卡牌加入 dest 牌堆頂端（前端），並依目的設事件與 system 觸發（進抽牌牌堆無事件）。供搬移與實例化（M9.3）共用。
func placeCard(game *Game, dest ContainerKind, card *Card) {
	switch dest {
	case ContainerHand:
		game.Hand.Push(card)
		game.EventDraw(card, cardGroup(game, card))
		fireTrigger(game, TriggerCardDraw) // 卡牌進手牌觸發

	case ContainerDeck:
		game.Deck.Push(card)

	case ContainerDrop:
		game.Drop.Push(card)
		game.EventDrop(card, cardGroup(game, card))
		fireTrigger(game, TriggerCardDrop) // 卡牌進棄牌牌堆觸發

	case ContainerExile:
		game.Exile.Push(card)
		game.EventExile(card, cardGroup(game, card))
		fireTrigger(game, TriggerCardExile) // 卡牌進流放牌堆觸發

	default:
		// 不可達：placeCard 僅以四牌堆 dest 呼叫
	} // switch
}

// cardGroup 取卡牌的群組編號（供 *Total 多重集合鍵）；靜態資料缺失回 0。
func cardGroup(game *Game, card *Card) int32 {
	meta := game.data.Card.Get(card.GetCardID())

	if meta == nil {
		return 0
	} // if

	return meta.Group
}
