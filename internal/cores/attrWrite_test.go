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
	game := NewGame(0, nil, nil, nil, nil)

	this.True(writeScore(game, AssignSet, 10)) // = 設值
	this.Equal(int32(10), game.GetScore().GetValue())
	this.True(writeScore(game, AssignAdd, 5)) // += 加
	this.Equal(int32(15), game.GetScore().GetValue())
	this.True(writeScore(game, AssignSub, 3)) // -= 減(score 無特例,走一般)
	this.Equal(int32(12), game.GetScore().GetValue())
	this.True(writeScore(game, AssignMul, 2)) // *= 乘
	this.Equal(int32(24), game.GetScore().GetValue())
	this.True(writeScore(game, AssignDiv, 5)) // /= 除(24/5=4.8 → half-away 5)
	this.Equal(int32(5), game.GetScore().GetValue())
	this.True(writeScore(game, AssignMod, 3)) // %= 取餘(5%3=2)
	this.Equal(int32(2), game.GetScore().GetValue())

	this.False(writeScore(game, AssignDiv, 0)) // 除 0 → no-op
	this.Equal(int32(2), game.GetScore().GetValue())
	this.False(writeScore(game, AssignMod, 0)) // 取餘 0 → no-op
	this.Equal(int32(2), game.GetScore().GetValue())
}

func (this *SuiteAttrWrite) TestAttrWriteLock() {
	game := NewGame(0, nil, nil, nil, nil)

	game.score = NewValue(9, 0) // 設初值 9(鎖定前)

	this.True(writeScore(game, AssignLock, 0)) // @ 計數 +1
	this.True(writeScore(game, AssignLock, 0))
	this.Equal(int32(2), game.GetScore().GetLock())

	this.False(writeScore(game, AssignSet, 100)) // Lock>0 → 帶值賦值 no-op
	this.Equal(int32(9), game.GetScore().GetValue())

	this.True(writeScore(game, AssignUnlock, 0)) // # 計數 -1
	this.True(writeScore(game, AssignUnlock, 0))
	this.Equal(int32(0), game.GetScore().GetLock())

	this.False(writeScore(game, AssignUnlock, 0)) // # 對 0 計數 → no-op(不降負)
	this.Equal(int32(0), game.GetScore().GetLock())

	this.True(writeScore(game, AssignSet, 100)) // 解鎖後帶值賦值恢復
	this.Equal(int32(100), game.GetScore().GetValue())
}

func (this *SuiteAttrWrite) TestAttrWriteClamp() {
	game := NewGame(0, nil, nil, nil, nil)

	game.moraleShield = NewValue(3, 0)
	this.True(writeMoraleShield(game, AssignSub, 10)) // 3-10=-7 → 夾 0
	this.Equal(int32(0), game.GetMoraleShield().GetValue())
	this.True(writeMoraleShield(game, AssignSet, 5)) // 正值不夾
	this.Equal(int32(5), game.GetMoraleShield().GetValue())

	this.True(writeMoraleBlock(game, AssignSet, -5)) // 夾 0
	this.Equal(int32(0), game.GetMoraleBlock().GetValue())
}

func (this *SuiteAttrWrite) TestAttrWriteLockOnly() {
	game := NewGame(0, nil, nil, nil, nil)

	this.True(writeEnergyKeep(game, AssignLock, 0)) // 純鎖屬性 @ 計數 +1
	this.Equal(int32(1), game.GetEnergyKeep().GetLock())
	this.False(writeEnergyKeep(game, AssignSet, 5)) // 帶值賦值對鎖屬性 no-op
	this.Equal(int32(0), game.GetEnergyKeep().GetValue())
	this.True(writeEnergyKeep(game, AssignUnlock, 0))
	this.Equal(int32(0), game.GetEnergyKeep().GetLock())
}

