package expr

// tokenKind 是 token 的種類。
type tokenKind int8

const (
	tokenEOF      tokenKind = iota // 輸入結束(哨兵)
	tokenNumber                    // 數字字面值
	tokenString                    // 字串字面值('…')
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
	kind    tokenKind
	text    string  // 原始文字(供錯誤訊息;識別子 / 字串則為其內容)
	number  float64 // tokenNumber 的數值
	boolean bool    // tokenBool 的值
}
