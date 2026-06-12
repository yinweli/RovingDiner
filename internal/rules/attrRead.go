package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// attrReadEntry 全域屬性讀詞條: 讀取行為 + 是否產出物件引用 + 參數數量。行為與 metadata 同詞條單一定義點
// (比照 selectorEntry; M28 R2 / R4A), 供 games.Validate / ValidateExpr 靜態查引用基底合法性與運算式詞彙。
type attrReadEntry struct {
	resolve cores.AttrReadFunc // 讀取行為(read* 具名函式)
	ref     bool               // 產出物件引用(對齊【營業規格書 | 二十三、屬性清單】類別欄為引用者; 引用左值的合法基底)
	arity   int                // 參數數量(查詢函式非 0; 一般屬性 0, 識別子與零參數呼叫等價)
}

// attrRead 全域屬性值讀取詞彙表(名稱 → 詞條); 服務 exprs.Resolver。
// 涵蓋【營業規格書 | 二十三、屬性清單】主表: 純值 / 容器大小 / 衍生 / 物件引用 / 查詢函式。
// 鎖定計數(Lock 後綴)以全名詞條登錄於本表尾段。
// 每一詞條對應一個獨立的 read* 函式(便於逐條單元測試); 本表僅作名稱 → 行為的索引。
var attrRead = map[string]attrReadEntry{
	// 餐廳 / 出牌全域數值屬性
	"morale":       {resolve: readMorale},
	"moraleMax":    {resolve: readMoraleMax},
	"moraleShield": {resolve: readMoraleShield},
	"moraleBlock":  {resolve: readMoraleBlock},
	"score":        {resolve: readScore},
	"energy":       {resolve: readEnergy},
	"energyMax":    {resolve: readEnergyMax},
	"energyKeep":   {resolve: readEnergyKeep},
	"handMax":      {resolve: readHandMax},
	"drawMax":      {resolve: readDrawMax},

	// 階段 / 回合
	"nextPhase": {resolve: readNextPhase},
	"round":     {resolve: readRound},
	"roundMax":  {resolve: readRoundMax},
	"roundLeft": {resolve: readRoundLeft},

	// 士氣受損事件
	"damageValue": {resolve: readDamageValue},
	"damageGuest": {resolve: readDamageGuest, ref: true},

	// 入座 / 離場 / 行動事件
	"seatLast":     {resolve: readSeatLast, ref: true},
	"seatCount":    {resolve: readSeatCount},
	"exitLast":     {resolve: readExitLast, ref: true},
	"exitLastSeat": {resolve: readExitLastSeat},
	"exitCount":    {resolve: readExitCount},
	"taskGuest":    {resolve: readTaskGuest, ref: true},
	"taskSkill":    {resolve: readTaskSkill},
	"taskCount":    {resolve: readTaskCount},

	// 卡牌事件: 回合計數 / 最後引用 / 變身編號
	"drawLast":   {resolve: readDrawLast, ref: true},
	"drawCount":  {resolve: readDrawCount},
	"dropLast":   {resolve: readDropLast, ref: true},
	"dropCount":  {resolve: readDropCount},
	"playLast":   {resolve: readPlayLast, ref: true},
	"playCount":  {resolve: readPlayCount},
	"exileLast":  {resolve: readExileLast, ref: true},
	"exileCount": {resolve: readExileCount},
	"morphLast":  {resolve: readMorphLast, ref: true},
	"morphCount": {resolve: readMorphCount},
	"morphOldID": {resolve: readMorphOldID},
	"morphNewID": {resolve: readMorphNewID},

	// 容器當下大小
	"seatSize":    {resolve: readSeatSize},
	"waitSize":    {resolve: readWaitSize},
	"roamSize":    {resolve: readRoamSize},
	"cardifySize": {resolve: readCardifySize},
	"taskSize":    {resolve: readTaskSize},

	// 衍生 / self
	"guestSize": {resolve: readGuestSize},
	"self":      {resolve: readSelf, ref: true},

	// 靜態座位佈局衍生
	"seatLeft":  {resolve: readSeatLeft},
	"tableSize": {resolve: readTableSize},

	// 查詢函式: 桌次
	"tableGuest": {resolve: readTableGuest, arity: 1},
	"tableCount": {resolve: readTableCount, arity: 2},

	// 查詢函式: 各牌堆 / 手牌的卡牌群組張數
	"handSize":  {resolve: readHandSize, arity: 1},
	"deckSize":  {resolve: readDeckSize, arity: 1},
	"dropSize":  {resolve: readDropSize, arity: 1},
	"exileSize": {resolve: readExileSize, arity: 1},

	// 查詢函式: 整場累積分組張數
	"drawTotal":  {resolve: readDrawTotal, arity: 1},
	"dropTotal":  {resolve: readDropTotal, arity: 1},
	"playTotal":  {resolve: readPlayTotal, arity: 1},
	"exileTotal": {resolve: readExileTotal, arity: 1},

	// 鎖定計數讀取(Lock 全名詞條; 僅【二十三】存取欄為「寫鎖 / 鎖」的屬性)
	"moraleLock":       {resolve: readMoraleLock},
	"moraleMaxLock":    {resolve: readMoraleMaxLock},
	"moraleShieldLock": {resolve: readMoraleShieldLock},
	"moraleBlockLock":  {resolve: readMoraleBlockLock},
	"scoreLock":        {resolve: readScoreLock},
	"energyLock":       {resolve: readEnergyLock},
	"energyMaxLock":    {resolve: readEnergyMaxLock},
	"energyKeepLock":   {resolve: readEnergyKeepLock},
	"handMaxLock":      {resolve: readHandMaxLock},
	"drawMaxLock":      {resolve: readDrawMaxLock},
}

