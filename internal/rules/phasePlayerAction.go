package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// phasePlayerAction 玩家行動階段(【營業規格書 | 十九、核心流程 | 3. 玩家行動階段】): 補牌至補牌張數上限(抽牌牌堆空則棄牌牌堆洗回)→
// 觸發 userStart → 玩家事件主迴圈——迴圈頂偵測階段跳轉(顧客行動 / 回合結束)視為玩家結束;
// Operator.PlayerAction 回手牌卡即出牌(playCard)、回 nil 即玩家結束(playerEnd)。
func phasePlayerAction(game *cores.Game) cores.PhaseKind {
	for int32(len(game.Hand)) < game.GetDrawMax().GetValue() && len(game.Deck)+len(game.Drop) > 0 {
		if len(game.Deck) == 0 { // 抽牌牌堆空 → 棄牌牌堆洗牌後整堆移入(此時棄牌牌堆轉空)
			shuffleCard(game, game.Drop)
			game.Deck = game.Drop
			game.Drop = cores.CardList{}
		} // if

		card := game.Deck[0] // 頂端 = index 0(M8 約定)
		removeCard(game, cores.ContainerDeck, card)
		placeCard(game, cores.ContainerHand, card) // 設抽牌事件 + cardDraw 觸發
	} // for

	fireTrigger(game, cores.TriggerUserStart) // 玩家開始觸發

	for {
		if next := game.GetNextPhase(); next == cores.PhaseGuestAction || next == cores.PhaseRoundEnd {
			return playerEnd(game) // 階段跳轉 → 視為玩家結束事件
		} // if

		card := game.GetOperator().PlayerAction(game)

		if card == nil {
			return playerEnd(game) // 玩家結束事件
		} // if

		playCard(game, card)
	} // for
}

// playCard 玩家出牌(【營業規格書 | 十九、核心流程 | 3. 玩家行動階段】出牌 case): 封印 / 卡牌化無空位 / 點數三閘 →
// 扣點、出牌事件、重複 (1 + 額外次數) 次啟動實例效果列表、卡牌化卡自動還原(restoreOne)、出牌後流放 / 棄牌路由 → 觸發 cardPlay。
// 閘門不過 / 回傳卡不在手牌(防禦)→ 不執行出牌, 主迴圈回到下一個玩家事件。
func playCard(game *cores.Game, card *cores.Card) {
	if _, where, ok := game.LocateCard(card.GetInstanceID()); ok == false || where != cores.ContainerHand {
		return // 防禦: 非本場手牌卡 → 不執行出牌
	} // if

	if card.GetSeal().IsLock() {
		return // 卡牌封印 → 不執行出牌
	} // if

	if card.GetCardify() != nil && len(emptySeat(game)) == 0 {
		return // 卡牌化卡無空位 → 不執行出牌
	} // if

	if game.GetEnergy().GetValue() < card.GetCost().GetValue() {
		return // 點數不足 → 不執行出牌
	} // if

	emitPlayTitle(game, card) // 範圍標題: 玩家出牌(操作元 = 卡牌 + 技能)

	game.GetEnergy().Sub(float64(card.GetCost().GetValue())) // 出牌耗能(鎖定 → 不扣、照出牌)
	cores.EmitProperty(game, 0, cores.NoneID, "energy", cores.AssignSub, float64(card.GetCost().GetValue()), float64(game.GetEnergy().GetValue()))
	game.EventPlay(card, cardGroup(game, card)) // 最後出牌 / 回合張數 / 整場累積
	extra := extraCount(game, card)

	for itor := int32(0); itor <= extra; itor++ { // 重複 (1 + 額外次數) 次
		runEffectList(game, card.GetEffectID().List(), game.CardSkillGroup(card.GetCardID())) // 啟動技能: 實例效果列表
	} // for

	if card.GetCardify() != nil {
		restoreOne(game, card) // 系統自動還原綁定顧客(效果若占滿座位 → 還原內部 no-op)
	} // if

	// 效果命令可能在出牌途中搬動本卡, 路由前重新定位(比照 cardRun)。
	if _, now, found := game.LocateCard(card.GetInstanceID()); found {
		dest := cores.ContainerDrop

		if card.GetPlayExile().IsLock() {
			dest = cores.ContainerExile // 出牌後流放
		} // if

		if now != dest {
			removeCard(game, now, card)
			placeCard(game, dest, card)
		} // if
	} // if

	fireTrigger(game, cores.TriggerCardPlay) // 玩家出牌觸發
}

// extraCount 額外發動次數: 於 [下限, 上限] 區間以 Rander 隨機取整數; 下限 > 上限 / 上限 < 0 時為 0, 隨機結果為負亦視為 0(防禦)。
func extraCount(game *cores.Game, card *cores.Card) int32 {
	low := card.GetExtraRunMin().GetValue()
	high := card.GetExtraRunMax().GetValue()

	if low > high || high < 0 {
		return 0
	} // if

	result := low + int32(game.GetRander().Intn(int(high-low+1)))

	if result < 0 {
		return 0
	} // if

	return result
}

// playerEnd 玩家結束流程(【營業規格書 | 十九、核心流程 | 3. 玩家行動階段】玩家結束 case): 發手動結束範圍標題
// (階段跳轉路徑亦視為玩家結束、照發)→ 觸發 userEnd → 剩餘手牌處理(未出牌流放 → 流放、不棄留手、其餘棄置)→
// 依下一階段分派(回合結束 / 預設顧客行動, 皆清除跳轉)。
func playerEnd(game *cores.Game) cores.PhaseKind {
	cores.EmitTitle(game, "手動結束")           // 範圍標題: 手動結束(流程、無操作元)
	fireTrigger(game, cores.TriggerUserEnd) // 玩家結束觸發

	for _, itor := range append(cores.CardList{}, game.Hand...) { // 快照: 處理本身與觸發效果都會搬動手牌
		if _, where, ok := game.LocateCard(itor.GetInstanceID()); ok == false || where != cores.ContainerHand {
			continue // 已被觸發效果搬離手牌 → 略過
		} // if

		if itor.GetUnplayExile().IsLock() {
			removeCard(game, cores.ContainerHand, itor)
			placeCard(game, cores.ContainerExile, itor) // 未出牌流放
			continue
		} // if

		if itor.GetKeep().IsLock() == false {
			removeCard(game, cores.ContainerHand, itor)
			placeCard(game, cores.ContainerDrop, itor) // 棄置; 不棄卡牌留手牌
		} // if
	} // for

	next := game.GetNextPhase()
	game.SetNextPhase(cores.PhaseNone) // 各分支皆清除跳轉

	if next == cores.PhaseRoundEnd {
		return cores.PhaseRoundEnd
	} // if

	return cores.PhaseGuestAction
}
