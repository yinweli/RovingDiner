package rules

import (
	"sort"

	"github.com/yinweli/RovingDiner/internal/cores"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

// selector 命令對象詞彙表(名稱 → 解析行為); 對應【營業規格書 | 二十四、命令對象清單】。
// 每一詞條對應一個獨立的 select* 具名函式(比照讀寫詞彙表), 容器來源 / filter / N 規則 / 座位鄰接 / auto-shuffle 共用 helper。
// 多數詞條只讀; deckTop 為唯一在解析時會修改狀態者(auto-shuffle 補牌)。
var selector = map[string]cores.SelectorFunc{
	// 無對象 / self 系
	cores.SelectorNone: selectNone,
	"self":             selectSelf,
	"selfNear":         selectSelfNear,
	"selfSame":         selectSelfSame,

	// 事件單例(取 Game 上最近一次事件的引用)
	"damageGuest": selectDamageGuest,
	"drawLast":    selectDrawLast,
	"dropLast":    selectDropLast,
	"exileLast":   selectExileLast,
	"exitLast":    selectExitLast,
	"morphLast":   selectMorphLast,
	"playLast":    selectPlayLast,
	"seatLast":    selectSeatLast,
	"taskGuest":   selectTaskGuest,

	// 顧客: 座位群與排隊
	"guestAll":  selectGuestAll,
	"guestPick": selectGuestPick,
	"guestRand": selectGuestRand,
	"guestWait": selectGuestWait,

	// 顧客: 鄰桌 / 同桌(先選 1 錨點顧客, 再展開其鄰 / 同桌)
	"nearPick": selectNearPick,
	"nearRand": selectNearRand,
	"samePick": selectSamePick,
	"sameRand": selectSameRand,

	// 卡牌容器: 手牌 / 抽牌 / 棄牌 / 流放(All 取全 + filter、Pick 玩家選、Rand 隨機)
	"handAll":   selectHandAll,
	"handPick":  selectHandPick,
	"handRand":  selectHandRand,
	"deckAll":   selectDeckAll,
	"deckPick":  selectDeckPick,
	"deckRand":  selectDeckRand,
	"deckTop":   selectDeckTop,
	"dropAll":   selectDropAll,
	"dropPick":  selectDropPick,
	"dropRand":  selectDropRand,
	"dropTop":   selectDropTop,
	"exileAll":  selectExileAll,
	"exilePick": selectExilePick,
	"exileRand": selectExileRand,
}

// HasSelector 回報命令對象詞彙表是否登錄 name; 供 games.Validate 檢查命令對象。
func HasSelector(name string) bool {
	_, ok := selector[name]
	return ok
}

// === 無對象 / self 系 ===

// selectNone 無命令對象(空集合); 保留名 cores.SelectorNone, ExecOperate 對它把 target 正規化為 nil(全域掃描記號)。
func selectNone(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return nil
}

// selectSelf 取建立時固定的 self; 未綁定 / 空物件回空集合, 綁定卡牌 / 顧客回該單一實例。
func selectSelf(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	if game.GetSelf() == nil {
		return nil
	} // if

	if game.GetSelf().GetCard() != nil {
		return []cores.InstanceID{game.GetSelf().GetCard().GetInstanceID()}
	} // if

	if game.GetSelf().GetGuest() != nil {
		return []cores.InstanceID{game.GetSelf().GetGuest().GetInstanceID()}
	} // if

	return nil
}

// selectSelfNear 取 self 的鄰桌顧客(不含自身與同桌); self 非顧客時為空集合。
func selectSelfNear(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	if game.GetSelf() == nil || game.GetSelf().GetGuest() == nil {
		return nil
	} // if

	return guestIDs(nearOf(game, game.GetSelf().GetGuest()))
}

// selectSelfSame 取 self 的同桌顧客(含自身); self 非顧客時為空集合。
func selectSelfSame(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	if game.GetSelf() == nil || game.GetSelf().GetGuest() == nil {
		return nil
	} // if

	return guestIDs(sameOf(game, game.GetSelf().GetGuest()))
}

// === 事件單例 ===

// selectDamageGuest 取最近一次士氣受損來源顧客; 來源非顧客 / 尚無來源時為空集合。
func selectDamageGuest(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return guestOne(game.GetDamageGuest())
}

// selectDrawLast 取最後抽出卡牌; 無則為空集合。
func selectDrawLast(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return cardOne(game.GetDrawLast())
}

// selectDropLast 取最後棄置卡牌; 無則為空集合。
func selectDropLast(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return cardOne(game.GetDropLast())
}

// selectExileLast 取最後流放卡牌; 無則為空集合。
func selectExileLast(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return cardOne(game.GetExileLast())
}

// selectExitLast 取最後離場顧客; 無則為空集合。
func selectExitLast(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return guestOne(game.GetExitLast())
}

// selectMorphLast 取最後變身卡牌; 無則為空集合。
func selectMorphLast(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return cardOne(game.GetMorphLast())
}

// selectPlayLast 取最後出牌卡牌; 無則為空集合。
func selectPlayLast(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return cardOne(game.GetPlayLast())
}

// selectSeatLast 取最後入座顧客; 無則為空集合。
func selectSeatLast(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return guestOne(game.GetSeatLast())
}

// selectTaskGuest 取最後行動顧客; 無則為空集合。
func selectTaskGuest(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return guestOne(game.GetTaskGuest())
}

// === 顧客: 座位群與排隊 ===

// selectGuestAll 取當下全部在座顧客(僅座位列表)。
func selectGuestAll(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return guestIDs(game.Seat.Sorted())
}

// selectGuestPick 暫停流程, 玩家從在座顧客挑最多 N 位(N 規則詳見【二十四｜N 規則】)。
func selectGuestPick(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	n, ok := oneInt(arg)

	if ok == false {
		return nil
	} // if

	return pickGuest(game, game.Seat.Sorted(), n, "guestPick")
}

// selectGuestRand 系統從在座顧客隨機選最多 N 位。
func selectGuestRand(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	n, ok := oneInt(arg)

	if ok == false {
		return nil
	} // if

	return randTake(game, guestIDs(game.Seat.Sorted()), n)
}

// selectGuestWait 取排隊佇列前 N 位(先進先出, 前端為隊首)。
func selectGuestWait(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	n, ok := oneInt(arg)

	if ok == false {
		return nil
	} // if

	return guestIDs(top(game.Wait, n))
}

// === 顧客: 鄰桌 / 同桌 ===

// selectNearPick 暫停流程, 玩家選 1 在座顧客, 再取其鄰桌(不含自身與同桌)。
func selectNearPick(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	anchor := pickGuestOne(game, game.Seat.Sorted(), "nearPick")

	if anchor == nil {
		return nil
	} // if

	return guestIDs(nearOf(game, anchor))
}

// selectNearRand 系統隨機選 1 在座顧客, 再取其鄰桌(不含自身與同桌)。
func selectNearRand(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	anchor := randGuestOne(game, game.Seat.Sorted())

	if anchor == nil {
		return nil
	} // if

	return guestIDs(nearOf(game, anchor))
}

// selectSamePick 暫停流程, 玩家選 1 在座顧客, 再取其同桌(含自身)。
func selectSamePick(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	anchor := pickGuestOne(game, game.Seat.Sorted(), "samePick")

	if anchor == nil {
		return nil
	} // if

	return guestIDs(sameOf(game, anchor))
}

// selectSameRand 系統隨機選 1 在座顧客, 再取其同桌(含自身)。
func selectSameRand(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	anchor := randGuestOne(game, game.Seat.Sorted())

	if anchor == nil {
		return nil
	} // if

	return guestIDs(sameOf(game, anchor))
}

// === 卡牌容器: 手牌 / 抽牌 / 棄牌 / 流放 ===

// selectHandAll 取手牌; 依卡牌編號 filter(0 取全部、X 篩出編號 X)。
func selectHandAll(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerAll(game.Hand, arg)
}

// selectHandPick 暫停流程, 玩家從手牌(依編號 filter 後)挑最多 N 張。
func selectHandPick(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerPick(game, game.Hand, arg, "handPick")
}

// selectHandRand 系統從手牌(依編號 filter 後)隨機選最多 N 張。
func selectHandRand(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerRand(game, game.Hand, arg)
}

// selectDeckAll 取抽牌牌堆; filter 規則同 handAll。
func selectDeckAll(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerAll(game.Deck, arg)
}

// selectDeckPick 暫停流程, 玩家從抽牌牌堆(依編號 filter 後)挑最多 N 張。
func selectDeckPick(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerPick(game, game.Deck, arg, "deckPick")
}

// selectDeckRand 系統從抽牌牌堆(依編號 filter 後)隨機選最多 N 張。
func selectDeckRand(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerRand(game, game.Deck, arg)
}

// selectDeckTop 取抽牌牌堆頂端(前端)N 張; 不足時依【二十四｜deckTop auto-shuffle 規則】洗棄牌補回。
// 唯一在解析時會修改狀態的命令對象: deck < N 且 drop 非空時, 將棄牌洗牌移至牌堆底端(尾端)再取頂端。
func selectDeckTop(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	n, ok := oneInt(arg)

	if ok == false || n <= 0 { // 壞參數 / N <= 0(Top 對 N <= 0 皆視為空集合)
		return nil
	} // if

	if int(n) > len(game.Deck) && len(game.Drop) > 0 {
		shuffleCard(game, game.Drop)
		back := game.Drop
		game.Deck = append(game.Deck, game.Drop...) // 洗後棄牌置底端(尾端)、保留原牌堆於頂端
		game.Drop = nil

		for _, itor := range back {
			emitCardMove(game, itor, cores.ContainerDrop, cores.ContainerDeck) // 洗回逐卡投影(成員真相, 發射序 = 洗後序; M21 拍板, 原 M18 洗回靜默作廢)
		} // for

		emitDeckOrder(game) // 重整快照(順序真相)
	} // if

	return cardIDs(top(game.Deck, n))
}

// selectDropAll 取棄牌牌堆; filter 規則同 handAll。
func selectDropAll(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerAll(game.Drop, arg)
}

// selectDropPick 暫停流程, 玩家從棄牌牌堆(依編號 filter 後)挑最多 N 張。
func selectDropPick(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerPick(game, game.Drop, arg, "dropPick")
}

// selectDropRand 系統從棄牌牌堆(依編號 filter 後)隨機選最多 N 張。
func selectDropRand(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerRand(game, game.Drop, arg)
}

// selectDropTop 取棄牌牌堆頂端(前端)N 張; 不足時取全部(不洗牌)。
func selectDropTop(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	n, ok := oneInt(arg)

	if ok == false {
		return nil
	} // if

	return cardIDs(top(game.Drop, n))
}

// selectExileAll 取流放牌堆; filter 規則同 handAll。
func selectExileAll(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerAll(game.Exile, arg)
}

// selectExilePick 暫停流程, 玩家從流放牌堆(依編號 filter 後)挑最多 N 張。
func selectExilePick(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerPick(game, game.Exile, arg, "exilePick")
}

// selectExileRand 系統從流放牌堆(依編號 filter 後)隨機選最多 N 張。
func selectExileRand(game *cores.Game, arg []exprs.Value) (result []cores.InstanceID) {
	return containerRand(game, game.Exile, arg)
}

// === 命令對象解析輔助 ===

// containerAll 取卡牌容器全量並依卡牌編號 filter(arg 為單一卡牌編號; 0 取全部、X 篩出編號 X)。
func containerAll(card []*cores.Card, arg []exprs.Value) (result []cores.InstanceID) {
	cardID, ok := oneInt(arg)

	if ok == false {
		return nil
	} // if

	return cardIDs(filterCardID(card, cardID))
}

// containerPick 暫停流程, 玩家從卡牌容器(依編號 filter 後)挑最多 N 張(arg 為 N、卡牌編號); source 為詞條鍵(玩家輸入紀錄用)。
func containerPick(game *cores.Game, card []*cores.Card, arg []exprs.Value, source string) (result []cores.InstanceID) {
	n, cardID, ok := twoInt(arg)

	if ok == false {
		return nil
	} // if

	return pickCard(game, filterCardID(card, cardID), n, source)
}

// containerRand 系統從卡牌容器(依編號 filter 後)隨機選最多 N 張(arg 為 N、卡牌編號)。
func containerRand(game *cores.Game, card []*cores.Card, arg []exprs.Value) (result []cores.InstanceID) {
	n, cardID, ok := twoInt(arg)

	if ok == false {
		return nil
	} // if

	return randTake(game, cardIDs(filterCardID(card, cardID)), n)
}

// filterCardID 依卡牌編號篩選卡牌容器: cardID == 0 視為任意(回原容器)、cardID == X 篩出 CardID == X 者。
func filterCardID(card []*cores.Card, cardID int32) (result []*cores.Card) {
	if cardID == 0 {
		return card
	} // if

	for _, itor := range card {
		if itor.GetCardID() == cardID {
			result = append(result, itor)
		} // if
	} // for

	return result
}

// cardIDs 取卡牌切片的實例編號集。
func cardIDs(card []*cores.Card) (result []cores.InstanceID) {
	for _, itor := range card {
		result = append(result, itor.GetInstanceID())
	} // for

	return result
}

// guestIDs 取顧客切片的實例編號集。
func guestIDs(guest []*cores.Guest) (result []cores.InstanceID) {
	for _, itor := range guest {
		result = append(result, itor.GetInstanceID())
	} // for

	return result
}

// cardOne 把單一卡牌包成身分集: nil 回空集合、否則回單元素集。供事件單例命令對象共用。
func cardOne(card *cores.Card) (result []cores.InstanceID) {
	if card == nil {
		return nil
	} // if

	return []cores.InstanceID{card.GetInstanceID()}
}

// guestOne 把單一顧客包成身分集: nil 回空集合、否則回單元素集。供事件單例命令對象共用。
func guestOne(guest *cores.Guest) (result []cores.InstanceID) {
	if guest == nil {
		return nil
	} // if

	return []cores.InstanceID{guest.GetInstanceID()}
}

// top 取切片頂端(前端)N 個: N <= 0 回空集合、N >= 長度回全部、否則回前 N 個。供 deckTop / dropTop / guestWait 共用。
func top[T any](item []T, n int32) (result []T) {
	if n <= 0 {
		return nil
	} // if

	if int(n) >= len(item) {
		return item
	} // if

	return item[:n]
}

// seatOccupant 取指定座位編號列表中有顧客占用者, 依座位編號升序(複製輸入後排序, 不動靜態座位表)。
func seatOccupant(game *cores.Game, seatID []int32) (result []*cores.Guest) {
	id := append([]int32(nil), seatID...)
	sort.Slice(id, func(i, j int) bool { return id[i] < id[j] })

	for _, itor := range id {
		if guest := game.Seat[itor]; guest != nil {
			result = append(result, guest)
		} // if
	} // for

	return result
}

// nearOf 取顧客鄰桌的占用顧客; 非入座(座位不存在)回空集合。
func nearOf(game *cores.Game, guest *cores.Guest) (result []*cores.Guest) {
	meta := game.GetSheet().Seat.Get(guest.GetSeatID())

	if meta == nil {
		return nil
	} // if

	return seatOccupant(game, meta.NearSeatID)
}

// sameOf 取顧客同桌的占用顧客; 非入座(座位不存在)回空集合。
func sameOf(game *cores.Game, guest *cores.Guest) (result []*cores.Guest) {
	meta := game.GetSheet().Seat.Get(guest.GetSeatID())

	if meta == nil {
		return nil
	} // if

	return seatOccupant(game, meta.SameSeatID)
}

// pickCard 套用 Pick 的 N 規則於卡牌候選: N <= 0 空集合; 候選 <= N 退化取全部(不彈介面); 否則委由 Operator 暫停玩家挑 N 張。
// 真選取發玩家輸入紀錄(source 為命令對象詞條鍵; 退化全取不經 Operator 不記; M18 拍板)。
func pickCard(game *cores.Game, card []*cores.Card, n int32, source string) (result []cores.InstanceID) {
	if n <= 0 { // N = 0 空集合; N < 0 對 Pick 視為空集合
		return nil
	} // if

	if int(n) >= len(card) {
		return cardIDs(card)
	} // if

	chosen := game.GetOperator().PickCard(card, int(n))
	emitSelect(game, source, 0, cardPickData(chosen))
	return cardIDs(chosen)
}

// pickGuest 套用 Pick 的 N 規則於顧客候選; 規則同 pickCard(含真選取的玩家輸入紀錄)。
func pickGuest(game *cores.Game, guest []*cores.Guest, n int32, source string) (result []cores.InstanceID) {
	if n <= 0 {
		return nil
	} // if

	if int(n) >= len(guest) {
		return guestIDs(guest)
	} // if

	chosen := game.GetOperator().PickGuest(guest, int(n))
	emitSelect(game, source, 0, guestPickData(chosen))
	return guestIDs(chosen)
}

// pickGuestOne 取單一錨點顧客(near / samePick 用): 空候選回 nil、單一候選退化直取(不彈介面)、否則 Operator 挑 1。
// 真選取發玩家輸入紀錄(規則同 pickCard)。
func pickGuestOne(game *cores.Game, guest []*cores.Guest, source string) (result *cores.Guest) {
	if len(guest) == 0 {
		return nil
	} // if

	if len(guest) == 1 {
		return guest[0]
	} // if

	chosen := game.GetOperator().PickGuest(guest, 1)
	emitSelect(game, source, 0, guestPickData(chosen))

	if len(chosen) == 0 {
		return nil
	} // if

	return chosen[0]
}

// randGuestOne 取單一錨點顧客(near / sameRand 用): 空候選回 nil、否則 Rander 隨機挑 1。
func randGuestOne(game *cores.Game, guest []*cores.Guest) (result *cores.Guest) {
	if len(guest) == 0 {
		return nil
	} // if

	return guest[game.GetRander().Intn(len(guest))]
}

// randTake 套用 Rand 的 N 規則於身分集候選:
// N = 0 空集合; N > 0 取 N(候選 <= N 退化取全部); N < 0 隨機保留 abs(N)、其餘取出(候選 <= abs(N) → 空集合)。
func randTake(game *cores.Game, candidate []cores.InstanceID, n int32) (result []cores.InstanceID) {
	if n == 0 {
		return nil
	} // if

	if n > 0 {
		if int(n) >= len(candidate) {
			return candidate
		} // if

		return randSubset(game, candidate, int(n))
	} // if

	keep := int(-n)

	if keep >= len(candidate) {
		return nil
	} // if

	return randSubset(game, candidate, len(candidate)-keep)
}

// randSubset 經 Rander 洗牌自候選隨機取 k 個, 輸出依候選原序(穩定), 不更動候選切片。
func randSubset[T any](game *cores.Game, candidate []T, k int) (result []T) {
	index := make([]int, len(candidate))

	for itor := range index {
		index[itor] = itor
	} // for

	game.GetRander().Shuffle(len(index), func(i, j int) { index[i], index[j] = index[j], index[i] })

	pick := index[:k]
	sort.Ints(pick)

	result = make([]T, 0, k)

	for _, itor := range pick {
		result = append(result, candidate[itor])
	} // for

	return result
}
