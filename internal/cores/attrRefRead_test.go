package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteAttrRefRead(t *testing.T) {
	suite.Run(t, new(SuiteAttrRefRead))
}

// SuiteAttrRefRead 驗證引用屬性讀取詞彙表(attrRefRead.go):卡牌 / 顧客引用屬性、鎖、共用 effect 查詢、型別不符。
type SuiteAttrRefRead struct {
	suite.Suite
}

func (this *SuiteAttrRefRead) TestAttrRefReadCard() {
	card := &Card{
		instanceID:  1,
		cardID:      101,
		cost:        NewValue(2, 0),
		extraRunMin: NewValue(1, 0),
		extraRunMax: NewValue(5, 0),
		effectID:    NewIDList(201, 201, 202),
	}
	game := NewGame(0, nil, nil, nil, nil)
	game.Hand = []*Card{card}
	game.Deck = []*Card{{instanceID: 2}}
	game.data = buildSheet()
	ref := NewRefCard(card)

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
	this.True(this.flag(game, NewRefCard(game.Deck[0]), "inDeck")) // 牌堆內卡牌 → true

	value, ok := game.AttrRef(ref, "cardify", nil) // 初值 nil → none
	this.True(ok)
	this.True(value.IsNone())

	_, ok = game.AttrRef(NewRefGuest(&Guest{}), "cardID", nil) // 顧客 ref 問卡牌屬性 → 失敗
	this.False(ok)
	_, ok = game.AttrRef(NewRefCard(&Card{cardID: 999}), "cardGroup", nil) // 卡牌資料不存在 → 失敗
	this.False(ok)
	_, ok = game.AttrRef(ref, "cardEffect", nil) // 缺參數 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadCardCardify() {
	guest := &Guest{instanceID: 7}
	card := &Card{instanceID: 1, cardify: guest}
	game := NewGame(0, nil, nil, nil, nil)

	value, ok := game.AttrRef(NewRefCard(card), "cardify", nil)
	this.True(ok)
	this.True(value.Ref().IsSame(NewRefGuest(guest)))
}

