package exprs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteSyntaxError(t *testing.T) {
	suite.Run(t, new(SuiteSyntaxError))
}

// SuiteSyntaxError 驗證 SyntaxError 的訊息格式:Pos 0 起算,顯示時以 1 起算「第 N 字附近」。
type SuiteSyntaxError struct {
	suite.Suite
}

func (this *SuiteSyntaxError) TestSyntaxErrorError() {
	err := &SyntaxError{Pos: 4, Msg: "括號未閉合,缺少 ')'"}
	this.Equal("第 5 字附近：括號未閉合,缺少 ')'", err.Error()) // Pos 4(0 起算)→ 顯示第 5 字
}

func (this *SuiteSyntaxError) TestNewError() {
	err := newError(0, "測試訊息")
	this.Require().Error(err)

	syntaxError, ok := err.(*SyntaxError)
	this.Require().True(ok) // newError 必回傳 *SyntaxError
	this.Equal(0, syntaxError.Pos)
	this.Equal("測試訊息", syntaxError.Msg)
	this.Equal("第 1 字附近：測試訊息", err.Error())
}
