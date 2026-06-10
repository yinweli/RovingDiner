package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteSelector(t *testing.T) {
	suite.Run(t, new(SuiteSelector))
}

// SuiteSelector 驗證命令對象詞彙表(selector.go):登錄查詢、無對象 / self / 事件單例 / 座位群 / 鄰桌同桌 / 容器 filter / N 規則 / deckTop auto-shuffle。
// 經 engine.selectObject 派發進入詞條;Pick 用 fakeOperator(取候選前綴)、Rand 用 fakeRander(恆等洗牌 + Intn 固定)以求決定性。
type SuiteSelector struct {
	suite.Suite
}

func (this *SuiteSelector) TestSelectorHas() {
	this.True(HasSelector("deckTop")) // M8 詞條已填入
	this.True(HasSelector("handAll"))
	this.False(HasSelector("nope")) // 未登錄
}

func (this *SuiteSelector) TestSelectNone() {
	eng := this.engine(NewRuntime(0), nil)
	this.Empty(this.must(eng, "none", nil))
}

func (this *SuiteSelector) TestSelectSelf() {
	this.Equal([]InstanceID{1}, this.must(this.engine(NewRuntime(0), &Ref{card: &Card{instanceID: 1}}), "self", nil))   // 綁定卡牌
	this.Equal([]InstanceID{2}, this.must(this.engine(NewRuntime(0), &Ref{guest: &Guest{instanceID: 2}}), "self", nil)) // 綁定顧客
	this.Empty(this.must(this.engine(NewRuntime(0), &Ref{}), "self", nil))                                              // 綁定空物件
	this.Empty(this.must(this.engine(NewRuntime(0), nil), "self", nil))                                                 // 未綁定
}

func (this *SuiteSelector) TestSelectSelfNearSame() {
	runtime, g1 := seatGuestRuntime()

	near := this.engine(runtime, &Ref{guest: g1})
	this.Equal([]InstanceID{13}, this.must(near, "selfNear", nil))     // g1(座1)鄰桌 = 座3 = g3
	this.Equal([]InstanceID{11, 12}, this.must(near, "selfSame", nil)) // g1 同桌 = 座1,2 = g1,g2

	card := this.engine(runtime, &Ref{card: &Card{instanceID: 9}})
	this.Empty(this.must(card, "selfNear", nil)) // self 非顧客 → 空集合
	this.Empty(this.must(card, "selfSame", nil))

	// self 顧客但未入座(SeatID=0、座位不存在)→ nearOf / sameOf meta 為 nil → 空集合
	roam := this.engine(runtime, &Ref{guest: &Guest{instanceID: 99, seatID: 0}})
	this.Empty(this.must(roam, "selfNear", nil))
	this.Empty(this.must(roam, "selfSame", nil))
}

func (this *SuiteSelector) TestSelectEventSingle() {
	card := &Card{instanceID: 1}
	guest := &Guest{instanceID: 2}
	runtime := NewRuntime(0)
	game := runtime.Game
	game.drawLast, game.dropLast, game.exileLast, game.morphLast, game.playLast = card, card, card, card, card
	game.damageGuest, game.exitLast, game.seatLast, game.taskGuest = guest, guest, guest, guest
	eng := this.engine(runtime, nil)

	for name, want := range map[string]InstanceID{
		"drawLast": 1, "dropLast": 1, "exileLast": 1, "morphLast": 1, "playLast": 1,
		"damageGuest": 2, "exitLast": 2, "seatLast": 2, "taskGuest": 2,
	} {
		this.Equal([]InstanceID{want}, this.must(eng, name, nil), name)
	} // for

	// 無事件來源(欄位 nil)→ 空集合(cardOne / guestOne 的 nil 分支)
	empty := this.engine(NewRuntime(0), nil)
	this.Empty(this.must(empty, "drawLast", nil))
	this.Empty(this.must(empty, "seatLast", nil))
}

func (this *SuiteSelector) TestSelectGuestAll() {
	runtime, _ := seatGuestRuntime()
	this.Equal([]InstanceID{11, 12, 13}, this.must(this.engine(runtime, nil), "guestAll", nil)) // 依座位編號升序
}

