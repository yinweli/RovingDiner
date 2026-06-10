package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteSelector(t *testing.T) {
	suite.Run(t, new(SuiteSelector))
}

// SuiteSelector 驗證命令對象詞彙表(selector.go):登錄查詢、無對象 / self / 事件單例 / 座位群 / 鄰桌同桌 / 容器 filter / N 規則 / deckTop auto-shuffle。
// Pick 用 tester.FakeOperator(取候選前綴)、Rand 用 tester.FakeRander(恆等洗牌 + Intn 固定)以求決定性。
type SuiteSelector struct {
	suite.Suite
}

func (this *SuiteSelector) TestSelectorHas() {
	this.True(HasSelector("deckTop"))
	this.True(HasSelector("handAll"))
	this.False(HasSelector("nope")) // 未登錄
}

func (this *SuiteSelector) TestSelectNone() {
	this.Empty(this.must(newGame(), "none", nil))
}

func (this *SuiteSelector) TestSelectSelf() {
	game := newGame()
	card := cores.NewCard(game, 101)
	guest := cores.NewGuest(game, 501)

	cardRef := cores.NewRefCard(card) // 綁定卡牌
	game.SetSelf(&cardRef)
	this.Equal(cardIDs([]*cores.Card{card}), this.must(game, "self", nil))

	guestRef := cores.NewRefGuest(guest) // 綁定顧客
	game.SetSelf(&guestRef)
	this.Equal(guestIDs([]*cores.Guest{guest}), this.must(game, "self", nil))

	none := cores.Ref{} // 綁定空物件
	game.SetSelf(&none)
	this.Empty(this.must(game, "self", nil))

	game.SetSelf(nil) // 未綁定
	this.Empty(this.must(game, "self", nil))
}

func (this *SuiteSelector) TestSelectSelfNearSame() {
	game := newGame()
	g1, g2, g3 := seatGuest(game)

	near := cores.NewRefGuest(g1)
	game.SetSelf(&near)
	this.Equal(guestIDs([]*cores.Guest{g3}), this.must(game, "selfNear", nil))     // g1(座1)鄰桌 = 座3 = g3
	this.Equal(guestIDs([]*cores.Guest{g1, g2}), this.must(game, "selfSame", nil)) // g1 同桌 = 座1,2 = g1,g2

	card := cores.NewRefCard(cores.NewCard(game, 101)) // self 非顧客 → 空集合
	game.SetSelf(&card)
	this.Empty(this.must(game, "selfNear", nil))
	this.Empty(this.must(game, "selfSame", nil))

	// self 顧客但未入座(SeatID=0、座位不存在)→ nearOf / sameOf meta 為 nil → 空集合
	roam := cores.NewRefGuest(&cores.Guest{})
	game.SetSelf(&roam)
	this.Empty(this.must(game, "selfNear", nil))
	this.Empty(this.must(game, "selfSame", nil))
}

func (this *SuiteSelector) TestSelectEventSingle() {
	game := newGame()
	card := cores.NewCard(game, 101)
	guest := cores.NewGuest(game, 501)
	game.EventDraw(card, 0)
	game.EventDrop(card, 0)
	game.EventExile(card, 0)
	game.EventMorph(card, 0, 0)
	game.EventPlay(card, 0)
	game.EventDamage(1, guest)
	game.EventExit(guest)
	game.EventSeat(guest)
	game.EventTask(guest, 0)

	for name, want := range map[string]cores.InstanceID{
		"drawLast": card.GetInstanceID(), "dropLast": card.GetInstanceID(), "exileLast": card.GetInstanceID(),
		"morphLast": card.GetInstanceID(), "playLast": card.GetInstanceID(),
		"damageGuest": guest.GetInstanceID(), "exitLast": guest.GetInstanceID(),
		"seatLast": guest.GetInstanceID(), "taskGuest": guest.GetInstanceID(),
	} {
		this.Equal([]cores.InstanceID{want}, this.must(game, name, nil), name)
	} // for

	// 無事件來源(欄位 nil)→ 空集合(cardOne / guestOne 的 nil 分支)
	empty := newGame()
	this.Empty(this.must(empty, "drawLast", nil))
	this.Empty(this.must(empty, "seatLast", nil))
}

