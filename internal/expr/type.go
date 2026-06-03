package expr

// Resolver 是 expr 與 game 的唯一接縫:expr 解析「條件對象」時呼叫它取值,
// 自身不認識任何遊戲屬性。介面對齊【營業規格書 | 二十三、屬性清單】的表結構——
// 主表(全域屬性 / 查詢函式 / 物件引用)走 Attr,子表(卡牌 / 顧客引用屬性)走 AttrRef。
//
// 兩個方法的第二回傳值 ok=false 代表條件對象解析失敗(引用不存在 / 型別不符 / self 未固定…),
// 會被 evaluator 視為【營業規格書 | 二十七、運算式 | 8】評估失敗並沿樹上傳。
type Resolver interface {
	// Attr 解析主表條目:arg 為 nil 代表裸屬性 / 物件引用(如 morale、round、self、drawLast);
	// arg 非 nil 代表查詢函式(如 tableCount('>=', 2)、handSize(2))。
	Attr(name string, arg []Value) (value Value, ok bool)

	// AttrRef 解析子表引用屬性:以 target 物件引用為主體存取 name 屬性。
	// arg 為 nil 代表 target.name(如 self.calm、drawLast.cardID);
	// arg 非 nil 代表 target.name(arg)查詢函式(如 self.effectStack(101))。
	AttrRef(target Value, name string, arg []Value) (value Value, ok bool)
}

// valueKind 標示 Value 當前承載的型別。
type valueKind int8

const (
	valueNum  valueKind = iota // 數值(整數 / 小數,內部一律 float64)
	valueText                  // 字串
	valueBool                  // 布林
	valueRef                   // 物件引用(卡牌 / 顧客 / 空物件)
)

// tokenKind 是 token 的種類。
type tokenKind int8

const (
	tokenEOF      tokenKind = iota // 輸入結束(哨兵)
	tokenNum                       // 數字字面值
	tokenText                      // 字串字面值('…')
	tokenBool                      // 布林字面值(true / false)
	tokenIdent                     // 識別子(屬性 / 引用 / 函式名)
	tokenAnd                       // AND
	tokenOr                        // OR
	tokenNot                       // !
	tokenPlus                      // +
	tokenMinus                     // -
	tokenStar                      // *
	tokenSlash                     // /
	tokenPercent                   // %
	tokenLT                        // <
	tokenGT                        // >
	tokenLE                        // <=
	tokenGE                        // >=
	tokenEQ                        // ==
	tokenNE                        // !=
	tokenLParen                    // (
	tokenRParen                    // )
	tokenComma                     // ,
	tokenDot                       // .
	tokenQuestion                  // ?
	tokenColon                     // :
)

// token 是詞法分析產出的最小單位。
type token struct {
	kind tokenKind
	text string  // 原始文字(供錯誤訊息;識別子 / 字串則為其內容)
	num  float64 // tokenNum 的數值
	flag bool    // tokenBool 的值
	pos  int     // 在來源字串的 rune 起始索引(0 起算;供錯誤定位)
}
