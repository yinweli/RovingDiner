package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteAttr(t *testing.T) {
	suite.Run(t, new(SuiteAttr))
}

// SuiteAttr 驗證全域屬性詞彙表(attr.go)的登錄查詢 HasAttr。M5 表為空,故任何名稱皆未登錄(詞條於 M6 填入後生效)。
type SuiteAttr struct {
	suite.Suite
}

func (this *SuiteAttr) TestAttrHas() {
	this.False(HasAttr("morale"))
	this.False(HasAttr("score"))
}
