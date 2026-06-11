package exprs

import (
	"math"
)

// Value 是運算式的求值結果(數值 / 布林 / 字串 / 空物件)。
// 中間值一律以 float64 表達、過程不四捨五入; 捨入由呼叫方以 Round 處理
// (對齊【營業規格書 | 二十七、運算式】總則)。
type Value struct {
	kind valueKind
	num  float64 // valueNum 的數值
	flag bool    // valueBool 的值
	text string  // valueText 的內容
	ref  Ref     // valueRef 的物件引用
}

// NewNum 建立數值。
func NewNum(num float64) Value {
	return Value{kind: valueNum, num: num}
}

// NewBool 建立布林值。
func NewBool(flag bool) Value {
	return Value{kind: valueBool, flag: flag}
}

// NewText 建立字串值。
func NewText(text string) Value {
	return Value{kind: valueText, text: text}
}

// NewNone 建立空物件值(none); 對齊【營業規格書 | 二十七、運算式 | 4】空物件字面值。
func NewNone() Value {
	return Value{kind: valueNone}
}

// NewRef 建立物件引用值; ref 由 Resolver 提供, exprs 不解讀其內容, 僅在比較時用 Ref.IsSame。
func NewRef(ref Ref) Value {
	return Value{kind: valueRef, ref: ref}
}

// IsNum 回傳是否為數值。
func (this Value) IsNum() bool {
	return this.kind == valueNum
}

// Num 取數值(僅 IsNum 為真時有意義)。
func (this Value) Num() float64 {
	return this.num
}

// IsBool 回傳是否為布林值。
func (this Value) IsBool() bool {
	return this.kind == valueBool
}

// Bool 取布林值(僅 IsBool 為真時有意義)。
func (this Value) Bool() bool {
	return this.flag
}

// IsText 回傳是否為字串值。
func (this Value) IsText() bool {
	return this.kind == valueText
}

// Text 取字串值(僅 IsText 為真時有意義)。
func (this Value) Text() string {
	return this.text
}

// IsNone 回傳是否為空物件(none)。
func (this Value) IsNone() bool {
	return this.kind == valueNone
}

// IsRef 回傳是否為物件引用。
func (this Value) IsRef() bool {
	return this.kind == valueRef
}

// Ref 取物件引用(僅 IsRef 為真時有意義)。
func (this Value) Ref() Ref {
	return this.ref
}

// isObject 回傳是否為物件值(空物件 none 或物件引用); 供相等比較歸類
// (對齊【營業規格書 | 二十七、運算式 | 2】物件引用 / 空物件比較)。
func (this Value) isObject() bool {
	return this.kind == valueNone || this.kind == valueRef
}

// Truthy 依【營業規格書 | 二十七、運算式 | 6】把值轉成布林判定:
// 布林直接取用、非 0 數值為真、0 為假; 其餘型別(字串 / 空物件)評估失敗(ok == false)。
// 邏輯運算元、三元條件與「省略比較符」的真假判定皆以此為準。
func (this Value) Truthy() (result, ok bool) {
	switch this.kind {
	case valueBool:
		return this.flag, true

	case valueNum:
		return this.num != 0, true

	default:
		return false, false
	} // switch
}

// Round 把運算式中間值四捨五入為整數, 採 half-away-from-zero
// (2.4→2、2.5→3、-2.4→-2、-2.5→-3), 對齊【營業規格書 | 二十七、運算式】總則;
// 由呼叫方把求值結果寫回屬性時使用(exprs 過程本身不四捨五入)。
func Round(num float64) int32 {
	return int32(math.Round(num))
}

// valueKind 區分運算式求值結果的型別; 對齊【營業規格書 | 二十七、運算式 | 4】字面值
// 與【營業規格書 | 二十七、運算式 | 2】物件引用。M4 起新增物件引用值。
type valueKind int8

const (
	valueNum  valueKind = iota // 數值(float64; 中間值可為小數)
	valueBool                  // 布林
	valueText                  // 字串
	valueNone                  // 空物件(none; 只與空物件相等)
	valueRef                   // 物件引用(卡牌 / 顧客實例; 比實例編號)
)
