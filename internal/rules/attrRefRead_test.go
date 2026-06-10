package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
	"github.com/yinweli/RovingDiner/internal/tester"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteAttrRefRead(t *testing.T) {
	suite.Run(t, new(SuiteAttrRefRead))
}

// SuiteAttrRefRead 驗證引用屬性讀取詞彙表(attrRefRead.go):卡牌 / 顧客引用屬性、鎖、共用 effect 查詢、型別不符。
type SuiteAttrRefRead struct {
	suite.Suite
}

func (this *SuiteAttrRefRead) TestAttrRefReadCard() {
	game := newGame()
	card := cores.NewCard(game, 101) // 卡 101:群組 1
	card.GetCost().Set(2)
	card.GetExtraRunMin().Set(1)
	card.GetExtraRunMax().Set(5)
	card.GetEffectID().Add(201, 201, 202)
	game.Hand = cores.CardList{card}
	game.Deck = cores.CardList{cores.NewCard(game, 101)}
	ref := cores.NewRefCard(card)

	this.Equal(float64(101), this.num(game, ref, "cardID"))
	this.Equal(float64(2), this.num(game, ref, "cost"))
	this.Equal(float64(1), this.num(game, ref, "extraRunMin"))
	this.Equal(float64(5), this.num(game, ref, "extraRunMax"))
	this.Equal(float64(0), this.num(game, ref, "cardSeal"))    // 值固定 0
	this.Equal(float64(0), this.num(game, ref, "keep"))        // 值固定 0
	this.Equal(float64(0), this.num(game, ref, "playExile"))   // 值固定 0
	this.Equal(float64(0), this.num(game, ref, "unplayExile")) // 值固定 0
	this.Equal(float64(1), this.num(game, ref, "cardGroup"))   // 卡 101 群組 1
	this.Equal(float64(2), this.num(game, ref, "cardEffect", exprs.NewNum(201)))
	this.Equal(float64(1), this.num(game, ref, "cardEffect", exprs.NewNum(202)))
	this.Equal(float64(0), this.num(game, ref, "cardEffect", exprs.NewNum(0))) // N==0 不命中

	this.True(this.flag(game, ref, "inHand"))
	this.False(this.flag(game, ref, "inDeck"))
	this.False(this.flag(game, ref, "inDrop"))
	this.False(this.flag(game, ref, "inExile"))
	this.True(this.flag(game, cores.NewRefCard(game.Deck[0]), "inDeck")) // 牌堆內卡牌 → true

	value, ok := game.AttrRef(ref, "cardify", nil) // 初值 nil → none
	this.True(ok)
	this.True(value.IsNone())

	_, ok = game.AttrRef(cores.NewRefGuest(&cores.Guest{}), "cardID", nil) // 顧客 ref 問卡牌屬性 → 失敗
	this.False(ok)
	_, ok = game.AttrRef(cores.NewRefCard(strayCard(999)), "cardGroup", nil) // 卡牌資料不存在 → 失敗
	this.False(ok)
	_, ok = game.AttrRef(ref, "cardEffect", nil) // 缺參數 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadCardCardify() {
	game := newGame()
	guest := cores.NewGuest(game, 501)
	card := cores.NewCard(game, 101)
	card.CardifyBind(guest)

	value, ok := game.AttrRef(cores.NewRefCard(card), "cardify", nil)
	this.True(ok)
	this.True(value.Ref().IsSame(cores.NewRefGuest(guest)))
}

