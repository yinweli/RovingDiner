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
		InstanceID:  1,
		CardID:      101,
		Cost:        Value{Value: 2},
		ExtraRunMin: Value{Value: 1},
		ExtraRunMax: Value{Value: 5},
		EffectID:    []int32{201, 201, 202},
	}
	runtime := NewRuntime(0)
	runtime.Hand = []*Card{card}
	runtime.Deck = []*Card{{InstanceID: 2}}
	eng := &Engine{runtime: runtime, data: buildSheet()}
	ref := cardRef{card: card}

	this.Equal(float64(101), this.num(eng, ref, "cardID"))
	this.Equal(float64(2), this.num(eng, ref, "cost"))
	this.Equal(float64(1), this.num(eng, ref, "extraRunMin"))
	this.Equal(float64(5), this.num(eng, ref, "extraRunMax"))
	this.Equal(float64(0), this.num(eng, ref, "cardSeal"))    // 值固定 0
	this.Equal(float64(0), this.num(eng, ref, "keep"))        // 值固定 0
	this.Equal(float64(0), this.num(eng, ref, "playExile"))   // 值固定 0
	this.Equal(float64(0), this.num(eng, ref, "unplayExile")) // 值固定 0
	this.Equal(float64(1), this.num(eng, ref, "cardGroup"))   // 卡 101 群組 1
	this.Equal(float64(2), this.num(eng, ref, "cardEffect", exprs.NewNum(201)))
	this.Equal(float64(1), this.num(eng, ref, "cardEffect", exprs.NewNum(202)))
	this.Equal(float64(0), this.num(eng, ref, "cardEffect", exprs.NewNum(0))) // N==0 不命中

	this.True(this.flag(eng, ref, "inHand"))
	this.False(this.flag(eng, ref, "inDeck"))
	this.False(this.flag(eng, ref, "inDrop"))
	this.False(this.flag(eng, ref, "inExile"))
	this.True(this.flag(eng, cardRef{card: runtime.Deck[0]}, "inDeck")) // 牌堆內卡牌 → true

	value, ok := eng.AttrRef(ref, "cardify", nil) // 初值 nil → none
	this.True(ok)
	this.True(value.IsNone())

	_, ok = eng.AttrRef(guestRef{guest: &Guest{}}, "cardID", nil) // 顧客 ref 問卡牌屬性 → 失敗
	this.False(ok)
	_, ok = eng.AttrRef(cardRef{card: &Card{CardID: 999}}, "cardGroup", nil) // 卡牌資料不存在 → 失敗
	this.False(ok)
	_, ok = eng.AttrRef(ref, "cardEffect", nil) // 缺參數 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadCardCardify() {
	guest := &Guest{InstanceID: 7}
	card := &Card{InstanceID: 1, Cardify: guest}
	eng := &Engine{runtime: NewRuntime(0)}

	value, ok := eng.AttrRef(cardRef{card: card}, "cardify", nil)
	this.True(ok)
	this.True(value.Ref().Same(guestRef{guest: guest}))
}

