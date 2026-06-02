package expr

import (
	"errors"
	"strconv"
	"strings"
)

// lex 把運算式原始字串切成 token 序列(尾端必附 tokenEOF);遇非法字元 / 未結束字串 / 無效數字回傳錯誤。
// 以 rune 切片掃描,字串字面值內可含 CJK;識別子僅限 ASCII 英數底線。
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
			tok, width, errNumber := lexNumber(char, i)
			if errNumber != nil {
				return nil, errNumber
			} // if
			result = append(result, tok)
			i += width

		case c == '\'':
			tok, width, errString := lexString(char, i)
			if errString != nil {
				return nil, errString
			} // if
			result = append(result, tok)
			i += width

		case isIdentStart(c):
			start := i
			for i < size && isIdentPart(char[i]) {
				i++
			} // for
			result = append(result, identToken(string(char[start:i])))

		default:
			tok, width, errOperator := lexOperator(char, i)
			if errOperator != nil {
				return nil, errOperator
			} // if
			result = append(result, tok)
			i += width
		} // switch
	} // for

	result = append(result, token{kind: tokenEOF, text: "<eof>"})
	return result, nil
}

// lexNumber 掃描整數 / 小數字面值;小數點後須緊接數字才視為小數部分。
func lexNumber(char []rune, start int) (result token, width int, err error) {
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
		return token{}, 0, errors.New("expr: 無效的數字: " + text)
	} // if

	return token{kind: tokenNumber, text: text, number: number}, i - start, nil
}

// lexString 掃描單引號字串字面值;未遇結尾單引號則回傳錯誤。
func lexString(char []rune, start int) (result token, width int, err error) {
	size := len(char)
	i := start + 1 // 跳過開頭單引號

	for i < size && char[i] != '\'' {
		i++
	} // for

	if i >= size {
		return token{}, 0, errors.New("expr: 字串未結束: " + string(char[start:]))
	} // if

	text := string(char[start+1 : i])
	return token{kind: tokenString, text: text}, i - start + 1, nil // +1 涵蓋結尾單引號
}

// lexOperator 掃描運算符 / 標點;處理 <= >= == != 等雙字元符號,單獨的 '=' 視為非法(賦值不屬於運算式)。
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
		return token{}, 0, errors.New("expr: 未預期的字元 '='(賦值不屬於運算式)")
	case '!':
		if next == '=' {
			return token{kind: tokenNE, text: "!="}, 2, nil
		} // if
		return token{kind: tokenNot, text: "!"}, 1, nil
	default:
		return token{}, 0, errors.New("expr: 未預期的字元: " + string(c))
	} // switch
}

// identToken 把識別子文字轉成對應 token;AND / OR / true / false 不區分大小寫,其餘為 tokenIdent。
func identToken(text string) token {
	switch strings.ToLower(text) {
	case "and":
		return token{kind: tokenAnd, text: text}
	case "or":
		return token{kind: tokenOr, text: text}
	case "true":
		return token{kind: tokenBool, text: text, boolean: true}
	case "false":
		return token{kind: tokenBool, text: text, boolean: false}
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
