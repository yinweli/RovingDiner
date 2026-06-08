package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteAttrRef(t *testing.T) {
	suite.Run(t, new(SuiteAttrRef))
}

// SuiteAttrRef 驗證引用屬性詞彙表(attrRef.go)的登錄查詢 HasAttrRef。M5 表為空,故任何名稱皆未登錄(詞條於 M6 填入後生效)。
type SuiteAttrRef struct {
	suite.Suite
}

func (this *SuiteAttrRef) TestAttrRefHas() {
	this.False(HasAttrRef("calm"))
	this.False(HasAttrRef("sate"))
}
