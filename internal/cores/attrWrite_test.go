package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteAttrWrite(t *testing.T) {
	suite.Run(t, new(SuiteAttrWrite))
}

// SuiteAttrWrite 驗證全域屬性寫入詞彙表(attrWrite.go):寫鎖各賦值符 / Lock>0 no-op、護盾 / 格擋夾 0、
// 純鎖屬性、round / roundMax / roundLeft 衍生、morale -= 特例(格擋 / 護盾 / morale / 鎖定 / 受損欄位 / 來源)。
type SuiteAttrWrite struct {
	suite.Suite
}

func (this *SuiteAttrWrite) TestAttrWriteValue() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}

	this.True(writeScore(eng, AssignSet, 10)) // = 設值
	this.Equal(int32(10), runtime.Game.Score.GetValue())
	this.True(writeScore(eng, AssignAdd, 5)) // += 加
	this.Equal(int32(15), runtime.Game.Score.GetValue())
	this.True(writeScore(eng, AssignSub, 3)) // -= 減(score 無特例,走一般)
	this.Equal(int32(12), runtime.Game.Score.GetValue())
	this.True(writeScore(eng, AssignMul, 2)) // *= 乘
	this.Equal(int32(24), runtime.Game.Score.GetValue())
	this.True(writeScore(eng, AssignDiv, 5)) // /= 除(24/5=4.8 → half-away 5)
	this.Equal(int32(5), runtime.Game.Score.GetValue())
	this.True(writeScore(eng, AssignMod, 3)) // %= 取餘(5%3=2)
	this.Equal(int32(2), runtime.Game.Score.GetValue())

	this.False(writeScore(eng, AssignDiv, 0)) // 除 0 → no-op
	this.Equal(int32(2), runtime.Game.Score.GetValue())
	this.False(writeScore(eng, AssignMod, 0)) // 取餘 0 → no-op
	this.Equal(int32(2), runtime.Game.Score.GetValue())
}

func (this *SuiteAttrWrite) TestAttrWriteLock() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}

	runtime.Game.Score = NewValue(9, 0) // 設初值 9(鎖定前)

	this.True(writeScore(eng, AssignLock, 0)) // @ 計數 +1
	this.True(writeScore(eng, AssignLock, 0))
	this.Equal(int32(2), runtime.Game.Score.GetLock())

	this.False(writeScore(eng, AssignSet, 100)) // Lock>0 → 帶值賦值 no-op
	this.Equal(int32(9), runtime.Game.Score.GetValue())

	this.True(writeScore(eng, AssignUnlock, 0)) // # 計數 -1
	this.True(writeScore(eng, AssignUnlock, 0))
	this.Equal(int32(0), runtime.Game.Score.GetLock())

	this.False(writeScore(eng, AssignUnlock, 0)) // # 對 0 計數 → no-op(不降負)
	this.Equal(int32(0), runtime.Game.Score.GetLock())

	this.True(writeScore(eng, AssignSet, 100)) // 解鎖後帶值賦值恢復
	this.Equal(int32(100), runtime.Game.Score.GetValue())
}

func (this *SuiteAttrWrite) TestAttrWriteClamp() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}

	runtime.Game.MoraleShield = NewValue(3, 0)
	this.True(writeMoraleShield(eng, AssignSub, 10)) // 3-10=-7 → 夾 0
	this.Equal(int32(0), runtime.Game.MoraleShield.GetValue())
	this.True(writeMoraleShield(eng, AssignSet, 5)) // 正值不夾
	this.Equal(int32(5), runtime.Game.MoraleShield.GetValue())

	this.True(writeMoraleBlock(eng, AssignSet, -5)) // 夾 0
	this.Equal(int32(0), runtime.Game.MoraleBlock.GetValue())
}

func (this *SuiteAttrWrite) TestAttrWriteLockOnly() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}

	this.True(writeEnergyKeep(eng, AssignLock, 0)) // 純鎖屬性 @ 計數 +1
	this.Equal(int32(1), runtime.Game.EnergyKeep.GetLock())
	this.False(writeEnergyKeep(eng, AssignSet, 5)) // 帶值賦值對鎖屬性 no-op
	this.Equal(int32(0), runtime.Game.EnergyKeep.GetValue())
	this.True(writeEnergyKeep(eng, AssignUnlock, 0))
	this.Equal(int32(0), runtime.Game.EnergyKeep.GetLock())
}