func (this *SuiteSelector) TestSelectGuestPick() {
	runtime, _ := seatGuestRuntime()
	eng := this.engine(runtime, nil)

	this.Equal([]InstanceID{11, 12}, this.must(eng, "guestPick", nums(2)))     // 候選 3 > 2 → Operator 取前 2
	this.Equal([]InstanceID{11, 12, 13}, this.must(eng, "guestPick", nums(5))) // 候選 3 <= 5 → 退化取全部(不彈介面)
	this.Empty(this.must(eng, "guestPick", nums(0)))                           // N = 0 → 空集合
	this.Empty(this.must(eng, "guestPick", nums(-1)))                          // N < 0 對 Pick → 空集合
	this.Empty(this.must(eng, "guestPick", nil))                               // 缺參數 → 空集合
}

func (this *SuiteSelector) TestSelectGuestRand() {
	runtime, _ := seatGuestRuntime()
	eng := this.engine(runtime, nil)

	this.Equal([]InstanceID{11, 12, 13}, this.must(eng, "guestRand", nums(3))) // 候選 3 <= 3 → 退化取全部
	this.Equal([]InstanceID{11}, this.must(eng, "guestRand", nums(1)))         // 候選 3 > 1 → 隨機取 1(恆等洗牌取首位)
	this.Empty(this.must(eng, "guestRand", nums(0)))                           // N = 0 → 空集合
	this.Equal([]InstanceID{11, 12}, this.must(eng, "guestRand", nums(-1)))    // N < 0 → 保留 1、取出其餘 2
	this.Empty(this.must(eng, "guestRand", nums(-5)))                          // 候選 <= abs(N) → 空集合
	this.Empty(this.must(eng, "guestRand", nil))                               // 缺參數 → 空集合
}

func (this *SuiteSelector) TestSelectGuestWait() {
	runtime := NewRuntime(0)
	runtime.Wait = []*Guest{{instanceID: 21}, {instanceID: 22}, {instanceID: 23}}
	eng := this.engine(runtime, nil)

	this.Equal([]InstanceID{21, 22}, this.must(eng, "guestWait", nums(2)))     // 取隊首 2 位
	this.Equal([]InstanceID{21, 22, 23}, this.must(eng, "guestWait", nums(5))) // N >= 長度 → 取全部
	this.Empty(this.must(eng, "guestWait", nums(0)))                           // N <= 0 → 空集合
	this.Empty(this.must(eng, "guestWait", nil))                               // 缺參數 → 空集合
}

func (this *SuiteSelector) TestSelectNearSame() {
	runtime, _ := seatGuestRuntime()
	eng := this.engine(runtime, nil)

	this.Equal([]InstanceID{13}, this.must(eng, "nearPick", nil))     // Operator 選 g1(座1)→ 鄰桌座3 = g3
	this.Equal([]InstanceID{13}, this.must(eng, "nearRand", nil))     // Rander 選首位 g1 → 鄰桌 g3
	this.Equal([]InstanceID{11, 12}, this.must(eng, "samePick", nil)) // g1 同桌 = g1,g2
	this.Equal([]InstanceID{11, 12}, this.must(eng, "sameRand", nil))

	// 空座位:無錨點可選 → 空集合(pickGuestOne / randGuestOne 的空候選分支)
	noSeat := this.engine(NewRuntime(0), nil)
	this.Empty(this.must(noSeat, "nearPick", nil))
	this.Empty(this.must(noSeat, "nearRand", nil))
	this.Empty(this.must(noSeat, "samePick", nil))
	this.Empty(this.must(noSeat, "sameRand", nil))

	// 單一在座顧客:pickGuestOne 退化直取(len==1)→ g2 同桌(座1,2)= g2
	one := NewRuntime(0)
	one.Seat[2] = &Guest{instanceID: 12, seatID: 2}
	this.Equal([]InstanceID{12}, this.must(this.engine(one, nil), "samePick", nil))

	// Operator 回空(防禦):候選 > 1 但選不出 → 空集合(pickGuestOne 的 chosen 空分支)
	emptyPick := NewEngine(runtime, nil, buildSheet(), fakeOperator{emptyPick: true}, fakeRander{}, nil)
	this.Empty(this.must(emptyPick, "nearPick", nil))
}