func (this *SuiteSelector) TestSelectGuestAll() {
	game := newGame()
	g1, g2, g3 := seatGuest(game)
	this.Equal(guestIDs([]*cores.Guest{g1, g2, g3}), this.must(game, "guestAll", nil)) // 依座位編號升序
}

func (this *SuiteSelector) TestSelectGuestPick() {
	game := newGame()
	g1, g2, g3 := seatGuest(game)

	this.Equal(guestIDs([]*cores.Guest{g1, g2}), this.must(game, "guestPick", nums(2)))     // 候選 3 > 2 → Operator 取前 2
	this.Equal(guestIDs([]*cores.Guest{g1, g2, g3}), this.must(game, "guestPick", nums(5))) // 候選 3 <= 5 → 退化取全部(不彈介面)
	this.Empty(this.must(game, "guestPick", nums(0)))                                       // N = 0 → 空集合
	this.Empty(this.must(game, "guestPick", nums(-1)))                                      // N < 0 對 Pick → 空集合
	this.Empty(this.must(game, "guestPick", nil))                                           // 缺參數 → 空集合
}

func (this *SuiteSelector) TestSelectGuestRand() {
	game := newGame()
	g1, g2, g3 := seatGuest(game)

	this.Equal(guestIDs([]*cores.Guest{g1, g2, g3}), this.must(game, "guestRand", nums(3))) // 候選 3 <= 3 → 退化取全部
	this.Equal(guestIDs([]*cores.Guest{g1}), this.must(game, "guestRand", nums(1)))         // 候選 3 > 1 → 隨機取 1(恆等洗牌取首位)
	this.Empty(this.must(game, "guestRand", nums(0)))                                       // N = 0 → 空集合
	this.Equal(guestIDs([]*cores.Guest{g1, g2}), this.must(game, "guestRand", nums(-1)))    // N < 0 → 保留 1、取出其餘 2
	this.Empty(this.must(game, "guestRand", nums(-5)))                                      // 候選 <= abs(N) → 空集合
	this.Empty(this.must(game, "guestRand", nil))                                           // 缺參數 → 空集合
}

func (this *SuiteSelector) TestSelectGuestWait() {
	game := newGame()
	w1 := cores.NewGuest(game, 501)
	w2 := cores.NewGuest(game, 501)
	w3 := cores.NewGuest(game, 501)
	game.Wait = cores.WaitList{w1, w2, w3}

	this.Equal(guestIDs([]*cores.Guest{w1, w2}), this.must(game, "guestWait", nums(2)))     // 取隊首 2 位
	this.Equal(guestIDs([]*cores.Guest{w1, w2, w3}), this.must(game, "guestWait", nums(5))) // N >= 長度 → 取全部
	this.Empty(this.must(game, "guestWait", nums(0)))                                       // N <= 0 → 空集合
	this.Empty(this.must(game, "guestWait", nil))                                           // 缺參數 → 空集合
}

func (this *SuiteSelector) TestSelectNearSame() {
	game := newGame()
	g1, g2, g3 := seatGuest(game)

	this.Equal(guestIDs([]*cores.Guest{g3}), this.must(game, "nearPick", nil))     // Operator 選 g1(座1)→ 鄰桌座3 = g3
	this.Equal(guestIDs([]*cores.Guest{g3}), this.must(game, "nearRand", nil))     // Rander 選首位 g1 → 鄰桌 g3
	this.Equal(guestIDs([]*cores.Guest{g1, g2}), this.must(game, "samePick", nil)) // g1 同桌 = g1,g2
	this.Equal(guestIDs([]*cores.Guest{g1, g2}), this.must(game, "sameRand", nil))

	// 空座位:無錨點可選 → 空集合(pickGuestOne / randGuestOne 的空候選分支)
	noSeat := newGame()
	this.Empty(this.must(noSeat, "nearPick", nil))
	this.Empty(this.must(noSeat, "nearRand", nil))
	this.Empty(this.must(noSeat, "samePick", nil))
	this.Empty(this.must(noSeat, "sameRand", nil))

	// 單一在座顧客:pickGuestOne 退化直取(len==1)→ g2 同桌(座1,2)= g2
	one := newGame()
	alone := cores.NewGuest(one, 501)
	one.Seat.Place(2, alone)
	this.Equal(guestIDs([]*cores.Guest{alone}), this.must(one, "samePick", nil))

	// Operator 回空(防禦):候選 > 1 但選不出 → 空集合(pickGuestOne 的 chosen 空分支)
	emptyPick := cores.NewGame(0, 0, tester.BuildData(), tester.FakeOperator{EmptyPick: true}, tester.FakeRander{})
	Register(emptyPick)
	seatGuest(emptyPick)
	this.Empty(this.must(emptyPick, "nearPick", nil))
}

