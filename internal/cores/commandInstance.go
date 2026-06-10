package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 實例化命令（【營業規格書 | 二十五、操作命令清單】*Add / *Copy / *Clone / *Roll / guestSpawn / waitAdd）：
// 依卡牌 / 顧客資料或抽獎結果建立新實例加入容器。新卡實例效果列表 = Card.SkillID → Skill.EffectID；
// 卡牌 bool 欄（Keep / Seal / PlayExile / UnplayExile）→ 鎖定計數。入容器重用 placeCard（事件 + 觸發）。

// === 加牌 *Add（none；卡牌編號；N）===

func commandHandAdd(game *Game, target []InstanceID, arg []exprs.Value) {
	cardAdd(game, arg, ContainerHand)
}

func commandDeckAdd(game *Game, target []InstanceID, arg []exprs.Value) {
	cardAdd(game, arg, ContainerDeck)
}

func commandDropAdd(game *Game, target []InstanceID, arg []exprs.Value) {
	cardAdd(game, arg, ContainerDrop)
}

func commandExileAdd(game *Game, target []InstanceID, arg []exprs.Value) {
	cardAdd(game, arg, ContainerExile)
}

// === 抽獎加牌 *Roll（none；抽獎群組編號；N）===

func commandHandRoll(game *Game, target []InstanceID, arg []exprs.Value) {
	cardRoll(game, arg, ContainerHand)
}

func commandDeckRoll(game *Game, target []InstanceID, arg []exprs.Value) {
	cardRoll(game, arg, ContainerDeck)
}

func commandDropRoll(game *Game, target []InstanceID, arg []exprs.Value) {
	cardRoll(game, arg, ContainerDrop)
}

func commandExileRoll(game *Game, target []InstanceID, arg []exprs.Value) {
	cardRoll(game, arg, ContainerExile)
}

// === 淺 / 深複製 *Copy / *Clone（對象；N；[洗牌(僅 deck)]；附加效果編號 1..M）===

func commandHandCopy(game *Game, target []InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, ContainerHand, false)
}

func commandDeckCopy(game *Game, target []InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, ContainerDeck, false)
}

func commandDropCopy(game *Game, target []InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, ContainerDrop, false)
}

func commandExileCopy(game *Game, target []InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, ContainerExile, false)
}

func commandHandClone(game *Game, target []InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, ContainerHand, true)
}

func commandDeckClone(game *Game, target []InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, ContainerDeck, true)
}

func commandDropClone(game *Game, target []InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, ContainerDrop, true)
}

func commandExileClone(game *Game, target []InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, ContainerExile, true)
}

// === 顧客實例化 ===

// commandGuestSpawn 依顧客編號實例化 1 位新顧客（【二十五 | guestSpawn】處理流程）：
// 座位編號 = 0 → 加入遊蕩列表並自動鎖（sate / sateSeal / calmSeal 各 +1）；
// 座位編號 = N>0 且該座位存在且空 → 設 seatID 加入座位列表（否則 no-op）；座位編號 < 0 / 顧客資料不存在 → no-op。
func commandGuestSpawn(game *Game, target []InstanceID, arg []exprs.Value) {
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

	if seatID > 0 && (game.data.Seat.Get(seatID) == nil || game.Seat[seatID] != nil) {
		return // 座位不存在 / 已占用 → no-op（先於實例化判定，不白費實例編號）
	} // if

	guest := NewGuest(game, guestID)

	if guest == nil {
		return // 顧客資料不存在 → no-op
	} // if

	if seatID == 0 {
		guest.RoamLock()
		game.Roam.Push(guest)
		return
	} // if

	game.Seat.Place(seatID, guest)
}

// commandWaitAdd 依顧客編號實例化 N 位新顧客插入排隊佇列前端（優先入座）；N <= 0 / 顧客資料不存在 → no-op（【二十五 | waitAdd】）。
func commandWaitAdd(game *Game, target []InstanceID, arg []exprs.Value) {
	guestID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	n, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	for itor := int32(0); itor < n; itor++ {
		guest := NewGuest(game, guestID)

		if guest == nil {
			return // 顧客資料不存在 → no-op
		} // if

		game.Wait.Insert(guest)
	} // for
}

// === 實例化輔助 ===

// cardAdd 依卡牌編號實例化 N 張加入 dest（none 命令對象）；N <= 0 / 卡牌資料不存在 → no-op。
func cardAdd(game *Game, arg []exprs.Value, dest ContainerKind) {
	cardID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	n, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	for itor := int32(0); itor < n; itor++ {
		card := NewCard(game, cardID)

		if card == nil {
			return // 卡牌資料不存在 → no-op
		} // if

		placeCard(game, dest, card)
	} // for
}

// cardRoll 對抽獎群組 weighted random 選編號實例化 N 張加入 dest（每張獨立 roll）；N <= 0 / 群組總權重 0 → no-op。
func cardRoll(game *Game, arg []exprs.Value, dest ContainerKind) {
	group, ok := argInt(arg)

	if ok == false {
		return
	} // if

	n, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	for itor := int32(0); itor < n; itor++ {
		cardID, rolled := rollCard(game, group)

		if rolled == false {
			return // 群組總權重 0 / 群組不存在 → no-op
		} // if

		card := NewCard(game, cardID)

		if card == nil {
			continue // 抽中編號無資料 → 跳過該張（防禦）
		} // if

		placeCard(game, dest, card)
	} // for
}

// copyClone 對命令對象每張卡牌（source）淺 / 深複製 N 張加入 dest 並附加 trailing 效果編號；
// dest 為抽牌牌堆時讀洗牌參數（index 1）、附加效果自 index 2，其餘容器附加效果自 index 1。
func copyClone(game *Game, target []InstanceID, arg []exprs.Value, dest ContainerKind, deep bool) {
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
		source, _, found := game.locateCard(itor)

		if found == false {
			continue // 非卡牌實例 → 該項 no-op
		} // if

		for count := int32(0); count < n; count++ {
			card := CopyCard(game, source, deep)

			if card == nil {
				continue // 淺複製載入失敗（卡牌資料不存在）→ 跳過該張
			} // if

			card.GetEffectID().Add(addEffect...)
			placeCard(game, dest, card)
		} // for
	} // for

	if shuffle {
		shuffleCard(game, game.Deck)
	} // if
}

// rollCard 自抽獎群組以 Rander weighted random 取 1 個卡牌編號;群組不存在 / 無正權重候選 → ok=false。
func rollCard(game *Game, group int32) (cardID int32, ok bool) {
	award, found := game.awardData[group]

	if found == false || len(award.weight) == 0 {
		return 0, false
	} // if

	return award.cardID[game.rander.Weighted(award.weight)], true
}