func (this *SuiteAttrRefRead) TestAttrRefReadCardLock() {
	card := &Card{
		cost:        NewValue(0, 3),
		extraRunMin: NewValue(0, 4),
		extraRunMax: NewValue(0, 6),
		seal:        NewValue(0, 7),
		keep:        NewValue(0, 8),
		playExile:   NewValue(0, 9),
		unplayExile: NewValue(0, 10),
	}
	game := NewGame(0, nil, nil, nil, nil)
	ref := NewRefCard(card)

	this.Equal(float64(3), this.num(game, ref, "costLock"))
	this.Equal(float64(4), this.num(game, ref, "extraRunMinLock"))
	this.Equal(float64(6), this.num(game, ref, "extraRunMaxLock"))
	this.Equal(float64(7), this.num(game, ref, "cardSealLock"))
	this.Equal(float64(8), this.num(game, ref, "keepLock"))
	this.Equal(float64(9), this.num(game, ref, "playExileLock"))
	this.Equal(float64(10), this.num(game, ref, "unplayExileLock"))

	_, ok := game.AttrRef(NewRefGuest(&Guest{}), "costLock", nil) // 型別不符 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadGuest() {
	guest := &Guest{
		instanceID:   1,
		guestID:      5,
		seatID:       1,
		freeze:       3,
		calm:         NewValue(8, 0),
		sate:         NewValue(4, 0),
		sateMax:      NewValue(20, 0),
		morale:       NewValue(6, 0),
		moraleMax:    NewValue(15, 0),
		score:        NewValue(9, 0),
		scoreMax:     NewValue(30, 0),
		sateHit:      Hit{hit: map[int32]bool{10: true, 20: true}},
		calmHit:      Hit{hit: map[int32]bool{5: true}},
		effectImmune: Immune{count: map[int32]int32{7: 2}},
		skillImmune:  Immune{count: map[int32]int32{3: 1}},
	}
	game := NewGame(0, nil, nil, nil, nil)
	game.Seat[1] = guest
	game.Seat[2] = &Guest{instanceID: 2}
	game.Seat[3] = &Guest{instanceID: 3}
	game.data = buildSheet()
	ref := NewRefGuest(guest)

	this.Equal(float64(5), this.num(game, ref, "guestID"))
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

	_, ok := game.AttrRef(NewRefCard(&Card{}), "calm", nil) // 卡牌 ref 問顧客屬性 → 失敗
	this.False(ok)
	_, ok = game.AttrRef(ref, "effectImmune", nil) // 缺參數 → 失敗
	this.False(ok)
	_, ok = game.AttrRef(ref, "skillImmune", nil) // 缺參數 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadGuestLock() {
	guest := &Guest{
		calm:      NewValue(0, 11),
		sate:      NewValue(0, 12),
		sateMax:   NewValue(0, 13),
		morale:    NewValue(0, 14),
		moraleMax: NewValue(0, 15),
		score:     NewValue(0, 16),
		scoreMax:  NewValue(0, 17),
		sateSeal:  NewValue(0, 18),
		calmSeal:  NewValue(0, 19),
	}
	game := NewGame(0, nil, nil, nil, nil)
	ref := NewRefGuest(guest)

	this.Equal(float64(11), this.num(game, ref, "calmLock"))
	this.Equal(float64(12), this.num(game, ref, "sateLock"))
	this.Equal(float64(13), this.num(game, ref, "sateMaxLock"))
	this.Equal(float64(14), this.num(game, ref, "moraleLock"))
	this.Equal(float64(15), this.num(game, ref, "moraleMaxLock"))
	this.Equal(float64(16), this.num(game, ref, "scoreLock"))
	this.Equal(float64(17), this.num(game, ref, "scoreMaxLock"))
	this.Equal(float64(18), this.num(game, ref, "sateSealLock"))
	this.Equal(float64(19), this.num(game, ref, "calmSealLock"))

	_, ok := game.AttrRef(NewRefCard(&Card{}), "calmLock", nil) // 型別不符 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadGuestNoSeat() {
	guest := &Guest{instanceID: 1, seatID: 0} // 遊蕩 / 卡牌化
	game := NewGame(0, buildSheet(), nil, nil, nil)
	ref := NewRefGuest(guest)

	this.Equal(float64(0), this.num(game, ref, "sameSize"))
	this.Equal(float64(0), this.num(game, ref, "nearSize"))
}

func (this *SuiteAttrRefRead) TestAttrRefReadEffect() {
	guest := &Guest{instanceID: 1}
	card := &Card{instanceID: 2}
	game := NewGame(0, nil, nil, nil, nil)
	game.Effect = EffectList{
		{effectID: 201, stack: 3, self: NewRefGuest(guest)}, // 群組 5
		{effectID: 202, stack: 1, self: NewRefGuest(guest)},
		{effectID: 201, stack: 9, self: NewRefCard(card)}, // 不同 self
	}
	data := buildSheet()
	game.data = data
	game.effectData = prepareEffect(data, nil)
	guestR := NewRefGuest(guest)
	cardR := NewRefCard(card)

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
	game := NewGame(0, buildSheet(), nil, nil, nil)
	wrong := NewRefGuest(&Guest{}) // 以顧客引用問卡牌屬性,全應失敗

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
	game := NewGame(0, buildSheet(), nil, nil, nil)
	wrong := NewRefCard(&Card{}) // 以卡牌引用問顧客屬性,全應失敗

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
func (this *SuiteAttrRefRead) num(game *Game, ref exprs.Ref, name string, arg ...exprs.Value) float64 {
	value, ok := game.AttrRef(ref, name, arg)
	this.Require().True(ok)
	this.Require().True(value.IsNum())
	return value.Num()
}

// flag 取引用屬性求值結果的布林;斷言命中且為布林。
func (this *SuiteAttrRefRead) flag(game *Game, ref exprs.Ref, name string) bool {
	value, ok := game.AttrRef(ref, name, nil)
	this.Require().True(ok)
	this.Require().True(value.IsBool())
	return value.Bool()
}