func (this *SuiteAttrRefRead) TestAttrRefReadCardLock() {
	card := &Card{
		Cost:        Value{Lock: 3},
		ExtraRunMin: Value{Lock: 4},
		ExtraRunMax: Value{Lock: 6},
		Seal:        Value{Lock: 7},
		Keep:        Value{Lock: 8},
		PlayExile:   Value{Lock: 9},
		UnplayExile: Value{Lock: 10},
	}
	eng := &Engine{runtime: NewRuntime(0)}
	ref := cardRef{card: card}

	this.Equal(float64(3), this.num(eng, ref, "costLock"))
	this.Equal(float64(4), this.num(eng, ref, "extraRunMinLock"))
	this.Equal(float64(6), this.num(eng, ref, "extraRunMaxLock"))
	this.Equal(float64(7), this.num(eng, ref, "cardSealLock"))
	this.Equal(float64(8), this.num(eng, ref, "keepLock"))
	this.Equal(float64(9), this.num(eng, ref, "playExileLock"))
	this.Equal(float64(10), this.num(eng, ref, "unplayExileLock"))

	_, ok := eng.AttrRef(guestRef{guest: &Guest{}}, "costLock", nil) // 型別不符 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadGuest() {
	guest := &Guest{
		InstanceID:   1,
		GuestID:      5,
		SeatID:       1,
		Freeze:       3,
		Calm:         Value{Value: 8},
		Sate:         Value{Value: 4},
		SateMax:      Value{Value: 20},
		Morale:       Value{Value: 6},
		MoraleMax:    Value{Value: 15},
		Score:        Value{Value: 9},
		ScoreMax:     Value{Value: 30},
		SateHit:      map[int32]bool{10: true, 20: true},
		CalmHit:      map[int32]bool{5: true},
		EffectImmune: map[int32]int32{7: 2},
		SkillImmune:  map[int32]int32{3: 1},
	}
	runtime := NewRuntime(0)
	runtime.Seat[1] = guest
	runtime.Seat[2] = &Guest{InstanceID: 2}
	runtime.Seat[3] = &Guest{InstanceID: 3}
	eng := &Engine{runtime: runtime, data: buildSheet()}
	ref := guestRef{guest: guest}

	this.Equal(float64(5), this.num(eng, ref, "guestID"))
	this.Equal(float64(1), this.num(eng, ref, "seatID"))
	this.Equal(float64(3), this.num(eng, ref, "freeze"))
	this.Equal(float64(8), this.num(eng, ref, "calm"))
	this.Equal(float64(4), this.num(eng, ref, "sate"))
	this.Equal(float64(20), this.num(eng, ref, "sateMax"))
	this.Equal(float64(6), this.num(eng, ref, "morale"))
	this.Equal(float64(15), this.num(eng, ref, "moraleMax"))
	this.Equal(float64(9), this.num(eng, ref, "score"))
	this.Equal(float64(30), this.num(eng, ref, "scoreMax"))
	this.Equal(float64(0), this.num(eng, ref, "sateSeal")) // 值固定 0
	this.Equal(float64(0), this.num(eng, ref, "calmSeal")) // 值固定 0
	this.Equal(float64(1), this.num(eng, ref, "calmHit"))  // 命中 1
	this.Equal(float64(2), this.num(eng, ref, "sateHit"))  // 命中 2
	this.Equal(float64(2), this.num(eng, ref, "effectImmune", exprs.NewNum(7)))
	this.Equal(float64(0), this.num(eng, ref, "effectImmune", exprs.NewNum(99)))
	this.Equal(float64(1), this.num(eng, ref, "skillImmune", exprs.NewNum(3)))
	this.Equal(float64(2), this.num(eng, ref, "sameSize")) // 座 1,2 占用、含自身
	this.Equal(float64(1), this.num(eng, ref, "nearSize")) // 座 3 占用

	_, ok := eng.AttrRef(cardRef{card: &Card{}}, "calm", nil) // 卡牌 ref 問顧客屬性 → 失敗
	this.False(ok)
	_, ok = eng.AttrRef(ref, "effectImmune", nil) // 缺參數 → 失敗
	this.False(ok)
	_, ok = eng.AttrRef(ref, "skillImmune", nil) // 缺參數 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadGuestLock() {
	guest := &Guest{
		Calm:      Value{Lock: 11},
		Sate:      Value{Lock: 12},
		SateMax:   Value{Lock: 13},
		Morale:    Value{Lock: 14},
		MoraleMax: Value{Lock: 15},
		Score:     Value{Lock: 16},
		ScoreMax:  Value{Lock: 17},
		SateSeal:  Value{Lock: 18},
		CalmSeal:  Value{Lock: 19},
	}
	eng := &Engine{runtime: NewRuntime(0)}
	ref := guestRef{guest: guest}

	this.Equal(float64(11), this.num(eng, ref, "calmLock"))
	this.Equal(float64(12), this.num(eng, ref, "sateLock"))
	this.Equal(float64(13), this.num(eng, ref, "sateMaxLock"))
	this.Equal(float64(14), this.num(eng, ref, "moraleLock"))
	this.Equal(float64(15), this.num(eng, ref, "moraleMaxLock"))
	this.Equal(float64(16), this.num(eng, ref, "scoreLock"))
	this.Equal(float64(17), this.num(eng, ref, "scoreMaxLock"))
	this.Equal(float64(18), this.num(eng, ref, "sateSealLock"))
	this.Equal(float64(19), this.num(eng, ref, "calmSealLock"))

	_, ok := eng.AttrRef(cardRef{card: &Card{}}, "calmLock", nil) // 型別不符 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadGuestNoSeat() {
	guest := &Guest{InstanceID: 1, SeatID: 0} // 遊蕩 / 卡牌化
	eng := &Engine{runtime: NewRuntime(0), data: buildSheet()}
	ref := guestRef{guest: guest}

	this.Equal(float64(0), this.num(eng, ref, "sameSize"))
	this.Equal(float64(0), this.num(eng, ref, "nearSize"))
}