func (this *SuiteSelector) TestSelectContainerAll() {
	game := newGame()
	h1 := cores.NewCard(game, 101)
	h2 := cores.NewCard(game, 102)
	deck := cores.NewCard(game, 101)
	drop := cores.NewCard(game, 103)
	exiled := strayCard(104)
	game.Hand = cores.CardList{h1, h2}
	game.Deck = cores.CardList{deck}
	game.Drop = cores.CardList{drop}
	game.Exile = cores.CardList{exiled}

	this.Equal(cardIDs([]*cores.Card{h1, h2}), this.must(game, "handAll", nums(0))) // 編號 0 → 取全部
	this.Equal(cardIDs([]*cores.Card{h1}), this.must(game, "handAll", nums(101)))   // 編號 101 → 篩出
	this.Empty(this.must(game, "handAll", nums(999)))                               // 無此編號 → 空集合
	this.Empty(this.must(game, "handAll", nil))                                     // 缺參數 → 空集合

	this.Equal(cardIDs([]*cores.Card{deck}), this.must(game, "deckAll", nums(0))) // 各容器 All 變體一一驗證
	this.Equal(cardIDs([]*cores.Card{drop}), this.must(game, "dropAll", nums(0)))
	this.Equal(cardIDs([]*cores.Card{exiled}), this.must(game, "exileAll", nums(0)))
}

func (this *SuiteSelector) TestSelectContainerPick() {
	game := newGame()
	h1 := cores.NewCard(game, 101)
	h2 := cores.NewCard(game, 101)
	h3 := cores.NewCard(game, 101)
	deck := cores.NewCard(game, 101)
	drop := cores.NewCard(game, 101)
	exiled := cores.NewCard(game, 101)
	game.Hand = cores.CardList{h1, h2, h3}
	game.Deck = cores.CardList{deck}
	game.Drop = cores.CardList{drop}
	game.Exile = cores.CardList{exiled}

	this.Equal(cardIDs([]*cores.Card{h1, h2}), this.must(game, "handPick", nums(2, 0)))     // 候選 3 > 2 → Operator 取前 2
	this.Equal(cardIDs([]*cores.Card{h1, h2, h3}), this.must(game, "handPick", nums(5, 0))) // 候選 <= N → 退化取全部
	this.Empty(this.must(game, "handPick", nums(0, 0)))                                     // N = 0 → 空集合
	this.Empty(this.must(game, "handPick", nums(1)))                                        // 參數不足(缺卡牌編號)→ 空集合

	this.Equal(cardIDs([]*cores.Card{deck}), this.must(game, "deckPick", nums(1, 0))) // 各容器 Pick 變體一一驗證
	this.Equal(cardIDs([]*cores.Card{drop}), this.must(game, "dropPick", nums(1, 0)))
	this.Equal(cardIDs([]*cores.Card{exiled}), this.must(game, "exilePick", nums(1, 0)))
}

func (this *SuiteSelector) TestSelectContainerRand() {
	game := newGame()
	h1 := cores.NewCard(game, 101)
	h2 := cores.NewCard(game, 101)
	h3 := cores.NewCard(game, 101)
	deck := cores.NewCard(game, 101)
	drop := cores.NewCard(game, 101)
	exiled := cores.NewCard(game, 101)
	game.Hand = cores.CardList{h1, h2, h3}
	game.Deck = cores.CardList{deck}
	game.Drop = cores.CardList{drop}
	game.Exile = cores.CardList{exiled}

	this.Equal(cardIDs([]*cores.Card{h1, h2}), this.must(game, "handRand", nums(2, 0))) // 候選 3 > 2 → 隨機取 2(恆等洗牌取前綴)
	this.Empty(this.must(game, "handRand", nums(1)))                                    // 參數不足 → 空集合

	this.Equal(cardIDs([]*cores.Card{deck}), this.must(game, "deckRand", nums(1, 0))) // 各容器 Rand 變體一一驗證
	this.Equal(cardIDs([]*cores.Card{drop}), this.must(game, "dropRand", nums(1, 0)))
	this.Equal(cardIDs([]*cores.Card{exiled}), this.must(game, "exileRand", nums(1, 0)))
}

