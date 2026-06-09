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
	this.Equal(int32(10), runtime.Game.Score.Value)
	this.True(writeScore(eng, AssignAdd, 5)) // += 加
	this.Equal(int32(15), runtime.Game.Score.Value)
	this.True(writeScore(eng, AssignSub, 3)) // -= 減(score 無特例,走一般)
	this.Equal(int32(12), runtime.Game.Score.Value)
	this.True(writeScore(eng, AssignMul, 2)) // *= 乘
	this.Equal(int32(24), runtime.Game.Score.Value)
	this.True(writeScore(eng, AssignDiv, 5)) // /= 除(24/5=4.8 → half-away 5)
	this.Equal(int32(5), runtime.Game.Score.Value)
	this.True(writeScore(eng, AssignMod, 3)) // %= 取餘(5%3=2)
	this.Equal(int32(2), runtime.Game.Score.Value)

	this.False(writeScore(eng, AssignDiv, 0)) // 除 0 → no-op
	this.Equal(int32(2), runtime.Game.Score.Value)
	this.False(writeScore(eng, AssignMod, 0)) // 取餘 0 → no-op
	this.Equal(int32(2), runtime.Game.Score.Value)
}

func (this *SuiteAttrWrite) TestAttrWriteLock() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}

	this.True(writeScore(eng, AssignLock, 0)) // @ 計數 +1
	this.True(writeScore(eng, AssignLock, 0))
	this.Equal(int32(2), runtime.Game.Score.Lock)

	runtime.Game.Score.Value = 9
	this.False(writeScore(eng, AssignSet, 100)) // Lock>0 → 帶值賦值 no-op
	this.Equal(int32(9), runtime.Game.Score.Value)

	this.True(writeScore(eng, AssignUnlock, 0)) // # 計數 -1
	this.True(writeScore(eng, AssignUnlock, 0))
	this.Equal(int32(0), runtime.Game.Score.Lock)

	this.False(writeScore(eng, AssignUnlock, 0)) // # 對 0 計數 → no-op(不降負)
	this.Equal(int32(0), runtime.Game.Score.Lock)

	this.True(writeScore(eng, AssignSet, 100)) // 解鎖後帶值賦值恢復
	this.Equal(int32(100), runtime.Game.Score.Value)
}

func (this *SuiteAttrWrite) TestAttrWriteClamp() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}

	runtime.Game.MoraleShield.Value = 3
	this.True(writeMoraleShield(eng, AssignSub, 10)) // 3-10=-7 → 夾 0
	this.Equal(int32(0), runtime.Game.MoraleShield.Value)
	this.True(writeMoraleShield(eng, AssignSet, 5)) // 正值不夾
	this.Equal(int32(5), runtime.Game.MoraleShield.Value)

	this.True(writeMoraleBlock(eng, AssignSet, -5)) // 夾 0
	this.Equal(int32(0), runtime.Game.MoraleBlock.Value)
}

func (this *SuiteAttrWrite) TestAttrWriteLockOnly() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}

	this.True(writeEnergyKeep(eng, AssignLock, 0)) // 純鎖屬性 @ 計數 +1
	this.Equal(int32(1), runtime.Game.EnergyKeep.Lock)
	this.False(writeEnergyKeep(eng, AssignSet, 5)) // 帶值賦值對鎖屬性 no-op
	this.Equal(int32(0), runtime.Game.EnergyKeep.Value)
	this.True(writeEnergyKeep(eng, AssignUnlock, 0))
	this.Equal(int32(0), runtime.Game.EnergyKeep.Lock)
}

func (this *SuiteAttrWrite) TestAttrWriteRound() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Round = 3
	runtime.Game.RoundMax = 10

	this.True(writeRound(eng, AssignAdd, 2)) // 回合 5
	this.Equal(int32(5), runtime.Game.Round)
	this.True(writeRoundMax(eng, AssignSet, 12)) // 回合上限 12
	this.Equal(int32(12), runtime.Game.RoundMax)
	this.False(writeRound(eng, AssignLock, 0)) // 寫屬性無鎖 → @ no-op(回合不變)
	this.Equal(int32(5), runtime.Game.Round)
}

