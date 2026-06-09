package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// attrReadFunc 全域屬性詞條的讀取行為:以 Engine 為 context 求值;arg 供查詢函式型屬性(deckSize…),純屬性忽略。
type attrReadFunc func(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool)

// attrRead 全域屬性值讀取詞彙表(名稱 → 讀取行為);服務 exprs.Resolver。
// 涵蓋【營業規格書 | 二十三、屬性清單】主表:純值 / 容器大小 / 衍生 / 物件引用 / 查詢函式。
// 鎖定計數(Lock 後綴)另置 attrLockRead,由 Engine.Attr 剝後綴路由。
// 每一詞條對應一個獨立的 read* 函式(便於逐條單元測試);本表僅作名稱 → 行為的索引。
var attrRead = map[string]attrReadFunc{
	// 餐廳 / 出牌全域數值屬性
	"morale":       readMorale,
	"moraleMax":    readMoraleMax,
	"moraleShield": readMoraleShield,
	"moraleBlock":  readMoraleBlock,
	"score":        readScore,
	"energy":       readEnergy,
	"energyMax":    readEnergyMax,
	"energyKeep":   readEnergyKeep,
	"handMax":      readHandMax,
	"drawMax":      readDrawMax,

	// 階段 / 回合
	"nextPhase": readNextPhase,
	"round":     readRound,
	"roundMax":  readRoundMax,
	"roundLeft": readRoundLeft,

	// 士氣受損事件
	"damageValue": readDamageValue,
	"damageGuest": readDamageGuest,

	// 入座 / 離場 / 行動事件
	"seatLast":     readSeatLast,
	"seatCount":    readSeatCount,
	"exitLast":     readExitLast,
	"exitLastSeat": readExitLastSeat,
	"exitCount":    readExitCount,
	"taskGuest":    readTaskGuest,
	"taskSkill":    readTaskSkill,
	"taskCount":    readTaskCount,

	// 卡牌事件:回合計數 / 最後引用 / 變身編號
	"drawLast":   readDrawLast,
	"drawCount":  readDrawCount,
	"dropLast":   readDropLast,
	"dropCount":  readDropCount,
	"playLast":   readPlayLast,
	"playCount":  readPlayCount,
	"exileLast":  readExileLast,
	"exileCount": readExileCount,
	"morphLast":  readMorphLast,
	"morphCount": readMorphCount,
	"morphOldID": readMorphOldID,
	"morphNewID": readMorphNewID,

	// 容器當下大小
	"seatSize":    readSeatSize,
	"waitSize":    readWaitSize,
	"roamSize":    readRoamSize,
	"cardifySize": readCardifySize,
	"taskSize":    readTaskSize,

	// 衍生 / self
	"guestSize": readGuestSize,
	"self":      readSelf,

	// 靜態座位佈局衍生
	"seatLeft":  readSeatLeft,
	"tableSize": readTableSize,

	// 查詢函式:桌次
	"tableGuest": readTableGuest,
	"tableCount": readTableCount,

	// 查詢函式:各牌堆 / 手牌的卡牌群組張數
	"handSize":  readHandSize,
	"deckSize":  readDeckSize,
	"dropSize":  readDropSize,
	"exileSize": readExileSize,

	// 查詢函式:整場累積分組張數
	"drawTotal":  readDrawTotal,
	"dropTotal":  readDropTotal,
	"playTotal":  readPlayTotal,
	"exileTotal": readExileTotal,
}

// attrLockRead 全域屬性鎖定計數讀取詞彙表(基底名 → 讀取行為);僅含【二十三】存取欄為「寫鎖 / 鎖」的屬性。
// Engine.Attr 於值表未命中且名稱以 Lock 結尾時,剝後綴查本表。
var attrLockRead = map[string]attrReadFunc{
	"morale":       readMoraleLock,
	"moraleMax":    readMoraleMaxLock,
	"moraleShield": readMoraleShieldLock,
	"moraleBlock":  readMoraleBlockLock,
	"score":        readScoreLock,
	"energy":       readEnergyLock,
	"energyMax":    readEnergyMaxLock,
	"energyKeep":   readEnergyKeepLock,
	"handMax":      readHandMaxLock,
	"drawMax":      readDrawMaxLock,
}

// === 餐廳 / 出牌全域數值屬性 ===

// readMorale 讀餐廳士氣值。
func readMorale(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.Morale.Value)), true
}

// readMoraleMax 讀餐廳士氣上限。
func readMoraleMax(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.MoraleMax.Value)), true
}

// readMoraleShield 讀餐廳士氣護盾。
func readMoraleShield(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.MoraleShield.Value)), true
}

// readMoraleBlock 讀餐廳士氣阻擋。
func readMoraleBlock(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.MoraleBlock.Value)), true
}

// readScore 讀餐廳分數。
func readScore(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.Score.Value)), true
}

// readEnergy 讀出牌能量。
func readEnergy(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.Energy.Value)), true
}

// readEnergyMax 讀出牌能量上限。
func readEnergyMax(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.EnergyMax.Value)), true
}

// readEnergyKeep 讀能量保留量。
func readEnergyKeep(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.EnergyKeep.Value)), true
}