func (this *SuiteSelector) TestSelectContainerAll() {
	runtime := NewRuntime(0)
	runtime.Hand = []*Card{{instanceID: 1, cardID: 101}, {instanceID: 2, cardID: 102}}
	runtime.Deck = []*Card{{instanceID: 3, cardID: 101}}
	runtime.Drop = []*Card{{instanceID: 4, cardID: 103}}
	runtime.Exile = []*Card{{instanceID: 5, cardID: 104}}
	eng := this.engine(runtime, nil)

	this.Equal([]InstanceID{1, 2}, this.must(eng, "handAll", nums(0))) // 編號 0 → 取全部
	this.Equal([]InstanceID{1}, this.must(eng, "handAll", nums(101)))  // 編號 101 → 篩出
	this.Empty(this.must(eng, "handAll", nums(999)))                   // 無此編號 → 空集合
	this.Empty(this.must(eng, "handAll", nil))                         // 缺參數 → 空集合

	this.Equal([]InstanceID{3}, this.must(eng, "deckAll", nums(0))) // 各容器 All 變體一一驗證
	this.Equal([]InstanceID{4}, this.must(eng, "dropAll", nums(0)))
	this.Equal([]InstanceID{5}, this.must(eng, "exileAll", nums(0)))
}

func (this *SuiteSelector) TestSelectContainerPick() {
	runtime := NewRuntime(0)
	runtime.Hand = []*Card{{instanceID: 1}, {instanceID: 2}, {instanceID: 3}}
	runtime.Deck = []*Card{{instanceID: 4}}
	runtime.Drop = []*Card{{instanceID: 5}}
	runtime.Exile = []*Card{{instanceID: 6}}
	eng := this.engine(runtime, nil)

	this.Equal([]InstanceID{1, 2}, this.must(eng, "handPick", nums(2, 0)))    // 候選 3 > 2 → Operator 取前 2
	this.Equal([]InstanceID{1, 2, 3}, this.must(eng, "handPick", nums(5, 0))) // 候選 <= N → 退化取全部
	this.Empty(this.must(eng, "handPick", nums(0, 0)))                        // N = 0 → 空集合
	this.Empty(this.must(eng, "handPick", nums(1)))                           // 參數不足(缺卡牌編號)→ 空集合

	this.Equal([]InstanceID{4}, this.must(eng, "deckPick", nums(1, 0))) // 各容器 Pick 變體一一驗證
	this.Equal([]InstanceID{5}, this.must(eng, "dropPick", nums(1, 0)))
	this.Equal([]InstanceID{6}, this.must(eng, "exilePick", nums(1, 0)))
}

func (this *SuiteSelector) TestSelectContainerRand() {
	runtime := NewRuntime(0)
	runtime.Hand = []*Card{{instanceID: 1}, {instanceID: 2}, {instanceID: 3}}
	runtime.Deck = []*Card{{instanceID: 4}}
	runtime.Drop = []*Card{{instanceID: 5}}
	runtime.Exile = []*Card{{instanceID: 6}}
	eng := this.engine(runtime, nil)

	this.Equal([]InstanceID{1, 2}, this.must(eng, "handRand", nums(2, 0))) // 候選 3 > 2 → 隨機取 2(恆等洗牌取前綴)
	this.Empty(this.must(eng, "handRand", nums(1)))                        // 參數不足 → 空集合

	this.Equal([]InstanceID{4}, this.must(eng, "deckRand", nums(1, 0))) // 各容器 Rand 變體一一驗證
	this.Equal([]InstanceID{5}, this.must(eng, "dropRand", nums(1, 0)))
	this.Equal([]InstanceID{6}, this.must(eng, "exileRand", nums(1, 0)))
}

func (this *SuiteSelector) TestSelectDeckTop() {
	// deck >= N:取頂端(前端)N,不洗牌
	deckGe := NewRuntime(0)
	deckGe.Deck = cards(1, 2, 3)
	deckGe.Drop = cards(4)
	this.Equal([]InstanceID{1, 2}, this.must(this.engine(deckGe, nil), "deckTop", nums(2)))
	this.Len(deckGe.Drop, 1) // 牌堆足夠 → 未洗牌移動

	// deck < N 且 deck+drop >= N:洗棄牌移至牌堆底端再取頂端 N
	deckShort := NewRuntime(0)
	deckShort.Deck = cards(1)
	deckShort.Drop = cards(4, 5)
	this.Equal([]InstanceID{1, 4}, this.must(this.engine(deckShort, nil), "deckTop", nums(2))) // 原牌堆 1 在頂、洗入的 4 接其後
	this.Empty(deckShort.Drop)                                                                 // 棄牌已全數移入

	// deck+drop < N:洗棄牌移入後取全部
	deckTiny := NewRuntime(0)
	deckTiny.Deck = cards(1)
	deckTiny.Drop = cards(4)
	this.Equal([]InstanceID{1, 4}, this.must(this.engine(deckTiny, nil), "deckTop", nums(5)))

	// deck < N 但棄牌為空:不洗牌、取全部
	deckNoDrop := NewRuntime(0)
	deckNoDrop.Deck = cards(1)
	this.Equal([]InstanceID{1}, this.must(this.engine(deckNoDrop, nil), "deckTop", nums(5)))

	// N <= 0 / 缺參數 → 空集合
	this.Empty(this.must(this.engine(deckGe, nil), "deckTop", nums(0)))
	this.Empty(this.must(this.engine(deckGe, nil), "deckTop", nil))
}

