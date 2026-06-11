package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// Ref 實例引用(卡牌兼顧客);至多一欄非 nil、皆 nil 為空物件(none),零值即空物件。
// 統一效果系統的 self 綁定(【營業規格書 | 八、目標類型】)與運算式的物件引用
// (【營業規格書 | 二十三、屬性清單】);實作 exprs.Ref,運算式 == / != 經 IsSame 比實例編號。
type Ref struct {
	card  *Card  // 引用卡牌時非 nil
	guest *Guest // 引用顧客時非 nil
}

// NewRefCard 以卡牌實例建構引用;card 為 nil 時即空物件(編碼點不需自帶 nil 守衛)。
func NewRefCard(card *Card) Ref {
	return Ref{card: card}
}

// NewRefGuest 以顧客實例建構引用;guest 為 nil 時即空物件。
func NewRefGuest(guest *Guest) Ref {
	return Ref{guest: guest}
}

// GetCard 讀引用的卡牌實例(非卡牌引用回 nil)。
func (this Ref) GetCard() *Card {
	return this.card
}

// GetGuest 讀引用的顧客實例(非顧客引用回 nil)。
func (this Ref) GetGuest() *Guest {
	return this.guest
}

// IsNone 回傳是否為空物件。
func (this Ref) IsNone() bool {
	return this.card == nil && this.guest == nil
}

// IsSame 回傳是否與另一引用指向同一實例:同為卡牌 / 顧客比實例編號、同為空物件相等、型別不符 false。
func (this Ref) IsSame(other exprs.Ref) bool {
	ref, ok := other.(Ref)

	if ok == false {
		return false
	} // if

	if this.card != nil && ref.card != nil {
		return this.card.GetInstanceID() == ref.card.GetInstanceID()
	} // if

	if this.guest != nil && ref.guest != nil {
		return this.guest.GetInstanceID() == ref.guest.GetInstanceID()
	} // if

	return this.IsNone() && ref.IsNone()
}

// Value 編碼成運算式值:空物件 → none、否則 → 物件引用。唯一編碼門——
// 空 Ref 經此路由為 none,不會被包進 exprs 的引用比較(對齊【營業規格書 | 二十七、運算式 | 2】)。
func (this Ref) Value() exprs.Value {
	if this.IsNone() {
		return exprs.NewNone()
	} // if

	return exprs.NewRef(this)
}

// AsCard 把引用拆為卡牌實例(Value 編碼的逆運算);非卡牌引用回 ok=false(對齊【二十七、運算式 | 7】型別不符即失敗)。
func AsCard(ref exprs.Ref) (card *Card, ok bool) {
	value, ok := ref.(Ref)

	if ok == false || value.card == nil {
		return nil, false
	} // if

	return value.card, true
}

// AsGuest 把引用拆為顧客實例;非顧客引用回 ok=false。
func AsGuest(ref exprs.Ref) (guest *Guest, ok bool) {
	value, ok := ref.(Ref)

	if ok == false || value.guest == nil {
		return nil, false
	} // if

	return value.guest, true
}

// RefTarget 自引用取投影事件的對象編號（資料編號 + 實例編號;【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 四之二】）;
// 非卡牌 / 顧客引用回零值。屬性事件（ExecAssign 收口）與效果事件（self 對象欄）共用。
func RefTarget(ref exprs.Ref) (dataID int32, instanceID InstanceID) {
	if card, ok := AsCard(ref); ok {
		return card.GetCardID(), card.GetInstanceID()
	} // if

	if guest, ok := AsGuest(ref); ok {
		return guest.GetGuestID(), guest.GetInstanceID()
	} // if

	return 0, NoneID
}
