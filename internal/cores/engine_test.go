package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteEngine(t *testing.T) {
	suite.Run(t, new(SuiteEngine))
}

// SuiteEngine 驗證驅動引擎對 exprs.Resolver 的委派(engine.go),讀全域詞彙表。
// M5 詞彙表為空,故全域 / 引用屬性查詢一律未命中(詞條於 M6 填入後生效)。
type SuiteEngine struct {
	suite.Suite
}

func (this *SuiteEngine) TestEngineAttr() {
	eng := &engine{}

	_, ok := eng.Attr("morale", nil)
	this.False(ok)
}

func (this *SuiteEngine) TestEngineAttrRef() {
	eng := &engine{}

	_, ok := eng.AttrRef(nil, "calm", nil)
	this.False(ok)
}
