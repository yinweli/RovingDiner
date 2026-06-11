package exprs

import (
	"strconv"
	"strings"
)

// lex 把運算式原始字串切成 token 序列(尾端必附 tokenEOF); 遇非法字元 / 未結束字串 / 無效數字回傳錯誤。
// 以 rune 切片掃描, 字串字面值內可含 CJK; 識別子僅限 ASCII 英數底線。
func lex(source string) (result []token, err error) {
	char := []rune(source)
	size := len(char)
	result = []token{}

	for i := 0; i < size; {
		c := char[i]

		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++ // 略過空白

		case c >= '0' && c <= '9':
			tok, width, errNumber := lexNum(char, i)

			if errNumber != nil {
				return nil, errNumber
			} // if

			tok.pos = i
			result = append(result, tok)
			i += width

		case c == '\'':
			tok, width, errText := lexText(char, i)

			if errText != nil {
				return nil, errText
			} // if

			tok.pos = i
			result = append(result, tok)
			i += width

		case isIdentStart(c):
			start := i

			for i < size && isIdentPart(char[i]) {
				i++
			} // for

			tok := identToken(string(char[start:i]))
			tok.pos = start
			result = append(result, tok)

		default:
			tok, width, errOperator := lexOperator(char, i)

			if errOperator != nil {
				return nil, errOperator
			} // if

			tok.pos = i
			result = append(result, tok)
			i += width
		} // switch
	} // for

	result = append(result, token{kind: tokenEOF, text: "<eof>", pos: size})
	return result, nil
}

// lexNum 掃描整數 / 小數字面值; 小數點後須緊接數字才視為小數部分(對齊【營業規格書 | 二十七、運算式 | 4】)。
func lexNum(char []rune, start int) (result token, width int, err error) {
	size := len(char)
	i := start

	for i < size && char[i] >= '0' && char[i] <= '9' {
		i++
	} // for

	if i+1 < size && char[i] == '.' && char[i+1] >= '0' && char[i+1] <= '9' {
		i++ // 吃小數點

		for i < size && char[i] >= '0' && char[i] <= '9' {
			i++
		} // for
	} // if

	text := string(char[start:i])
	number, errParse := strconv.ParseFloat(text, 64)

	if errParse != nil {
		return token{}, 0, newError(start, "無效的數字:"+text)
	} // if

	return token{kind: tokenNum, text: text, num: number}, i - start, nil
}

// lexText 掃描單引號字串字面值; 未遇結尾單引號則回傳錯誤。
func lexText(char []rune, start int) (result token, width int, err error) {
	size := len(char)
	i := start + 1 // 跳過開頭單引號

	for i < size && char[i] != '\'' {
		i++
	} // for

	if i >= size {
		return token{}, 0, newError(start, "字串未結束, 缺少結尾單引號 '")
	} // if

	text := string(char[start+1 : i])
	return token{kind: tokenText, text: text}, i - start + 1, nil // +1 涵蓋結尾單引號
}

// lexOperator 掃描運算符 / 標點; 處理 <= >= == != 等雙字元符號, 單獨的 '=' 視為非法(賦值不屬於運算式)。
func lexOperator(char []rune, i int) (result token, width int, err error) {
	c := char[i]
	next := rune(0)

	if i+1 < len(char) {
		next = char[i+1]
	} // if

	switch c {
	case '+':
		return token{kind: tokenPlus, text: "+"}, 1, nil

	case '-':
		return token{kind: tokenMinus, text: "-"}, 1, nil

	case '*':
		return token{kind: tokenStar, text: "*"}, 1, nil

	case '/':
		return token{kind: tokenSlash, text: "/"}, 1, nil

	case '%':
		return token{kind: tokenPercent, text: "%"}, 1, nil

	case '(':
		return token{kind: tokenLParen, text: "("}, 1, nil

	case ')':
		return token{kind: tokenRParen, text: ")"}, 1, nil

	case ',':
		return token{kind: tokenComma, text: ","}, 1, nil

	case '.':
		return token{kind: tokenDot, text: "."}, 1, nil

	case '?':
		return token{kind: tokenQuestion, text: "?"}, 1, nil

	case ':':
		return token{kind: tokenColon, text: ":"}, 1, nil

	case '<':
		if next == '=' {
			return token{kind: tokenLE, text: "<="}, 2, nil
		} // if

		return token{kind: tokenLT, text: "<"}, 1, nil

	case '>':
		if next == '=' {
			return token{kind: tokenGE, text: ">="}, 2, nil
		} // if

		return token{kind: tokenGT, text: ">"}, 1, nil

	case '=':
		if next == '=' {
			return token{kind: tokenEQ, text: "=="}, 2, nil
		} // if

		return token{}, 0, newError(i, "不該出現的字元 '='(賦值不屬於運算式)")

	case '!':
		if next == '=' {
			return token{kind: tokenNE, text: "!="}, 2, nil
		} // if

		return token{kind: tokenNot, text: "!"}, 1, nil

	default:
		return token{}, 0, newError(i, "不認得的字元:"+string(c))
	} // switch
}

// identToken 把識別子文字轉成對應 token; AND / OR / true / false / none 不區分大小寫, 其餘為 tokenIdent。
func identToken(text string) token {
	switch strings.ToLower(text) {
	case "and":
		return token{kind: tokenAnd, text: text}

	case "or":
		return token{kind: tokenOr, text: text}

	case "true":
		return token{kind: tokenBool, text: text, flag: true}

	case "false":
		return token{kind: tokenBool, text: text, flag: false}

	case "none":
		return token{kind: tokenNone, text: text}

	default:
		return token{kind: tokenIdent, text: text}
	} // switch
}

// isIdentStart 回傳 rune 是否可作識別子起始字元(ASCII 英文字母或底線)。
func isIdentStart(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// isIdentPart 回傳 rune 是否可作識別子後續字元(起始字元或數字)。
func isIdentPart(c rune) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

// token 是詞法分析產出的最小單位。
type token struct {
	kind tokenKind
	text string  // 原始文字(供錯誤訊息; 識別子 / 字串則為其內容)
	num  float64 // tokenNum 的數值
	flag bool    // tokenBool 的值
	pos  int     // 在來源字串的 rune 起始索引(0 起算; 供錯誤定位)
}

// tokenKind 是 token 的種類, 對齊【營業規格書 | 二十七、運算式 | 2】運算符與【營業規格書 | 二十七、運算式 | 4】字面值。
type tokenKind int8

const (
	tokenEOF      tokenKind = iota // 輸入結束(哨兵)
	tokenNum                       // 數字字面值(整數 / 小數)
	tokenText                      // 字串字面值('…')
	tokenBool                      // 布林字面值(true / false)
	tokenNone                      // 空物件字面值(none)
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
