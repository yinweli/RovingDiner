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
	card := &Card{instanceID: 1}
	game := NewGame(0, nil, nil, nil, nil)
	ref := NewRefCard(card)

	this.True(writeRefCost(game, ref, AssignSet, 4)) // 寫鎖:帶值
	this.Equal(int32(4), card.GetCost().GetValue())
	this.True(writeRefExtraRunMin(game, ref, AssignAdd, 2))
	this.Equal(int32(2), card.GetExtraRunMin().GetValue())
	this.True(writeRefExtraRunMax(game, ref, AssignSet, 9))
	this.Equal(int32(9), card.GetExtraRunMax().GetValue())

	this.True(writeRefCardSeal(game, ref, AssignLock, 0)) // 純鎖:@
	this.Equal(int32(1), card.GetSeal().GetLock())
	this.False(writeRefCardSeal(game, ref, AssignSet, 5)) // 純鎖:帶值賦值 no-op
	this.Equal(int32(0), card.GetSeal().GetValue())
	this.True(writeRefKeep(game, ref, AssignLock, 0))
	this.Equal(int32(1), card.GetKeep().GetLock())
	this.True(writeRefPlayExile(game, ref, AssignLock, 0))
	this.Equal(int32(1), card.GetPlayExile().GetLock())
	this.True(writeRefUnplayExile(game, ref, AssignLock, 0))
	this.Equal(int32(1), card.GetUnplayExile().GetLock())
}

func (this *SuiteAttrRefWrite) TestAttrRefWriteGuest() {
	guest := &Guest{instanceID: 1}
	game := NewGame(0, nil, nil, nil, nil)
	ref := NewRefGuest(guest)

	this.True(writeRefCalm(game, ref, AssignSet, 8))
	this.Equal(int32(8), guest.GetCalm().GetValue())
	this.True(writeRefSate(game, ref, AssignSet, 5))
	this.Equal(int32(5), guest.GetSate().GetValue())
	this.True(writeRefSateMax(game, ref, AssignSet, 20))
	this.Equal(int32(20), guest.GetSateMax().GetValue())
	this.True(writeRefMorale(game, ref, AssignSet, 6)) // 引用 morale 走一般運算(無餐廳特例)
	this.Equal(int32(6), guest.GetMorale().GetValue())
	this.True(writeRefMorale(game, ref, AssignSub, 2)) // 引用 -= 為一般減
	this.Equal(int32(4), guest.GetMorale().GetValue())
	this.True(writeRefMoraleMax(game, ref, AssignSet, 15))
	this.Equal(int32(15), guest.GetMoraleMax().GetValue())
	this.True(writeRefScore(game, ref, AssignSet, 9))
	this.Equal(int32(9), guest.GetScore().GetValue())
	this.True(writeRefScoreMax(game, ref, AssignSet, 30))
	this.Equal(int32(30), guest.GetScoreMax().GetValue())

	this.True(writeRefSateSeal(game, ref, AssignLock, 0)) // 純鎖
	this.Equal(int32(1), guest.GetSateSeal().GetLock())
	this.True(writeRefCalmSeal(game, ref, AssignLock, 0))
	this.Equal(int32(1), guest.GetCalmSeal().GetLock())
}

func (this *SuiteAttrRefWrite) TestAttrRefWriteTypeMismatch() {
	game := NewGame(0, nil, nil, nil, nil)
	guestWrong := NewRefGuest(&Guest{}) // 以顧客引用寫卡牌屬性,全應失敗

	for _, name := range []string{
		"cost", "extraRunMin", "extraRunMax", "cardSeal", "keep", "playExile", "unplayExile",
	} {
		this.False(attrRefWrite[name](game, guestWrong, AssignSet, 1), name)
	} // for

	cardWrong := NewRefCard(&Card{}) // 以卡牌引用寫顧客屬性,全應失敗

	for _, name := range []string{
		"calm", "sate", "sateMax", "morale", "moraleMax", "score", "scoreMax", "sateSeal", "calmSeal",
	} {
		this.False(attrRefWrite[name](game, cardWrong, AssignSet, 1), name)
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
