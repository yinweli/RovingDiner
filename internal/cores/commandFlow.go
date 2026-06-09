package cores

import (
	"sort"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 處理流程命令（【營業規格書 | 二十五、操作命令清單】各「處理流程」）：cardRun / *Morph / cardify / restore / guest*。
// 命令本體於 M9 完成；觸發（fireTrigger，M11）與啟動效果列表（runEffectList，M12）已接呼叫;
// 清理效果（cleanupEffect）留 M13,插入點以 // TODO(M13) 標（空 stub 無語句、無法覆蓋，故不先建）。

// commandCardRun 強制發動卡牌（cardRun 處理流程）：消耗點數、設出牌事件、啟動實例效果列表、進棄牌堆、觸發 cardPlay。
// 卡牌可位於 手牌 / 抽牌 / 棄牌；流放牌堆視為位置不符 → 該項 no-op。參數：消耗點數（bool）、進棄牌堆（bool）。
func commandCardRun(eng *Engine, target []InstanceID, arg []exprs.Value) {
	useEnergy := argBool(arg)
	toDrop := argBool(argTail(arg, 1))

	for _, itor := range target {
		card, where, ok := eng.locateCard(itor)

		if ok == false || where == ContainerExile {
			continue // 不存在 / 流放牌堆（位置不符）→ 該項 no-op
		} // if

		if card.Seal.Locked() {
			continue // 封印閘門 → no-op
		} // if

		game := eng.runtime.Game

		if useEnergy {
			if game.Energy.Value < card.Cost.Value {
				continue // 點數不足 → no-op
			} // if

			game.Energy.Value -= card.Cost.Value
		} // if

		game.PlayLast = card
		game.PlayCount++
		game.PlayTotal[cardGroup(eng, card)]++                              // §二十五 step4 列 最後出牌 / 回合張數;整場累積出牌於此補（與其他 *Total 一致）
		runEffectList(eng, card.EffectID, cardSkillGroup(eng, card.CardID)) // 啟動實例效果列表（§二十五 step4 出牌前）

		if toDrop && where != ContainerDrop {
			removeCard(eng, where, card)
			placeCard(eng, ContainerDrop, card) // 進棄牌牌堆（設 dropLast 等 + M11 cardDrop 觸發）
		} // if

		fireTrigger(eng, TriggerCardPlay) // 玩家出牌觸發
	} // for
}

// === 變身 *Morph（對象；抽獎群組編號）===

func commandHandMorph(eng *Engine, target []InstanceID, arg []exprs.Value) {
	morph(eng, target, arg, ContainerHand)
}

func commandDeckMorph(eng *Engine, target []InstanceID, arg []exprs.Value) {
	morph(eng, target, arg, ContainerDeck)
}

func commandDropMorph(eng *Engine, target []InstanceID, arg []exprs.Value) {
	morph(eng, target, arg, ContainerDrop)
}

func commandExileMorph(eng *Engine, target []InstanceID, arg []exprs.Value) {
	morph(eng, target, arg, ContainerExile)
}

// morph 變身處理流程：抽獎取新 cardID、就地換 cardID + 重分配實例編號 + 依新卡資料載入部分欄位、設變身事件；
// 位置不符 / 群組總權重 0 該項 no-op。source 留原牌堆位置、不觸發 cardDraw / Drop / Exile。
func morph(eng *Engine, target []InstanceID, arg []exprs.Value, where ContainerKind) {
	group, ok := argInt(arg)

	if ok == false {
		return
	} // if

	for _, itor := range target {
		card, at, found := eng.locateCard(itor)

		if found == false || at != where {
			continue // 位置不符 → 該項 no-op
		} // if

		newID, rolled := rollCard(eng, group)

		if rolled == false {
			continue // 群組總權重 0 / 群組不存在 → 該項 no-op
		} // if

		meta := eng.data.Card.Get(newID)

		if meta == nil {
			continue // 變身後卡牌資料不存在 → 跳過（防禦）
		} // if

		oldID := card.CardID
		card.CardID = newID
		card.InstanceID = eng.runtime.NextID() // 重分配實例編號（舊編號失效）
		card.Cost = Value{Value: meta.Cost}    // §二十五 step4 僅載 出牌費用 / 不棄鎖 / 封印鎖 / 實例效果列表
		card.Keep = boolLock(meta.Keep)
		card.Seal = boolLock(meta.Seal)
		card.EffectID = skillEffect(eng, meta.SkillID)
		// TODO(M13)：cleanupEffect 以變身前舊實例編號清理佇列（M13 實作須在重分配 InstanceID 前先存舊編號）

		game := eng.runtime.Game
		game.MorphLast = card
		game.MorphOldID = oldID
		game.MorphNewID = newID
		game.MorphCount++
		fireTrigger(eng, TriggerCardMorph) // 卡牌變身觸發
	} // for
}

// commandCardify 卡牌化顧客（cardify 處理流程）：將座位顧客移入卡牌化列表、凍結、實例化對應卡牌（不棄 + 綁來源）加入手牌。
// 顧客不在座位列表 / 卡牌資料不存在 → 該項 no-op。參數：卡牌編號。
func commandCardify(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	for _, itor := range target {
		guest, where, found := eng.locateGuest(itor)

		if found == false || where != ContainerSeat {
			continue // 不在座位列表 → 該項 no-op
		} // if

		card := newCard(eng, cardID)

		if card == nil {
			continue // 卡牌資料不存在 → 該項 no-op
		} // if

		delete(eng.runtime.Seat, guest.SeatID) // 1. 自座位列表移除
		guest.SeatID = 0
		eng.runtime.Cardify = append(eng.runtime.Cardify, guest) // 2. 加入卡牌化列表
		guest.Freeze = eng.runtime.Game.Round                    // 3. 凍結起始回合
		card.Cardify = guest                                     // 5. 綁卡牌化來源
		card.Keep.Lock++                                         // 6. 不棄卡牌鎖定 + 1
		placeCard(eng, ContainerHand, card)                      // 7. 加入手牌
	} // for
}

// commandRestore 卡牌化還原（restore 處理流程）：把卡牌綁定的卡牌化顧客送回隨機空座位、調整其效果到期、解凍、解綁、不棄回退。
// self.cardify = none / 剩餘座位 = 0 → 該項 no-op。參數：無。
func commandRestore(eng *Engine, target []InstanceID, arg []exprs.Value) {
	round := eng.runtime.Game.Round

	for _, itor := range target {
		card, _, found := eng.locateCard(itor)

		if found == false || card.Cardify == nil {
			continue // 非卡牌 / self.cardify = none → 該項 no-op
		} // if

		seatID, ok := randomEmptySeat(eng)

		if ok == false {
			continue // 剩餘座位 = 0 → 該項 no-op
		} // if

		guest := card.Cardify
		eng.runtime.Cardify = removeGuest(eng.runtime.Cardify, guest) // 2. 自卡牌化列表移除
		guest.SeatID = seatID
		eng.runtime.Seat[seatID] = guest // 3. 加入座位列表（隨機空位）

		for _, effect := range eng.runtime.Effect { // 4. 調整其效果結束回合（補回凍結期間）
			if effect.Self.Guest == guest && effect.Expire > 0 {
				effect.Expire += round - guest.Freeze
			} // if
		} // for

		guest.Freeze = 0    // 5. 解凍
		card.Cardify = nil  // 6. self.cardify = none
		lockDec(&card.Keep) // 7. 不棄卡牌鎖定 - 1
	} // for
}

// === 顧客流程 ===

// commandGuestExit 顧客離場（guestExit 處理流程）：設離場事件、（M11）觸發離場時機、自所在容器移除、（M13）清理效果、給滿意值、扣士氣。
// 參數：是否給滿意值（bool）、是否扣士氣（bool）。
func commandGuestExit(eng *Engine, target []InstanceID, arg []exprs.Value) {
	giveScore := argBool(arg)
	dropMorale := argBool(argTail(arg, 1))

	for _, itor := range target {
		guest, where, found := eng.locateGuest(itor)

		if found == false {
			continue // 非顧客實例 → 該項 no-op
		} // if

		guestExitOne(eng, guest, where, giveScore, dropMorale)
	} // for
}

// commandGuestReturn 顧客回座（guestReturn 處理流程）：解入列自動鎖;有空位回隨機座位、無空位走 guestExit(false, true)。
// 限位於遊蕩列表者;其他位置 → 該項 no-op。參數：無。
func commandGuestReturn(eng *Engine, target []InstanceID, arg []exprs.Value) {
	for _, itor := range target {
		guest, where, found := eng.locateGuest(itor)

		if found == false || where != ContainerRoam {
			continue // 限遊蕩列表 → 該項 no-op
		} // if

		lockDec(&guest.Sate) // 1. 解入列自動鎖（sate / sateSeal / calmSeal 各 -1）
		lockDec(&guest.SateSeal)
		lockDec(&guest.CalmSeal)

		seatID, ok := randomEmptySeat(eng)

		if ok == false {
			guestExitOne(eng, guest, ContainerRoam, false, true) // 3. 剩餘座位 = 0 → 離場
			continue
		} // if

		eng.runtime.Roam = removeGuest(eng.runtime.Roam, guest) // 2. 有空位 → 回座
		guest.SeatID = seatID
		eng.runtime.Seat[seatID] = guest
	} // for
}

// commandGuestRoam 顧客遊蕩（guestRoam）：座位顧客移入遊蕩列表並自動鎖（sate / sateSeal / calmSeal 各 +1）;限座位列表者。
// 參數：無。
func commandGuestRoam(eng *Engine, target []InstanceID, arg []exprs.Value) {
	for _, itor := range target {
		guest, where, found := eng.locateGuest(itor)

		if found == false || where != ContainerSeat {
			continue // 限座位列表 → 該項 no-op
		} // if

		delete(eng.runtime.Seat, guest.SeatID)
		guest.SeatID = 0
		eng.runtime.Roam = append(eng.runtime.Roam, guest)
		guest.Sate.Lock++ // 自動鎖
		guest.SateSeal.Lock++
		guest.CalmSeal.Lock++
	} // for
}

// commandGuestSeat 顧客入座（guestSeat）：自排隊佇列彈出隊首 1 位加入隨機空座位、設入座事件、（M11）觸發 guestSeat。
// 命令對象固定 none;排隊佇列空 / 無空位 → no-op。參數：無。
func commandGuestSeat(eng *Engine, target []InstanceID, arg []exprs.Value) {
	if len(eng.runtime.Wait) == 0 {
		return // 排隊佇列空 → no-op
	} // if

	seatID, ok := randomEmptySeat(eng)

	if ok == false {
		return // 無空座位 → no-op
	} // if

	guest := eng.runtime.Wait[0]
	eng.runtime.Wait = eng.runtime.Wait[1:] // 彈出隊首
	guest.SeatID = seatID
	eng.runtime.Seat[seatID] = guest

	game := eng.runtime.Game
	game.SeatLast = guest
	game.SeatCount++
	fireTrigger(eng, TriggerGuestSeat) // 顧客入座觸發
}

// === 流程輔助 ===

// guestExitOne 對單一顧客執行 guestExit 處理流程（供 commandGuestExit 與 guestReturn 無座位分支共用）。
func guestExitOne(eng *Engine, guest *Guest, where ContainerKind, giveScore, dropMorale bool) {
	game := eng.runtime.Game
	game.ExitLast = guest // 1. 離場事件
	game.ExitLastSeat = guest.SeatID
	game.ExitCount++
	fireTrigger(eng, TriggerExitAny) // 2. 顧客離場時機

	if giveScore {
		fireTrigger(eng, TriggerExitSate) // 3. 飽食離場時機（提供滿意值前）
	} // if

	if dropMorale {
		fireTrigger(eng, TriggerExitCalm) // 4. 生氣離場時機（扣士氣前）
	} // if

	removeGuestContainer(eng, where, guest) // 5. 自所在容器移除
	fireTrigger(eng, TriggerExitDone)       // 6. 顧客離場後時機
	// TODO(M13)：cleanupEffect(Self{Guest: guest})（7. 清理離場顧客殘留效果）

	if giveScore {
		game.Score.Value += guest.Score.Value // 8. 給滿意值
	} // if

	if dropMorale {
		moraleDamage(eng, float64(guest.Morale.Value), guest) // 9. 扣士氣（morale -= 特例,以離場顧客為來源）
	} // if
}

// removeGuestContainer 自指定容器移除顧客（座位以 seatID 刪鍵;排隊 / 遊蕩 / 卡牌化以實例編號剔除）。
func removeGuestContainer(eng *Engine, where ContainerKind, guest *Guest) {
	switch where {
	case ContainerSeat:
		delete(eng.runtime.Seat, guest.SeatID)

	case ContainerWait:
		eng.runtime.Wait = removeGuest(eng.runtime.Wait, guest)

	case ContainerRoam:
		eng.runtime.Roam = removeGuest(eng.runtime.Roam, guest)

	case ContainerCardify:
		eng.runtime.Cardify = removeGuest(eng.runtime.Cardify, guest)

	default:
		// 不可達：顧客僅存在於上述四容器
	} // switch
}

// randomEmptySeat 自座位表格取一個隨機空座位編號（依座位編號排序後以 Rander 抽,確保決定性）;無空位回 ok=false。
func randomEmptySeat(eng *Engine) (seatID int32, ok bool) {
	id := eng.data.Seat.Keys()
	sort.Slice(id, func(i, j int) bool { return id[i] < id[j] })

	empty := []int32{}

	for _, itor := range id {
		if eng.runtime.Seat[itor] == nil {
			empty = append(empty, itor)
		} // if
	} // for

	if len(empty) == 0 {
		return 0, false
	} // if

	return empty[eng.rander.Intn(len(empty))], true
}
