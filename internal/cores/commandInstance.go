package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 實例化命令（【營業規格書 | 二十五、操作命令清單】*Add / *Copy / *Clone / *Roll / guestSpawn / waitAdd）：
// 依卡牌 / 顧客資料或抽獎結果建立新實例加入容器。新卡實例效果列表 = Card.SkillID → Skill.EffectID；
// 卡牌 bool 欄（Keep / Seal / PlayExile / UnplayExile）→ 鎖定計數。入容器重用 placeCard（事件 + 觸發）。

// === 加牌 *Add（none；卡牌編號；N）===

func commandHandAdd(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardAdd(eng, arg, ContainerHand)
}

func commandDeckAdd(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardAdd(eng, arg, ContainerDeck)
}

func commandDropAdd(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardAdd(eng, arg, ContainerDrop)
}

func commandExileAdd(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardAdd(eng, arg, ContainerExile)
}

// === 抽獎加牌 *Roll（none；抽獎群組編號；N）===

func commandHandRoll(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardRoll(eng, arg, ContainerHand)
}

func commandDeckRoll(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardRoll(eng, arg, ContainerDeck)
}

func commandDropRoll(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardRoll(eng, arg, ContainerDrop)
}

func commandExileRoll(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardRoll(eng, arg, ContainerExile)
}

// === 淺 / 深複製 *Copy / *Clone（對象；N；[洗牌(僅 deck)]；附加效果編號 1..M）===

func commandHandCopy(eng *Engine, target []InstanceID, arg []exprs.Value) {
	copyClone(eng, target, arg, ContainerHand, false)
}

func commandDeckCopy(eng *Engine, target []InstanceID, arg []exprs.Value) {
	copyClone(eng, target, arg, ContainerDeck, false)
}

func commandDropCopy(eng *Engine, target []InstanceID, arg []exprs.Value) {
	copyClone(eng, target, arg, ContainerDrop, false)
}

func commandExileCopy(eng *Engine, target []InstanceID, arg []exprs.Value) {
	copyClone(eng, target, arg, ContainerExile, false)
}

func commandHandClone(eng *Engine, target []InstanceID, arg []exprs.Value) {
	copyClone(eng, target, arg, ContainerHand, true)
}

func commandDeckClone(eng *Engine, target []InstanceID, arg []exprs.Value) {
	copyClone(eng, target, arg, ContainerDeck, true)
}

func commandDropClone(eng *Engine, target []InstanceID, arg []exprs.Value) {
	copyClone(eng, target, arg, ContainerDrop, true)
}

func commandExileClone(eng *Engine, target []InstanceID, arg []exprs.Value) {
	copyClone(eng, target, arg, ContainerExile, true)
}

// === 顧客實例化 ===

// commandGuestSpawn 依顧客編號實例化 1 位新顧客（【二十五 | guestSpawn】處理流程）：
// 座位編號 = 0 → 加入遊蕩列表並自動鎖（sate / sateSeal / calmSeal 各 +1）；
// 座位編號 = N>0 且該座位存在且空 → 設 seatID 加入座位列表（否則 no-op）；座位編號 < 0 / 顧客資料不存在 → no-op。
func commandGuestSpawn(eng *Engine, target []InstanceID, arg []exprs.Value) {
	guestID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	seatID, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	if seatID < 0 {
		return // 座位編號 < 0 → no-op
	} // if

	if seatID > 0 && (eng.data.Seat.Get(seatID) == nil || eng.runtime.Seat[seatID] != nil) {
		return // 座位不存在 / 已占用 → no-op（先於實例化判定，不白費實例編號）
	} // if

	guest := newGuest(eng, guestID)

	if guest == nil {
		return // 顧客資料不存在 → no-op
	} // if

	if seatID == 0 {
		guest.Sate.Lock++
		guest.SateSeal.Lock++
		guest.CalmSeal.Lock++
		eng.runtime.Roam = append(eng.runtime.Roam, guest)
		return
	} // if

	guest.SeatID = seatID
	eng.runtime.Seat[seatID] = guest
}

// commandWaitAdd 依顧客編號實例化 N 位新顧客插入排隊佇列前端（優先入座）；N <= 0 / 顧客資料不存在 → no-op（【二十五 | waitAdd】）。
func commandWaitAdd(eng *Engine, target []InstanceID, arg []exprs.Value) {
	guestID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	n, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	for itor := int32(0); itor < n; itor++ {
		guest := newGuest(eng, guestID)

		if guest == nil {
			return // 顧客資料不存在 → no-op
		} // if

		eng.runtime.Wait = append([]*Guest{guest}, eng.runtime.Wait...)
	} // for
}

// === 實例化輔助 ===

// cardAdd 依卡牌編號實例化 N 張加入 dest（none 命令對象）；N <= 0 / 卡牌資料不存在 → no-op。
func cardAdd(eng *Engine, arg []exprs.Value, dest ContainerKind) {
	cardID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	n, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	for itor := int32(0); itor < n; itor++ {
		card := newCard(eng, cardID)

		if card == nil {
			return // 卡牌資料不存在 → no-op
		} // if

		placeCard(eng, dest, card)
	} // for
}

// cardRoll 對抽獎群組 weighted random 選編號實例化 N 張加入 dest（每張獨立 roll）；N <= 0 / 群組總權重 0 → no-op。
func cardRoll(eng *Engine, arg []exprs.Value, dest ContainerKind) {
	group, ok := argInt(arg)

	if ok == false {
		return
	} // if

	n, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	for itor := int32(0); itor < n; itor++ {
		cardID, rolled := rollCard(eng, group)

		if rolled == false {
			return // 群組總權重 0 / 群組不存在 → no-op
		} // if

		card := newCard(eng, cardID)

		if card == nil {
			continue // 抽中編號無資料 → 跳過該張（防禦）
		} // if

		placeCard(eng, dest, card)
	} // for
}

// copyClone 對命令對象每張卡牌（source）淺 / 深複製 N 張加入 dest 並附加 trailing 效果編號；
// dest 為抽牌牌堆時讀洗牌參數（index 1）、附加效果自 index 2，其餘容器附加效果自 index 1。
func copyClone(eng *Engine, target []InstanceID, arg []exprs.Value, dest ContainerKind, deep bool) {
	n, ok := argInt(arg)

	if ok == false {
		return
	} // if

	effectStart := 1
	shuffle := false

	if dest == ContainerDeck {
		shuffle = argBool(argTail(arg, 1))
		effectStart = 2
	} // if

	addEffect := intList(argTail(arg, effectStart))

	for _, itor := range target {
		source, _, found := eng.locateCard(itor)

		if found == false {
			continue // 非卡牌實例 → 該項 no-op
		} // if

		for count := int32(0); count < n; count++ {
			card := copyCard(eng, source, deep)

			if card == nil {
				continue // 淺複製載入失敗（卡牌資料不存在）→ 跳過該張
			} // if

			card.EffectID = append(card.EffectID, addEffect...)
			placeCard(eng, dest, card)
		} // for
	} // for

	if shuffle {
		shuffleCard(eng, eng.runtime.Deck)
	} // if
}

// rollCard 自抽獎群組以 Rander weighted random 取 1 個卡牌編號;群組不存在 / 無正權重候選 → ok=false。
func rollCard(eng *Engine, group int32) (cardID int32, ok bool) {
	award, found := eng.award[group]

	if found == false || len(award.weight) == 0 {
		return 0, false
	} // if

	return award.cardID[eng.rander.Weighted(award.weight)], true
}
