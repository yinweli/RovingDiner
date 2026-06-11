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

// TestNewBarKey 驗證建構: M25 綁定 = space 模式循環 + n 步進前進 + q 離開(列示)+ ctrl+c 逃生(不列示),
// 順序照【營業顯示規格書 | 6、畫面規格 | 6.11】常態行 2 終態。
func (this *SuiteBarKey) TestNewBarKey() {
	target := newBarKey()
	this.Require().Len(target.bind, 4)
	this.Equal(" ", target.bind[0].key)
	this.Equal("n", target.bind[1].key)
	this.Equal("q", target.bind[2].key)
	this.Equal("ctrl+c", target.bind[3].key)
	this.Equal("", target.bind[3].label) // 逃生鍵有作用不列示
}

// TestBarKeyView 驗證渲染: 固定 2 行、行位照終態安排、空標籤跳過、超寬依預算截斷。
func (this *SuiteBarKey) TestBarKeyView() {
	this.Equal("\n[Space]快/慢/步進 [N]前進 [Q]離開", newBarKey().View(100)) // 第 1 行尚無鍵 → 留白(M26 導覽鍵進場)

	target := barKey{bind: []keyBind{ // 行位驗證用假表(M26 起第 1 行才有真鍵)
		{key: "t", label: "[T]導覽", row: 1},
		{key: "u", label: "[U]導覽", row: 1},
		{key: "q", label: "[Q]離開", row: 2},
	}}
	this.Equal("[T]導覽 [U]導覽\n[Q]離開", target.View(100))
	this.Equal("[T]\n[Q]", target.View(3)) // 超寬截斷
}

// TestBarKeyFind 驗證按鍵分派: 綁定鍵回行為、未綁定回 nil。
func (this *SuiteBarKey) TestBarKeyFind() {
	target := newBarKey()
	this.Require().NotNil(target.Find(" "))
	this.Equal(cycleMsg{}, target.Find(" ")())
	this.Require().NotNil(target.Find("n"))
	this.Equal(stepMsg{}, target.Find("n")())
	this.Require().NotNil(target.Find("q"))
	this.Equal(tea.QuitMsg{}, target.Find("q")())
	this.Require().NotNil(target.Find("ctrl+c"))
	this.Equal(tea.QuitMsg{}, target.Find("ctrl+c")())
	this.Nil(target.Find("x"))
}
