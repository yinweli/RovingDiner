package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteRef(t *testing.T) {
	suite.Run(t, new(SuiteRef))
}

// SuiteRef 驗證實例引用(ref.go): 建構 / 取值 / 空物件與同一性判定 / 運算式值編碼, 與 asCard / asGuest 解碼。
type SuiteRef struct {
	suite.Suite
}

// TestNewRefCard 驗證 NewRefCard 以卡牌建構引用、nil 即空物件。
func (this *SuiteRef) TestNewRefCard() {
	card := &Card{instanceID: 1}
	this.Same(card, NewRefCard(card).GetCard())
	this.True(NewRefCard(nil).IsNone()) // nil → 空物件
}

// TestNewRefGuest 驗證 NewRefGuest 以顧客建構引用、nil 即空物件。
func (this *SuiteRef) TestNewRefGuest() {
	guest := &Guest{instanceID: 2}
	this.Same(guest, NewRefGuest(guest).GetGuest())
	this.True(NewRefGuest(nil).IsNone()) // nil → 空物件
}

// TestRefGetCard 驗證 GetCard 取回卡牌實例、非卡牌引用回 nil。
func (this *SuiteRef) TestRefGetCard() {
	card := &Card{instanceID: 3}
	this.Same(card, NewRefCard(card).GetCard())
	this.Nil(NewRefGuest(&Guest{}).GetCard()) // 非卡牌引用 → nil
}

// TestRefGetGuest 驗證 GetGuest 取回顧客實例、非顧客引用回 nil。
func (this *SuiteRef) TestRefGetGuest() {
	guest := &Guest{instanceID: 4}
	this.Same(guest, NewRefGuest(guest).GetGuest())
	this.Nil(NewRefCard(&Card{}).GetGuest()) // 非顧客引用 → nil
}

// TestRefIsNone 驗證 IsNone 空物件判定(零值即空物件)。
func (this *SuiteRef) TestRefIsNone() {
	this.True(Ref{}.IsNone()) // 零值 → 空物件
	this.False(NewRefCard(&Card{}).IsNone())
	this.False(NewRefGuest(&Guest{}).IsNone())
}

// TestRefIsSame 驗證 IsSame 同一性: 同型別比實例編號、空物件彼此相等、跨型別 / 外來引用不相等。
func (this *SuiteRef) TestRefIsSame() {
	card1 := NewRefCard(&Card{instanceID: 1})
	card2 := NewRefCard(&Card{instanceID: 2})
	guest1 := NewRefGuest(&Guest{instanceID: 1})
	guest2 := NewRefGuest(&Guest{instanceID: 2})

	this.True(card1.IsSame(NewRefCard(&Card{instanceID: 1}))) // 不同指標、同實例編號
	this.False(card1.IsSame(card2))
	this.True(guest1.IsSame(NewRefGuest(&Guest{instanceID: 1})))
	this.False(guest1.IsSame(guest2))
	this.False(card1.IsSame(guest1))       // 跨型別(同實例編號)不相等
	this.True(Ref{}.IsSame(Ref{}))         // 空物件彼此相等
	this.False(card1.IsSame(Ref{}))        // 引用 vs 空物件
	this.False(card1.IsSame(foreignRef{})) // 非本引擎引用 → 型別斷言不符
}

// TestRefValue 驗證 Value 編碼: 空物件 → none、卡牌 / 顧客 → 物件引用。
func (this *SuiteRef) TestRefValue() {
	this.True(Ref{}.Value().IsNone()) // 空物件 → none
	this.True(NewRefCard(nil).Value().IsNone())

	card := NewRefCard(&Card{instanceID: 5})
	value := card.Value()
	this.True(value.IsRef())
	this.True(card.IsSame(value.Ref()))

	guest := NewRefGuest(&Guest{instanceID: 7})
	value = guest.Value()
	this.True(value.IsRef())
	this.True(guest.IsSame(value.Ref()))
}

// TestAsCard 驗證 asCard 解碼: 卡牌引用拆回實例、非卡牌引用 / 外來引用失敗。
func (this *SuiteRef) TestAsCard() {
	card := &Card{instanceID: 3}

	value, ok := AsCard(NewRefCard(card))
	this.True(ok)
	this.Same(card, value) // 拆回同一卡牌實例

	_, ok = AsCard(NewRefGuest(&Guest{})) // 非卡牌引用 → 失敗
	this.False(ok)

	_, ok = AsCard(foreignRef{}) // 非本引擎引用 → 失敗
	this.False(ok)
}

// TestAsGuest 驗證 asGuest 解碼: 顧客引用拆回實例、非顧客引用 / 外來引用失敗。
func (this *SuiteRef) TestAsGuest() {
	guest := &Guest{instanceID: 4}

	value, ok := AsGuest(NewRefGuest(guest))
	this.True(ok)
	this.Same(guest, value) // 拆回同一顧客實例

	_, ok = AsGuest(NewRefCard(&Card{})) // 非顧客引用 → 失敗
	this.False(ok)

	_, ok = AsGuest(foreignRef{}) // 非本引擎引用 → 失敗
	this.False(ok)
}

// TestRefTarget 驗證 RefTarget 取投影事件對象編號: 卡牌 / 顧客引用回(資料編號, 實例編號)、空引用回零值。
func (this *SuiteRef) TestRefTarget() {
	dataID, instanceID := RefTarget(NewRefCard(&Card{cardID: 103, instanceID: 1}))
	this.Equal(int32(103), dataID)
	this.Equal(InstanceID(1), instanceID)

	dataID, instanceID = RefTarget(NewRefGuest(&Guest{guestID: 501, instanceID: 4}))
	this.Equal(int32(501), dataID)
	this.Equal(InstanceID(4), instanceID)

	dataID, instanceID = RefTarget(Ref{}) // 空引用 → 零值
	this.Equal(int32(0), dataID)
	this.Equal(NoneID, instanceID)
}

// foreignRef 測試用外來引用(非 cores.Ref 的 exprs.Ref 實作); 驗證型別斷言不符分支。
type foreignRef struct{}

func (this foreignRef) IsSame(other exprs.Ref) bool {
	return false
}