func (this *SuiteAttrWrite) TestAttrWriteRound() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Round = NewValue(3, 0)
	runtime.Game.RoundMax = NewValue(10, 0)

	this.True(writeRound(eng, AssignAdd, 2)) // 回合 5
	this.Equal(int32(5), runtime.Game.Round.GetValue())
	this.True(writeRoundMax(eng, AssignSet, 12)) // 回合上限 12
	this.Equal(int32(12), runtime.Game.RoundMax.GetValue())
	this.False(writeRound(eng, AssignLock, 0)) // 寫屬性無鎖 → @ no-op(回合與鎖定計數皆不變)
	this.Equal(int32(5), runtime.Game.Round.GetValue())
	this.Equal(int32(0), runtime.Game.Round.GetLock())
	this.False(writeRound(eng, AssignUnlock, 0)) // 寫屬性無鎖 → # no-op
	this.Equal(int32(0), runtime.Game.Round.GetLock())

	this.True(writeRound(eng, AssignSub, 1)) // 回合 4
	this.Equal(int32(4), runtime.Game.Round.GetValue())
	this.True(writeRound(eng, AssignMul, 3)) // 回合 12
	this.Equal(int32(12), runtime.Game.Round.GetValue())
	this.True(writeRound(eng, AssignDiv, 2)) // 回合 6
	this.Equal(int32(6), runtime.Game.Round.GetValue())
	this.True(writeRound(eng, AssignMod, 4)) // 回合 6 % 4 = 2
	this.Equal(int32(2), runtime.Game.Round.GetValue())
	this.False(writeRound(eng, AssignMod, 0)) // 取餘 0 → no-op
	this.Equal(int32(2), runtime.Game.Round.GetValue())
}

func (this *SuiteAttrWrite) TestAttrWriteRoundLeft() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Round = NewValue(4, 0)
	runtime.Game.RoundMax = NewValue(10, 0)

	this.True(writeRoundLeft(eng, AssignSet, 3)) // 剩餘 = 3 → 回合上限 = 回合 + 3 = 7
	this.Equal(int32(7), runtime.Game.RoundMax.GetValue())
	this.True(writeRoundLeft(eng, AssignAdd, 2)) // 剩餘 += 2 → 回合上限 += 2 = 9(基準 7-4=3、+2=5、4+5=9)
	this.Equal(int32(9), runtime.Game.RoundMax.GetValue())
	this.True(writeRoundLeft(eng, AssignSet, -5)) // 回合上限 = 4 + (-5) = -1 → 夾 >= 回合(4)
	this.Equal(int32(4), runtime.Game.RoundMax.GetValue())
	this.False(writeRoundLeft(eng, AssignDiv, 0)) // 除 0 → no-op
	this.Equal(int32(4), runtime.Game.RoundMax.GetValue())
	this.False(writeRoundLeft(eng, AssignLock, 0)) // @ → no-op(衍生寫屬性無鎖)
	this.Equal(int32(4), runtime.Game.RoundMax.GetValue())
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleNormal() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Morale = NewValue(20, 0)

	this.True(writeMorale(eng, AssignSet, 30)) // 非 -= → 一般寫鎖運算
	this.Equal(int32(30), runtime.Game.Morale.GetValue())
	this.True(writeMorale(eng, AssignAdd, 5))
	this.Equal(int32(35), runtime.Game.Morale.GetValue())
	this.True(writeMorale(eng, AssignLock, 0)) // @ 鎖定
	this.Equal(int32(1), runtime.Game.Morale.GetLock())
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageBlock() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Morale = NewValue(20, 0)
	runtime.Game.MoraleBlock = NewValue(2, 0)

	this.True(writeMorale(eng, AssignSub, 8)) // 格擋 > 0 → 格擋 -1、無視 N、morale 不動
	this.Equal(int32(1), runtime.Game.MoraleBlock.GetValue())
	this.Equal(int32(20), runtime.Game.Morale.GetValue())
	this.Equal(int32(0), runtime.Game.DamageValue)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageBlockLocked() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Morale = NewValue(20, 0)
	runtime.Game.MoraleBlock = NewValue(2, 1)

	this.False(writeMorale(eng, AssignSub, 8)) // 格擋鎖定:仍無視 N、但不減層 → 無變更
	this.Equal(int32(2), runtime.Game.MoraleBlock.GetValue())
	this.Equal(int32(20), runtime.Game.Morale.GetValue())
	this.Equal(int32(0), runtime.Game.DamageValue)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageShield() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Morale = NewValue(20, 0)
	runtime.Game.MoraleShield = NewValue(5, 0)

	this.True(writeMorale(eng, AssignSub, 3)) // 護盾 5 >= N=3 → 護盾 -3=2、N 歸 0、morale 不動
	this.Equal(int32(2), runtime.Game.MoraleShield.GetValue())
	this.Equal(int32(20), runtime.Game.Morale.GetValue())
	this.Equal(int32(0), runtime.Game.DamageValue)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageShieldLocked() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Morale = NewValue(20, 0)
	runtime.Game.MoraleShield = NewValue(5, 1)

	this.True(writeMorale(eng, AssignSub, 3)) // 護盾鎖定:不消耗、殘餘全進 morale → 20 - 3 = 17
	this.Equal(int32(5), runtime.Game.MoraleShield.GetValue())
	this.Equal(int32(17), runtime.Game.Morale.GetValue())
	this.Equal(int32(3), runtime.Game.DamageValue)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageThrough() {
	guest := &Guest{InstanceID: 9}
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime, self: &Self{Guest: guest}}
	runtime.Game.Morale = NewValue(20, 0)
	runtime.Game.MoraleShield = NewValue(2, 0)

	this.True(writeMorale(eng, AssignSub, 6)) // 護盾 2 < N=6 → 護盾歸 0、N=4 → morale -4=16、實扣 4
	this.Equal(int32(0), runtime.Game.MoraleShield.GetValue())
	this.Equal(int32(16), runtime.Game.Morale.GetValue())
	this.Equal(int32(4), runtime.Game.DamageValue)
	this.Require().NotNil(runtime.Game.DamageGuest) // 來源 = self 顧客
	this.Equal(InstanceID(9), runtime.Game.DamageGuest.InstanceID)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageLocked() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Morale = NewValue(20, 1)
	runtime.Game.MoraleShield = NewValue(2, 0)

	this.True(writeMorale(eng, AssignSub, 6)) // morale 鎖定:護盾仍消耗(2→0),morale 不動、無受損
	this.Equal(int32(0), runtime.Game.MoraleShield.GetValue())
	this.Equal(int32(20), runtime.Game.Morale.GetValue())
	this.Equal(int32(0), runtime.Game.DamageValue)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageNoSource() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime, self: &Self{Card: &Card{}}} // self 為卡牌 → 非顧客來源
	runtime.Game.Morale = NewValue(20, 0)

	this.True(writeMorale(eng, AssignSub, 5)) // 無格擋 / 護盾、N=5 → morale 15、實扣 5
	this.Equal(int32(15), runtime.Game.Morale.GetValue())
	this.Equal(int32(5), runtime.Game.DamageValue)
	this.Nil(runtime.Game.DamageGuest) // self 非顧客 → 空物件

	this.False(writeMorale(eng, AssignSub, 0)) // N=0 → 無消耗、無變更
	this.Equal(int32(15), runtime.Game.Morale.GetValue())
}

