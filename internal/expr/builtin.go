package expr

// builtinFunc 是內建函式的實作簽章:純運算函式,只對已求值引數運算、不觸及 Resolver
// (見【營業規格書 | 二十六、內建函式清單】)。
type builtinFunc func(arg []Value) (value Value, ok bool)

// builtin 是內建函式註冊表,對齊【營業規格書 | 二十六、內建函式清單】;
// 新增內建函式只需在此註冊一筆並補上對應 builtinFunc,nodeCall 與 parser 皆無須改動。
var builtin = map[string]builtinFunc{
	"max": builtinMax,
	"min": builtinMin,
}

// builtinMax 內建函式 max:至少 2 引數、全為數值,回傳最大者,否則失敗。
func builtinMax(arg []Value) (value Value, ok bool) {
	num, valid := numericArg(arg, 2)
	if valid == false {
		return Value{}, false
	} // if

	result := num[0]

	for _, itor := range num[1:] {
		if itor > result {
			result = itor
		} // if
	} // for

	return NewNum(result), true
}

// builtinMin 內建函式 min:至少 2 引數、全為數值,回傳最小者,否則失敗。
func builtinMin(arg []Value) (value Value, ok bool) {
	num, valid := numericArg(arg, 2)
	if valid == false {
		return Value{}, false
	} // if

	result := num[0]

	for _, itor := range num[1:] {
		if itor < result {
			result = itor
		} // if
	} // for

	return NewNum(result), true
}

// numericArg 校驗內建函式引數:至少 least 個且全為數值,回傳數值切片;不足或含非數值即失敗。
func numericArg(arg []Value, least int) (num []float64, ok bool) {
	if len(arg) < least {
		return nil, false
	} // if

	num = make([]float64, 0, len(arg))

	for _, itor := range arg {
		if itor.kind != valueNum {
			return nil, false
		} // if

		num = append(num, itor.num)
	} // for

	return num, true
}