func (this *SuiteAttrRefRead) TestAttrRefReadEffect() {
	guest := &Guest{InstanceID: 1}
	card := &Card{InstanceID: 2}
	runtime := NewRuntime(0)
	runtime.Effect = []*Effect{
		{EffectID: 201, Stack: 3, Self: Self{Guest: guest}}, // 群組 5
		{EffectID: 202, Stack: 1, Self: Self{Guest: guest}},
		{EffectID: 201, Stack: 9, Self: Self{Card: card}}, // 不同 self
	}
	eng := &Engine{runtime: runtime, data: buildSheet()}
	guestR := guestRef{guest: guest}
	cardR := cardRef{card: card}

	this.Equal(float64(3), this.num(eng, guestR, "effectStack", exprs.NewNum(201))) // 顧客 self
	this.Equal(float64(9), this.num(eng, cardR, "effectStack", exprs.NewNum(201)))  // 卡牌 self
	this.Equal(float64(1), this.num(eng, guestR, "effectGroup", exprs.NewNum(5)))   // 群組 5 = effect 201
	this.Equal(float64(0), this.num(eng, guestR, "effectGroup", exprs.NewNum(0)))   // N==0 不命中
	this.Equal(float64(0), this.num(eng, guestR, "effectStack", exprs.NewNum(999))) // 不存在 → 0

	_, ok := eng.AttrRef(guestR, "effectStack", nil) // 缺參數 → 失敗
	this.False(ok)
	_, ok = eng.AttrRef(guestR, "effectGroup", nil)
	this.False(ok)
}

func (this *SuiteAttrRefRead) TestAttrRefReadCardTypeMismatch() {
	eng := &Engine{runtime: NewRuntime(0), data: buildSheet()}
	wrong := guestRef{guest: &Guest{}} // 以顧客引用問卡牌屬性,全應失敗

	for _, name := range []string{
		"cardID", "cost", "extraRunMin", "extraRunMax", "cardSeal", "keep",
		"playExile", "unplayExile", "cardify", "cardGroup", "cardEffect",
		"inHand", "inDeck", "inDrop", "inExile",
	} {
		_, ok := eng.AttrRef(wrong, name, nil)
		this.False(ok, name)
	} // for
}

func (this *SuiteAttrRefRead) TestAttrRefReadGuestTypeMismatch() {
	eng := &Engine{runtime: NewRuntime(0), data: buildSheet()}
	wrong := cardRef{card: &Card{}} // 以卡牌引用問顧客屬性,全應失敗

	for _, name := range []string{
		"calm", "sate", "sateMax", "morale", "moraleMax", "score", "scoreMax",
		"sateSeal", "calmSeal", "seatID", "guestID", "freeze", "calmHit", "sateHit",
		"effectImmune", "skillImmune", "sameSize", "nearSize",
	} {
		_, ok := eng.AttrRef(wrong, name, nil)
		this.False(ok, name)
	} // for
}

// num 取引用屬性求值結果的數字;斷言命中且為數值。
func (this *SuiteAttrRefRead) num(eng *Engine, ref exprs.Ref, name string, arg ...exprs.Value) float64 {
	value, ok := eng.AttrRef(ref, name, arg)
	this.Require().True(ok)
	this.Require().True(value.IsNum())
	return value.Num()
}

// flag 取引用屬性求值結果的布林;斷言命中且為布林。
func (this *SuiteAttrRefRead) flag(eng *Engine, ref exprs.Ref, name string) bool {
	value, ok := eng.AttrRef(ref, name, nil)
	this.Require().True(ok)
	this.Require().True(value.IsBool())
	return value.Bool()
}