func (this *SuiteAttrWrite) TestAttrWriteRoundLeft() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Round = 4
	runtime.Game.RoundMax = 10

	this.True(writeRoundLeft(eng, AssignSet, 3)) // 剩餘 = 3 → 回合上限 = 回合 + 3 = 7
	this.Equal(int32(7), runtime.Game.RoundMax)
	this.True(writeRoundLeft(eng, AssignAdd, 2)) // 剩餘 += 2 → 回合上限 += 2 = 9(基準 7-4=3、+2=5、4+5=9)
	this.Equal(int32(9), runtime.Game.RoundMax)
	this.True(writeRoundLeft(eng, AssignSet, -5)) // 回合上限 = 4 + (-5) = -1 → 夾 >= 回合(4)
	this.Equal(int32(4), runtime.Game.RoundMax)
	this.False(writeRoundLeft(eng, AssignDiv, 0)) // 除 0 → no-op
	this.Equal(int32(4), runtime.Game.RoundMax)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleNormal() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Morale.Value = 20

	this.True(writeMorale(eng, AssignSet, 30)) // 非 -= → 一般寫鎖運算
	this.Equal(int32(30), runtime.Game.Morale.Value)
	this.True(writeMorale(eng, AssignAdd, 5))
	this.Equal(int32(35), runtime.Game.Morale.Value)
	this.True(writeMorale(eng, AssignLock, 0)) // @ 鎖定
	this.Equal(int32(1), runtime.Game.Morale.Lock)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageBlock() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Morale.Value = 20
	runtime.Game.MoraleBlock.Value = 2

	this.True(writeMorale(eng, AssignSub, 8)) // 格擋 > 0 → 格擋 -1、無視 N、morale 不動
	this.Equal(int32(1), runtime.Game.MoraleBlock.Value)
	this.Equal(int32(20), runtime.Game.Morale.Value)
	this.Equal(int32(0), runtime.Game.DamageValue)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageShield() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Morale.Value = 20
	runtime.Game.MoraleShield.Value = 5

	this.True(writeMorale(eng, AssignSub, 3)) // 護盾 5 >= N=3 → 護盾 -3=2、N 歸 0、morale 不動
	this.Equal(int32(2), runtime.Game.MoraleShield.Value)
	this.Equal(int32(20), runtime.Game.Morale.Value)
	this.Equal(int32(0), runtime.Game.DamageValue)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageThrough() {
	guest := &Guest{InstanceID: 9}
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime, self: &Self{Guest: guest}}
	runtime.Game.Morale.Value = 20
	runtime.Game.MoraleShield.Value = 2

	this.True(writeMorale(eng, AssignSub, 6)) // 護盾 2 < N=6 → 護盾歸 0、N=4 → morale -4=16、實扣 4
	this.Equal(int32(0), runtime.Game.MoraleShield.Value)
	this.Equal(int32(16), runtime.Game.Morale.Value)
	this.Equal(int32(4), runtime.Game.DamageValue)
	this.Require().NotNil(runtime.Game.DamageGuest) // 來源 = self 顧客
	this.Equal(InstanceID(9), runtime.Game.DamageGuest.InstanceID)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageLocked() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime}
	runtime.Game.Morale = Value{Value: 20, Lock: 1}
	runtime.Game.MoraleShield.Value = 2

	this.True(writeMorale(eng, AssignSub, 6)) // morale 鎖定:護盾仍消耗(2→0),morale 不動、無受損
	this.Equal(int32(0), runtime.Game.MoraleShield.Value)
	this.Equal(int32(20), runtime.Game.Morale.Value)
	this.Equal(int32(0), runtime.Game.DamageValue)
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageNoSource() {
	runtime := NewRuntime(0)
	eng := &Engine{runtime: runtime, self: &Self{Card: &Card{}}} // self 為卡牌 → 非顧客來源
	runtime.Game.Morale.Value = 20

	this.True(writeMorale(eng, AssignSub, 5)) // 無格擋 / 護盾、N=5 → morale 15、實扣 5
	this.Equal(int32(15), runtime.Game.Morale.Value)
	this.Equal(int32(5), runtime.Game.DamageValue)
	this.Nil(runtime.Game.DamageGuest) // self 非顧客 → 空物件

	this.False(writeMorale(eng, AssignSub, 0)) // N=0 → 無消耗、無變更
	this.Equal(int32(15), runtime.Game.Morale.Value)
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
	this.Equal(int32(50), runtime.Game.MoraleMax.Value)
	this.True(writeEnergy(eng, AssignSet, 3))
	this.Equal(int32(3), runtime.Game.Energy.Value)
	this.True(writeEnergyMax(eng, AssignSet, 6))
	this.Equal(int32(6), runtime.Game.EnergyMax.Value)
	this.True(writeHandMax(eng, AssignSet, 10))
	this.Equal(int32(10), runtime.Game.HandMax.Value)
	this.True(writeDrawMax(eng, AssignSet, 5))
	this.Equal(int32(5), runtime.Game.DrawMax.Value)
}

func (this *SuiteAttrWrite) TestHasAttrWrite() {
	this.True(HasAttrWrite("morale"))
	this.True(HasAttrWrite("roundLeft"))
	this.True(HasAttrWrite("energyKeep"))
	this.False(HasAttrWrite("nextPhase")) // 唯讀(僅 phaseJump 可設)
	this.False(HasAttrWrite("seatSize"))  // 唯讀衍生
	this.False(HasAttrWrite("nope"))
}
