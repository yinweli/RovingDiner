package exprs

// Resolver 是 exprs 與 games 的唯一接縫:把運算式中的條件對象解析成值,對齊
// 【營業規格書 | 二十三、屬性清單】表結構——主表(全域屬性 / 查詢函式 / 物件引用)走 Attr,
// 子表(卡牌 / 顧客引用屬性、引用查詢函式)走 AttrRef。由 games 實作(M6)。
// 任一解析不成立(屬性不存在 / 型別不符 / self 未綁定 / 引用為空物件…)回傳 ok == false,
// 評估即失敗(對齊【營業規格書 | 二十七、運算式 | 7】)。
type Resolver interface {
	// Attr 解析主表條目:全域屬性(arg 為空)、查詢函式(arg 帶參數)、物件引用(回傳引用值)。
	Attr(name string, arg []Value) (result Value, ok bool)

	// AttrRef 以引用 ref 為主體解析子表條目:引用屬性(arg 為空)或引用查詢函式(arg 帶參數)。
	AttrRef(ref Ref, name string, arg []Value) (result Value, ok bool)
}

// Ref 是物件引用(卡牌 / 顧客實例)在運算式中的不透明代表;exprs 不解讀其內容,
// 僅在 == / != 比較時以 Same 判定是否同一實例(對齊【營業規格書 | 二十七、運算式 | 2】比實例編號)。
// 由 games 實作。
type Ref interface {
	// Same 回傳是否與另一引用指向同一實例。
	Same(other Ref) bool
}

// Builtin 是內建純運算函式(min / max…),只對輸入參數運算、不讀系統狀態
// (對齊【營業規格書 | 二十六、內建函式清單】);由 games 注入,exprs 自身不內建任何函式。
// 任一參數非數值 / 數量不符 / 語意失敗回傳 ok == false。
type Builtin func(arg []Value) (result Value, ok bool)

// Env 是求值期環境:條件對象解析器 Resolver 與內建函式註冊表 Builtin,於 Eval 時由 games 帶入,
// 不持任何全域可變狀態。零值(Resolver 為 nil、Builtin 為 nil)代表純語言求值——
// 遇條件對象 / 函式即評估失敗。
type Env struct {
	Resolver Resolver
	Builtin  map[string]Builtin
}
