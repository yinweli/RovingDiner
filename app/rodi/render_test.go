package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteRender(t *testing.T) {
	suite.Run(t, new(SuiteRender))
}

// SuiteRender 驗證渲染工具(render.go): 顯示寬補白 / 截斷 / 兩行對齊表, 全形 2 格半形 1 格混排不歪。
type SuiteRender struct {
	suite.Suite
}

// TestPadTo 驗證右補空白至顯示寬: 全形以 2 格計; 已達 / 超出寬度原樣回傳。
func (this *SuiteRender) TestPadTo() {
	this.Equal("abc  ", padTo("abc", 5))
	this.Equal("中文 ", padTo("中文", 5)) // 全形 2 格: 中文 = 4 格 + 補 1
	this.Equal("abc", padTo("abc", 3))
	this.Equal("abcd", padTo("abcd", 3)) // 超寬不截斷(截斷歸 truncTo)
}

// TestTruncTo 驗證截斷至顯示寬: 全形不切半字; 未超寬原樣回傳。
func (this *SuiteRender) TestTruncTo() {
	this.Equal("abc", truncTo("abc", 5))
	this.Equal("ab", truncTo("abcd", 2))
	this.Equal("中", truncTo("中文字", 3)) // 3 格只容 1 個全形, 不切半字
}

// TestAlignPair 驗證兩行對齊表: 欄寬 = 標籤 / 數值較寬者、欄距 2 空白、行尾不留補白。
func (this *SuiteRender) TestAlignPair() {
	row1, row2 := alignPair([]string{"回合", "士氣值", "x"}, []string{"3/10", "25", "1250"})
	this.Equal("回合  士氣值  x", row1)
	this.Equal("3/10  25      1250", row2)
}
