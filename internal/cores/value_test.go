package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteValue(t *testing.T) {
	suite.Run(t, new(SuiteValue))
}

// SuiteValue 驗證 value.go 的數值容器讀取、變更原語、賦值符派發與夾值。
type SuiteValue struct {
	suite.Suite
}

// TestNewValue 驗證 NewValue 以數值與鎖定計數建構。
func (this *SuiteValue) TestNewValue() {
	v := NewValue(5, 2)
	this.Equal(int32(5), v.GetValue())
	this.Equal(int32(2), v.GetLock())
}

// TestNewValueLock 驗證 NewValueLock 以 bool 旗標建構鎖屬性初值(true → 計數 1、false → 0, 數值固定 0)。
func (this *SuiteValue) TestNewValueLock() {
	on := NewValueLock(true)
	this.Equal(int32(0), on.GetValue())
	this.Equal(int32(1), on.GetLock())

	off := NewValueLock(false)
	this.Equal(int32(0), off.GetValue())
	this.Equal(int32(0), off.GetLock())
}

// TestValueGetValue 驗證 GetValue 讀回當前數值。
func (this *SuiteValue) TestValueGetValue() {
	v := NewValue(7, 0)
	this.Equal(int32(7), v.GetValue())
}

// TestValueGetLock 驗證 GetLock 讀回鎖定計數。
func (this *SuiteValue) TestValueGetLock() {
	v := NewValue(0, 3)
	this.Equal(int32(3), v.GetLock())
}

// TestValueIsLock 驗證 IsLock 僅在鎖定計數 > 0 時回報鎖定。
func (this *SuiteValue) TestValueIsLock() {
	unlock := NewValue(5, 0)
	this.False(unlock.IsLock()) // 未鎖定

	one := NewValue(0, 1)
	this.True(one.IsLock()) // 鎖定一層

	many := NewValue(9, 3)
	this.True(many.IsLock()) // 鎖定多層, 數值不影響鎖定判定

	neg := NewValue(0, -1)
	this.False(neg.IsLock()) // 計數非正不算鎖定
}

// TestValueSet 驗證 Set 帶值賦值 =、捨入 half-away-from-zero、鎖定 no-op。
func (this *SuiteValue) TestValueSet() {
	v := NewValue(1, 0)
	this.True(v.Set(2.5)) // 2.5 → 3
	this.Equal(int32(3), v.GetValue())

	this.True(v.Set(-2.5)) // -2.5 → -3
	this.Equal(int32(-3), v.GetValue())

	lock := NewValue(1, 1)
	this.False(lock.Set(9)) // 鎖定 → no-op
	this.Equal(int32(1), lock.GetValue())
}

// TestValueAdd 驗證 Add 帶值賦值 +=、捨入、鎖定 no-op。
func (this *SuiteValue) TestValueAdd() {
	v := NewValue(5, 0)
	this.True(v.Add(0.5)) // 5 + 0.5 = 5.5 → 6
	this.Equal(int32(6), v.GetValue())

	lock := NewValue(5, 1)
	this.False(lock.Add(3)) // 鎖定 → no-op
	this.Equal(int32(5), lock.GetValue())
}

// TestValueSub 驗證 Sub 帶值賦值 -=、鎖定 no-op。
func (this *SuiteValue) TestValueSub() {
	v := NewValue(5, 0)
	this.True(v.Sub(2))
	this.Equal(int32(3), v.GetValue())

	lock := NewValue(5, 1)
	this.False(lock.Sub(2)) // 鎖定 → no-op
	this.Equal(int32(5), lock.GetValue())
}

// TestValueMul 驗證 Mul 帶值賦值 *=、捨入。
func (this *SuiteValue) TestValueMul() {
	v := NewValue(3, 0)
	this.True(v.Mul(1.5)) // 3 * 1.5 = 4.5 → 5
	this.Equal(int32(5), v.GetValue())
}

// TestValueDiv 驗證 Div 帶值賦值 /=、除零 no-op、捨入。
func (this *SuiteValue) TestValueDiv() {
	v := NewValue(5, 0)
	this.True(v.Div(2)) // 5 / 2 = 2.5 → 3
	this.Equal(int32(3), v.GetValue())

	this.False(v.Div(0)) // 除零 → no-op
	this.Equal(int32(3), v.GetValue())
}

