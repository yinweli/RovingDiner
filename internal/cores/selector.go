package cores

import (
	"sort"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

// selectorFunc 命令對象詞條的解析行為:以 Engine 為 context、arg 為 [...] 內已求值參數,產出作用對象集合(身分集)。
type selectorFunc func(eng *Engine, arg []exprs.Value) (result []InstanceID)

// selector 命令對象詞彙表(名稱 → 解析行為);對應【營業規格書 | 二十四、命令對象清單】。
// 每一詞條對應一個獨立的 select* 具名函式(比照讀寫詞彙表),容器來源 / filter / N 規則 / 座位鄰接 / auto-shuffle 共用 helper。
// 多數詞條只讀;deckTop 為唯一在解析時會修改狀態者(auto-shuffle 補牌)。
var selector = map[string]selectorFunc{
	// 無對象 / self 系
	"none":     selectNone,
	"self":     selectSelf,
	"selfNear": selectSelfNear,
	"selfSame": selectSelfSame,

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

	// 顧客:座位群與排隊
	"guestAll":  selectGuestAll,
	"guestPick": selectGuestPick,
	"guestRand": selectGuestRand,
	"guestWait": selectGuestWait,

	// 顧客:鄰桌 / 同桌(先選 1 錨點顧客,再展開其鄰 / 同桌)
	"nearPick": selectNearPick,
	"nearRand": selectNearRand,
	"samePick": selectSamePick,
	"sameRand": selectSameRand,

	// 卡牌容器:手牌 / 抽牌 / 棄牌 / 流放(All 取全 + filter、Pick 玩家選、Rand 隨機)
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

// HasSelector 回報命令對象詞彙表是否登錄 name;供 games.Validate 檢查命令對象。
func HasSelector(name string) bool {
	_, ok := selector[name]
	return ok
}

// === 無對象 / self 系 ===

// selectNone 無命令對象(空集合)。
func selectNone(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return nil
}

// selectSelf 取建立時固定的 self;未綁定 / 空物件回空集合,綁定卡牌 / 顧客回該單一實例。
func selectSelf(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	if eng.self == nil {
		return nil
	} // if

	if eng.self.Card != nil {
		return []InstanceID{eng.self.Card.InstanceID}
	} // if

	if eng.self.Guest != nil {
		return []InstanceID{eng.self.Guest.InstanceID}
	} // if

	return nil
}

// selectSelfNear 取 self 的鄰桌顧客(不含自身與同桌);self 非顧客時為空集合。
func selectSelfNear(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	if eng.self == nil || eng.self.Guest == nil {
		return nil
	} // if

	return guestIDs(eng.nearOf(eng.self.Guest))
}

// selectSelfSame 取 self 的同桌顧客(含自身);self 非顧客時為空集合。
func selectSelfSame(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	if eng.self == nil || eng.self.Guest == nil {
		return nil
	} // if

	return guestIDs(eng.sameOf(eng.self.Guest))
}

// === 事件單例 ===

// selectDamageGuest 取最近一次士氣受損來源顧客;來源非顧客 / 尚無來源時為空集合。
func selectDamageGuest(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return guestOne(eng.runtime.Game.DamageGuest)
}

// selectDrawLast 取最後抽出卡牌;無則為空集合。
func selectDrawLast(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return cardOne(eng.runtime.Game.DrawLast)
}

// selectDropLast 取最後棄置卡牌;無則為空集合。
func selectDropLast(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return cardOne(eng.runtime.Game.DropLast)
}

// selectExileLast 取最後流放卡牌;無則為空集合。
func selectExileLast(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return cardOne(eng.runtime.Game.ExileLast)
}

// selectExitLast 取最後離場顧客;無則為空集合。
func selectExitLast(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return guestOne(eng.runtime.Game.ExitLast)
}

// selectMorphLast 取最後變身卡牌;無則為空集合。
func selectMorphLast(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return cardOne(eng.runtime.Game.MorphLast)
}

// selectPlayLast 取最後出牌卡牌;無則為空集合。
func selectPlayLast(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return cardOne(eng.runtime.Game.PlayLast)
}

// selectSeatLast 取最後入座顧客;無則為空集合。
func selectSeatLast(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return guestOne(eng.runtime.Game.SeatLast)
}

// selectTaskGuest 取最後行動顧客;無則為空集合。
func selectTaskGuest(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return guestOne(eng.runtime.Game.TaskGuest)
}

// === 顧客:座位群與排隊 ===

// selectGuestAll 取當下全部在座顧客(僅座位列表)。
func selectGuestAll(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return guestIDs(seatGuest(eng))
}

// selectGuestPick 暫停流程,玩家從在座顧客挑最多 N 位(N 規則詳見【二十四｜N 規則】)。
func selectGuestPick(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	n, ok := oneInt(arg)

	if ok == false {
		return nil
	} // if

	return eng.pickGuest(seatGuest(eng), n)
}

// selectGuestRand 系統從在座顧客隨機選最多 N 位。
func selectGuestRand(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	n, ok := oneInt(arg)

	if ok == false {
		return nil
	} // if

	return eng.randTake(guestIDs(seatGuest(eng)), n)
}

// selectGuestWait 取排隊佇列前 N 位(先進先出,前端為隊首)。
func selectGuestWait(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	n, ok := oneInt(arg)

	if ok == false {
		return nil
	} // if

	return guestIDs(top(eng.runtime.Wait, n))
}

// === 顧客:鄰桌 / 同桌 ===

// selectNearPick 暫停流程,玩家選 1 在座顧客,再取其鄰桌(不含自身與同桌)。
func selectNearPick(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	anchor := eng.pickGuestOne(seatGuest(eng))

	if anchor == nil {
		return nil
	} // if

	return guestIDs(eng.nearOf(anchor))
}

// selectNearRand 系統隨機選 1 在座顧客,再取其鄰桌(不含自身與同桌)。
func selectNearRand(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	anchor := eng.randGuestOne(seatGuest(eng))

	if anchor == nil {
		return nil
	} // if

	return guestIDs(eng.nearOf(anchor))
}

// selectSamePick 暫停流程,玩家選 1 在座顧客,再取其同桌(含自身)。
func selectSamePick(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	anchor := eng.pickGuestOne(seatGuest(eng))

	if anchor == nil {
		return nil
	} // if

	return guestIDs(eng.sameOf(anchor))
}

// selectSameRand 系統隨機選 1 在座顧客,再取其同桌(含自身)。
func selectSameRand(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	anchor := eng.randGuestOne(seatGuest(eng))

	if anchor == nil {
		return nil
	} // if

	return guestIDs(eng.sameOf(anchor))
}

// === 卡牌容器:手牌 / 抽牌 / 棄牌 / 流放 ===

// selectHandAll 取手牌;依卡牌編號 filter(0 取全部、X 篩出編號 X)。
func selectHandAll(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return containerAll(eng.runtime.Hand, arg)
}

// selectHandPick 暫停流程,玩家從手牌(依編號 filter 後)挑最多 N 張。
func selectHandPick(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return eng.containerPick(eng.runtime.Hand, arg)
}

// selectHandRand 系統從手牌(依編號 filter 後)隨機選最多 N 張。
func selectHandRand(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return eng.containerRand(eng.runtime.Hand, arg)
}

// selectDeckAll 取抽牌牌堆;filter 規則同 handAll。
func selectDeckAll(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return containerAll(eng.runtime.Deck, arg)
}

// selectDeckPick 暫停流程,玩家從抽牌牌堆(依編號 filter 後)挑最多 N 張。
func selectDeckPick(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return eng.containerPick(eng.runtime.Deck, arg)
}

// selectDeckRand 系統從抽牌牌堆(依編號 filter 後)隨機選最多 N 張。
func selectDeckRand(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return eng.containerRand(eng.runtime.Deck, arg)
}

// selectDeckTop 取抽牌牌堆頂端(前端)N 張;不足時依【二十四｜deckTop auto-shuffle 規則】洗棄牌補回。
// 唯一在解析時會修改狀態的命令對象:deck < N 且 drop 非空時,將棄牌洗牌移至牌堆底端(尾端)再取頂端。
func selectDeckTop(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	n, ok := oneInt(arg)

	if ok == false || n <= 0 { // 壞參數 / N <= 0(Top 對 N <= 0 皆視為空集合)
		return nil
	} // if

	runtime := eng.runtime

	if int(n) > len(runtime.Deck) && len(runtime.Drop) > 0 {
		eng.shuffleCard(runtime.Drop)
		runtime.Deck = append(runtime.Deck, runtime.Drop...) // 洗後棄牌置底端(尾端)、保留原牌堆於頂端
		runtime.Drop = nil
	} // if

	return cardIDs(top(runtime.Deck, n))
}

// selectDropAll 取棄牌牌堆;filter 規則同 handAll。
func selectDropAll(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return containerAll(eng.runtime.Drop, arg)
}

// selectDropPick 暫停流程,玩家從棄牌牌堆(依編號 filter 後)挑最多 N 張。
func selectDropPick(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return eng.containerPick(eng.runtime.Drop, arg)
}

// selectDropRand 系統從棄牌牌堆(依編號 filter 後)隨機選最多 N 張。
func selectDropRand(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return eng.containerRand(eng.runtime.Drop, arg)
}

// selectDropTop 取棄牌牌堆頂端(前端)N 張;不足時取全部(不洗牌)。
func selectDropTop(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	n, ok := oneInt(arg)

	if ok == false {
		return nil
	} // if

	return cardIDs(top(eng.runtime.Drop, n))
}

// selectExileAll 取流放牌堆;filter 規則同 handAll。
func selectExileAll(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return containerAll(eng.runtime.Exile, arg)
}

// selectExilePick 暫停流程,玩家從流放牌堆(依編號 filter 後)挑最多 N 張。
func selectExilePick(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return eng.containerPick(eng.runtime.Exile, arg)
}

// selectExileRand 系統從流放牌堆(依編號 filter 後)隨機選最多 N 張。
func selectExileRand(eng *Engine, arg []exprs.Value) (result []InstanceID) {
	return eng.containerRand(eng.runtime.Exile, arg)
}

// === 命令對象解析輔助 ===

// containerAll 取卡牌容器全量並依卡牌編號 filter(arg 為單一卡牌編號;0 取全部、X 篩出編號 X)。
func containerAll(card []*Card, arg []exprs.Value) (result []InstanceID) {
	cardID, ok := oneInt(arg)

	if ok == false {
		return nil
	} // if

	return cardIDs(filterCardID(card, cardID))
}

// containerPick 暫停流程,玩家從卡牌容器(依編號 filter 後)挑最多 N 張(arg 為 N、卡牌編號)。
func (this *Engine) containerPick(card []*Card, arg []exprs.Value) (result []InstanceID) {
	n, cardID, ok := twoInt(arg)

	if ok == false {
		return nil
	} // if

	return this.pickCard(filterCardID(card, cardID), n)
}

// containerRand 系統從卡牌容器(依編號 filter 後)隨機選最多 N 張(arg 為 N、卡牌編號)。
func (this *Engine) containerRand(card []*Card, arg []exprs.Value) (result []InstanceID) {
	n, cardID, ok := twoInt(arg)

	if ok == false {
		return nil
	} // if

	return this.randTake(cardIDs(filterCardID(card, cardID)), n)
}

// filterCardID 依卡牌編號篩選卡牌容器:cardID == 0 視為任意(回原容器)、cardID == X 篩出 CardID == X 者。
func filterCardID(card []*Card, cardID int32) (result []*Card) {
	if cardID == 0 {
		return card
	} // if

	for _, itor := range card {
		if itor.CardID == cardID {
			result = append(result, itor)
		} // if
	} // for

	return result
}

// cardIDs 取卡牌切片的實例編號集。
func cardIDs(card []*Card) (result []InstanceID) {
	for _, itor := range card {
		result = append(result, itor.InstanceID)
	} // for

	return result
}

// guestIDs 取顧客切片的實例編號集。
func guestIDs(guest []*Guest) (result []InstanceID) {
	for _, itor := range guest {
		result = append(result, itor.InstanceID)
	} // for

	return result
}

// cardOne 把單一卡牌包成身分集:nil 回空集合、否則回單元素集。供事件單例命令對象共用。
func cardOne(card *Card) (result []InstanceID) {
	if card == nil {
		return nil
	} // if

	return []InstanceID{card.InstanceID}
}

// guestOne 把單一顧客包成身分集:nil 回空集合、否則回單元素集。供事件單例命令對象共用。
func guestOne(guest *Guest) (result []InstanceID) {
	if guest == nil {
		return nil
	} // if

	return []InstanceID{guest.InstanceID}
}

// top 取切片頂端(前端)N 個:N <= 0 回空集合、N >= 長度回全部、否則回前 N 個。供 deckTop / dropTop / guestWait 共用。
func top[T any](item []T, n int32) (result []T) {
	if n <= 0 {
		return nil
	} // if

	if int(n) >= len(item) {
		return item
	} // if

	return item[:n]
}

// seatGuest 取座位列表中全部顧客,依座位編號升序(消除 map 迭代無序、確保 Pick / Rand 候選決定性)。
func seatGuest(eng *Engine) (result []*Guest) {
	seatID := make([]int32, 0, len(eng.runtime.Seat))

	for k := range eng.runtime.Seat {
		seatID = append(seatID, k)
	} // for

	sort.Slice(seatID, func(i, j int) bool { return seatID[i] < seatID[j] })

	for _, itor := range seatID {
		result = append(result, eng.runtime.Seat[itor])
	} // for

	return result
}

// seatOccupant 取指定座位編號列表中有顧客占用者,依座位編號升序(複製輸入後排序,不動靜態座位表)。
func seatOccupant(eng *Engine, seatID []int32) (result []*Guest) {
	id := append([]int32(nil), seatID...)
	sort.Slice(id, func(i, j int) bool { return id[i] < id[j] })

	for _, itor := range id {
		if guest := eng.runtime.Seat[itor]; guest != nil {
			result = append(result, guest)
		} // if
	} // for

	return result
}

// nearOf 取顧客鄰桌的占用顧客;非入座(座位不存在)回空集合。
func (this *Engine) nearOf(guest *Guest) (result []*Guest) {
	meta := this.data.Seat.Get(guest.SeatID)

	if meta == nil {
		return nil
	} // if

	return seatOccupant(this, meta.NearSeatID)
}

// sameOf 取顧客同桌的占用顧客;非入座(座位不存在)回空集合。
func (this *Engine) sameOf(guest *Guest) (result []*Guest) {
	meta := this.data.Seat.Get(guest.SeatID)

	if meta == nil {
		return nil
	} // if

	return seatOccupant(this, meta.SameSeatID)
}

// pickCard 套用 Pick 的 N 規則於卡牌候選:N <= 0 空集合;候選 <= N 退化取全部(不彈介面);否則委由 Operator 暫停玩家挑 N 張。
func (this *Engine) pickCard(card []*Card, n int32) (result []InstanceID) {
	if n <= 0 { // N = 0 空集合;N < 0 對 Pick 視為空集合
		return nil
	} // if

	if int(n) >= len(card) {
		return cardIDs(card)
	} // if

	return cardIDs(this.operator.PickCard(card, int(n)))
}

// pickGuest 套用 Pick 的 N 規則於顧客候選;規則同 pickCard。
func (this *Engine) pickGuest(guest []*Guest, n int32) (result []InstanceID) {
	if n <= 0 {
		return nil
	} // if

	if int(n) >= len(guest) {
		return guestIDs(guest)
	} // if

	return guestIDs(this.operator.PickGuest(guest, int(n)))
}

// pickGuestOne 取單一錨點顧客(near / samePick 用):空候選回 nil、單一候選退化直取(不彈介面)、否則 Operator 挑 1。
func (this *Engine) pickGuestOne(guest []*Guest) (result *Guest) {
	if len(guest) == 0 {
		return nil
	} // if

	if len(guest) == 1 {
		return guest[0]
	} // if

	chosen := this.operator.PickGuest(guest, 1)

	if len(chosen) == 0 {
		return nil
	} // if

	return chosen[0]
}

// randGuestOne 取單一錨點顧客(near / sameRand 用):空候選回 nil、否則 Rander 隨機挑 1。
func (this *Engine) randGuestOne(guest []*Guest) (result *Guest) {
	if len(guest) == 0 {
		return nil
	} // if

	return guest[this.rander.Intn(len(guest))]
}

// randTake 套用 Rand 的 N 規則於身分集候選:
// N = 0 空集合;N > 0 取 N(候選 <= N 退化取全部);N < 0 隨機保留 abs(N)、其餘取出(候選 <= abs(N) → 空集合)。
func (this *Engine) randTake(candidate []InstanceID, n int32) (result []InstanceID) {
	if n == 0 {
		return nil
	} // if

	if n > 0 {
		if int(n) >= len(candidate) {
			return candidate
		} // if

		return this.randSubset(candidate, int(n))
	} // if

	keep := int(-n)

	if keep >= len(candidate) {
		return nil
	} // if

	return this.randSubset(candidate, len(candidate)-keep)
}

// randSubset 經 Rander 洗牌自候選隨機取 k 個,輸出依候選原序(穩定),不更動候選切片。
func (this *Engine) randSubset(candidate []InstanceID, k int) (result []InstanceID) {
	index := make([]int, len(candidate))

	for itor := range index {
		index[itor] = itor
	} // for

	this.rander.Shuffle(len(index), func(i, j int) { index[i], index[j] = index[j], index[i] })

	pick := index[:k]
	sort.Ints(pick)

	result = make([]InstanceID, 0, k)

	for _, itor := range pick {
		result = append(result, candidate[itor])
	} // for

	return result
}

// shuffleCard 經 Rander 就地洗牌卡牌切片(供 deckTop auto-shuffle 用)。
func (this *Engine) shuffleCard(card []*Card) {
	this.rander.Shuffle(len(card), func(i, j int) { card[i], card[j] = card[j], card[i] })
}