func (this *SuiteAttrRefRead) TestAttrRefReadCardLock() {
	game := newGame()
	card := &cores.Card{}
	lock(card.GetCost(), 3)
	lock(card.GetExtraRunMin(), 4)
	lock(card.GetExtraRunMax(), 6)
	lock(card.GetSeal(), 7)
	lock(card.GetKeep(), 8)
	lock(card.GetPlayExile(), 9)
	lock(card.GetUnplayExile(), 10)
	ref := cores.NewRefCard(card)

	this.Equal(float64(3), this.num(game, ref, "costLock"))
	this.Equal(float64(4), this.num(game, ref, "extraRunMinLock"))
	this.Equal(float64(6), this.num(game, ref, "extraRunMaxLock"))
	this.Equal(float64(7), this.num(game, ref, "cardSealLock"))
	this.Equal(float64(8), this.num(game, ref, "keepLock"))
	this.Equal(float64(9), this.num(game, ref, "playExileLock"))
	this.Equal(float64(10), this.num(game, ref, "unplayExileLock"))

	_, ok := game.AttrRef(cores.NewRefGuest(&cores.Guest{}), "costLock", nil) // 型別不符 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadGuest() {
	game := newGame()
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(1, guest)
	game.Seat[2] = &cores.Guest{}
	game.Seat[3] = &cores.Guest{}
	guest.SetFreeze(3)
	guest.GetCalm().Set(8)
	guest.GetSate().Set(4)
	guest.GetSateMax().Set(20)
	guest.GetMorale().Set(6)
	guest.GetMoraleMax().Set(15)
	guest.GetScore().Set(9)
	guest.GetScoreMax().Set(30)
	guest.GetSateHit().Add(10)
	guest.GetSateHit().Add(20)
	guest.GetCalmHit().Add(5)
	guest.GetEffectImmune().Add(7)
	guest.GetEffectImmune().Add(7)
	guest.GetSkillImmune().Add(3)
	ref := cores.NewRefGuest(guest)

	this.Equal(float64(501), this.num(game, ref, "guestID"))
	this.Equal(float64(1), this.num(game, ref, "seatID"))
	this.Equal(float64(3), this.num(game, ref, "freeze"))
	this.Equal(float64(8), this.num(game, ref, "calm"))
	this.Equal(float64(4), this.num(game, ref, "sate"))
	this.Equal(float64(20), this.num(game, ref, "sateMax"))
	this.Equal(float64(6), this.num(game, ref, "morale"))
	this.Equal(float64(15), this.num(game, ref, "moraleMax"))
	this.Equal(float64(9), this.num(game, ref, "score"))
	this.Equal(float64(30), this.num(game, ref, "scoreMax"))
	this.Equal(float64(0), this.num(game, ref, "sateSeal")) // 值固定 0
	this.Equal(float64(0), this.num(game, ref, "calmSeal")) // 值固定 0
	this.Equal(float64(1), this.num(game, ref, "calmHit"))  // 命中 1
	this.Equal(float64(2), this.num(game, ref, "sateHit"))  // 命中 2
	this.Equal(float64(2), this.num(game, ref, "effectImmune", exprs.NewNum(7)))
	this.Equal(float64(0), this.num(game, ref, "effectImmune", exprs.NewNum(99)))
	this.Equal(float64(1), this.num(game, ref, "skillImmune", exprs.NewNum(3)))
	this.Equal(float64(2), this.num(game, ref, "sameSize")) // 座 1,2 占用、含自身
	this.Equal(float64(1), this.num(game, ref, "nearSize")) // 座 3 占用

	_, ok := game.AttrRef(cores.NewRefCard(&cores.Card{}), "calm", nil) // 卡牌 ref 問顧客屬性 → 失敗
	this.False(ok)
	_, ok = game.AttrRef(ref, "effectImmune", nil) // 缺參數 → 失敗
	this.False(ok)
	_, ok = game.AttrRef(ref, "skillImmune", nil) // 缺參數 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadGuestLock() {
	game := newGame()
	guest := &cores.Guest{}
	lock(guest.GetCalm(), 11)
	lock(guest.GetSate(), 12)
	lock(guest.GetSateMax(), 13)
	lock(guest.GetMorale(), 14)
	lock(guest.GetMoraleMax(), 15)
	lock(guest.GetScore(), 16)
	lock(guest.GetScoreMax(), 17)
	lock(guest.GetSateSeal(), 18)
	lock(guest.GetCalmSeal(), 19)
	ref := cores.NewRefGuest(guest)

	this.Equal(float64(11), this.num(game, ref, "calmLock"))
	this.Equal(float64(12), this.num(game, ref, "sateLock"))
	this.Equal(float64(13), this.num(game, ref, "sateMaxLock"))
	this.Equal(float64(14), this.num(game, ref, "moraleLock"))
	this.Equal(float64(15), this.num(game, ref, "moraleMaxLock"))
	this.Equal(float64(16), this.num(game, ref, "scoreLock"))
	this.Equal(float64(17), this.num(game, ref, "scoreMaxLock"))
	this.Equal(float64(18), this.num(game, ref, "sateSealLock"))
	this.Equal(float64(19), this.num(game, ref, "calmSealLock"))

	_, ok := game.AttrRef(cores.NewRefCard(&cores.Card{}), "calmLock", nil) // 型別不符 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadGuestNoSeat() {
	game := newGame()
	guest := &cores.Guest{} // 遊蕩 / 卡牌化(seatID 0)

	this.Equal(float64(0), this.num(game, cores.NewRefGuest(guest), "sameSize"))
	this.Equal(float64(0), this.num(game, cores.NewRefGuest(guest), "nearSize"))
}

