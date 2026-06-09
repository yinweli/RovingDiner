package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// cardRef 卡牌實例在運算式中的物件引用;實作 exprs.Ref,== / != 比較時比實例編號。
// 對應【營業規格書 | 二十三、屬性清單 | 卡牌引用屬性】。
type cardRef struct {
	card *Card
}

// Same 回傳是否與另一引用指向同一卡牌實例(型別不符回 false)。
func (this cardRef) Same(other exprs.Ref) bool {
	ref, ok := other.(cardRef)

	if ok == false {
		return false
	} // if

	return ref.card.InstanceID == this.card.InstanceID
}

// guestRef 顧客實例在運算式中的物件引用;實作 exprs.Ref,== / != 比較時比實例編號。
// 對應【營業規格書 | 二十三、屬性清單 | 顧客引用屬性】。
type guestRef struct {
	guest *Guest
}

// Same 回傳是否與另一引用指向同一顧客實例(型別不符回 false)。
func (this guestRef) Same(other exprs.Ref) bool {
	ref, ok := other.(guestRef)

	if ok == false {
		return false
	} // if

	return ref.guest.InstanceID == this.guest.InstanceID
}

// cardValue 把卡牌實例包成運算式值:nil 回空物件、否則回卡牌引用。供物件引用型全域屬性(drawLast…)使用。
func cardValue(card *Card) exprs.Value {
	if card == nil {
		return exprs.NewNone()
	} // if

	return exprs.NewRef(cardRef{card: card})
}

// guestValue 把顧客實例包成運算式值:nil 回空物件、否則回顧客引用。供物件引用型全域屬性(seatLast…)使用。
func guestValue(guest *Guest) exprs.Value {
	if guest == nil {
		return exprs.NewNone()
	} // if

	return exprs.NewRef(guestRef{guest: guest})
}

// selfValue 解析 self:未綁定(nil)回失敗、綁定空物件回 none、綁定卡牌 / 顧客回對應引用。
// 對應【營業規格書 | 二十七、運算式 | 7】「self 未固定」與「綁定空物件」兩態之別。
func selfValue(self *Self) (result exprs.Value, ok bool) {
	if self == nil {
		return exprs.Value{}, false
	} // if

	if self.Card != nil {
		return exprs.NewRef(cardRef{card: self.Card}), true
	} // if

	if self.Guest != nil {
		return exprs.NewRef(guestRef{guest: self.Guest}), true
	} // if

	return exprs.NewNone(), true
}

// asCard 把引用拆為卡牌實例(cardValue 的逆運算);非卡牌引用回 ok=false(對齊【二十七、運算式 | 7】型別不符即失敗)。
func asCard(ref exprs.Ref) (card *Card, ok bool) {
	value, ok := ref.(cardRef)

	if ok == false {
		return nil, false
	} // if

	return value.card, true
}

// asGuest 把引用拆為顧客實例(guestValue 的逆運算);非顧客引用回 ok=false。
func asGuest(ref exprs.Ref) (guest *Guest, ok bool) {
	value, ok := ref.(guestRef)

	if ok == false {
		return nil, false
	} // if

	return value.guest, true
}
