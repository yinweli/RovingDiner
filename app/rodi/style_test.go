package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteStyle(t *testing.T) {
	suite.Run(t, new(SuiteStyle))
}

// SuiteStyle 驗證樣式層(style.go): 日誌行角色樣式分派。
type SuiteStyle struct {
	suite.Suite
}

// TestLineStyle 驗證行角色樣式分派(首字即角色、欄 2 原色)。
func (this *SuiteStyle) TestLineStyle() {
	this.Equal(styleTitle, lineStyle("[R3 玩家行動] 玩家出牌"))
	this.Equal(styleOperand, lineStyle("* 301@開朗"))
	this.Equal(styleEffect, lineStyle("- 401@加耐#7"))
	this.Equal(styleFlow, lineStyle("$ 出牌點數 -= 2 >> 8"))
	this.Equal(styleField, lineStyle("  飽食值 -= 2 >> 3"))
}