func (this *SuiteSelector) TestSelectDeckTop() {
	// deck >= N:取頂端(前端)N,不洗牌
	deckGe := newGame()
	d1 := cores.NewCard(deckGe, 101)
	d2 := cores.NewCard(deckGe, 101)
	deckGe.Deck = cores.CardList{d1, d2, cores.NewCard(deckGe, 101)}
	deckGe.Drop = cores.CardList{cores.NewCard(deckGe, 101)}
	this.Equal(cardIDs([]*cores.Card{d1, d2}), this.must(deckGe, "deckTop", nums(2)))
	this.Len(deckGe.Drop, 1) // 牌堆足夠 → 未洗牌移動

	// deck < N 且 deck+drop >= N:洗棄牌移至牌堆底端再取頂端 N
	deckShort := newGame()
	s1 := cores.NewCard(deckShort, 101)
	s4 := cores.NewCard(deckShort, 101)
	deckShort.Deck = cores.CardList{s1}
	deckShort.Drop = cores.CardList{s4, cores.NewCard(deckShort, 101)}
	this.Equal(cardIDs([]*cores.Card{s1, s4}), this.must(deckShort, "deckTop", nums(2))) // 原牌堆 s1 在頂、洗入的 s4 接其後
	this.Empty(deckShort.Drop)                                                           // 棄牌已全數移入

	// deck+drop < N:洗棄牌移入後取全部
	deckTiny := newGame()
	t1 := cores.NewCard(deckTiny, 101)
	t4 := cores.NewCard(deckTiny, 101)
	deckTiny.Deck = cores.CardList{t1}
	deckTiny.Drop = cores.CardList{t4}
	this.Equal(cardIDs([]*cores.Card{t1, t4}), this.must(deckTiny, "deckTop", nums(5)))

	// deck < N 但棄牌為空:不洗牌、取全部
	deckNoDrop := newGame()
	n1 := cores.NewCard(deckNoDrop, 101)
	deckNoDrop.Deck = cores.CardList{n1}
	this.Equal(cardIDs([]*cores.Card{n1}), this.must(deckNoDrop, "deckTop", nums(5)))

	// N <= 0 / 缺參數 → 空集合
	this.Empty(this.must(deckGe, "deckTop", nums(0)))
	this.Empty(this.must(deckGe, "deckTop", nil))
}

func (this *SuiteSelector) TestSelectDropTop() {
	game := newGame()
	d1 := cores.NewCard(game, 101)
	d2 := cores.NewCard(game, 101)
	game.Drop = cores.CardList{d1, d2, cores.NewCard(game, 101)}

	this.Equal(cardIDs([]*cores.Card{d1, d2}), this.must(game, "dropTop", nums(2)))               // 取頂端 2
	this.Equal(cardIDs([]*cores.Card{d1, d2, game.Drop[2]}), this.must(game, "dropTop", nums(9))) // 不足 → 取全部
	this.Empty(this.must(game, "dropTop", nil))                                                   // 缺參數 → 空集合
}

// === 測試輔助(置尾) ===

// must 自詞彙表取出命令對象詞條並斷言已登錄,派發回作用對象集合(聚焦於結果斷言)。
func (this *SuiteSelector) must(game *cores.Game, name string, arg []exprs.Value) (result []cores.InstanceID) {
	resolve, ok := selector[name]
	this.Require().True(ok, name)
	return resolve(game, arg)
}

// seatGuest 對營業實例佈置 3 位在座顧客(座1=g1 / 座2=g2 / 座3=g3;桌1=座1,2、桌2=座3,對齊迷你表座位佈局)。
func seatGuest(game *cores.Game) (g1, g2, g3 *cores.Guest) {
	g1 = cores.NewGuest(game, 501)
	g2 = cores.NewGuest(game, 501)
	g3 = cores.NewGuest(game, 501)
	game.Seat.Place(1, g1)
	game.Seat.Place(2, g2)
	game.Seat.Place(3, g3)
	return g1, g2, g3
}

// nums 把整數批次包成運算式數值參數列表(模擬 [...] 內已求值參數)。
func nums(n ...int) (result []exprs.Value) {
	for _, itor := range n {
		result = append(result, exprs.NewNum(float64(itor)))
	} // for

	return result
}
