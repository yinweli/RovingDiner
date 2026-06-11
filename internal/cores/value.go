package cores

import (
	"math"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

// Value 數值實例; 執行期可變屬性的統一容器。
// 對應【營業規格書 | 五、實例結構 | 數值（Value）實例】。
//
// 一般運算(= += -= *= /= %=)改數值、鎖定 / 解鎖(@ #)改鎖定計數;
// 鎖定計數 > 0 時一般運算對該屬性為 no-op。純計數型屬性數值固定為 0、狀態全表達於鎖定計數。
//
// 欄位私有: 讀取走 GetValue / GetLock / IsLock, 變更走方法(Set / Add … / Lock / Unlock), 無繞鎖通道。
type Value struct {
	value int32 // 屬性當前數值(整數)
	lock  int32 // 屬性鎖定計數; 由 @ / # 控制;> 0 時一般運算為 no-op
}

// NewValue 以數值與鎖定計數建構 Value。
func NewValue(value, lock int32) Value {
	return Value{value: value, lock: lock}
}

// NewValueLock 以靜態 bool 旗標建構鎖屬性初值: true → 鎖定計數 1、false → 0(數值固定 0)。供新實例化卡牌 / 顧客載入鎖型欄位。
func NewValueLock(lock bool) Value {
	if lock {
		return NewValue(0, 1)
	} // if

	return NewValue(0, 0)
}

// GetValue 讀屬性當前數值。
func (this *Value) GetValue() int32 {
	return this.value
}

// GetLock 讀屬性鎖定計數。
func (this *Value) GetLock() int32 {
	return this.lock
}

// IsLock 回傳屬性是否處於鎖定狀態(鎖定計數 > 0)。
func (this *Value) IsLock() bool {
	return this.lock > 0
}

// Set 帶值賦值 =; 鎖定時 no-op。
func (this *Value) Set(n float64) bool {
	return this.write(n)
}

// Add 帶值賦值 +=; 鎖定時 no-op。
func (this *Value) Add(n float64) bool {
	return this.write(float64(this.value) + n)
}

// Sub 帶值賦值 -=; 鎖定時 no-op。
func (this *Value) Sub(n float64) bool {
	return this.write(float64(this.value) - n)
}

// Mul 帶值賦值 *=; 鎖定時 no-op。
func (this *Value) Mul(n float64) bool {
	return this.write(float64(this.value) * n)
}

// Div 帶值賦值 /=; 除零 no-op、鎖定時 no-op。
func (this *Value) Div(n float64) bool {
	if n == 0 {
		return false
	} // if

	return this.write(float64(this.value) / n)
}

// Mod 帶值賦值 %=; 取餘零 no-op、鎖定時 no-op。
func (this *Value) Mod(n float64) bool {
	if n == 0 {
		return false
	} // if

	return this.write(math.Mod(float64(this.value), n))
}

// Lock 鎖定運算符 @; 鎖定計數 + 1(一律生效)。
func (this *Value) Lock() {
	this.lock++
}

// Unlock 解鎖運算符 #; 鎖定計數 > 0 時 - 1, 否則 no-op。
func (this *Value) Unlock() bool {
	if this.lock > 0 {
		this.lock--
		return true
	} // if

	return false
}

// Apply 依賦值符派發到對應原語; 命令路徑單一入口(流程內部碼直接用具名方法)。
func (this *Value) Apply(op AssignKind, n float64) bool {
	switch op {
	case AssignSet:
		return this.Set(n)

	case AssignAdd:
		return this.Add(n)

	case AssignSub:
		return this.Sub(n)

	case AssignMul:
		return this.Mul(n)

	case AssignDiv:
		return this.Div(n)

	case AssignMod:
		return this.Mod(n)

	case AssignLock:
		this.Lock()
		return true

	case AssignUnlock:
		return this.Unlock()
	} // switch

	return false
}

// ApplyLockOnly 鎖屬性(純計數封印, 數值固定 0)的賦值派發: 僅鎖定 / 解鎖; 帶值賦值不適用故 no-op(可寫性由 Validate 先擋)。
func (this *Value) ApplyLockOnly(op AssignKind) bool {
	switch op {
	case AssignLock:
		this.Lock()
		return true

	case AssignUnlock:
		return this.Unlock()

	default:
		return false // 帶值賦值不適用純鎖屬性 → no-op
	} // switch
}

// ApplyValueOnly 寫屬性(無鎖定語意, 如 round / roundMax, 鎖定計數恆 0)的賦值派發: 僅帶值賦值; @ # 不適用故 no-op。
func (this *Value) ApplyValueOnly(op AssignKind, n float64) bool {
	if op == AssignLock || op == AssignUnlock {
		return false // 鎖定 / 解鎖不適用寫屬性 → no-op
	} // if

	return this.Apply(op, n)
}

// Clamp 夾下限至 low(運算後低於 low 時夾為 low); 夾值政策在詞條、僅護盾 / 格擋 chain。
func (this *Value) Clamp(low int32) {
	if this.value < low {
		this.value = low
	} // if
}

// write 帶值賦值的共通寫入: 鎖定 → no-op; 否則 half-away-from-zero 捨入後寫回。
func (this *Value) write(n float64) bool {
	if this.IsLock() {
		return false
	} // if

	this.value = exprs.Round(n)
	return true
}
