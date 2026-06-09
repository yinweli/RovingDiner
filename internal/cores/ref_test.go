package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteRef(t *testing.T) {
	suite.Run(t, new(SuiteRef))
}

// SuiteRef 驗證物件引用基元(ref.go):cardRef / guestRef 的 Same 同一性,編碼 cardValue / guestValue / selfValue 與解碼 asCard / asGuest。
type SuiteRef struct {
	suite.Suite
}

func (this *SuiteRef) TestCardRefSame() {
	card1 := &Card{InstanceID: 1}
	card2 := &Card{InstanceID: 2}

	this.True(cardRef{card: card1}.Same(cardRef{card: card1}))
	this.True(cardRef{card: card1}.Same(cardRef{card: &Card{InstanceID: 1}})) // 不同指標、同實例編號
	this.False(cardRef{card: card1}.Same(cardRef{card: card2}))
	this.False(cardRef{card: card1}.Same(guestRef{guest: &Guest{InstanceID: 1}})) // 跨型別
}

func (this *SuiteRef) TestGuestRefSame() {
	guest1 := &Guest{InstanceID: 1}
	guest2 := &Guest{InstanceID: 2}

	this.True(guestRef{guest: guest1}.Same(guestRef{guest: guest1}))
	this.True(guestRef{guest: guest1}.Same(guestRef{guest: &Guest{InstanceID: 1}}))
	this.False(guestRef{guest: guest1}.Same(guestRef{guest: guest2}))
	this.False(guestRef{guest: guest1}.Same(cardRef{card: &Card{InstanceID: 1}}))
}

func (this *SuiteRef) TestCardValue() {
	this.True(cardValue(nil).IsNone())

	value := cardValue(&Card{InstanceID: 5})
	this.True(value.IsRef())
	this.True(value.Ref().(cardRef).Same(cardRef{card: &Card{InstanceID: 5}}))
}

func (this *SuiteRef) TestGuestValue() {
	this.True(guestValue(nil).IsNone())

	value := guestValue(&Guest{InstanceID: 7})
	this.True(value.IsRef())
	this.True(value.Ref().(guestRef).Same(guestRef{guest: &Guest{InstanceID: 7}}))
}

func (this *SuiteRef) TestSelfValue() {
	// 未綁定 → 失敗
	_, ok := selfValue(nil)
	this.False(ok)

	// 綁定空物件 → none
	value, ok := selfValue(&Self{})
	this.True(ok)
	this.True(value.IsNone())

	// 綁定卡牌 → 卡牌引用
	value, ok = selfValue(&Self{Card: &Card{InstanceID: 3}})
	this.True(ok)
	this.True(value.IsRef())
	this.True(value.Ref().(cardRef).Same(cardRef{card: &Card{InstanceID: 3}}))

	// 綁定顧客 → 顧客引用
	value, ok = selfValue(&Self{Guest: &Guest{InstanceID: 4}})
	this.True(ok)
	this.True(value.IsRef())
	this.True(value.Ref().(guestRef).Same(guestRef{guest: &Guest{InstanceID: 4}}))
}

func (this *SuiteRef) TestAsCard() {
	card := &Card{InstanceID: 3}

	value, ok := asCard(cardRef{card: card})
	this.True(ok)
	this.Same(card, value) // 拆回同一卡牌實例

	_, ok = asCard(guestRef{guest: &Guest{}}) // 非卡牌引用 → 失敗
	this.False(ok)
}

func (this *SuiteRef) TestAsGuest() {
	guest := &Guest{InstanceID: 4}

	value, ok := asGuest(guestRef{guest: guest})
	this.True(ok)
	this.Same(guest, value) // 拆回同一顧客實例

	_, ok = asGuest(cardRef{card: &Card{}}) // 非顧客引用 → 失敗
	this.False(ok)
}