// HasObjectRef 回報全域屬性讀詞彙表是否登錄 name 且詞條產出物件引用; 供 games.Validate 檢查
// 引用左值的引用基底合法性(如 morale.cost 的 morale 非引用、未知名稱皆回 false; 執行期仍寬鬆 no-op)。
func HasObjectRef(name string) bool {
	entry, ok := attrRead[name]
	return ok && entry.ref
}

// AttrReadArity 回報全域屬性讀詞條的參數數量(查無回 ok=false); 供 games.ValidateExpr 靜態校驗
// 運算式識別子 / 查詢函式(M28 R4A; 一般屬性 0、查詢函式依詞條, 執行期仍寬鬆評估失敗)。
func AttrReadArity(name string) (arity int, ok bool) {
	entry, okEntry := attrRead[name]

	if okEntry == false {
		return 0, false
	} // if

	return entry.arity, true
}

// === 餐廳 / 出牌全域數值屬性 ===

// readMorale 讀餐廳士氣值。
func readMorale(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMorale().GetValue())), true
}

// readMoraleMax 讀餐廳士氣上限。
func readMoraleMax(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleMax().GetValue())), true
}

// readMoraleShield 讀餐廳士氣護盾。
func readMoraleShield(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleShield().GetValue())), true
}

// readMoraleBlock 讀餐廳士氣阻擋。
func readMoraleBlock(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleBlock().GetValue())), true
}

// readScore 讀餐廳滿意值。
func readScore(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetScore().GetValue())), true
}

// readEnergy 讀出牌點數。
func readEnergy(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergy().GetValue())), true
}

// readEnergyMax 讀出牌點數上限。
func readEnergyMax(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergyMax().GetValue())), true
}

// readEnergyKeep 讀出牌點數保留。
func readEnergyKeep(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergyKeep().GetValue())), true
}

// readHandMax 讀手牌上限。
func readHandMax(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetHandMax().GetValue())), true
}

// readDrawMax 讀每回合補牌上限。
func readDrawMax(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetDrawMax().GetValue())), true
}

// === 階段 / 回合 ===

// readNextPhase 讀下一階段(文字)。
func readNextPhase(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewText(string(game.GetNextPhase())), true
}

// readRound 讀當前回合數。
func readRound(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetRound().GetValue())), true
}

// readRoundMax 讀回合上限。
func readRoundMax(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetRoundMax().GetValue())), true
}

// readRoundLeft 讀剩餘回合數(上限 - 當前, 夾 0)。
func readRoundLeft(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(max(int32(0), game.GetRoundMax().GetValue()-game.GetRound().GetValue()))), true
}

// === 士氣受損事件 ===

// readDamageValue 讀本次士氣受損量。
func readDamageValue(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetDamageValue())), true
}

// readDamageGuest 讀造成士氣受損的顧客引用。
func readDamageGuest(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cores.NewRefGuest(game.GetDamageGuest()).Value(), true
}

// === 入座 / 離場 / 行動事件 ===

// readSeatLast 讀最後入座的顧客引用。
func readSeatLast(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cores.NewRefGuest(game.GetSeatLast()).Value(), true
}

