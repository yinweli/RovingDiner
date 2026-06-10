package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteEffect(t *testing.T) {
	suite.Run(t, new(SuiteEffect))
}

// SuiteEffect 驗證效果實例與效果佇列(effect.go):建構 / 取值 / 結束回合與堆疊寫入,佇列入列 / 移除 / 同份查找 / 作用順序排序。
type SuiteEffect struct {
	suite.Suite
}

// TestNewEffect 驗證 NewEffect 建構效果實例、結束回合依作用回合（0 整場 / N 期限）;資料不存在回 nil。
func (this *SuiteEffect) TestNewEffect() {
	runtime := NewRuntime(0)
	runtime.Game.round = NewValue(5, 0)
	eng := this.engine(runtime)

	guest := &Guest{instanceID: 99}
	effect := NewEffect(eng, 401, NewRefGuest(guest), 3) // RunRound 2 → Expire = 5 + 2 − 1
	this.Require().NotNil(effect)
	this.Equal(int32(401), effect.GetEffectID())
	this.Equal(int32(6), effect.GetExpire())
	this.Equal(int32(3), effect.GetStack())
	this.Equal(guest, effect.GetSelf().GetGuest())
	this.NotEqual(InstanceID(0), effect.GetInstanceID()) // 配發實例編號

	zero := NewEffect(eng, 402, Ref{}, 1) // RunRound 0 → Expire 0（整場保留）
	this.Require().NotNil(zero)
	this.Equal(int32(0), zero.GetExpire())

	this.Nil(NewEffect(eng, 999, Ref{}, 1)) // 查無效果資料 → nil
}

// TestEffectGetInstanceID 驗證 GetInstanceID 取回實例編號。
func (this *SuiteEffect) TestEffectGetInstanceID() {
	this.Equal(InstanceID(7), (&Effect{instanceID: 7}).GetInstanceID())
}

// TestEffectGetEffectID 驗證 GetEffectID 取回效果編號。
func (this *SuiteEffect) TestEffectGetEffectID() {
	this.Equal(int32(401), (&Effect{effectID: 401}).GetEffectID())
}

// TestEffectGetExpire 驗證 GetExpire 取回結束回合。
func (this *SuiteEffect) TestEffectGetExpire() {
	this.Equal(int32(6), (&Effect{expire: 6}).GetExpire())
}

// TestEffectGetStack 驗證 GetStack 取回當前堆疊層數。
func (this *SuiteEffect) TestEffectGetStack() {
	this.Equal(int32(3), (&Effect{stack: 3}).GetStack())
}

// TestEffectGetSelf 驗證 GetSelf 取回 self 物件。
func (this *SuiteEffect) TestEffectGetSelf() {
	guest := &Guest{instanceID: 9}
	this.Equal(guest, (&Effect{self: NewRefGuest(guest)}).GetSelf().GetGuest())
	this.True((&Effect{}).GetSelf().IsNone()) // 零值 → 空物件 self
}

// TestEffectSetExpire 驗證 SetExpire 覆寫結束回合。
func (this *SuiteEffect) TestEffectSetExpire() {
	effect := &Effect{expire: 5}
	effect.SetExpire(11)
	this.Equal(int32(11), effect.GetExpire())
}

// TestEffectStackAdd 驗證 StackAdd 疊層、上限夾制（0 = 無上限）與實際增量回傳。
func (this *SuiteEffect) TestEffectStackAdd() {
	effect := &Effect{stack: 2}

	this.Equal(int32(3), effect.StackAdd(3, 0)) // 無上限 → 全加
	this.Equal(int32(5), effect.GetStack())

	this.Equal(int32(1), effect.StackAdd(4, 6)) // 夾上限 6 → 實際 +1
	this.Equal(int32(6), effect.GetStack())

	this.Equal(int32(0), effect.StackAdd(1, 6)) // 已達上限 → +0
	this.Equal(int32(6), effect.GetStack())
}

// TestEffectListPush 驗證 Push 加入佇列尾端。
func (this *SuiteEffect) TestEffectListPush() {
	list := EffectList{}

	list.Push(&Effect{instanceID: 1, effectID: 401})
	list.Push(&Effect{instanceID: 2, effectID: 402})
	this.Require().Len(list, 2)
	this.Equal(InstanceID(1), list[0].GetInstanceID())
	this.Equal(InstanceID(2), list[1].GetInstanceID())
}

// TestEffectListRemove 驗證 Remove 依實例編號移除、未命中不變。
func (this *SuiteEffect) TestEffectListRemove() {
	keep := &Effect{instanceID: 1, effectID: 401}
	drop := &Effect{instanceID: 2, effectID: 402}
	list := EffectList{keep, drop}

	list.Remove(drop.GetInstanceID()) // 命中 → 移除
	this.Require().Len(list, 1)
	this.Equal(InstanceID(1), list[0].GetInstanceID())

	list.Remove(InstanceID(9)) // 未命中 → 不變
	this.Len(list, 1)
}

// TestEffectListFind 驗證 Find 同份查找:效果編號同 && self 同,空物件 self 彼此相同。
func (this *SuiteEffect) TestEffectListFind() {
	guest := &Guest{instanceID: 11}
	bound := &Effect{instanceID: 1, effectID: 401, self: NewRefGuest(guest)}
	unbound := &Effect{instanceID: 2, effectID: 402} // 零值 self = 空物件
	list := EffectList{bound, unbound}

	this.Same(bound, list.Find(NewRefGuest(guest), 401))
	this.Same(unbound, list.Find(Ref{}, 402))                     // 無目標 self 彼此相同
	this.Nil(list.Find(NewRefGuest(&Guest{instanceID: 12}), 401)) // self 不同 → 無
	this.Nil(list.Find(NewRefGuest(guest), 402))                  // 效果編號不同 → 無
	this.Nil(list.Find(Ref{}, 999))                               // 查無 → nil
}

// TestEffectSort 驗證 effectSort 依作用順序排序:大者優先、同序效果編號小者優先、查無資料殿後。
func (this *SuiteEffect) TestEffectSort() {
	runtime := NewRuntime(0)
	eng := this.engine(runtime)

	effect := []*Effect{
		{instanceID: 1, effectID: 402}, // RunOrder 10
		{instanceID: 2, effectID: 403}, // RunOrder 20（最大 → 最先）
		{instanceID: 3, effectID: 999}, // 查無資料 → RunOrder 0（最後）
		{instanceID: 4, effectID: 401}, // RunOrder 10、同序 EffectID 401 < 402 → 先於 402
	}
	effectSort(eng, effect)

	order := []int32{effect[0].GetEffectID(), effect[1].GetEffectID(), effect[2].GetEffectID(), effect[3].GetEffectID()}
	this.Equal([]int32{403, 401, 402, 999}, order)
}

// === 測試輔助（置尾） ===

func (this *SuiteEffect) engine(runtime *Runtime) *Engine {
	return NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{}, nil)
}
