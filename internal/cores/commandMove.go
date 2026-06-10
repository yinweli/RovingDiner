package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 容器搬移命令（【營業規格書 | 二十五、操作命令清單】手牌 / 牌堆間移動）：
// 將命令對象中「位於來源容器」的卡牌移至目的容器（位置不符該項 no-op）；
// 進手牌 / 棄牌 / 流放各設 *Last / *Count / *Total（cardGroup 鍵）並觸發 cardDraw / cardDrop / cardExile；進抽牌牌堆無事件。
// 帶「洗牌」參數者（源或目的為抽牌牌堆）於操作後依參數洗抽牌牌堆。

func commandHandToDeck(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerHand, ContainerDeck, argBool(arg))
}

func commandHandToDrop(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerHand, ContainerDrop, false)
}

func commandHandToExile(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerHand, ContainerExile, false)
}

func commandDeckToDrop(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerDeck, ContainerDrop, argBool(arg))
}

func commandDeckToExile(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerDeck, ContainerExile, argBool(arg))
}

func commandDeckToHand(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerDeck, ContainerHand, argBool(arg))
}

func commandDropToDeck(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerDrop, ContainerDeck, argBool(arg))
}

func commandDropToExile(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerDrop, ContainerExile, false)
}

func commandDropToHand(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerDrop, ContainerHand, false)
}

func commandExileToDeck(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerExile, ContainerDeck, argBool(arg))
}

func commandExileToDrop(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerExile, ContainerDrop, false)
}

func commandExileToHand(eng *Engine, target []InstanceID, arg []exprs.Value) {
	moveCards(eng, target, ContainerExile, ContainerHand, false)
}

// moveCards 把身分集中位於 source 容器的卡牌移至 dest；位置不符 / 不存在該項 no-op；shuffle 為真時操作後洗抽牌牌堆。
func moveCards(eng *Engine, target []InstanceID, source, dest ContainerKind, shuffle bool) {
	for _, itor := range target {
		card, where, ok := eng.locateCard(itor)

		if ok == false || where != source {
			continue // 不存在 / 位置不符 → 該項 no-op
		} // if

		removeCard(eng, source, card)
		placeCard(eng, dest, card)
	} // for

	if shuffle {
		shuffleCard(eng, eng.runtime.Deck)
	} // if
}

// removeCard 自 source 牌堆移除指定卡牌（以實例編號比對）。
func removeCard(eng *Engine, source ContainerKind, card *Card) {
	switch source {
	case ContainerHand:
		eng.runtime.Hand.Remove(card.GetInstanceID())

	case ContainerDeck:
		eng.runtime.Deck.Remove(card.GetInstanceID())

	case ContainerDrop:
		eng.runtime.Drop.Remove(card.GetInstanceID())

	case ContainerExile:
		eng.runtime.Exile.Remove(card.GetInstanceID())

	default:
		// 不可達：removeCard 僅以四牌堆 source 呼叫
	} // switch
}

// placeCard 把卡牌加入 dest 牌堆頂端（前端），並依目的設事件與 system 觸發（進抽牌牌堆無事件）。供搬移與實例化（M9.3）共用。
func placeCard(eng *Engine, dest ContainerKind, card *Card) {
	switch dest {
	case ContainerHand:
		eng.runtime.Hand.Push(card)
		eng.runtime.Game.EventDraw(card, cardGroup(eng, card))
		fireTrigger(eng, TriggerCardDraw) // 卡牌進手牌觸發

	case ContainerDeck:
		eng.runtime.Deck.Push(card)

	case ContainerDrop:
		eng.runtime.Drop.Push(card)
		eng.runtime.Game.EventDrop(card, cardGroup(eng, card))
		fireTrigger(eng, TriggerCardDrop) // 卡牌進棄牌牌堆觸發

	case ContainerExile:
		eng.runtime.Exile.Push(card)
		eng.runtime.Game.EventExile(card, cardGroup(eng, card))
		fireTrigger(eng, TriggerCardExile) // 卡牌進流放牌堆觸發

	default:
		// 不可達：placeCard 僅以四牌堆 dest 呼叫
	} // switch
}

// cardGroup 取卡牌的群組編號（供 *Total 多重集合鍵）；靜態資料缺失回 0。
func cardGroup(eng *Engine, card *Card) int32 {
	meta := eng.data.Card.Get(card.GetCardID())

	if meta == nil {
		return 0
	} // if

	return meta.Group
}