func (this *SuiteSelector) TestSelectDropTop() {
	runtime := NewRuntime(0)
	runtime.Drop = cards(1, 2, 3)
	eng := this.engine(runtime, nil)

	this.Equal([]InstanceID{1, 2}, this.must(eng, "dropTop", nums(2)))    // 取頂端 2
	this.Equal([]InstanceID{1, 2, 3}, this.must(eng, "dropTop", nums(9))) // 不足 → 取全部
	this.Empty(this.must(eng, "dropTop", nil))                            // 缺參數 → 空集合
}

// === 測試輔助(置尾) ===

// engine 組裝測試引擎:注入 runtime / self、共用 buildSheet 靜態表、決定性 fake Operator / Rander。
func (this *SuiteSelector) engine(runtime *Runtime, self *Ref) *Engine {
	return NewEngine(runtime, self, buildSheet(), fakeOperator{}, fakeRander{}, nil)
}

// must 派發命令對象並斷言名稱已登錄,回傳作用對象集合(聚焦於結果斷言)。
func (this *SuiteSelector) must(eng *Engine, name string, arg []exprs.Value) (result []InstanceID) {
	result, ok := eng.selectObject(name, arg)
	this.Require().True(ok, name)
	return result
}

// seatGuestRuntime 組裝 3 位在座顧客(座1=g1=11 / 座2=12 / 座3=13;桌1=座1,2、桌2=座3,對齊 buildSheet);回傳 g1 供 self 綁定。
func seatGuestRuntime() (runtime *Runtime, g1 *Guest) {
	runtime = NewRuntime(0)
	g1 = &Guest{instanceID: 11, seatID: 1}
	runtime.Seat[1] = g1
	runtime.Seat[2] = &Guest{instanceID: 12, seatID: 2}
	runtime.Seat[3] = &Guest{instanceID: 13, seatID: 3}
	return runtime, g1
}

// cards 以實例編號批次建卡牌切片(供容器 / 牌堆測試)。
func cards(id ...int) (result []*Card) {
	for _, itor := range id {
		result = append(result, &Card{instanceID: InstanceID(itor)})
	} // for

	return result
}

// nums 把整數批次包成運算式數值參數列表(模擬 [...] 內已求值參數)。
func nums(n ...int) (result []exprs.Value) {
	for _, itor := range n {
		result = append(result, exprs.NewNum(float64(itor)))
	} // for

	return result
}

// fakeOperator 決定性玩家輸入替身:Pick 取候選前綴;emptyPick 為真時 PickGuest 回空(驗證防禦分支)。
type fakeOperator struct {
	emptyPick bool
}

func (this fakeOperator) PlayerAction(runtime *Runtime) Action {
	return Action{}
}

func (this fakeOperator) PickGuest(source []*Guest, count int) []*Guest {
	if this.emptyPick {
		return nil
	} // if

	return source[:count]
}

func (this fakeOperator) PickCard(source []*Card, count int) []*Card {
	return source[:count]
}

func (this fakeOperator) PickDiscard(source []*Card, over int) []*Card {
	return nil
}

// fakeRander 決定性亂數替身:Intn 固定回 intn(預設 0、取候選首位);Shuffle 恆等(randSubset 取候選前綴,結果可預期)。
type fakeRander struct {
	intn int
}

func (this fakeRander) Intn(n int) int {
	return this.intn
}

func (this fakeRander) Shuffle(n int, swap func(i, j int)) {
	if n > 0 {
		swap(0, 0) // 恆等交換:執行 caller 的 swap 閉包(覆蓋),但不改變順序,使 randSubset / deckTop 結果可預期
	} // if
}

func (this fakeRander) Weighted(weight []int32) int {
	return 0
}