func (this *SuiteAttrWrite) TestDamageSource() {
	guest := &Guest{InstanceID: 3}
	this.Equal(guest, damageSource(&Self{Guest: guest})) // 顧客 self → 該顧客
	this.Nil(damageSource(&Self{Card: &Card{}}))         // 卡牌 self → 空物件
	this.Nil(damageSource(nil))                          // self 未綁定 → 空物件
}

func (this *SuiteAttrWrite) TestAttrWriteFields() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}

	this.True(writeMoraleMax(eng, AssignSet, 50))
	this.Equal(int32(50), runtime.Game.MoraleMax.GetValue())
	this.True(writeEnergy(eng, AssignSet, 3))
	this.Equal(int32(3), runtime.Game.Energy.GetValue())
	this.True(writeEnergyMax(eng, AssignSet, 6))
	this.Equal(int32(6), runtime.Game.EnergyMax.GetValue())
	this.True(writeHandMax(eng, AssignSet, 10))
	this.Equal(int32(10), runtime.Game.HandMax.GetValue())
	this.True(writeDrawMax(eng, AssignSet, 5))
	this.Equal(int32(5), runtime.Game.DrawMax.GetValue())
}

func (this *SuiteAttrWrite) TestHasAttrWrite() {
	this.True(HasAttrWrite("morale"))
	this.True(HasAttrWrite("roundLeft"))
	this.True(HasAttrWrite("energyKeep"))
	this.False(HasAttrWrite("nextPhase")) // 唯讀(僅 phaseJump 可設)
	this.False(HasAttrWrite("seatSize"))  // 唯讀衍生
	this.False(HasAttrWrite("nope"))
}