// readHandMax 讀手牌上限。
func readHandMax(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.HandMax.Value)), true
}

// readDrawMax 讀每回合補牌上限。
func readDrawMax(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.DrawMax.Value)), true
}

// === 階段 / 回合 ===

// readNextPhase 讀下一階段(文字)。
func readNextPhase(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewText(string(eng.runtime.Game.NextPhase)), true
}

// readRound 讀當前回合數。
func readRound(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.Round)), true
}

// readRoundMax 讀回合上限。
func readRoundMax(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.RoundMax)), true
}

// readRoundLeft 讀剩餘回合數(上限 - 當前,夾 0)。
func readRoundLeft(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	game := eng.runtime.Game
	return exprs.NewNum(float64(max(int32(0), game.RoundMax-game.Round))), true
}

// === 士氣受損事件 ===

// readDamageValue 讀本次士氣受損量。
func readDamageValue(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.DamageValue)), true
}

// readDamageGuest 讀造成士氣受損的顧客引用。
func readDamageGuest(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestValue(eng.runtime.Game.DamageGuest), true
}

// === 入座 / 離場 / 行動事件 ===

// readSeatLast 讀最後入座的顧客引用。
func readSeatLast(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestValue(eng.runtime.Game.SeatLast), true
}

// readSeatCount 讀本回合入座計數。
func readSeatCount(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.SeatCount)), true
}

// readExitLast 讀最後離場的顧客引用。
func readExitLast(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestValue(eng.runtime.Game.ExitLast), true
}

// readExitLastSeat 讀最後離場顧客的座位編號。
func readExitLastSeat(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.ExitLastSeat)), true
}

// readExitCount 讀本回合離場計數。
func readExitCount(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.ExitCount)), true
}

// readTaskGuest 讀當前行動的顧客引用。
func readTaskGuest(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestValue(eng.runtime.Game.TaskGuest), true
}

// readTaskSkill 讀當前行動的技能編號。
func readTaskSkill(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.TaskSkill)), true
}

// readTaskCount 讀本回合行動計數。
func readTaskCount(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.TaskCount)), true
}

// === 卡牌事件:回合計數 / 最後引用 / 變身編號 ===

// readDrawLast 讀最後抽到的卡牌引用。
func readDrawLast(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardValue(eng.runtime.Game.DrawLast), true
}

// readDrawCount 讀本回合抽牌計數。
func readDrawCount(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.DrawCount)), true
}

// readDropLast 讀最後棄置的卡牌引用。
func readDropLast(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardValue(eng.runtime.Game.DropLast), true
}

// readDropCount 讀本回合棄牌計數。
func readDropCount(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.DropCount)), true
}

// readPlayLast 讀最後打出的卡牌引用。
func readPlayLast(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardValue(eng.runtime.Game.PlayLast), true
}

// readPlayCount 讀本回合出牌計數。
func readPlayCount(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.PlayCount)), true
}

// readExileLast 讀最後流放的卡牌引用。
func readExileLast(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardValue(eng.runtime.Game.ExileLast), true
}

// readExileCount 讀本回合流放計數。
func readExileCount(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.ExileCount)), true
}

// readMorphLast 讀最後變身的卡牌引用。
func readMorphLast(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardValue(eng.runtime.Game.MorphLast), true
}

// readMorphCount 讀本回合變身計數。
func readMorphCount(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.MorphCount)), true
}

// readMorphOldID 讀變身前的卡牌編號。
func readMorphOldID(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.MorphOldID)), true
}

// readMorphNewID 讀變身後的卡牌編號。
func readMorphNewID(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.MorphNewID)), true
}

// === 容器當下大小 ===

// readSeatSize 讀入座顧客數。
func readSeatSize(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(eng.runtime.Seat))), true
}

// readWaitSize 讀排隊顧客數。
func readWaitSize(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(eng.runtime.Wait))), true
}

// readRoamSize 讀遊蕩顧客數。
func readRoamSize(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(eng.runtime.Roam))), true
}

// readCardifySize 讀卡牌化顧客數。
func readCardifySize(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(eng.runtime.Cardify))), true
}

// readTaskSize 讀行動佇列長度。
func readTaskSize(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(eng.runtime.Action))), true
}

// === 衍生 / self ===

// readGuestSize 讀店內顧客總數(座位 + 排隊 + 遊蕩 + 卡牌化)。
func readGuestSize(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	r := eng.runtime
	return exprs.NewNum(float64(len(r.Seat) + len(r.Wait) + len(r.Roam) + len(r.Cardify))), true
}

// readSelf 讀 self:綁定顧客 / 卡牌回引用、綁定空物件回 none、未固定回失敗。
func readSelf(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return selfValue(eng.self)
}

// === 靜態座位佈局衍生 ===

// readSeatLeft 讀剩餘空座位數(座位總數 - 已占用,夾 0)。
func readSeatLeft(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	total := int32(len(eng.data.Seat.Keys()))
	return exprs.NewNum(float64(max(int32(0), total-int32(len(eng.runtime.Seat))))), true
}