// readSeatCount 讀本回合入座計數。
func readSeatCount(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetSeatCount())), true
}

// readExitLast 讀最後離場的顧客引用。
func readExitLast(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cores.NewRefGuest(game.GetExitLast()).Value(), true
}

// readExitLastSeat 讀最後離場顧客的座位編號。
func readExitLastSeat(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetExitLastSeat())), true
}

// readExitCount 讀本回合離場計數。
func readExitCount(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetExitCount())), true
}

// readTaskGuest 讀當前行動的顧客引用。
func readTaskGuest(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cores.NewRefGuest(game.GetTaskGuest()).Value(), true
}

// readTaskSkill 讀當前行動的技能編號。
func readTaskSkill(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetTaskSkill())), true
}

// readTaskCount 讀本回合行動計數。
func readTaskCount(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetTaskCount())), true
}

// === 卡牌事件: 回合計數 / 最後引用 / 變身編號 ===

// readDrawLast 讀最後抽到的卡牌引用。
func readDrawLast(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cores.NewRefCard(game.GetDrawLast()).Value(), true
}

// readDrawCount 讀本回合抽牌計數。
func readDrawCount(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetDrawCount())), true
}

// readDropLast 讀最後棄置的卡牌引用。
func readDropLast(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cores.NewRefCard(game.GetDropLast()).Value(), true
}

// readDropCount 讀本回合棄牌計數。
func readDropCount(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetDropCount())), true
}

// readPlayLast 讀最後打出的卡牌引用。
func readPlayLast(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cores.NewRefCard(game.GetPlayLast()).Value(), true
}

// readPlayCount 讀本回合出牌計數。
func readPlayCount(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetPlayCount())), true
}

// readExileLast 讀最後流放的卡牌引用。
func readExileLast(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cores.NewRefCard(game.GetExileLast()).Value(), true
}

// readExileCount 讀本回合流放計數。
func readExileCount(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetExileCount())), true
}

// readMorphLast 讀最後變身的卡牌引用。
func readMorphLast(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cores.NewRefCard(game.GetMorphLast()).Value(), true
}

// readMorphCount 讀本回合變身計數。
func readMorphCount(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMorphCount())), true
}

// readMorphOldID 讀變身前的卡牌編號。
func readMorphOldID(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMorphOldID())), true
}

// readMorphNewID 讀變身後的卡牌編號。
func readMorphNewID(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMorphNewID())), true
}

// === 容器當下大小 ===

// readSeatSize 讀入座顧客數。
func readSeatSize(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Seat))), true
}

// readWaitSize 讀排隊顧客數。
func readWaitSize(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Wait))), true
}

// readRoamSize 讀遊蕩顧客數。
func readRoamSize(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Roam))), true
}

// readCardifySize 讀卡牌化顧客數。
func readCardifySize(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Cardify))), true
}

// readTaskSize 讀行動佇列長度。
func readTaskSize(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Action))), true
}

// === 衍生 / self ===

// readGuestSize 讀店內顧客總數(座位 + 排隊 + 遊蕩 + 卡牌化)。
func readGuestSize(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(len(game.Seat) + len(game.Wait) + len(game.Roam) + len(game.Cardify))), true
}

// readSelf 讀 self: 綁定顧客 / 卡牌回引用、綁定空物件回 none、未固定(nil)回失敗。
// 對應【營業規格書 | 二十七、運算式 | 7】「self 未固定」與「綁定空物件」兩態之別。
func readSelf(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	if game.GetSelf() == nil {
		return exprs.Value{}, false
	} // if

	return game.GetSelf().Value(), true
}

// === 靜態座位佈局衍生 ===

// readSeatLeft 讀剩餘空座位數(座位總數 - 已占用, 夾 0)。
func readSeatLeft(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	total := int32(len(game.GetSheet().Seat.Keys()))
	return exprs.NewNum(float64(max(int32(0), total-int32(len(game.Seat))))), true
}

// readTableSize 讀桌次總數(座位表中相異 TableID 數)。
func readTableSize(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	table := map[int32]bool{}

	for _, seatID := range game.GetSheet().Seat.Keys() {
		meta := game.GetSheet().Seat.Get(seatID)

		if meta != nil {
			table[meta.TableID] = true
		} // if
	} // for

	return exprs.NewNum(float64(len(table))), true
}

// === 查詢函式: 桌次 ===