func (this *SuiteAttrRefRead) TestAttrRefReadEffect() {
	data := tester.BuildData()
	data.SetEffect(202, cores.EffectData{}) // 補登迷你表沒有的效果 202
	game := cores.NewGame(0, 0, data, nil, nil)
	Register(game)
	guest := &cores.Guest{}
	card := &cores.Card{}
	guestR := cores.NewRefGuest(guest)
	cardR := cores.NewRefCard(card)
	game.Effect.Push(cores.NewEffect(game, 201, guestR, 3)) // 群組 5
	game.Effect.Push(cores.NewEffect(game, 202, guestR, 1))
	game.Effect.Push(cores.NewEffect(game, 201, cardR, 9)) // 不同 self

	this.Equal(float64(3), this.num(game, guestR, "effectStack", exprs.NewNum(201))) // 顧客 self
	this.Equal(float64(9), this.num(game, cardR, "effectStack", exprs.NewNum(201)))  // 卡牌 self
	this.Equal(float64(1), this.num(game, guestR, "effectGroup", exprs.NewNum(5)))   // 群組 5 = effect 201
	this.Equal(float64(0), this.num(game, guestR, "effectGroup", exprs.NewNum(0)))   // N==0 不命中
	this.Equal(float64(0), this.num(game, guestR, "effectStack", exprs.NewNum(999))) // 不存在 → 0

	_, ok := game.AttrRef(guestR, "effectStack", nil) // 缺參數 → 失敗
	this.False(ok)
	_, ok = game.AttrRef(guestR, "effectGroup", nil)
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadCardTypeMismatch() {
	game := newGame()
	wrong := cores.NewRefGuest(&cores.Guest{}) // 以顧客引用問卡牌屬性,全應失敗

	for _, name := range []string{
		"cardID", "cost", "extraRunMin", "extraRunMax", "cardSeal", "keep",
		"playExile", "unplayExile", "cardify", "cardGroup", "cardEffect",
		"inHand", "inDeck", "inDrop", "inExile",
	} {
		_, ok := game.AttrRef(wrong, name, nil)
		this.False(ok, name)
	} // for
}

func (this *SuiteAttrRefRead) TestAttrRefReadGuestTypeMismatch() {
	game := newGame()
	wrong := cores.NewRefCard(&cores.Card{}) // 以卡牌引用問顧客屬性,全應失敗

	for _, name := range []string{
		"calm", "sate", "sateMax", "morale", "moraleMax", "score", "scoreMax",
		"sateSeal", "calmSeal", "seatID", "guestID", "freeze", "calmHit", "sateHit",
		"effectImmune", "skillImmune", "sameSize", "nearSize",
	} {
		_, ok := game.AttrRef(wrong, name, nil)
		this.False(ok, name)
	} // for
}

// num 取引用屬性求值結果的數字;斷言命中且為數值。
func (this *SuiteAttrRefRead) num(game *cores.Game, ref exprs.Ref, name string, arg ...exprs.Value) float64 {
	value, ok := game.AttrRef(ref, name, arg)
	this.Require().True(ok)
	this.Require().True(value.IsNum())
	return value.Num()
}

// flag 取引用屬性求值結果的布林;斷言命中且為布林。
func (this *SuiteAttrRefRead) flag(game *cores.Game, ref exprs.Ref, name string) bool {
	value, ok := game.AttrRef(ref, name, nil)
	this.Require().True(ok)
	this.Require().True(value.IsBool())
	return value.Bool()
}

// lock 對屬性組件鎖定 n 次（公開介面佈置鎖定計數）。
func lock(value *cores.Value, n int32) {
	for i := int32(0); i < n; i++ {
		value.Lock()
	} // for
}

// strayCard 以獨立迷你表建構主表沒有的卡牌（製造「卡牌資料不存在」的查詢對象）;
// 實例編號自獨立序列錯開（自 1000 起）,避免與呼叫端 Game 的編號相撞。
func strayCard(cardID int32) *cores.Card {
	sheet := &sheeter.Sheeter{}
	sheet.Card.Data = map[int32]*sheeter.Card{cardID: {ID: cardID}}
	game := cores.NewGame(0, 0, cores.NewData(sheet, nil), nil, nil)

	for i := 0; i < 999; i++ {
		game.NextID()
	} // for

	return cores.NewCard(game, cardID)
}