// readTableSize 讀桌次總數(座位表中相異 TableID 數)。
func readTableSize(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	table := map[int32]bool{}

	for _, seatID := range eng.data.Seat.Keys() {
		meta := eng.data.Seat.Get(seatID)

		if meta != nil {
			table[meta.TableID] = true
		} // if
	} // for
	return exprs.NewNum(float64(len(table))), true
}

// === 查詢函式:桌次 ===

// readTableGuest 讀桌次 N 的入座顧客數(N == 0 不命中任何桌次)。
func readTableGuest(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	if n == 0 { // N == 0 不命中任何桌次
		return exprs.NewNum(0), true
	} // if
	count := int32(0)

	for seatID := range eng.runtime.Seat {
		meta := eng.data.Seat.Get(seatID)

		if meta != nil && meta.TableID == n {
			count++
		} // if
	} // for
	return exprs.NewNum(float64(count)), true
}

// readTableCount 讀占用人數與 K 滿足運算符 op 的桌數(arg = [op 文字, K 數值])。
func readTableCount(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	if len(arg) != 2 {
		return exprs.Value{}, false
	} // if

	if arg[0].IsText() == false || arg[1].IsNum() == false {
		return exprs.Value{}, false
	} // if
	op := arg[0].Text()
	k := int32(arg[1].Num())

	table := map[int32]bool{} // 列舉全部桌次(含 0 人桌)

	for _, seatID := range eng.data.Seat.Keys() {
		meta := eng.data.Seat.Get(seatID)

		if meta != nil {
			table[meta.TableID] = true
		} // if
	} // for
	occupied := map[int32]int32{} // 每桌占用人數

	for seatID := range eng.runtime.Seat {
		meta := eng.data.Seat.Get(seatID)

		if meta != nil {
			occupied[meta.TableID]++
		} // if
	} // for

	count := int32(0)

	for tableID := range table {
		match, valid := compareOp(op, occupied[tableID], k)

		if valid == false {
			return exprs.Value{}, false
		} // if

		if match {
			count++
		} // if
	} // for
	return exprs.NewNum(float64(count)), true
}

// === 查詢函式:各牌堆 / 手牌的卡牌群組張數 ===

// readHandSize 讀手牌中卡牌群組 N 的張數(N == 0 回全量)。
func readHandSize(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(eng.runtime.Hand, eng.data, arg)
}

// readDeckSize 讀抽牌牌堆中卡牌群組 N 的張數(N == 0 回全量)。
func readDeckSize(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(eng.runtime.Deck, eng.data, arg)
}

// readDropSize 讀棄牌牌堆中卡牌群組 N 的張數(N == 0 回全量)。
func readDropSize(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(eng.runtime.Drop, eng.data, arg)
}

// readExileSize 讀流放牌堆中卡牌群組 N 的張數(N == 0 回全量)。
func readExileSize(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(eng.runtime.Exile, eng.data, arg)
}

// === 查詢函式:整場累積分組張數 ===

// readDrawTotal 讀整場抽牌累積中群組 N 的張數(N == 0 回全加總)。
func readDrawTotal(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(eng.runtime.Game.DrawTotal, arg)
}

// readDropTotal 讀整場棄牌累積中群組 N 的張數(N == 0 回全加總)。
func readDropTotal(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(eng.runtime.Game.DropTotal, arg)
}

// readPlayTotal 讀整場出牌累積中群組 N 的張數(N == 0 回全加總)。
func readPlayTotal(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(eng.runtime.Game.PlayTotal, arg)
}

// readExileTotal 讀整場流放累積中群組 N 的張數(N == 0 回全加總)。
func readExileTotal(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(eng.runtime.Game.ExileTotal, arg)
}

// === 全域屬性鎖定計數(基底名 → .Lock) ===

// readMoraleLock 讀餐廳士氣的鎖定計數。
func readMoraleLock(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.Morale.Lock)), true
}

// readMoraleMaxLock 讀餐廳士氣上限的鎖定計數。
func readMoraleMaxLock(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.MoraleMax.Lock)), true
}

// readMoraleShieldLock 讀餐廳士氣護盾的鎖定計數。
func readMoraleShieldLock(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.MoraleShield.Lock)), true
}

// readMoraleBlockLock 讀餐廳士氣阻擋的鎖定計數。
func readMoraleBlockLock(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.MoraleBlock.Lock)), true
}

// readScoreLock 讀餐廳分數的鎖定計數。
func readScoreLock(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.Score.Lock)), true
}

// readEnergyLock 讀出牌能量的鎖定計數。
func readEnergyLock(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.Energy.Lock)), true
}

// readEnergyMaxLock 讀出牌能量上限的鎖定計數。
func readEnergyMaxLock(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.EnergyMax.Lock)), true
}

// readEnergyKeepLock 讀能量保留量的鎖定計數。
func readEnergyKeepLock(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.EnergyKeep.Lock)), true
}

// readHandMaxLock 讀手牌上限的鎖定計數。
func readHandMaxLock(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.HandMax.Lock)), true
}

// readDrawMaxLock 讀補牌上限的鎖定計數。
func readDrawMaxLock(eng *Engine, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(eng.runtime.Game.DrawMax.Lock)), true
}