// readTableGuest 讀桌次 N 的入座顧客數(N == 0 不命中任何桌次)。
func readTableGuest(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	if n == 0 { // N == 0 不命中任何桌次
		return exprs.NewNum(0), true
	} // if
	count := int32(0)

	for seatID := range game.Seat {
		meta := game.GetSheet().Seat.Get(seatID)

		if meta != nil && meta.TableID == n {
			count++
		} // if
	} // for

	return exprs.NewNum(float64(count)), true
}

// readTableCount 讀占用人數與 K 滿足運算符 op 的桌數(arg = [op 文字, K 數值])。
func readTableCount(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	if len(arg) != 2 {
		return exprs.Value{}, false
	} // if

	if arg[0].IsText() == false || arg[1].IsNum() == false {
		return exprs.Value{}, false
	} // if
	op := arg[0].Text()
	k := int32(arg[1].Num())

	table := map[int32]bool{} // 列舉全部桌次(含 0 人桌)

	for _, seatID := range game.GetSheet().Seat.Keys() {
		meta := game.GetSheet().Seat.Get(seatID)

		if meta != nil {
			table[meta.TableID] = true
		} // if
	} // for
	occupied := map[int32]int32{} // 每桌占用人數

	for seatID := range game.Seat {
		meta := game.GetSheet().Seat.Get(seatID)

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

// === 查詢函式: 各牌堆 / 手牌的卡牌群組張數 ===

// readHandSize 讀手牌中卡牌群組 N 的張數(N == 0 回全量)。
func readHandSize(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(game.Hand, game.GetSheet(), arg)
}

// readDeckSize 讀抽牌牌堆中卡牌群組 N 的張數(N == 0 回全量)。
func readDeckSize(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(game.Deck, game.GetSheet(), arg)
}

// readDropSize 讀棄牌牌堆中卡牌群組 N 的張數(N == 0 回全量)。
func readDropSize(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(game.Drop, game.GetSheet(), arg)
}

// readExileSize 讀流放牌堆中卡牌群組 N 的張數(N == 0 回全量)。
func readExileSize(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupSize(game.Exile, game.GetSheet(), arg)
}

// === 查詢函式: 整場累積分組張數 ===

// readDrawTotal 讀整場抽牌累積中群組 N 的張數(N == 0 回全加總)。
func readDrawTotal(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(game.GetDrawTotal(), arg)
}

// readDropTotal 讀整場棄牌累積中群組 N 的張數(N == 0 回全加總)。
func readDropTotal(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(game.GetDropTotal(), arg)
}

// readPlayTotal 讀整場出牌累積中群組 N 的張數(N == 0 回全加總)。
func readPlayTotal(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(game.GetPlayTotal(), arg)
}

// readExileTotal 讀整場流放累積中群組 N 的張數(N == 0 回全加總)。
func readExileTotal(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return groupTotal(game.GetExileTotal(), arg)
}

// === 全域屬性鎖定計數(基底名 → .GetLock()) ===

// readMoraleLock 讀餐廳士氣的鎖定計數。
func readMoraleLock(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMorale().GetLock())), true
}

// readMoraleMaxLock 讀餐廳士氣上限的鎖定計數。
func readMoraleMaxLock(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleMax().GetLock())), true
}

// readMoraleShieldLock 讀餐廳士氣護盾的鎖定計數。
func readMoraleShieldLock(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleShield().GetLock())), true
}

// readMoraleBlockLock 讀餐廳士氣阻擋的鎖定計數。
func readMoraleBlockLock(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetMoraleBlock().GetLock())), true
}

// readScoreLock 讀餐廳滿意值的鎖定計數。
func readScoreLock(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetScore().GetLock())), true
}

// readEnergyLock 讀出牌點數的鎖定計數。
func readEnergyLock(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergy().GetLock())), true
}

// readEnergyMaxLock 讀出牌點數上限的鎖定計數。
func readEnergyMaxLock(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergyMax().GetLock())), true
}

// readEnergyKeepLock 讀出牌點數保留的鎖定計數。
func readEnergyKeepLock(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetEnergyKeep().GetLock())), true
}

// readHandMaxLock 讀手牌上限的鎖定計數。
func readHandMaxLock(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetHandMax().GetLock())), true
}

// readDrawMaxLock 讀補牌上限的鎖定計數。
func readDrawMaxLock(game *cores.Game, arg []exprs.Value) (result exprs.Value, ok bool) {
	return exprs.NewNum(float64(game.GetDrawMax().GetLock())), true
}
