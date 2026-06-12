package exprs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteLexer(t *testing.T) {
	suite.Run(t, new(SuiteLexer))
}

// SuiteLexer 驗證詞法分析: 各 token、雙字元符號、字串(含 CJK)、數字、關鍵字大小寫、空白與非法輸入。
type SuiteLexer struct {
	suite.Suite
}

func (this *SuiteLexer) TestLexWhitespaceIgnored() {
	this.Equal(this.kindOf("1+2"), this.kindOf("  1 \t + \n 2  "))
}

func (this *SuiteLexer) TestLexEmpty() {
	result, err := lex("")
	this.Require().NoError(err)
	this.Len(result, 1)
	this.Equal(tokenEOF, result[0].kind) // 空輸入只有 EOF
}

func (this *SuiteLexer) TestLexPosition() {
	result, err := lex("12 + ab")
	this.Require().NoError(err)
	this.Equal(0, result[0].pos) // 12
	this.Equal(3, result[1].pos) // +
	this.Equal(5, result[2].pos) // ab
	this.Equal(7, result[3].pos) // EOF(置於字串長度處)
}

func (this *SuiteLexer) TestLexNum() {
	result, err := lex("5 1.5 100")
	this.Require().NoError(err)
	this.Equal(tokenNum, result[0].kind)
	this.Equal(5.0, result[0].num)
	this.Equal(1.5, result[1].num)
	this.Equal(100.0, result[2].num)
}

func (this *SuiteLexer) TestLexNumMultiDot() {
	// 多個小數點: 1.2 為一數、第二個點獨立、3 為另一數(lexer 只組合合法數字串, 不產生無效數字)
	kind := this.kindOf("1.2.3")
	this.Equal([]tokenKind{tokenNum, tokenDot, tokenNum}, kind)
}

func (this *SuiteLexer) TestLexNumDotNotFraction() {
	// 小數點後非數字 → 數字止於整數, 點獨立為 tokenDot
	kind := this.kindOf("2.foo")
	this.Equal([]tokenKind{tokenNum, tokenDot, tokenIdent}, kind)
}

func (this *SuiteLexer) TestLexText() {
	result, err := lex("'玩家行動'")
	this.Require().NoError(err)
	this.Equal(tokenText, result[0].kind)
	this.Equal("玩家行動", result[0].text) // 內含 CJK、不含引號
}

func (this *SuiteLexer) TestLexTextUnterminated() {
	_, err := lex("'未結束")
	this.Require().Error(err)
}

func (this *SuiteLexer) TestLexOperatorSingle() {
	kind := this.kindOf("+ - * / % ( ) , . ? : < >")
	this.Equal([]tokenKind{
		tokenPlus, tokenMinus, tokenStar, tokenSlash, tokenPercent,
		tokenLParen, tokenRParen, tokenComma, tokenDot, tokenQuestion, tokenColon,
		tokenLT, tokenGT,
	}, kind)
}

func (this *SuiteLexer) TestLexOperatorDouble() {
	kind := this.kindOf("<= >= == != !")
	this.Equal([]tokenKind{tokenLE, tokenGE, tokenEQ, tokenNE, tokenNot}, kind)
}

func (this *SuiteLexer) TestLexOperatorNotVersusNe() {
	// ! 與 != 須正確區分
	this.Equal([]tokenKind{tokenNot, tokenIdent}, this.kindOf("!a"))
	this.Equal([]tokenKind{tokenIdent, tokenNE, tokenIdent}, this.kindOf("a!=b"))
}

func (this *SuiteLexer) TestLexOperatorBareAssignIllegal() {
	_, err := lex("a = 1") // 賦值不屬於運算式
	this.Require().Error(err)
}

func (this *SuiteLexer) TestLexOperatorUnexpectedChar() {
	_, err := lex("a & b")
	this.Require().Error(err)
}

func (this *SuiteLexer) TestIdentToken() {
	kind := this.kindOf("morale self drawLast")
	this.Equal([]tokenKind{tokenIdent, tokenIdent, tokenIdent}, kind)
}

func (this *SuiteLexer) TestIdentTokenKeyword() {
	result, err := lex("AND or And TRUE false none NONE")
	this.Require().NoError(err)
	this.Equal(tokenAnd, result[0].kind)
	this.Equal(tokenOr, result[1].kind)
	this.Equal(tokenAnd, result[2].kind) // 不區分大小寫
	this.Equal(tokenBool, result[3].kind)
	this.True(result[3].flag)
	this.Equal(tokenBool, result[4].kind)
	this.False(result[4].flag)
	this.Equal(tokenNone, result[5].kind) // none 為空物件字面值關鍵字
	this.Equal(tokenNone, result[6].kind) // 不區分大小寫
}

// kindOf 取出 token 序列的 tokenKind 列表(略去尾端 tokenEOF), 方便比對。
func (this *SuiteLexer) kindOf(source string) []tokenKind {
	result, err := lex(source)
	this.Require().NoError(err)
	this.Require().NotEmpty(result)
	this.Require().Equal(tokenEOF, result[len(result)-1].kind) // 尾端必為 EOF
	kind := []tokenKind{}

	for _, itor := range result[:len(result)-1] {
		kind = append(kind, itor.kind)
	} // for

	return kind
}