// TestValueMod 驗證 Mod 帶值賦值 %=、取餘零 no-op。
func (this *SuiteValue) TestValueMod() {
	v := NewValue(5, 0)
	this.True(v.Mod(3)) // 5 % 3 = 2
	this.Equal(int32(2), v.GetValue())

	this.False(v.Mod(0)) // 取餘零 → no-op
	this.Equal(int32(2), v.GetValue())
}

// TestValueLock 驗證 Lock 鎖定計數 + 1。
func (this *SuiteValue) TestValueLock() {
	v := NewValue(0, 0)
	v.Lock()
	this.Equal(int32(1), v.GetLock())

	v.Lock()
	this.Equal(int32(2), v.GetLock())
}

// TestValueUnlock 驗證 Unlock 鎖定計數 > 0 時 - 1、對 0 計數 no-op。
func (this *SuiteValue) TestValueUnlock() {
	v := NewValue(0, 2)
	this.True(v.Unlock())
	this.Equal(int32(1), v.GetLock())

	this.True(v.Unlock())
	this.Equal(int32(0), v.GetLock())

	this.False(v.Unlock()) // 對 0 計數 → no-op、不降為負
	this.Equal(int32(0), v.GetLock())
}

// TestValueApply 驗證 Apply 依賦值符派發至對應原語(八路 + 未知賦值符)。
func (this *SuiteValue) TestValueApply() {
	v := NewValue(4, 0)
	this.True(v.Apply(AssignSet, 10)) // = 10
	this.Equal(int32(10), v.GetValue())

	this.True(v.Apply(AssignAdd, 2)) // += → 12
	this.True(v.Apply(AssignSub, 3)) // -= → 9
	this.True(v.Apply(AssignMul, 2)) // *= → 18
	this.True(v.Apply(AssignDiv, 3)) // /= → 6
	this.True(v.Apply(AssignMod, 4)) // %= → 2
	this.Equal(int32(2), v.GetValue())

	this.False(v.Apply(AssignDiv, 0)) // 除零 → no-op

	this.True(v.Apply(AssignLock, 0)) // @ → lock++
	this.Equal(int32(1), v.GetLock())

	this.False(v.Apply(AssignAdd, 1)) // 鎖定 → 帶值賦值 no-op
	this.Equal(int32(2), v.GetValue())

	this.True(v.Apply(AssignUnlock, 0)) // # → lock--
	this.Equal(int32(0), v.GetLock())

	this.False(v.Apply(AssignKind(99), 0)) // 未知賦值符 → default no-op
}

// TestValueApplyLockOnly 驗證 ApplyLockOnly 純鎖屬性派發: 僅 @ # 生效、帶值賦值 no-op。
func (this *SuiteValue) TestValueApplyLockOnly() {
	v := NewValue(0, 0)

	this.True(v.ApplyLockOnly(AssignLock)) // @ → lock++
	this.Equal(int32(1), v.GetLock())

	this.True(v.ApplyLockOnly(AssignUnlock)) // # → lock--
	this.Equal(int32(0), v.GetLock())

	this.False(v.ApplyLockOnly(AssignUnlock)) // 對 0 計數 → no-op

	this.False(v.ApplyLockOnly(AssignSet)) // 帶值賦值不適用 → no-op
	this.Equal(int32(0), v.GetValue())
}

// TestValueApplyValueOnly 驗證 ApplyValueOnly 寫屬性派發: 僅帶值賦值生效、@ # no-op。
func (this *SuiteValue) TestValueApplyValueOnly() {
	v := NewValue(4, 0)

	this.True(v.ApplyValueOnly(AssignAdd, 2)) // += → 6
	this.Equal(int32(6), v.GetValue())

	this.False(v.ApplyValueOnly(AssignLock, 0)) // @ 不適用 → no-op
	this.False(v.ApplyValueOnly(AssignUnlock, 0))
	this.Equal(int32(0), v.GetLock())

	this.False(v.ApplyValueOnly(AssignDiv, 0)) // 除零 → no-op
	this.Equal(int32(6), v.GetValue())
}

// TestValueClamp 驗證 Clamp 夾下限。
func (this *SuiteValue) TestValueClamp() {
	low := NewValue(-2, 0)
	low.Clamp(0)
	this.Equal(int32(0), low.GetValue()) // 低於下限 → 夾

	high := NewValue(5, 0)
	high.Clamp(0)
	this.Equal(int32(5), high.GetValue()) // 不低於下限 → 不動
}
