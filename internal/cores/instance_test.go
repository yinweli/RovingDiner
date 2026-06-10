package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteInstance(t *testing.T) {
	suite.Run(t, new(SuiteInstance))
}

// SuiteInstance 驗證 instance.go 各實例型別的方法行為與建立函式。
type SuiteInstance struct {
	suite.Suite
}

// TestSelfIsNone 驗證 Self.IsNone 僅在卡牌與顧客皆為 nil 時回報空物件。
func (this *SuiteInstance) TestSelfIsNone() {
	this.True(Self{}.IsNone())                                // 空物件
	this.False(Self{Card: &Card{}}.IsNone())                  // self 為卡牌
	this.False(Self{Guest: &Guest{}}.IsNone())                // self 為顧客
	this.False(Self{Card: &Card{}, Guest: &Guest{}}.IsNone()) // 兩欄皆非 nil（防禦性，正常不應發生）
}

// TestNewCard 驗證 newCard 載入卡牌資料初始值、bool 欄轉鎖、SkillID 取技能效果列表;資料不存在回 nil。
func (this *SuiteInstance) TestNewCard() {
	runtime := NewRuntime(0)
	eng := this.engine(runtime)

	card := newCard(eng, 103) // 卡 103：Cost 2、Keep / Seal bool → 鎖、SkillID 301 → 效果列表
	this.Require().NotNil(card)
	this.Equal(int32(103), card.CardID)
	this.Equal(int32(2), card.Cost.GetValue())
	this.Equal(int32(1), card.Keep.GetLock())      // bool 欄 → 鎖定計數
	this.Equal(int32(1), card.Seal.GetLock())      //
	this.Equal(int32(0), card.PlayExile.GetLock()) // sheet 未設 → 0
	this.Equal([]int32{401, 402}, card.EffectID)   // SkillID 301 → Skill.EffectID
	this.NotEqual(InstanceID(0), card.InstanceID)  // 配發實例編號

	this.Nil(newCard(eng, 999)) // 卡牌資料不存在 → nil
}

// TestCopyCard 驗證 copyCard 淺複製載初始值、深複製複製 source 當前狀態與效果列表;實例編號皆重生。
func (this *SuiteInstance) TestCopyCard() {
	runtime := NewRuntime(0)
	source := &Card{InstanceID: runtime.NextID(), CardID: 103, Cost: NewValue(9, 0), EffectID: []int32{777}}
	eng := this.engine(runtime)

	shallow := copyCard(eng, source, false) // 淺複製：載卡牌資料初始值（非 source 當前狀態）
	this.Require().NotNil(shallow)
	this.Equal(int32(2), shallow.Cost.GetValue())   // 初始費用 2（非 source 的 9）
	this.Equal([]int32{401, 402}, shallow.EffectID) // 技能效果列表
	this.NotEqual(source.InstanceID, shallow.InstanceID)

	deep := copyCard(eng, source, true) // 深複製：複製 source 當前狀態 + 效果列表
	this.Require().NotNil(deep)
	this.Equal(int32(9), deep.Cost.GetValue())        // source 當前費用
	this.Equal([]int32{777}, deep.EffectID)           // source 效果列表深複製
	this.NotEqual(source.InstanceID, deep.InstanceID) // 實例編號重生
}

// TestNewGuest 驗證 newGuest 載入顧客數值與封印鎖、飽食值初值 0、免疫表初始化;資料不存在回 nil。
func (this *SuiteInstance) TestNewGuest() {
	runtime := NewRuntime(0)
	eng := this.engine(runtime)

	guest := newGuest(eng, 501) // 顧客 501：載數值 + SateSeal bool → 鎖
	this.Require().NotNil(guest)
	this.Equal(int32(501), guest.GuestID)
	this.Equal(int32(10), guest.ScoreMax.GetValue())
	this.Equal(int32(5), guest.Morale.GetValue())
	this.Equal(int32(3), guest.Calm.GetValue())
	this.Equal(int32(12), guest.SateMax.GetValue()) // 飽食值離場線取自顧客資料
	this.Equal(int32(0), guest.Sate.GetValue())     // 飽食值初值 0（顧客資料無此欄）
	this.Equal(int32(0), guest.Sate.GetLock())      // 不自動鎖（自動鎖屬 guestSpawn）
	this.Equal(int32(1), guest.SateSeal.GetLock())  // sheet true → 鎖定計數 1
	this.Equal(int32(0), guest.CalmSeal.GetLock())  // sheet 未設 → 0
	this.NotNil(guest.EffectImmune)                 // 免疫表初始化
	this.NotEqual(InstanceID(0), guest.InstanceID)  // 配發實例編號

	this.Nil(newGuest(eng, 999)) // 顧客資料不存在 → nil
}

// TestNewEffect 驗證 newEffect 建構效果實例、結束回合依作用回合（0 整場 / N 期限）;資料不存在回 nil。
func (this *SuiteInstance) TestNewEffect() {
	runtime := NewRuntime(0)
	runtime.Game.Round = NewValue(5, 0)
	eng := this.engine(runtime)

	guest := &Guest{InstanceID: 99}
	effect := newEffect(eng, 401, Self{Guest: guest}, 3) // RunRound 2 → Expire = 5 + 2 − 1
	this.Require().NotNil(effect)
	this.Equal(int32(401), effect.EffectID)
	this.Equal(int32(6), effect.Expire)
	this.Equal(int32(3), effect.Stack)
	this.Equal(guest, effect.Self.Guest)
	this.NotEqual(InstanceID(0), effect.InstanceID) // 配發實例編號

	zero := newEffect(eng, 402, Self{}, 1) // RunRound 0 → Expire 0（整場保留）
	this.Require().NotNil(zero)
	this.Equal(int32(0), zero.Expire)

	this.Nil(newEffect(eng, 999, Self{}, 1)) // 查無效果資料 → nil
}

// === 測試輔助（置尾） ===

func (this *SuiteInstance) engine(runtime *Runtime) *Engine {
	return NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{}, nil)
}
