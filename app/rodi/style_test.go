package rodi

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"
)

func TestSuiteStyle(t *testing.T) {
	suite.Run(t, new(SuiteStyle))
}

// SuiteStyle 驗證樣式層(style.go): 聚焦高亮與日誌行角色樣式分派。
type SuiteStyle struct {
	suite.Suite
}

// TestFocusView 驗證聚焦高亮: 僅首行(標題列)上反白、其餘行原樣; 首行既有樣式(標題青 / 提醒黃 /
// 框線暗)先剝除再反白(讓位)。臨時升 ANSI profile 使樣式可見、測畢還原 Ascii(無 TTY 預設),
// 不擾其他釘字串測試。
func (this *SuiteStyle) TestFocusView() {
	lipgloss.SetColorProfile(termenv.ANSI)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	this.Equal(styleFocus.Render("+- 座位 -+")+"\n| 內容 |", focusView("+- 座位 -+\n| 內容 |"))
	this.Equal(styleFocus.Render("單行"), focusView("單行")) // 無第二行(防禦): 整輸出即首行
	this.Equal(styleFocus.Render("+- 手牌 (請出牌) -+")+"\n| 內容 |",
		focusView(styleLine.Render("+- ")+styleTitle.Render("手牌")+styleNote.Render(" (請出牌)")+styleLine.Render(" -+")+"\n| 內容 |")) // 剝色讓位
}

// TestRestyle 驗證整段重上樣式: 段內既有樣式剝除後上新樣式(嵌套 reset 不切斷整段態); 純文字等同直上。
func (this *SuiteStyle) TestRestyle() {
	lipgloss.SetColorProfile(termenv.ANSI)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	this.Equal(styleCursor.Render("不棄 封印"), restyle(&styleCursor, "不棄 "+styleBad.Render("封印")))
	this.Equal(styleDim.Render("素文字"), restyle(&styleDim, "素文字"))
}

// TestLineStyle 驗證行角色樣式分派(首字即角色、欄 2 原色)。
func (this *SuiteStyle) TestLineStyle() {
	this.Equal(styleTitle, lineStyle("[R3 玩家行動] 玩家出牌"))
	this.Equal(styleOperand, lineStyle("* 301@開朗"))
	this.Equal(styleEffect, lineStyle("- 401@加耐#7"))
	this.Equal(styleFlow, lineStyle("$ 出牌點數 -= 2 >> 8"))
	this.Equal(styleField, lineStyle("  飽食值 -= 2 >> 3"))
}
