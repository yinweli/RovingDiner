package expr

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteLexer(t *testing.T) {
	suite.Run(t, new(SuiteLexer))
}

// SuiteLexer 驗證詞法分析:各 token、雙字元符號、字串(含 CJK)、數字、關鍵字大小寫、空白與非法輸入。
type SuiteLexer struct {
	suite.Suite
}

// kindOf 取出 token 序列的 kind 列表(略去尾端 tokenEOF),方便比對。
func (this *SuiteLexer) kindOf(source string) []tokenKind {
	token, err := lex(source)
	this.Require().NoError(err)
	this.Require().NotEmpty(token)
	this.Require().Equal(tokenEOF, token[len(token)-1].kind) // 尾端必為 EOF

	kind := []tokenKind{}
	for _, itor := range token[:len(token)-1] {
		kind = append(kind, itor.kind)
	} // for

	return kind
}

func (this *SuiteLexer) TestNumber() {
	token, err := lex("5 1.5 100")
	this.Require().NoError(err)
	this.Equal(tokenNumber, token[0].kind)
	this.Equal(5.0, token[0].number)
	this.Equal(1.5, token[1].number)
	this.Equal(100.0, token[2].number)
}

func (this *SuiteLexer) TestNumberDotNotFraction() {
	// 小數點後非數字 → 數字止於整數,點獨立為 tokenDot(如 drawLast 之外的 2.cardID 情境)
	kind := this.kindOf("2.foo")
	this.Equal([]tokenKind{tokenNumber, tokenDot, tokenIdent}, kind)
}

func (this *SuiteLexer) TestString() {
	token, err := lex("'玩家行動'")
	this.Require().NoError(err)
	this.Equal(tokenString, token[0].kind)
	this.Equal("玩家行動", token[0].text) // 內含 CJK、不含引號
}

func (this *SuiteLexer) TestStringUnterminated() {
	_, err := lex("'未結束")
	this.Error(err)
}

func (this *SuiteLexer) TestKeywordCaseInsensitive() {
	token, err := lex("AND or And TRUE false")
	this.Require().NoError(err)
	this.Equal(tokenAnd, token[0].kind)
	this.Equal(tokenOr, token[1].kind)
	this.Equal(tokenAnd, token[2].kind)
	this.Equal(tokenBool, token[3].kind)
	this.True(token[3].boolean)
	this.Equal(tokenBool, token[4].kind)
	this.False(token[4].boolean)
}

func (this *SuiteLexer) TestIdent() {
	kind := this.kindOf("morale self drawLast")
	this.Equal([]tokenKind{tokenIdent, tokenIdent, tokenIdent}, kind)
}

func (this *SuiteLexer) TestOperatorSingle() {
	kind := this.kindOf("+ - * / % ( ) , . ? : < >")
	this.Equal([]tokenKind{
		tokenPlus, tokenMinus, tokenStar, tokenSlash, tokenPercent,
		tokenLParen, tokenRParen, tokenComma, tokenDot, tokenQuestion, tokenColon,
		tokenLT, tokenGT,
	}, kind)
}

func (this *SuiteLexer) TestOperatorDouble() {
	kind := this.kindOf("<= >= == != !")
	this.Equal([]tokenKind{tokenLE, tokenGE, tokenEQ, tokenNE, tokenNot}, kind)
}

func (this *SuiteLexer) TestNotVersusNe() {
	// ! 與 != 須正確區分
	this.Equal([]tokenKind{tokenNot, tokenIdent}, this.kindOf("!a"))
	this.Equal([]tokenKind{tokenIdent, tokenNE, tokenIdent}, this.kindOf("a!=b"))
}

func (this *SuiteLexer) TestWhitespaceIgnored() {
	this.Equal(this.kindOf("1+2"), this.kindOf("  1 \t + \n 2  "))
}

func (this *SuiteLexer) TestBareAssignIllegal() {
	_, err := lex("a = 1") // 賦值不屬於運算式
	this.Error(err)
}

func (this *SuiteLexer) TestUnexpectedChar() {
	_, err := lex("a & b")
	this.Error(err)
}

func (this *SuiteLexer) TestEmpty() {
	token, err := lex("")
	this.Require().NoError(err)
	this.Len(token, 1)
	this.Equal(tokenEOF, token[0].kind) // 空輸入只有 EOF
}
