package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteGuest(t *testing.T) {
	suite.Run(t, new(SuiteGuest))
}

// SuiteGuest 驗證顧客實例與顧客容器(guest.go):建構 / 取值 / 凍結與自動鎖,排隊佇列 / 座位列表 / 順序無關列表。
type SuiteGuest struct {
	suite.Suite
}

// TestNewGuest 驗證 NewGuest 載入顧客數值與封印鎖、飽食值初值 0、免疫表初始化;資料不存在回 nil。
func (this *SuiteGuest) TestNewGuest() {
	game := NewGame(0, 0, NewData(buildSheet(), nil), nil, nil, nil)

	guest := NewGuest(game, 501) // 顧客 501：載數值 + SateSeal bool → 鎖
	this.Require().NotNil(guest)
	this.Equal(int32(501), guest.GetGuestID())
	this.Equal(int32(10), guest.GetScoreMax().GetValue())
	this.Equal(int32(5), guest.GetMorale().GetValue())
	this.Equal(int32(3), guest.GetCalm().GetValue())
	this.Equal(int32(12), guest.GetSateMax().GetValue()) // 飽食值離場線取自顧客資料
	this.Equal(int32(0), guest.GetSate().GetValue())     // 飽食值初值 0（顧客資料無此欄）
	this.Equal(int32(0), guest.GetSate().GetLock())      // 不自動鎖（自動鎖屬 guestSpawn）
	this.Equal(int32(1), guest.GetSateSeal().GetLock())  // sheet true → 鎖定計數 1
	this.Equal(int32(0), guest.GetCalmSeal().GetLock())  // sheet 未設 → 0
	this.Equal(int32(0), guest.GetEffectImmune().Get(1)) // 免疫表初始化（空表）
	this.NotEqual(InstanceID(0), guest.GetInstanceID())  // 配發實例編號

	this.Nil(NewGuest(game, 999)) // 顧客資料不存在 → nil
}

// TestGuestGetInstanceID 驗證 GetInstanceID 取回實例編號。
func (this *SuiteGuest) TestGuestGetInstanceID() {
	this.Equal(InstanceID(7), (&Guest{instanceID: 7}).GetInstanceID())
}

// TestGuestGetGuestID 驗證 GetGuestID 取回顧客編號。
func (this *SuiteGuest) TestGuestGetGuestID() {
	this.Equal(int32(501), (&Guest{guestID: 501}).GetGuestID())
}

// TestGuestGetSeatID 驗證 GetSeatID 取回座位編號。
func (this *SuiteGuest) TestGuestGetSeatID() {
	this.Equal(int32(3), (&Guest{seatID: 3}).GetSeatID())
}

// TestGuestGetScore 驗證 GetScore 取回滿意值組件（寫入經組件可見）。
func (this *SuiteGuest) TestGuestGetScore() {
	guest := &Guest{score: NewValue(9, 0)}
	this.Equal(int32(9), guest.GetScore().GetValue())

	guest.GetScore().Add(1) // 經組件寫入 → 同一實體
	this.Equal(int32(10), guest.GetScore().GetValue())
}

// TestGuestGetScoreMax 驗證 GetScoreMax 取回滿意值上限。
func (this *SuiteGuest) TestGuestGetScoreMax() {
	this.Equal(int32(30), (&Guest{scoreMax: NewValue(30, 0)}).GetScoreMax().GetValue())
}

// TestGuestGetMorale 驗證 GetMorale 取回士氣值。
func (this *SuiteGuest) TestGuestGetMorale() {
	this.Equal(int32(6), (&Guest{morale: NewValue(6, 0)}).GetMorale().GetValue())
}

// TestGuestGetMoraleMax 驗證 GetMoraleMax 取回士氣值上限。
func (this *SuiteGuest) TestGuestGetMoraleMax() {
	this.Equal(int32(15), (&Guest{moraleMax: NewValue(15, 0)}).GetMoraleMax().GetValue())
}

// TestGuestGetSate 驗證 GetSate 取回飽食值。
func (this *SuiteGuest) TestGuestGetSate() {
	this.Equal(int32(4), (&Guest{sate: NewValue(4, 0)}).GetSate().GetValue())
}

// TestGuestGetSateMax 驗證 GetSateMax 取回飽食值離場線。
func (this *SuiteGuest) TestGuestGetSateMax() {
	this.Equal(int32(20), (&Guest{sateMax: NewValue(20, 0)}).GetSateMax().GetValue())
}

// TestGuestGetCalm 驗證 GetCalm 取回耐心值。
func (this *SuiteGuest) TestGuestGetCalm() {
	this.Equal(int32(8), (&Guest{calm: NewValue(8, 0)}).GetCalm().GetValue())
}

