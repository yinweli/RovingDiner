package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteSelector(t *testing.T) {
	suite.Run(t, new(SuiteSelector))
}

// SuiteSelector 驗證命令對象詞彙表(selector.go)的登錄查詢 HasSelector。M5 表為空,故任何名稱皆未登錄(詞條於 M8 填入後生效)。
type SuiteSelector struct {
	suite.Suite
}

func (this *SuiteSelector) TestSelectorHas() {
	this.False(HasSelector("deckTop"))
	this.False(HasSelector("handAll"))
}
