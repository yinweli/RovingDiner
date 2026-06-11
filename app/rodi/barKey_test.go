package rodi

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/suite"
)

func TestSuiteBarKey(t *testing.T) {
	suite.Run(t, new(SuiteBarKey))
}

// SuiteBarKey 驗證鍵位列組件(barKey.go): 誠實列鍵渲染與按鍵分派同表。
type SuiteBarKey struct {
	suite.Suite
}

// TestNewBarKey 驗證建構: 三模式表就位(選取 / modal 留待 M27 / M26 R3 填); 常態表 = tab 雙向切區
// (shift+tab 併入 tab 標籤不列示)+ 四方向鍵移游標(down/left/right 併入 up 標籤不列示)+ space 模式循環
// + n 步進前進 + q 離開 + ctrl+c 逃生(不列示), 順序照【營業顯示規格書 | 6、畫面規格 | 6.11】常態終態。
func (this *SuiteBarKey) TestNewBarKey() {
	target := newBarKey()
	this.Require().Len(target.bind, 3)
	this.Require().Len(target.bind[keyModeNormal], 10)
	this.Equal("tab", target.bind[keyModeNormal][0].key)
	this.Equal("shift+tab", target.bind[keyModeNormal][1].key)
	this.Equal("", target.bind[keyModeNormal][1].label) // 併入 tab 標籤, 有作用不列示
	this.Equal("up", target.bind[keyModeNormal][2].key)
	this.Equal("down", target.bind[keyModeNormal][3].key)
	this.Equal("left", target.bind[keyModeNormal][4].key)
	this.Equal("right", target.bind[keyModeNormal][5].key)
	this.Equal("", target.bind[keyModeNormal][3].label) // 併入 up 的 [Arrow] 標籤, 有作用不列示
	this.Equal(" ", target.bind[keyModeNormal][6].key)
	this.Equal("n", target.bind[keyModeNormal][7].key)
	this.Equal("q", target.bind[keyModeNormal][8].key)
	this.Equal("ctrl+c", target.bind[keyModeNormal][9].key)
	this.Equal("", target.bind[keyModeNormal][9].label) // 逃生鍵有作用不列示
	this.Empty(target.bind[keyModePick])
	this.Empty(target.bind[keyModeModal])
}

// TestBarKeyView 驗證渲染: 依模式查表固定 2 行、行位照終態安排、空標籤跳過、空表兩行留白(高度穩定)、
// 超寬依預算截斷。
func (this *SuiteBarKey) TestBarKeyView() {
	this.Equal("[Tab/Shift+Tab]切區 [Arrow]移動游標\n[Space]快/慢/步進 [N]前進 [Q]離開", newBarKey().View(keyModeNormal, 100))
	this.Equal("\n", newBarKey().View(keyModeModal, 100)) // 空表(R3 填): 兩行留白

	target := barKey{bind: map[keyMode][]keyBind{keyModeNormal: { // 行位驗證用假表
		{key: "t", label: "[T]導覽", row: 1},
		{key: "u", label: "[U]導覽", row: 1},
		{key: "q", label: "[Q]離開", row: 2},
	}}}
	this.Equal("[T]導覽 [U]導覽\n[Q]離開", target.View(keyModeNormal, 100))
	this.Equal("[T]\n[Q]", target.View(keyModeNormal, 3)) // 超寬截斷
}

// TestBarKeyFind 驗證按鍵分派: 依模式查表、綁定鍵回行為、未綁定回 nil(空表恆未綁定)。
func (this *SuiteBarKey) TestBarKeyFind() {
	target := newBarKey()
	this.Require().NotNil(target.Find(keyModeNormal, "tab"))
	this.Equal(tabMsg{delta: 1}, target.Find(keyModeNormal, "tab")())
	this.Require().NotNil(target.Find(keyModeNormal, "shift+tab"))
	this.Equal(tabMsg{delta: -1}, target.Find(keyModeNormal, "shift+tab")())

	for _, itor := range []string{"up", "down", "left", "right"} {
		this.Require().NotNil(target.Find(keyModeNormal, itor))
		this.Equal(moveMsg{key: itor}, target.Find(keyModeNormal, itor)())
	} // for

	this.Require().NotNil(target.Find(keyModeNormal, " "))
	this.Equal(cycleMsg{}, target.Find(keyModeNormal, " ")())
	this.Require().NotNil(target.Find(keyModeNormal, "n"))
	this.Equal(stepMsg{}, target.Find(keyModeNormal, "n")())
	this.Require().NotNil(target.Find(keyModeNormal, "q"))
	this.Equal(tea.QuitMsg{}, target.Find(keyModeNormal, "q")())
	this.Require().NotNil(target.Find(keyModeNormal, "ctrl+c"))
	this.Equal(tea.QuitMsg{}, target.Find(keyModeNormal, "ctrl+c")())
	this.Nil(target.Find(keyModeNormal, "x"))
	this.Nil(target.Find(keyModeModal, "q")) // 空表查無(modal 態 q 無作用)
}
