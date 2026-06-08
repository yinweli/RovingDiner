package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteCommand(t *testing.T) {
	suite.Run(t, new(SuiteCommand))
}

// SuiteCommand 驗證操作命令詞彙表(command.go)的登錄查詢 HasCommand。M5 表為空,故任何名稱皆未登錄(詞條於 M9 填入後生效)。
type SuiteCommand struct {
	suite.Suite
}

func (this *SuiteCommand) TestCommandHas() {
	this.False(HasCommand("handAdd"))
	this.False(HasCommand("deckShuffle"))
}
