package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// attrReadFunc 全域屬性詞條的讀取行為:以 Game 為 context 求值;arg 供查詢函式型屬性(deckSize…),純屬性忽略。
type attrReadFunc func(game *Game, arg []exprs.Value) (result exprs.Value, ok bool)

// attrRead 全域屬性值讀取詞彙表(名稱 → 讀取行為);服務 exprs.Resolver。
// 涵蓋【營業規格書 | 二十三、屬性清單】主表:純值 / 容器大小 / 衍生 / 物件引用 / 查詢函式。
// 鎖定計數(Lock 後綴)另置 attrLockRead,由 Game.Attr 剝後綴路由。
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
// Game.Attr 於值表未命中且名稱以 Lock 結尾時,剝後綴查本表。
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
func readMorale(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMorale().GetValue())), true
}

// readMoraleMax 讀餐廳士氣上限。
func readMoraleMax(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleMax().GetValue())), true
}

// readMoraleShield 讀餐廳士氣護盾。
func readMoraleShield(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleShield().GetValue())), true
}

// readMoraleBlock 讀餐廳士氣阻擋。
func readMoraleBlock(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleBlock().GetValue())), true
}

// readScore 讀餐廳分數。
func readScore(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetScore().GetValue())), true
}

// readEnergy 讀出牌能量。
func readEnergy(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergy().GetValue())), true
}

// readEnergyMax 讀出牌能量上限。
func readEnergyMax(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergyMax().GetValue())), true
}

// readEnergyKeep 讀能量保留量。
func readEnergyKeep(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergyKeep().GetValue())), true
}

// readHandMax 讀手牌上限。
func readHandMax(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetHandMax().GetValue())), true
}

// readDrawMax 讀每回合補牌上限。
func readDrawMax(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetDrawMax().GetValue())), true
}

// === 階段 / 回合 ===

// readNextPhase 讀下一階段(文字)。
func readNextPhase(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewText(string(game.GetNextPhase())), true
}

// readRound 讀當前回合數。
func readRound(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetRound().GetValue())), true
}

// readRoundMax 讀回合上限。
func readRoundMax(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetRoundMax().GetValue())), true
}

// readRoundLeft 讀剩餘回合數(上限 - 當前,夾 0)。
func readRoundLeft(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(max(int32(0), game.GetRoundMax().GetValue()-game.GetRound().GetValue()))), true
}

// === 士氣受損事件 ===

// readDamageValue 讀本次士氣受損量。
func readDamageValue(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetDamageValue())), true
}

// readDamageGuest 讀造成士氣受損的顧客引用。
func readDamageGuest(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return NewRefGuest(game.GetDamageGuest()).Value(), true
}

// === 入座 / 離場 / 行動事件 ===

// readSeatLast 讀最後入座的顧客引用。
func readSeatLast(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return NewRefGuest(game.GetSeatLast()).Value(), true
}

// readSeatCount 讀本回合入座計數。
func readSeatCount(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetSeatCount())), true
}

// readExitLast 讀最後離場的顧客引用。
func readExitLast(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return NewRefGuest(game.GetExitLast()).Value(), true
}

// readExitLastSeat 讀最後離場顧客的座位編號。
func readExitLastSeat(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetExitLastSeat())), true
}

// readExitCount 讀本回合離場計數。
func readExitCount(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetExitCount())), true
}

// readTaskGuest 讀當前行動的顧客引用。
func readTaskGuest(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return NewRefGuest(game.GetTaskGuest()).Value(), true
}

// readTaskSkill 讀當前行動的技能編號。
func readTaskSkill(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetTaskSkill())), true
}

// readTaskCount 讀本回合行動計數。
func readTaskCount(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetTaskCount())), true
}

// === 卡牌事件:回合計數 / 最後引用 / 變身編號 ===

// readDrawLast 讀最後抽到的卡牌引用。
func readDrawLast(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return NewRefCard(game.GetDrawLast()).Value(), true
}

// readDrawCount 讀本回合抽牌計數。
func readDrawCount(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetDrawCount())), true
}

// readDropLast 讀最後棄置的卡牌引用。
func readDropLast(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return NewRefCard(game.GetDropLast()).Value(), true
}

// readDropCount 讀本回合棄牌計數。
func readDropCount(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetDropCount())), true
}

// readPlayLast 讀最後打出的卡牌引用。
func readPlayLast(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return NewRefCard(game.GetPlayLast()).Value(), true
}

// readPlayCount 讀本回合出牌計數。
func readPlayCount(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetPlayCount())), true
}

// readExileLast 讀最後流放的卡牌引用。
func readExileLast(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return NewRefCard(game.GetExileLast()).Value(), true
}

// readExileCount 讀本回合流放計數。
func readExileCount(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetExileCount())), true
}

// readMorphLast 讀最後變身的卡牌引用。
func readMorphLast(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return NewRefCard(game.GetMorphLast()).Value(), true
}

// readMorphCount 讀本回合變身計數。
func readMorphCount(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMorphCount())), true
}

// readMorphOldID 讀變身前的卡牌編號。
func readMorphOldID(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMorphOldID())), true
}

// readMorphNewID 讀變身後的卡牌編號。
func readMorphNewID(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMorphNewID())), true
}

// === 容器當下大小 ===

// readSeatSize 讀入座顧客數。
func readSeatSize(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Seat))), true
}

// readWaitSize 讀排隊顧客數。
func readWaitSize(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Wait))), true
}

// readRoamSize 讀遊蕩顧客數。
func readRoamSize(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Roam))), true
}