// TestGuestGetSateSeal 驗證 GetSateSeal 取回封印飽食技能。
func (this *SuiteGuest) TestGuestGetSateSeal() {
	this.Equal(int32(1), (&Guest{sateSeal: NewValueLock(true)}).GetSateSeal().GetLock())
}

// TestGuestGetCalmSeal 驗證 GetCalmSeal 取回封印耐心技能。
func (this *SuiteGuest) TestGuestGetCalmSeal() {
	this.Equal(int32(1), (&Guest{calmSeal: NewValueLock(true)}).GetCalmSeal().GetLock())
}

// TestGuestGetSateHit 驗證 GetSateHit 取回飽食門檻集合（零值可用）。
func (this *SuiteGuest) TestGuestGetSateHit() {
	guest := &Guest{}
	guest.GetSateHit().Add(10)
	this.True(guest.GetSateHit().IsHit(10))
}

// TestGuestGetCalmHit 驗證 GetCalmHit 取回耐心門檻集合（零值可用）。
func (this *SuiteGuest) TestGuestGetCalmHit() {
	guest := &Guest{}
	guest.GetCalmHit().Add(5)
	this.Equal(int32(1), guest.GetCalmHit().Count())
}

// TestGuestGetEffectImmune 驗證 GetEffectImmune 取回效果免疫計數（零值可用）。
func (this *SuiteGuest) TestGuestGetEffectImmune() {
	guest := &Guest{}
	guest.GetEffectImmune().Add(7)
	this.Equal(int32(1), guest.GetEffectImmune().Get(7))
}

// TestGuestGetSkillImmune 驗證 GetSkillImmune 取回技能免疫計數（零值可用）。
func (this *SuiteGuest) TestGuestGetSkillImmune() {
	guest := &Guest{}
	guest.GetSkillImmune().Add(3)
	this.Equal(int32(1), guest.GetSkillImmune().Get(3))
}

// TestGuestGetFreeze 驗證 GetFreeze 取回凍結起始回合。
func (this *SuiteGuest) TestGuestGetFreeze() {
	this.Equal(int32(4), (&Guest{freeze: 4}).GetFreeze())
}

// TestGuestSetFreeze 驗證 SetFreeze 寫凍結起始回合（記錄 / 歸零）。
func (this *SuiteGuest) TestGuestSetFreeze() {
	guest := &Guest{}
	guest.SetFreeze(7)
	this.Equal(int32(7), guest.GetFreeze())
	guest.SetFreeze(0)
	this.Equal(int32(0), guest.GetFreeze())
}

// TestGuestRoamLock 驗證 RoamLock 自動鎖三連（sate / sateSeal / calmSeal 各 +1,其餘不動）。
func (this *SuiteGuest) TestGuestRoamLock() {
	guest := &Guest{}
	guest.RoamLock()
	this.Equal(int32(1), guest.GetSate().GetLock())
	this.Equal(int32(1), guest.GetSateSeal().GetLock())
	this.Equal(int32(1), guest.GetCalmSeal().GetLock())
	this.Equal(int32(0), guest.GetCalm().GetLock()) // 其餘屬性不受影響
}

// TestGuestRoamUnlock 驗證 RoamUnlock 解自動鎖三連（各 -1,夾 ≥ 0）。
func (this *SuiteGuest) TestGuestRoamUnlock() {
	guest := &Guest{}
	guest.RoamLock()
	guest.RoamUnlock()
	this.Equal(int32(0), guest.GetSate().GetLock())
	this.Equal(int32(0), guest.GetSateSeal().GetLock())
	this.Equal(int32(0), guest.GetCalmSeal().GetLock())

	guest.RoamUnlock() // 已歸零再解 → 夾 ≥ 0
	this.Equal(int32(0), guest.GetSate().GetLock())
}

// TestWaitListInsert 驗證 Insert 前端插隊（後插者居首）。
func (this *SuiteGuest) TestWaitListInsert() {
	first := &Guest{instanceID: 1}
	second := &Guest{instanceID: 2}
	list := WaitList{}

	list.Insert(first)
	list.Insert(second)
	this.Require().Len(list, 2)
	this.Same(second, list[0]) // 後插者居首（優先入座）
	this.Same(first, list[1])
}

// TestWaitListPop 驗證 Pop 彈出隊首、空佇列回 nil。
func (this *SuiteGuest) TestWaitListPop() {
	first := &Guest{instanceID: 1}
	second := &Guest{instanceID: 2}
	list := WaitList{first, second}

	this.Same(first, list.Pop()) // 先進先出
	this.Same(second, list.Pop())
	this.Empty(list)
	this.Nil(list.Pop()) // 空佇列 → nil
}

