package expr

import "math"

// kind 標示 Value 當前承載的型別。
type kind int8

const (
	kindNumber kind = iota // 數值(整數 / 小數,內部一律 float64)
	kindString             // 字串
	kindBool               // 布林
	kindRef                // 物件引用(卡牌 / 顧客 / 空物件)
)

// ref 是 expr 對物件引用的極簡抽象;expr 只比較引用相等性,不認識遊戲實例。
// game 端實作 Resolver 時把 defines.InstanceID 包成此型別(經 NewRef / NewNone)。
type ref struct {
	id   int64 // 實例編號(none 為 true 時無意義)
	none bool  // 是否為空物件
}

// Value 是運算式的值;以標籤聯合承載數值 / 字串 / 布林 / 物件引用。
// C# 移植時以 struct + enum tag 對應(不採 Go interface 慣用法)。
type Value struct {
	kind    kind
	number  float64
	str     string
	boolean bool
	ref     ref
}

// NewNumber 建立數值 Value。
func NewNumber(number float64) Value {
	return Value{kind: kindNumber, number: number}
}

// NewString 建立字串 Value。
func NewString(text string) Value {
	return Value{kind: kindString, str: text}
}

// NewBool 建立布林 Value。
func NewBool(boolean bool) Value {
	return Value{kind: kindBool, boolean: boolean}
}

// NewRef 建立非空物件引用 Value(綁定實例編號)。
func NewRef(id int64) Value {
	return Value{kind: kindRef, ref: ref{id: id}}
}

// NewNone 建立空物件引用 Value;空物件只與空物件相等(見【二十七〇2】)。
func NewNone() Value {
	return Value{kind: kindRef, ref: ref{none: true}}
}

// IsNumber 回傳是否為數值。
func (this Value) IsNumber() bool {
	return this.kind == kindNumber
}

// IsString 回傳是否為字串。
func (this Value) IsString() bool {
	return this.kind == kindString
}

// IsBool 回傳是否為布林。
func (this Value) IsBool() bool {
	return this.kind == kindBool
}

// IsRef 回傳是否為物件引用(含空物件)。
func (this Value) IsRef() bool {
	return this.kind == kindRef
}

// IsNone 回傳是否為空物件引用。
func (this Value) IsNone() bool {
	return this.kind == kindRef && this.ref.none == true
}

// Number 取數值;非數值型別回傳 0。
func (this Value) Number() float64 {
	return this.number
}

// Str 取字串;非字串型別回傳空字串。
func (this Value) Str() string {
	return this.str
}

// Bool 取布林;非布林型別回傳 false。
func (this Value) Bool() bool {
	return this.boolean
}

// RefID 取物件引用的實例編號;非引用 / 空物件回傳 0。
func (this Value) RefID() int64 {
	if this.kind == kindRef && this.ref.none == false {
		return this.ref.id
	} // if

	return 0
}

// AsBool 依【二十七〇6】省略比較符的布林判定:布林取自身、數值非 0 為真 / 0 為假、
// 其餘型別(字串 / 物件引用)評估失敗。供 NOT / AND / OR / 三元條件與呼叫方協調最終值使用。
func AsBool(value Value) (result, ok bool) {
	switch value.kind {
	case kindBool:
		return value.boolean, true
	case kindNumber:
		return value.number != 0, true
	case kindString, kindRef:
		return false, false
	} // switch

	return false, false
}

// Round 將小數四捨五入為整數,採 half-away-from-zero(2.5→3、-2.5→-3、2.4→2)。
// expr 內部不四捨五入;此工具供呼叫方(屬性修改命令)寫回屬性時使用,見【十七〇1】運算順序。
func Round(x float64) int32 {
	return int32(math.Round(x))
}
