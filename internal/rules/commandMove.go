package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 容器搬移命令（【營業規格書 | 二十五、操作命令清單】手牌 / 牌堆間移動）：
// 將命令對象中「位於來源容器」的卡牌移至目的容器（位置不符該項 no-op）；
// 進手牌 / 棄牌 / 流放各設 *Last / *Count / *Total（cardGroup 鍵）並觸發 cardDraw / cardDrop / cardExile；進抽牌牌堆無事件。
// 帶「洗牌」參數者（源或目的為抽牌牌堆）於操作後依參數洗抽牌牌堆。

func commandHandToDeck(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerHand, cores.ContainerDeck, argBool(arg))
}

func commandHandToDrop(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerHand, cores.ContainerDrop, false)
}

func commandHandToExile(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerHand, cores.ContainerExile, false)
}

func commandDeckToDrop(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerDeck, cores.ContainerDrop, argBool(arg))
}

func commandDeckToExile(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerDeck, cores.ContainerExile, argBool(arg))
}

func commandDeckToHand(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerDeck, cores.ContainerHand, argBool(arg))
}

func commandDropToDeck(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerDrop, cores.ContainerDeck, argBool(arg))
}

func commandDropToExile(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerDrop, cores.ContainerExile, false)
}

func commandDropToHand(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerDrop, cores.ContainerHand, false)
}

func commandExileToDeck(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerExile, cores.ContainerDeck, argBool(arg))
}

func commandExileToDrop(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerExile, cores.ContainerDrop, false)
}

func commandExileToHand(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	moveCards(game, target, cores.ContainerExile, cores.ContainerHand, false)
}

// moveCards 把身分集中位於 source 容器的卡牌移至 dest；位置不符 / 不存在該項 no-op；shuffle 為真時操作後洗抽牌牌堆。
func moveCards(game *cores.Game, target []cores.InstanceID, source, dest cores.ContainerKind, shuffle bool) {
	for _, itor := range target {
		card, where, ok := game.LocateCard(itor)

		if ok == false || where != source {
			continue // 不存在 / 位置不符 → 該項 no-op
		} // if

		removeCard(game, source, card)
		placeCard(game, source, dest, card)
	} // for

	if shuffle {
		shuffleCard(game, game.Deck)
	} // if
}