func (this *SuiteAttrWrite) TestAttrWriteRound() {
	game := NewGame(0, nil, nil, nil, nil)
	game.round = NewValue(3, 0)
	game.roundMax = NewValue(10, 0)

	this.True(writeRound(game, AssignAdd, 2)) // 回合 5
	this.Equal(int32(5), game.GetRound().GetValue())
	this.True(writeRoundMax(game, AssignSet, 12)) // 回合上限 12
	this.Equal(int32(12), game.GetRoundMax().GetValue())
	this.False(writeRound(game, AssignLock, 0)) // 寫屬性無鎖 → @ no-op(回合與鎖定計數皆不變)
	this.Equal(int32(5), game.GetRound().GetValue())
	this.Equal(int32(0), game.GetRound().GetLock())
	this.False(writeRound(game, AssignUnlock, 0)) // 寫屬性無鎖 → # no-op
	this.Equal(int32(0), game.GetRound().GetLock())

	this.True(writeRound(game, AssignSub, 1)) // 回合 4
	this.Equal(int32(4), game.GetRound().GetValue())
	this.True(writeRound(game, AssignMul, 3)) // 回合 12
	this.Equal(int32(12), game.GetRound().GetValue())
	this.True(writeRound(game, AssignDiv, 2)) // 回合 6
	this.Equal(int32(6), game.GetRound().GetValue())
	this.True(writeRound(game, AssignMod, 4)) // 回合 6 % 4 = 2
	this.Equal(int32(2), game.GetRound().GetValue())
	this.False(writeRound(game, AssignMod, 0)) // 取餘 0 → no-op
	this.Equal(int32(2), game.GetRound().GetValue())
}

