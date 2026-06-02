package expr

// Resolver 是 expr 與 game 的唯一接縫:expr 解析「條件對象」時呼叫它取值,
// 自身不認識任何遊戲屬性。介面對齊【營業規格書 | 二十三、屬性清單】的表結構——
// 主表(全域屬性 / 查詢函式 / 物件引用)走 Property,子表(卡牌 / 顧客引用屬性)走 Member。
//
// 兩個方法的第二回傳值 ok=false 代表條件對象解析失敗(引用不存在 / 型別不符 / self 未固定…),
// 會被 evaluator 視為【營業規格書 | 二十七、運算式 | 8】評估失敗並沿樹上傳。
type Resolver interface {
	// Property 解析主表條目:arg 為 nil 代表裸屬性 / 物件引用(如 morale、round、self、drawLast);
	// arg 非 nil 代表查詢函式(如 tableCount('>=', 2)、handSize(2))。
	Property(name string, arg []Value) (value Value, ok bool)

	// Member 解析子表引用屬性:以 target 物件引用為主體存取 name 屬性。
	// arg 為 nil 代表 target.name(如 self.calm、drawLast.cardID);
	// arg 非 nil 代表 target.name(arg)查詢函式(如 self.effectStack(101))。
	Member(target Value, name string, arg []Value) (value Value, ok bool)
}