// readCardifySize 讀卡牌化顧客數。
func readCardifySize(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Cardify))), true
}

// readTaskSize 讀行動佇列長度。
func readTaskSize(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Action))), true
}

// === 衍生 / self ===

// readGuestSize 讀店內顧客總數(座位 + 排隊 + 遊蕩 + 卡牌化)。
func readGuestSize(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Seat) + len(game.Wait) + len(game.Roam) + len(game.Cardify))), true
}

// readSelf 讀 self:綁定顧客 / 卡牌回引用、綁定空物件回 none、未固定(nil)回失敗。
// 對應【營業規格書 | 二十七、運算式 | 7】「self 未固定」與「綁定空物件」兩態之別。
func readSelf(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	if game.self == nil {
		return exprs.Value{}, false
	} // if

	return game.self.Value(), true
}

// === 靜態座位佈局衍生 ===

// readSeatLeft 讀剩餘空座位數(座位總數 - 已占用,夾 0)。
func readSeatLeft(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	total := int32(len(game.data.Seat.Keys()))
	return exprs.NewNum(float64(max(int32(0), total-int32(len(game.Seat))))), true
}

// readTableSize 讀桌次總數(座位表中相異 TableID 數)。
func readTableSize(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	table := map[int32]bool{}

	for _, seatID := range game.data.Seat.Keys() {
		meta := game.data.Seat.Get(seatID)

		if meta != nil {
			table[meta.TableID] = true
		} // if
	} // for
	return exprs.NewNum(float64(len(table))), true
}

// === 查詢函式:桌次 ===

// readTableGuest 讀桌次 N 的入座顧客數(N == 0 不命中任何桌次)。
func readTableGuest(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	if n == 0 { // N == 0 不命中任何桌次
		return exprs.NewNum(0), true
	} // if
	count := int32(0)

	for seatID := range game.Seat {
		meta := game.data.Seat.Get(seatID)

		if meta != nil && meta.TableID == n {
			count++
		} // if
	} // for
	return exprs.NewNum(float64(count)), true
}

// readTableCount 讀占用人數與 K 滿足運算符 op 的桌數(arg = [op 文字, K 數值])。
func readTableCount(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	if len(arg) != 2 {
		return exprs.Value{}, false
	} // if

	if arg[0].IsText() == false || arg[1].IsNum() == false {
		return exprs.Value{}, false
	} // if
	op := arg[0].Text()
	k := int32(arg[1].Num())

	table := map[int32]bool{} // 列舉全部桌次(含 0 人桌)

	for _, seatID := range game.data.Seat.Keys() {
		meta := game.data.Seat.Get(seatID)

		if meta != nil {
			table[meta.TableID] = true
		} // if
	} // for
	occupied := map[int32]int32{} // 每桌占用人數

	for seatID := range game.Seat {
		meta := game.data.Seat.Get(seatID)

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
func readHandSize(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(game.Hand, game.data, arg)
}

// readDeckSize 讀抽牌牌堆中卡牌群組 N 的張數(N == 0 回全量)。
func readDeckSize(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(game.Deck, game.data, arg)
}

// readDropSize 讀棄牌牌堆中卡牌群組 N 的張數(N == 0 回全量)。
func readDropSize(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(game.Drop, game.data, arg)
}

// readExileSize 讀流放牌堆中卡牌群組 N 的張數(N == 0 回全量)。
func readExileSize(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(game.Exile, game.data, arg)
}

// === 查詢函式:整場累積分組張數 ===

// readDrawTotal 讀整場抽牌累積中群組 N 的張數(N == 0 回全加總)。
func readDrawTotal(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(game.GetDrawTotal(), arg)
}

// readDropTotal 讀整場棄牌累積中群組 N 的張數(N == 0 回全加總)。
func readDropTotal(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(game.GetDropTotal(), arg)
}

// readPlayTotal 讀整場出牌累積中群組 N 的張數(N == 0 回全加總)。
func readPlayTotal(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(game.GetPlayTotal(), arg)
}

// readExileTotal 讀整場流放累積中群組 N 的張數(N == 0 回全加總)。
func readExileTotal(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(game.GetExileTotal(), arg)
}

// === 全域屬性鎖定計數(基底名 → .GetLock()) ===

// readMoraleLock 讀餐廳士氣的鎖定計數。
func readMoraleLock(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMorale().GetLock())), true
}

// readMoraleMaxLock 讀餐廳士氣上限的鎖定計數。
func readMoraleMaxLock(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleMax().GetLock())), true
}

// readMoraleShieldLock 讀餐廳士氣護盾的鎖定計數。
func readMoraleShieldLock(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleShield().GetLock())), true
}

// readMoraleBlockLock 讀餐廳士氣阻擋的鎖定計數。
func readMoraleBlockLock(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleBlock().GetLock())), true
}

// readScoreLock 讀餐廳分數的鎖定計數。
func readScoreLock(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetScore().GetLock())), true
}

// readEnergyLock 讀出牌能量的鎖定計數。
func readEnergyLock(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergy().GetLock())), true
}

// readEnergyMaxLock 讀出牌能量上限的鎖定計數。
func readEnergyMaxLock(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergyMax().GetLock())), true
}

// readEnergyKeepLock 讀能量保留量的鎖定計數。
func readEnergyKeepLock(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergyKeep().GetLock())), true
}

// readHandMaxLock 讀手牌上限的鎖定計數。
func readHandMaxLock(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetHandMax().GetLock())), true
}

// readDrawMaxLock 讀補牌上限的鎖定計數。
func readDrawMaxLock(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetDrawMax().GetLock())), true
}