func (this *SuiteAttrWrite) TestAttrWriteRoundLeft() {
	game := NewGame(0, nil, nil, nil, nil)
	game.round = NewValue(4, 0)
	game.roundMax = NewValue(10, 0)

	this.True(writeRoundLeft(game, AssignSet, 3)) // 剩餘 = 3 → 回合上限 = 回合 + 3 = 7
	this.Equal(int32(7), game.GetRoundMax().GetValue())
	this.True(writeRoundLeft(game, AssignAdd, 2)) // 剩餘 += 2 → 回合上限 += 2 = 9(基準 7-4=3、+2=5、4+5=9)
	this.Equal(int32(9), game.GetRoundMax().GetValue())
	this.True(writeRoundLeft(game, AssignSet, -5)) // 回合上限 = 4 + (-5) = -1 → 夾 >= 回合(4)
	this.Equal(int32(4), game.GetRoundMax().GetValue())
	this.False(writeRoundLeft(game, AssignDiv, 0)) // 除 0 → no-op
	this.Equal(int32(4), game.GetRoundMax().GetValue())
	this.False(writeRoundLeft(game, AssignLock, 0)) // @ → no-op(衍生寫屬性無鎖)
	this.Equal(int32(4), game.GetRoundMax().GetValue())
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleNormal() {
	game := NewGame(0, nil, nil, nil, nil)
	game.morale = NewValue(20, 0)

	this.True(writeMorale(game, AssignSet, 30)) // 非 -= → 一般寫鎖運算
	this.Equal(int32(30), game.GetMorale().GetValue())
	this.True(writeMorale(game, AssignAdd, 5))
	this.Equal(int32(35), game.GetMorale().GetValue())
	this.True(writeMorale(game, AssignLock, 0)) // @ 鎖定
	this.Equal(int32(1), game.GetMorale().GetLock())
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageBlock() {
	game := NewGame(0, nil, nil, nil, nil)
	game.morale = NewValue(20, 0)
	game.moraleBlock = NewValue(2, 0)

	this.True(writeMorale(game, AssignSub, 8)) // 格擋 > 0 → 格擋 -1、無視 N、morale 不動
	this.Equal(int32(1), game.GetMoraleBlock().GetValue())
	this.Equal(int32(20), game.GetMorale().GetValue())
	this.Equal(int32(0), game.GetDamageValue())
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageBlockLocked() {
	game := NewGame(0, nil, nil, nil, nil)
	game.morale = NewValue(20, 0)
	game.moraleBlock = NewValue(2, 1)

	this.False(writeMorale(game, AssignSub, 8)) // 格擋鎖定:仍無視 N、但不減層 → 無變更
	this.Equal(int32(2), game.GetMoraleBlock().GetValue())
	this.Equal(int32(20), game.GetMorale().GetValue())
	this.Equal(int32(0), game.GetDamageValue())
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageShield() {
	game := NewGame(0, nil, nil, nil, nil)
	game.morale = NewValue(20, 0)
	game.moraleShield = NewValue(5, 0)

	this.True(writeMorale(game, AssignSub, 3)) // 護盾 5 >= N=3 → 護盾 -3=2、N 歸 0、morale 不動
	this.Equal(int32(2), game.GetMoraleShield().GetValue())
	this.Equal(int32(20), game.GetMorale().GetValue())
	this.Equal(int32(0), game.GetDamageValue())
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageShieldLocked() {
	game := NewGame(0, nil, nil, nil, nil)
	game.morale = NewValue(20, 0)
	game.moraleShield = NewValue(5, 1)

	this.True(writeMorale(game, AssignSub, 3)) // 護盾鎖定:不消耗、殘餘全進 morale → 20 - 3 = 17
	this.Equal(int32(5), game.GetMoraleShield().GetValue())
	this.Equal(int32(17), game.GetMorale().GetValue())
	this.Equal(int32(3), game.GetDamageValue())
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageThrough() {
	guest := &Guest{instanceID: 9}
	game := NewGame(0, nil, nil, nil, nil)
	game.self = &Ref{guest: guest}
	game.morale = NewValue(20, 0)
	game.moraleShield = NewValue(2, 0)

	this.True(writeMorale(game, AssignSub, 6)) // 護盾 2 < N=6 → 護盾歸 0、N=4 → morale -4=16、實扣 4
	this.Equal(int32(0), game.GetMoraleShield().GetValue())
	this.Equal(int32(16), game.GetMorale().GetValue())
	this.Equal(int32(4), game.GetDamageValue())
	this.Require().NotNil(game.GetDamageGuest()) // 來源 = self 顧客
	this.Equal(InstanceID(9), game.GetDamageGuest().GetInstanceID())
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageLocked() {
	game := NewGame(0, nil, nil, nil, nil)
	game.morale = NewValue(20, 1)
	game.moraleShield = NewValue(2, 0)

	this.True(writeMorale(game, AssignSub, 6)) // morale 鎖定:護盾仍消耗(2→0),morale 不動、無受損
	this.Equal(int32(0), game.GetMoraleShield().GetValue())
	this.Equal(int32(20), game.GetMorale().GetValue())
	this.Equal(int32(0), game.GetDamageValue())
}

func (this *SuiteAttrWrite) TestAttrWriteMoraleDamageNoSource() {
	game := NewGame(0, nil, nil, nil, nil)
	game.self = &Ref{card: &Card{}} // self 為卡牌 → 非顧客來源
	game.morale = NewValue(20, 0)

	this.True(writeMorale(game, AssignSub, 5)) // 無格擋 / 護盾、N=5 → morale 15、實扣 5
	this.Equal(int32(15), game.GetMorale().GetValue())
	this.Equal(int32(5), game.GetDamageValue())
	this.Nil(game.GetDamageGuest()) // self 非顧客 → 空物件

	this.False(writeMorale(game, AssignSub, 0)) // N=0 → 無消耗、無變更
	this.Equal(int32(15), game.GetMorale().GetValue())
}

func (this *SuiteAttrWrite) TestDamageSource() {
	guest := &Guest{instanceID: 3}
	this.Equal(guest, damageSource(&Ref{guest: guest})) // 顧客 self → 該顧客
	this.Nil(damageSource(&Ref{card: &Card{}}))         // 卡牌 self → 空物件
	this.Nil(damageSource(nil))                         // self 未綁定 → 空物件
}

func (this *SuiteAttrWrite) TestAttrWriteFields() {
	game := NewGame(0, nil, nil, nil, nil)

	this.True(writeMoraleMax(game, AssignSet, 50))
	this.Equal(int32(50), game.GetMoraleMax().GetValue())
	this.True(writeEnergy(game, AssignSet, 3))
	this.Equal(int32(3), game.GetEnergy().GetValue())
	this.True(writeEnergyMax(game, AssignSet, 6))
	this.Equal(int32(6), game.GetEnergyMax().GetValue())
	this.True(writeHandMax(game, AssignSet, 10))
	this.Equal(int32(10), game.GetHandMax().GetValue())
	this.True(writeDrawMax(game, AssignSet, 5))
	this.Equal(int32(5), game.GetDrawMax().GetValue())
}

func (this *SuiteAttrWrite) TestHasAttrWrite() {
	this.True(HasAttrWrite("morale"))
	this.True(HasAttrWrite("roundLeft"))
	this.True(HasAttrWrite("energyKeep"))
	this.False(HasAttrWrite("nextPhase")) // 唯讀(僅 phaseJump 可設)
	this.False(HasAttrWrite("seatSize"))  // 唯讀衍生
	this.False(HasAttrWrite("nope"))
}
