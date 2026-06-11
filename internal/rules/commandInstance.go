package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 實例化命令(【營業規格書 | 二十五、操作命令清單】*Add / *Copy / *Clone / *Roll / guestSpawn / waitAdd):
// 依卡牌 / 顧客資料或抽獎結果建立新實例加入容器。新卡實例效果列表 = Card.SkillID → Skill.EffectID;
// 卡牌 bool 欄(Keep / Seal / PlayExile / UnplayExile)→ 鎖定計數。入容器重用 placeCard(事件 + 觸發)。

// === 加牌 *Add(none; 卡牌編號; N)===

func commandHandAdd(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardAdd(game, arg, cores.ContainerHand)
}

func commandDeckAdd(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardAdd(game, arg, cores.ContainerDeck)
}

func commandDropAdd(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardAdd(game, arg, cores.ContainerDrop)
}

func commandExileAdd(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardAdd(game, arg, cores.ContainerExile)
}

// === 抽獎加牌 *Roll(none; 抽獎群組編號; N)===

func commandHandRoll(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardRoll(game, arg, cores.ContainerHand)
}

func commandDeckRoll(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardRoll(game, arg, cores.ContainerDeck)
}

func commandDropRoll(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardRoll(game, arg, cores.ContainerDrop)
}

func commandExileRoll(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardRoll(game, arg, cores.ContainerExile)
}

// === 淺 / 深複製 *Copy / *Clone(對象; N; [洗牌(僅 deck)]; 附加效果編號 1..M)===

func commandHandCopy(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, cores.ContainerHand, false)
}

func commandDeckCopy(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, cores.ContainerDeck, false)
}

func commandDropCopy(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, cores.ContainerDrop, false)
}

func commandExileCopy(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, cores.ContainerExile, false)
}

func commandHandClone(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, cores.ContainerHand, true)
}

func commandDeckClone(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, cores.ContainerDeck, true)
}

func commandDropClone(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, cores.ContainerDrop, true)
}

func commandExileClone(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	copyClone(game, target, arg, cores.ContainerExile, true)
}

// === 顧客實例化 ===

// commandGuestSpawn 依顧客編號實例化 1 位新顧客(【二十五 | guestSpawn】處理流程):
// 座位編號 = 0 → 加入遊蕩列表並自動鎖(sate / sateSeal / calmSeal 各 +1);
// 座位編號 = N>0 且該座位存在且空 → 設 seatID 加入座位列表(否則 no-op); 座位編號 < 0 / 顧客資料不存在 → no-op。
func commandGuestSpawn(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
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

	if seatID > 0 && (game.GetSheet().Seat.Get(seatID) == nil || game.Seat[seatID] != nil) {
		return // 座位不存在 / 已占用 → no-op(先於實例化判定, 不白費實例編號)
	} // if

	guest := cores.NewGuest(game, guestID)

	if guest == nil {
		return // 顧客資料不存在 → no-op
	} // if

	if seatID == 0 {
		guest.RoamLock()
		game.Roam.Push(guest)
		emitGuestMove(game, guest, cores.ContainerNone, cores.ContainerRoam, 0) // 新建直入遊蕩
		return
	} // if

	game.Seat.Place(seatID, guest)
	emitGuestMove(game, guest, cores.ContainerNone, cores.ContainerSeat, seatID) // 新建直入座位
}

// commandWaitAdd 依顧客編號實例化 N 位新顧客插入排隊佇列前端(優先入座); N <= 0 / 顧客資料不存在 → no-op(【二十五 | waitAdd】)。
func commandWaitAdd(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	guestID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	n, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	for itor := int32(0); itor < n; itor++ {
		guest := cores.NewGuest(game, guestID)

		if guest == nil {
			return // 顧客資料不存在 → no-op
		} // if

		game.Wait.Insert(guest)
		emitGuestMove(game, guest, cores.ContainerNone, cores.ContainerWait, 0) // 新建直入排隊
	} // for
}

// === 實例化輔助 ===

// cardAdd 依卡牌編號實例化 N 張加入 dest(none 命令對象); N <= 0 / 卡牌資料不存在 → no-op。
func cardAdd(game *cores.Game, arg []exprs.Value, dest cores.ContainerKind) {
	cardID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	n, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	for itor := int32(0); itor < n; itor++ {
		card := cores.NewCard(game, cardID)

		if card == nil {
			return // 卡牌資料不存在 → no-op
		} // if

		placeCard(game, cores.ContainerNone, dest, card) // 新建直入
	} // for
}

// cardRoll 對抽獎群組 weighted random 選編號實例化 N 張加入 dest(每張獨立 roll); N <= 0 / 群組總權重 0 → no-op。
func cardRoll(game *cores.Game, arg []exprs.Value, dest cores.ContainerKind) {
	group, ok := argInt(arg)

	if ok == false {
		return
	} // if

	n, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	for itor := int32(0); itor < n; itor++ {
		cardID, rolled := game.RollCard(group)

		if rolled == false {
			return // 群組總權重 0 / 群組不存在 → no-op
		} // if

		card := cores.NewCard(game, cardID)

		if card == nil {
			continue // 抽中編號無資料 → 跳過該張(防禦)
		} // if

		placeCard(game, cores.ContainerNone, dest, card) // 新建直入
	} // for
}

// copyClone 對命令對象每張卡牌(source)淺 / 深複製 N 張加入 dest 並附加 trailing 效果編號;
// dest 為抽牌牌堆時讀洗牌參數(index 1)、附加效果自 index 2, 其餘容器附加效果自 index 1。
func copyClone(game *cores.Game, target []cores.InstanceID, arg []exprs.Value, dest cores.ContainerKind, deep bool) {
	n, ok := argInt(arg)

	if ok == false {
		return
	} // if

	effectStart := 1
	shuffle := false

	if dest == cores.ContainerDeck {
		shuffle = argBool(argTail(arg, 1))
		effectStart = 2
	} // if

	addEffect := intList(argTail(arg, effectStart))

	for _, itor := range target {
		source, _, found := game.LocateCard(itor)

		if found == false {
			continue // 非卡牌實例 → 該項 no-op
		} // if

		for count := int32(0); count < n; count++ {
			card := cores.CopyCard(game, source, deep)

			if card == nil {
				continue // 淺複製載入失敗(卡牌資料不存在)→ 跳過該張
			} // if

			card.GetEffectID().Add(addEffect...)
			placeCard(game, cores.ContainerNone, dest, card) // 新建直入
		} // for
	} // for

	if shuffle {
		shuffleCard(game, game.Deck)
	} // if
}
