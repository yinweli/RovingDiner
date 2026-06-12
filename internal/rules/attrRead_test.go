package rules

import (
	"testing"

	"github.com/yinweli/RovingDiner/internal/cores"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteAttrRead(t *testing.T) {
	suite.Run(t, new(SuiteAttrRead))
}

// SuiteAttrRead 驗證全域屬性讀取詞彙表(attrRead.go): 純值 / 容器 / 衍生 / 物件引用 / self / 鎖 / 靜態 / 查詢函式。
type SuiteAttrRead struct {
	suite.Suite
}

// TestAttrReadHasObjectRef 驗證引用基底述詞: 引用詞條(顧客 / 卡牌 / self)為真, 數值詞條與未登錄為假。
func (this *SuiteAttrRead) TestAttrReadHasObjectRef() {
	this.True(HasObjectRef("self"))
	this.True(HasObjectRef("damageGuest"))
	this.True(HasObjectRef("seatLast"))
	this.True(HasObjectRef("drawLast"))
	this.False(HasObjectRef("morale"))    // 數值詞條非引用基底
	this.False(HasObjectRef("drawCount")) // 同上
	this.False(HasObjectRef("nope"))      // 未登錄
}

// TestAttrReadArity 驗證讀詞條參數數量查詢: 一般屬性 0、查詢函式依詞條(1 / 2), 未登錄回 ok=false。
func (this *SuiteAttrRead) TestAttrReadArity() {
	this.Equal(0, this.arity("morale"))
	this.Equal(0, this.arity("self"))
	this.Equal(1, this.arity("tableGuest"))
	this.Equal(2, this.arity("tableCount"))
	this.Equal(1, this.arity("handSize"))
	this.Equal(1, this.arity("drawTotal"))

	_, ok := AttrReadArity("nope")
	this.False(ok) // 未登錄
}

func (this *SuiteAttrRead) TestAttrReadValue() {
	game := newGame()
	game.GetMorale().Set(25)
	game.GetMoraleMax().Set(50)
	game.GetMoraleShield().Set(9)
	game.GetMoraleBlock().Set(8)
	game.GetScore().Set(7)
	game.GetEnergy().Set(3)
	game.GetEnergyMax().Set(6)
	game.GetHandMax().Set(10)
	game.GetDrawMax().Set(5)
	game.GetRound().Set(4)
	game.GetRoundMax().Set(12)
	card := cores.NewCard(game, 101)
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(3, guest)
	game.EventDamage(6, guest) // 受損值 6
	game.EventSeat(guest)      // 入座 2
	game.EventSeat(guest)
	game.EventExit(guest)     // 離場 1(座位 3)
	game.EventTask(guest, 77) // 行動 4、技能 77
	game.EventTask(guest, 77)
	game.EventTask(guest, 77)
	game.EventTask(guest, 77)
	game.EventDraw(card, 0) // 抽牌 2
	game.EventDraw(card, 0)
	game.EventDrop(card, 0) // 棄牌 1
	game.EventPlay(card, 0) // 出牌 3
	game.EventPlay(card, 0)
	game.EventPlay(card, 0)
	game.EventExile(card, 0)        // 流放 1
	game.EventMorph(card, 100, 200) // 變身 2、前後編號 100 / 200
	game.EventMorph(card, 100, 200)
	game.SetNextPhase(cores.PhasePlayerAction)

	this.Equal(float64(25), this.num(game, "morale"))
	this.Equal(float64(50), this.num(game, "moraleMax"))
	this.Equal(float64(9), this.num(game, "moraleShield"))
	this.Equal(float64(8), this.num(game, "moraleBlock"))
	this.Equal(float64(7), this.num(game, "score"))
	this.Equal(float64(3), this.num(game, "energy"))
	this.Equal(float64(6), this.num(game, "energyMax"))
	this.Equal(float64(0), this.num(game, "energyKeep")) // 值固定 0
	this.Equal(float64(10), this.num(game, "handMax"))
	this.Equal(float64(5), this.num(game, "drawMax"))
	this.Equal(float64(4), this.num(game, "round"))
	this.Equal(float64(12), this.num(game, "roundMax"))
	this.Equal(float64(6), this.num(game, "damageValue"))
	this.Equal(float64(2), this.num(game, "seatCount"))
	this.Equal(float64(3), this.num(game, "exitLastSeat"))
	this.Equal(float64(1), this.num(game, "exitCount"))
	this.Equal(float64(77), this.num(game, "taskSkill"))
	this.Equal(float64(4), this.num(game, "taskCount"))
	this.Equal(float64(2), this.num(game, "drawCount"))
	this.Equal(float64(1), this.num(game, "dropCount"))
	this.Equal(float64(3), this.num(game, "playCount"))
	this.Equal(float64(1), this.num(game, "exileCount"))
	this.Equal(float64(2), this.num(game, "morphCount"))
	this.Equal(float64(100), this.num(game, "morphOldID"))
	this.Equal(float64(200), this.num(game, "morphNewID"))

	value, ok := game.Attr("nextPhase", nil)
	this.True(ok)
	this.Equal("玩家行動", value.Text())
}

func (this *SuiteAttrRead) TestAttrReadContainer() {
	game := newGame()
	game.Hand = []*cores.Card{{}, {}}
	game.Wait = []*cores.Guest{{}}
	game.Roam = []*cores.Guest{{}, {}, {}}
	game.Seat[1] = &cores.Guest{}
	game.Cardify = []*cores.Guest{{}}
	game.Action = cores.ActionList{{}}

	this.Equal(float64(1), this.num(game, "seatSize"))
	this.Equal(float64(1), this.num(game, "waitSize"))
	this.Equal(float64(3), this.num(game, "roamSize"))
	this.Equal(float64(1), this.num(game, "cardifySize"))
	this.Equal(float64(1), this.num(game, "taskSize"))
	this.Equal(float64(6), this.num(game, "guestSize")) // 1+1+3+1
}

func (this *SuiteAttrRead) TestAttrReadRoundLeft() {
	game := newGame()
	game.GetRound().Set(8)
	game.GetRoundMax().Set(10)
	this.Equal(float64(2), this.num(game, "roundLeft"))

	game.GetRound().Set(12) // 超過上限 → 夾 0
	this.Equal(float64(0), this.num(game, "roundLeft"))
}

func (this *SuiteAttrRead) TestAttrReadObject() {
	game := newGame()
	card := cores.NewCard(game, 101)
	guest := cores.NewGuest(game, 501)
	game.EventDraw(card, 0)
	game.EventDrop(card, 0)
	game.EventExile(card, 0)
	game.EventMorph(card, 0, 0)
	game.EventSeat(guest)
	game.EventExit(guest)
	game.EventTask(guest, 0)
	game.EventDamage(1, guest)

	for _, name := range []string{"drawLast", "dropLast", "exileLast", "morphLast"} {
		value, ok := game.Attr(name, nil)
		this.True(ok)
		this.True(cores.NewRefCard(card).IsSame(value.Ref()), name)
	} // for

	for _, name := range []string{"seatLast", "exitLast", "taskGuest", "damageGuest"} {
		value, ok := game.Attr(name, nil)
		this.True(ok)
		this.True(cores.NewRefGuest(guest).IsSame(value.Ref()), name)
	} // for

	value, ok := game.Attr("playLast", nil) // 尚無 → 空物件
	this.True(ok)
	this.True(value.IsNone())
}

func (this *SuiteAttrRead) TestAttrReadSelf() {
	game := newGame()

	_, ok := game.Attr("self", nil) // 未綁定 → 失敗
	this.False(ok)

	none := cores.Ref{} // 綁定空物件 → none
	game.SetSelf(&none)
	value, ok := game.Attr("self", nil)
	this.True(ok)
	this.True(value.IsNone())

	guest := cores.NewGuest(game, 501) // 綁定顧客 → 引用
	bound := cores.NewRefGuest(guest)
	game.SetSelf(&bound)
	value, ok = game.Attr("self", nil)
	this.True(ok)
	this.True(cores.NewRefGuest(guest).IsSame(value.Ref()))
}

func (this *SuiteAttrRead) TestAttrReadLock() {
	game := newGame()
	game.GetEnergy().Set(3)
	lock(game.GetMorale(), 1)
	lock(game.GetMoraleMax(), 2)
	lock(game.GetMoraleShield(), 3)
	lock(game.GetMoraleBlock(), 4)
	lock(game.GetScore(), 5)
	lock(game.GetEnergy(), 6)
	lock(game.GetEnergyMax(), 7)
	lock(game.GetEnergyKeep(), 8)
	lock(game.GetHandMax(), 9)
	lock(game.GetDrawMax(), 10)

	this.Equal(float64(1), this.num(game, "moraleLock"))
	this.Equal(float64(2), this.num(game, "moraleMaxLock"))
	this.Equal(float64(3), this.num(game, "moraleShieldLock"))
	this.Equal(float64(4), this.num(game, "moraleBlockLock"))
	this.Equal(float64(5), this.num(game, "scoreLock"))
	this.Equal(float64(6), this.num(game, "energyLock"))
	this.Equal(float64(7), this.num(game, "energyMaxLock"))
	this.Equal(float64(8), this.num(game, "energyKeepLock"))
	this.Equal(float64(9), this.num(game, "handMaxLock"))
	this.Equal(float64(10), this.num(game, "drawMaxLock"))
}

func (this *SuiteAttrRead) TestAttrReadStatic() {
	game := newGame()
	game.Seat[1] = &cores.Guest{}
	game.Seat[2] = &cores.Guest{}

	this.Equal(float64(1), this.num(game, "seatLeft"))  // 3 座位 - 占用 2
	this.Equal(float64(2), this.num(game, "tableSize")) // 桌 1、2
	this.Equal(float64(2), this.num(game, "tableGuest", exprs.NewNum(1)))
	this.Equal(float64(0), this.num(game, "tableGuest", exprs.NewNum(2)))
	this.Equal(float64(0), this.num(game, "tableGuest", exprs.NewNum(0))) // N==0 不命中

	_, ok := game.Attr("tableGuest", nil) // 缺參數 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRead) TestAttrReadTableCount() {
	game := newGame()
	game.Seat[1] = &cores.Guest{}
	game.Seat[2] = &cores.Guest{} // 桌1 = 2 人; 桌2 = 0 人

	this.Equal(float64(1), this.num(game, "tableCount", exprs.NewText(">="), exprs.NewNum(2)))
	this.Equal(float64(1), this.num(game, "tableCount", exprs.NewText("=="), exprs.NewNum(0)))
	this.Equal(float64(1), this.num(game, "tableCount", exprs.NewText(">"), exprs.NewNum(0)))
	this.Equal(float64(2), this.num(game, "tableCount", exprs.NewText("<="), exprs.NewNum(2)))
	this.Equal(float64(0), this.num(game, "tableCount", exprs.NewText("<"), exprs.NewNum(0)))
	this.Equal(float64(2), this.num(game, "tableCount", exprs.NewText("!="), exprs.NewNum(1)))

	_, ok := game.Attr("tableCount", []exprs.Value{exprs.NewText("~="), exprs.NewNum(0)}) // 未知運算符
	this.False(ok)
	_, ok = game.Attr("tableCount", []exprs.Value{exprs.NewNum(1)}) // 參數數量不符
	this.False(ok)
	_, ok = game.Attr("tableCount", []exprs.Value{exprs.NewNum(1), exprs.NewNum(2)}) // 型別不符
	this.False(ok)
}

func (this *SuiteAttrRead) TestAttrReadGroupQuery() {
	game := newGame()
	game.Hand = cores.CardList{cores.NewCard(game, 101), cores.NewCard(game, 101), cores.NewCard(game, 102)}
	game.Deck = cores.CardList{cores.NewCard(game, 101)}
	game.Drop = cores.CardList{cores.NewCard(game, 102), cores.NewCard(game, 102)}
	game.Exile = cores.CardList{cores.NewCard(game, 101)}
	tally(game.GetDrawTotal(), 1, 3)
	tally(game.GetDrawTotal(), 2, 5)
	tally(game.GetDropTotal(), 1, 1)
	tally(game.GetPlayTotal(), 2, 2)
	tally(game.GetExileTotal(), 1, 4)

	this.Equal(float64(2), this.num(game, "handSize", exprs.NewNum(1))) // 群組 1 = 卡 101 兩張
	this.Equal(float64(1), this.num(game, "handSize", exprs.NewNum(2))) // 群組 2 = 卡 102 一張
	this.Equal(float64(3), this.num(game, "handSize", exprs.NewNum(0))) // N==0 全量
	this.Equal(float64(1), this.num(game, "deckSize", exprs.NewNum(1)))
	this.Equal(float64(2), this.num(game, "dropSize", exprs.NewNum(0)))
	this.Equal(float64(1), this.num(game, "exileSize", exprs.NewNum(1)))

	this.Equal(float64(3), this.num(game, "drawTotal", exprs.NewNum(1)))
	this.Equal(float64(8), this.num(game, "drawTotal", exprs.NewNum(0))) // 全加總
	this.Equal(float64(1), this.num(game, "dropTotal", exprs.NewNum(1)))
	this.Equal(float64(2), this.num(game, "playTotal", exprs.NewNum(2)))
	this.Equal(float64(4), this.num(game, "exileTotal", exprs.NewNum(1)))

	_, ok := game.Attr("handSize", []exprs.Value{exprs.NewText("x")}) // 參數型別不符
	this.False(ok)
	_, ok = game.Attr("drawTotal", nil) // 缺參數
	this.False(ok)
}

// arity 查讀詞條 arity 並斷言已登錄(聚焦於數量斷言)。
func (this *SuiteAttrRead) arity(name string) int {
	arity, ok := AttrReadArity(name)
	this.Require().True(ok, name)
	return arity
}

// num 取全域屬性求值結果的數字; 斷言命中且為數值。
func (this *SuiteAttrRead) num(game *cores.Game, name string, arg ...exprs.Value) float64 {
	value, ok := game.Attr(name, arg)
	this.Require().True(ok)
	this.Require().True(value.IsNum())
	return value.Num()
}

// tally 對累積計數的指定群組加 n 次(公開介面佈置整場累積)。
func tally(total *cores.Tally, group, n int32) {
	for i := int32(0); i < n; i++ {
		total.Add(group)
	} // for
}