// TestWaitListRemove 驗證 Remove 依實例編號移除、未命中不變。
func (this *SuiteGuest) TestWaitListRemove() {
	keep := &Guest{instanceID: 1}
	list := WaitList{keep, {instanceID: 2}}

	list.Remove(InstanceID(2)) // 命中 → 移除
	this.Require().Len(list, 1)
	this.Same(keep, list[0])

	list.Remove(InstanceID(9)) // 未命中 → 不變
	this.Len(list, 1)
}

// TestWaitListFind 驗證 Find 依實例編號查找、未命中回 nil。
func (this *SuiteGuest) TestWaitListFind() {
	guest := &Guest{instanceID: 5}
	list := WaitList{{instanceID: 1}, guest}

	this.Same(guest, list.Find(InstanceID(5)))
	this.Nil(list.Find(InstanceID(9))) // 未命中 → nil
}

// TestSeatListPlace 驗證 Place 成對寫入:設顧客座位編號 + 登記座位。
func (this *SuiteGuest) TestSeatListPlace() {
	guest := &Guest{instanceID: 1}
	seat := SeatList{}

	seat.Place(3, guest)
	this.Same(guest, seat[3])
	this.Equal(int32(3), guest.GetSeatID())
}

// TestSeatListRemove 驗證 Remove 成對寫入:刪座位鍵 + 歸零顧客座位編號;未入座者 no-op。
func (this *SuiteGuest) TestSeatListRemove() {
	guest := &Guest{instanceID: 1}
	seat := SeatList{}
	seat.Place(3, guest)

	seat.Remove(guest)
	this.Nil(seat[3])
	this.Equal(int32(0), guest.GetSeatID())

	seat.Remove(&Guest{instanceID: 2}) // 未入座（seatID 0 無鍵）→ no-op
	this.Empty(seat)
}

// TestSeatListFind 驗證 Find 依實例編號查找在座顧客;未命中回 nil。
func (this *SuiteGuest) TestSeatListFind() {
	guest := &Guest{instanceID: 1}
	seat := SeatList{}
	seat.Place(3, guest)

	this.Same(guest, seat.Find(1))
	this.Nil(seat.Find(99)) // 未命中
}

// TestSeatListSorted 驗證 Sorted 依座位編號升序取顧客（決定性）。
func (this *SuiteGuest) TestSeatListSorted() {
	g1 := &Guest{instanceID: 1}
	g2 := &Guest{instanceID: 2}
	g3 := &Guest{instanceID: 3}
	seat := SeatList{}
	seat.Place(5, g1)
	seat.Place(2, g2)
	seat.Place(9, g3)

	this.Equal([]*Guest{g2, g1, g3}, seat.Sorted()) // 座 2 < 5 < 9
	this.Empty(SeatList{}.Sorted())                 // 空列表 → 空集合
}

// TestSeatListOccupied 驗證 Occupied 計座位編號列表中的占用數。
func (this *SuiteGuest) TestSeatListOccupied() {
	seat := SeatList{}
	seat.Place(1, &Guest{instanceID: 1})
	seat.Place(3, &Guest{instanceID: 3})

	this.Equal(int32(2), seat.Occupied([]int32{1, 2, 3})) // 座 1、3 占用,座 2 空
	this.Equal(int32(0), seat.Occupied([]int32{2, 4}))    // 皆空
	this.Equal(int32(0), seat.Occupied(nil))              // 空列表
}

// TestGuestListPush 驗證 Push 加入列表尾端。
func (this *SuiteGuest) TestGuestListPush() {
	first := &Guest{instanceID: 1}
	second := &Guest{instanceID: 2}
	list := GuestList{}

	list.Push(first)
	list.Push(second)
	this.Require().Len(list, 2)
	this.Same(first, list[0])
	this.Same(second, list[1])
}

// TestGuestListRemove 驗證 Remove 依實例編號移除、未命中不變。
func (this *SuiteGuest) TestGuestListRemove() {
	keep := &Guest{instanceID: 1}
	list := GuestList{keep, {instanceID: 2}}

	list.Remove(InstanceID(2)) // 命中 → 移除
	this.Require().Len(list, 1)
	this.Same(keep, list[0])

	list.Remove(InstanceID(9)) // 未命中 → 不變
	this.Len(list, 1)
}

// TestGuestListFind 驗證 Find 依實例編號查找、未命中回 nil。
func (this *SuiteGuest) TestGuestListFind() {
	guest := &Guest{instanceID: 5}
	list := GuestList{{instanceID: 1}, guest}

	this.Same(guest, list.Find(InstanceID(5)))
	this.Nil(list.Find(InstanceID(9))) // 未命中 → nil
}

// === 測試輔助（置尾） ===
