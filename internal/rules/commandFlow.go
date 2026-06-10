package rules

import (
	"sort"

	"github.com/yinweli/RovingDiner/internal/cores"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 處理流程命令（【營業規格書 | 二十五、操作命令清單】各「處理流程」）：cardRun / *Morph / cardify / restore / guest*。
// 命令本體於 M9 完成；觸發（fireTrigger，M11）與啟動效果列表（runEffectList，M12）已接呼叫;
// 清理效果（cleanupEffect）留 M13,插入點以 // TODO(M13) 標（空 stub 無語句、無法覆蓋，故不先建）。

// commandCardRun 強制發動卡牌（cardRun 處理流程）：消耗點數、設出牌事件、啟動實例效果列表、進棄牌堆、觸發 cardPlay。
// 卡牌可位於 手牌 / 抽牌 / 棄牌；流放牌堆視為位置不符 → 該項 no-op。參數：消耗點數（bool）、進棄牌堆（bool）。
func commandCardRun(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	useEnergy := argBool(arg)
	toDrop := argBool(argTail(arg, 1))

	for _, itor := range target {
		card, where, ok := game.LocateCard(itor)

		if ok == false || where == cores.ContainerExile {
			continue // 不存在 / 流放牌堆（位置不符）→ 該項 no-op
		} // if

		if card.GetSeal().IsLock() {
			continue // 封印閘門 → no-op
		} // if

		if useEnergy {
			if game.GetEnergy().GetValue() < card.GetCost().GetValue() {
				continue // 點數不足 → no-op
			} // if

			game.GetEnergy().Sub(float64(card.GetCost().GetValue())) // 出牌耗能(鎖定 → 不扣、照出牌)
		} // if

		game.EventPlay(card, cardGroup(game, card))                                           // §二十五 step4 列 最後出牌 / 回合張數;整場累積出牌於此補（與其他 *Total 一致）
		runEffectList(game, card.GetEffectID().List(), game.CardSkillGroup(card.GetCardID())) // 啟動實例效果列表（§二十五 step4 出牌前）

		// 效果命令可能在出牌途中搬動本卡（甚至移出四牌堆），故進棄牌堆前重新定位當前容器,不沿用上方 where。
		if toDrop {
			if _, now, found := game.LocateCard(itor); found && now != cores.ContainerDrop {
				removeCard(game, now, card)
				placeCard(game, cores.ContainerDrop, card) // 進棄牌牌堆（設 dropLast 等 + M11 cardDrop 觸發）
			} // if
		} // if

		fireTrigger(game, cores.TriggerCardPlay) // 玩家出牌觸發
	} // for
}

// === 變身 *Morph（對象；抽獎群組編號）===

func commandHandMorph(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	morph(game, target, arg, cores.ContainerHand)
}

func commandDeckMorph(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	morph(game, target, arg, cores.ContainerDeck)
}

func commandDropMorph(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	morph(game, target, arg, cores.ContainerDrop)
}

func commandExileMorph(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	morph(game, target, arg, cores.ContainerExile)
}

// morph 變身處理流程：抽獎取新 cardID、就地換 cardID + 重分配實例編號 + 依新卡資料載入部分欄位、設變身事件；
// 位置不符 / 群組總權重 0 該項 no-op。source 留原牌堆位置、不觸發 cardDraw / Drop / Exile。
func morph(game *cores.Game, target []cores.InstanceID, arg []exprs.Value, where cores.ContainerKind) {
	group, ok := argInt(arg)

	if ok == false {
		return
	} // if

	for _, itor := range target {
		card, at, found := game.LocateCard(itor)

		if found == false || at != where {
			continue // 位置不符 → 該項 no-op
		} // if

		newID, rolled := game.RollCard(group)

		if rolled == false {
			continue // 群組總權重 0 / 群組不存在 → 該項 no-op
		} // if

		oldID := card.GetCardID()
		// TODO(M13)：cleanupEffect 以變身前舊實例編號清理佇列（M13 實作須在呼叫 Morph 前先存 GetInstanceID）

		if card.Morph(game, newID) == false {
			continue // 變身後卡牌資料不存在 → 跳過（防禦）
		} // if

		game.EventMorph(card, oldID, newID)
		fireTrigger(game, cores.TriggerCardMorph) // 卡牌變身觸發
	} // for
}

// commandCardify 卡牌化顧客（cardify 處理流程）：將座位顧客移入卡牌化列表、凍結、實例化對應卡牌（不棄 + 綁來源）加入手牌。
// 顧客不在座位列表 / 卡牌資料不存在 → 該項 no-op。參數：卡牌編號。
func commandCardify(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	for _, itor := range target {
		guest, where, found := game.LocateGuest(itor)

		if found == false || where != cores.ContainerSeat {
			continue // 不在座位列表 → 該項 no-op
		} // if

		card := cores.NewCard(game, cardID)

		if card == nil {
			continue // 卡牌資料不存在 → 該項 no-op
		} // if

		game.Seat.Remove(guest)                     // 1. 自座位列表移除
		game.Cardify.Push(guest)                    // 2. 加入卡牌化列表
		guest.SetFreeze(game.GetRound().GetValue()) // 3. 凍結起始回合
		card.CardifyBind(guest)                     // 5+6. 綁卡牌化來源 + 不棄卡牌鎖定 + 1
		placeCard(game, cores.ContainerHand, card)  // 7. 加入手牌
	} // for
}

// commandRestore 卡牌化還原（restore 處理流程）：把卡牌綁定的卡牌化顧客送回隨機空座位、調整其效果到期、解凍、解綁、不棄回退。
// self.cardify = none / 剩餘座位 = 0 → 該項 no-op。參數：無。
func commandRestore(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	round := game.GetRound().GetValue()

	for _, itor := range target {
		card, _, found := game.LocateCard(itor)

		if found == false || card.GetCardify() == nil {
			continue // 非卡牌 / self.cardify = none → 該項 no-op
		} // if

		seatID, ok := randomEmptySeat(game)

		if ok == false {
			continue // 剩餘座位 = 0 → 該項 no-op
		} // if

		guest := card.GetCardify()
		game.Cardify.Remove(guest.GetInstanceID()) // 2. 自卡牌化列表移除
		game.Seat.Place(seatID, guest)             // 3. 加入座位列表（隨機空位）

		for _, effect := range game.Effect { // 4. 調整其效果結束回合（補回凍結期間）
			if effect.GetSelf().GetGuest() == guest && effect.GetExpire() > 0 {
				effect.SetExpire(effect.GetExpire() + round - guest.GetFreeze())
			} // if
		} // for

		guest.SetFreeze(0) // 5. 解凍
		card.CardifyFree() // 6+7. 解綁 + 不棄卡牌鎖定 - 1（Unlock 夾 ≥ 0）
	} // for
}

// === 顧客流程 ===

// commandGuestExit 顧客離場（guestExit 處理流程）：設離場事件、（M11）觸發離場時機、自所在容器移除、（M13）清理效果、給滿意值、扣士氣。
// 參數：是否給滿意值（bool）、是否扣士氣（bool）。
func commandGuestExit(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	giveScore := argBool(arg)
	dropMorale := argBool(argTail(arg, 1))

	for _, itor := range target {
		guest, where, found := game.LocateGuest(itor)

		if found == false {
			continue // 非顧客實例 → 該項 no-op
		} // if

		guestExitOne(game, guest, where, giveScore, dropMorale)
	} // for
}

// commandGuestReturn 顧客回座（guestReturn 處理流程）：解入列自動鎖;有空位回隨機座位、無空位走 guestExit(false, true)。
// 限位於遊蕩列表者;其他位置 → 該項 no-op。參數：無。
func commandGuestReturn(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	for _, itor := range target {
		guest, where, found := game.LocateGuest(itor)

		if found == false || where != cores.ContainerRoam {
			continue // 限遊蕩列表 → 該項 no-op
		} // if

		guest.RoamUnlock() // 1. 解入列自動鎖（sate / sateSeal / calmSeal 各 -1）

		seatID, ok := randomEmptySeat(game)

		if ok == false {
			guestExitOne(game, guest, cores.ContainerRoam, false, true) // 3. 剩餘座位 = 0 → 離場
			continue
		} // if

		game.Roam.Remove(guest.GetInstanceID()) // 2. 有空位 → 回座
		game.Seat.Place(seatID, guest)
	} // for
}

// commandGuestRoam 顧客遊蕩（guestRoam）：座位顧客移入遊蕩列表並自動鎖（sate / sateSeal / calmSeal 各 +1）;限座位列表者。
// 參數：無。
func commandGuestRoam(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	for _, itor := range target {
		guest, where, found := game.LocateGuest(itor)

		if found == false || where != cores.ContainerSeat {
			continue // 限座位列表 → 該項 no-op
		} // if

		game.Seat.Remove(guest)
		game.Roam.Push(guest)
		guest.RoamLock() // 自動鎖
	} // for
}

// commandGuestSeat 顧客入座（guestSeat）：自排隊佇列彈出隊首 1 位加入隨機空座位、設入座事件、（M11）觸發 guestSeat。
// 命令對象固定 none;排隊佇列空 / 無空位 → no-op。參數：無。
func commandGuestSeat(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	if len(game.Wait) == 0 {
		return // 排隊佇列空 → no-op
	} // if

	seatID, ok := randomEmptySeat(game)

	if ok == false {
		return // 無空座位 → no-op
	} // if

	guest := game.Wait.Pop() // 彈出隊首
	game.Seat.Place(seatID, guest)

	game.EventSeat(guest)
	fireTrigger(game, cores.TriggerGuestSeat) // 顧客入座觸發
}

// === 流程輔助 ===

// guestExitOne 對單一顧客執行 guestExit 處理流程（供 commandGuestExit 與 guestReturn 無座位分支共用）。
func guestExitOne(game *cores.Game, guest *cores.Guest, where cores.ContainerKind, giveScore, dropMorale bool) {
	game.EventExit(guest)                   // 1. 離場事件（離場座位取自當下 seatID,故在容器移除前）
	fireTrigger(game, cores.TriggerExitAny) // 2. 顧客離場時機

	if giveScore {
		fireTrigger(game, cores.TriggerExitSate) // 3. 飽食離場時機（提供滿意值前）
	} // if

	if dropMorale {
		fireTrigger(game, cores.TriggerExitCalm) // 4. 生氣離場時機（扣士氣前）
	} // if

	removeGuestContainer(game, where, guest) // 5. 自所在容器移除
	fireTrigger(game, cores.TriggerExitDone) // 6. 顧客離場後時機
	// TODO(M13)：cleanupEffect(NewRefGuest(guest))（7. 清理離場顧客殘留效果）

	if giveScore {
		game.GetScore().Add(float64(guest.GetScore().GetValue())) // 8. 給滿意值(鎖定 → 不給)
	} // if

	if dropMorale {
		moraleDamage(game, float64(guest.GetMorale().GetValue()), guest) // 9. 扣士氣（morale -= 特例,以離場顧客為來源）
	} // if
}

// removeGuestContainer 自指定容器移除顧客（座位走 SeatList.Remove 含 seatID 歸零;排隊 / 遊蕩 / 卡牌化以實例編號剔除）。
func removeGuestContainer(game *cores.Game, where cores.ContainerKind, guest *cores.Guest) {
	switch where {
	case cores.ContainerSeat:
		game.Seat.Remove(guest)

	case cores.ContainerWait:
		game.Wait.Remove(guest.GetInstanceID())

	case cores.ContainerRoam:
		game.Roam.Remove(guest.GetInstanceID())

	case cores.ContainerCardify:
		game.Cardify.Remove(guest.GetInstanceID())

	default:
		// 不可達：顧客僅存在於上述四容器
	} // switch
}

// randomEmptySeat 自座位表格取一個隨機空座位編號（依座位編號排序後以 Rander 抽,確保決定性）;無空位回 ok=false。
func randomEmptySeat(game *cores.Game) (seatID int32, ok bool) {
	id := game.GetSheet().Seat.Keys()
	sort.Slice(id, func(i, j int) bool { return id[i] < id[j] })

	empty := []int32{}

	for _, itor := range id {
		if game.Seat[itor] == nil {
			empty = append(empty, itor)
		} // if
	} // for

	if len(empty) == 0 {
		return 0, false
	} // if

	return empty[game.GetRander().Intn(len(empty))], true
}
