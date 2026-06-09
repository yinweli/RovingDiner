package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteAttrRefWrite(t *testing.T) {
	suite.Run(t, new(SuiteAttrRefWrite))
}

// SuiteAttrRefWrite 驗證引用屬性寫入詞彙表(attrRefWrite.go):卡牌 / 顧客寫鎖屬性、純鎖屬性、型別不符 no-op。
type SuiteAttrRefWrite struct {
	suite.Suite
}

func (this *SuiteAttrRefWrite) TestAttrRefWriteCard() {
	card := &Card{InstanceID: 1}
	eng := &Engine{runtime: NewRuntime(0)}
	ref := cardRef{card: card}

	this.True(writeRefCost(eng, ref, AssignSet, 4)) // 寫鎖:帶值
	this.Equal(int32(4), card.Cost.Value)
	this.True(writeRefExtraRunMin(eng, ref, AssignAdd, 2))
	this.Equal(int32(2), card.ExtraRunMin.Value)
	this.True(writeRefExtraRunMax(eng, ref, AssignSet, 9))
	this.Equal(int32(9), card.ExtraRunMax.Value)

	this.True(writeRefCardSeal(eng, ref, AssignLock, 0)) // 純鎖:@
	this.Equal(int32(1), card.Seal.Lock)
	this.False(writeRefCardSeal(eng, ref, AssignSet, 5)) // 純鎖:帶值賦值 no-op
	this.Equal(int32(0), card.Seal.Value)
	this.True(writeRefKeep(eng, ref, AssignLock, 0))
	this.Equal(int32(1), card.Keep.Lock)
	this.True(writeRefPlayExile(eng, ref, AssignLock, 0))
	this.Equal(int32(1), card.PlayExile.Lock)
	this.True(writeRefUnplayExile(eng, ref, AssignLock, 0))
	this.Equal(int32(1), card.UnplayExile.Lock)
}

func (this *SuiteAttrRefWrite) TestAttrRefWriteGuest() {
	guest := &Guest{InstanceID: 1}
	eng := &Engine{runtime: NewRuntime(0)}
	ref := guestRef{guest: guest}

	this.True(writeRefCalm(eng, ref, AssignSet, 8))
	this.Equal(int32(8), guest.Calm.Value)
	this.True(writeRefSate(eng, ref, AssignSet, 5))
	this.Equal(int32(5), guest.Sate.Value)
	this.True(writeRefSateMax(eng, ref, AssignSet, 20))
	this.Equal(int32(20), guest.SateMax.Value)
	this.True(writeRefMorale(eng, ref, AssignSet, 6)) // 引用 morale 走一般運算(無餐廳特例)
	this.Equal(int32(6), guest.Morale.Value)
	this.True(writeRefMorale(eng, ref, AssignSub, 2)) // 引用 -= 為一般減
	this.Equal(int32(4), guest.Morale.Value)
	this.True(writeRefMoraleMax(eng, ref, AssignSet, 15))
	this.Equal(int32(15), guest.MoraleMax.Value)
	this.True(writeRefScore(eng, ref, AssignSet, 9))
	this.Equal(int32(9), guest.Score.Value)
	this.True(writeRefScoreMax(eng, ref, AssignSet, 30))
	this.Equal(int32(30), guest.ScoreMax.Value)

	this.True(writeRefSateSeal(eng, ref, AssignLock, 0)) // 純鎖
	this.Equal(int32(1), guest.SateSeal.Lock)
	this.True(writeRefCalmSeal(eng, ref, AssignLock, 0))
	this.Equal(int32(1), guest.CalmSeal.Lock)
}

func (this *SuiteAttrRefWrite) TestAttrRefWriteTypeMismatch() {
	eng := &Engine{runtime: NewRuntime(0)}
	guestWrong := guestRef{guest: &Guest{}} // 以顧客引用寫卡牌屬性,全應失敗

	for _, name := range []string{
		"cost", "extraRunMin", "extraRunMax", "cardSeal", "keep", "playExile", "unplayExile",
	} {
		this.False(attrRefWrite[name](eng, guestWrong, AssignSet, 1), name)
	} // for

	cardWrong := cardRef{card: &Card{}} // 以卡牌引用寫顧客屬性,全應失敗

	for _, name := range []string{
		"calm", "sate", "sateMax", "morale", "moraleMax", "score", "scoreMax", "sateSeal", "calmSeal",
	} {
		this.False(attrRefWrite[name](eng, cardWrong, AssignSet, 1), name)
	} // for
}

func (this *SuiteAttrRefWrite) TestHasAttrRefWrite() {
	this.True(HasAttrRefWrite("cost"))
	this.True(HasAttrRefWrite("calm"))
	this.True(HasAttrRefWrite("cardSeal"))
	this.False(HasAttrRefWrite("cardID"))  // 唯讀
	this.False(HasAttrRefWrite("inHand"))  // 唯讀
	this.False(HasAttrRefWrite("guestID")) // 唯讀
	this.False(HasAttrRefWrite("nope"))
}
