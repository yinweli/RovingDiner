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

// TestNewBarKey 驗證建構: 三模式表就位(M27 R4 全鍵終態); 常態表 = tab 雙向切區(shift+tab 併入 tab 標籤
// 不列示)+ 四方向鍵移游標(down/left/right 併入 up 標籤不列示)+ enter 檢視 + f1 說明 + f2 計數
// + space 模式循環 + n 步進前進 + p 出牌 + e 結束(M27 R3 點亮)+ q 離開 + ctrl+c 逃生(不列示);
// 選取表 = tab 雙向 + 四方向 + space 加選/取消 + enter 確認 + ctrl+c 逃生(M27 R4);
// modal 態表 = 上下捲動 + esc 關閉 + ctrl+c 逃生, 順序照【營業顯示規格書 | 6、畫面規格 | 6.11】終態。
func (this *SuiteBarKey) TestNewBarKey() {
	target := newBarKey()
	this.Require().Len(target.bind, 3)
	this.Require().Len(target.bind[keyModeNormal], 15)
	this.Equal("tab", target.bind[keyModeNormal][0].key)
	this.Equal("shift+tab", target.bind[keyModeNormal][1].key)
	this.Equal("", target.bind[keyModeNormal][1].label) // 併入 tab 標籤, 有作用不列示
	this.Equal("up", target.bind[keyModeNormal][2].key)
	this.Equal("down", target.bind[keyModeNormal][3].key)
	this.Equal("left", target.bind[keyModeNormal][4].key)
	this.Equal("right", target.bind[keyModeNormal][5].key)
	this.Equal("", target.bind[keyModeNormal][3].label) // 併入 up 的 [Arrow] 標籤, 有作用不列示
	this.Equal("enter", target.bind[keyModeNormal][6].key)
	this.Equal("[Enter]檢視", target.bind[keyModeNormal][6].label) // R4 點亮(檢視 modal 全區到位)
	this.Equal("f1", target.bind[keyModeNormal][7].key)
	this.Equal("[F1]說明", target.bind[keyModeNormal][7].label) // M26A 改綁格式說明
	this.Equal("f2", target.bind[keyModeNormal][8].key)
	this.Equal("[F2]計數", target.bind[keyModeNormal][8].label) // M26A 計數讓位至 F2
	this.Equal(" ", target.bind[keyModeNormal][9].key)
	this.Equal("n", target.bind[keyModeNormal][10].key)
	this.Equal("p", target.bind[keyModeNormal][11].key)
	this.Equal("[P]出牌", target.bind[keyModeNormal][11].label) // M27 R3 點亮(玩家行動等待態到位)
	this.Equal("e", target.bind[keyModeNormal][12].key)
	this.Equal("[E]結束", target.bind[keyModeNormal][12].label)
	this.Equal("q", target.bind[keyModeNormal][13].key)
	this.Equal("ctrl+c", target.bind[keyModeNormal][14].key)
	this.Equal("", target.bind[keyModeNormal][14].label) // 逃生鍵有作用不列示
	this.Require().Len(target.bind[keyModePick], 9)
	this.Equal("tab", target.bind[keyModePick][0].key)
	this.Equal("shift+tab", target.bind[keyModePick][1].key)
	this.Equal("up", target.bind[keyModePick][2].key)
	this.Equal(" ", target.bind[keyModePick][6].key)
	this.Equal("[Space]加選/取消", target.bind[keyModePick][6].label)
	this.Equal("enter", target.bind[keyModePick][7].key)
	this.Equal("[Enter]確認", target.bind[keyModePick][7].label)
	this.Equal("ctrl+c", target.bind[keyModePick][8].key)
	this.Require().Len(target.bind[keyModeModal], 4)
	this.Equal("up", target.bind[keyModeModal][0].key)
	this.Equal("esc", target.bind[keyModeModal][2].key)
	this.Equal("ctrl+c", target.bind[keyModeModal][3].key)
}

// TestBarKeyView 驗證渲染: 依模式查表固定 2 行、行位照終態安排、空標籤跳過、提示文字取代行 2(選取模式)、
// 超寬依預算截斷。
func (this *SuiteBarKey) TestBarKeyView() {
	this.Equal("[Tab/Shift+Tab]切區 [Arrow]移動游標 [Enter]檢視 [F1]說明 [F2]計數\n[Space]快/慢/步進 [N]前進 [P]出牌 [E]結束 [Q]離開", newBarKey().View(keyModeNormal, 100, ""))
	this.Equal("[Up/Down]欄位捲動 [Esc]關閉\n", newBarKey().View(keyModeModal, 100, ""))                                                                 // modal 態: 行 2 留白
	this.Equal("[Tab/Shift+Tab]切區 [Arrow]移動游標 [Space]加選/取消 [Enter]確認\n開朗 要求選顧客 (已選 1/2)", newBarKey().View(keyModePick, 100, "開朗 要求選顧客 (已選 1/2)")) // 選取: 行 1 鍵位、行 2 提示

	target := barKey{bind: map[keyMode][]keyBind{keyModeNormal: { // 行位驗證用假表
		{key: "t", label: "[T]導覽", row: 1},
		{key: "u", label: "[U]導覽", row: 1},
		{key: "q", label: "[Q]離開", row: 2},
	}}}
	this.Equal("[T]導覽 [U]導覽\n[Q]離開", target.View(keyModeNormal, 100, ""))
	this.Equal("[T]\n[Q]", target.View(keyModeNormal, 3, "")) // 超寬截斷
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

	this.Require().NotNil(target.Find(keyModeNormal, "enter"))
	this.Equal(enterMsg{}, target.Find(keyModeNormal, "enter")())
	this.Require().NotNil(target.Find(keyModeNormal, "f1"))
	this.Equal(helpMsg{}, target.Find(keyModeNormal, "f1")())
	this.Require().NotNil(target.Find(keyModeNormal, "f2"))
	this.Equal(countMsg{}, target.Find(keyModeNormal, "f2")())
	this.Require().NotNil(target.Find(keyModeModal, "esc"))
	this.Equal(popMsg{}, target.Find(keyModeModal, "esc")())
	this.Require().NotNil(target.Find(keyModeModal, "up"))
	this.Equal(moveMsg{key: "up"}, target.Find(keyModeModal, "up")())
	this.Require().NotNil(target.Find(keyModePick, " "))
	this.Equal(toggleMsg{}, target.Find(keyModePick, " ")())
	this.Require().NotNil(target.Find(keyModePick, "enter"))
	this.Equal(confirmMsg{}, target.Find(keyModePick, "enter")())
	this.Require().NotNil(target.Find(keyModePick, "tab"))
	this.Equal(tabMsg{delta: 1}, target.Find(keyModePick, "tab")())
	this.Require().NotNil(target.Find(keyModePick, "left"))
	this.Equal(moveMsg{key: "left"}, target.Find(keyModePick, "left")())
	this.Require().NotNil(target.Find(keyModeNormal, " "))
	this.Equal(cycleMsg{}, target.Find(keyModeNormal, " ")())
	this.Require().NotNil(target.Find(keyModeNormal, "n"))
	this.Equal(stepMsg{}, target.Find(keyModeNormal, "n")())
	this.Require().NotNil(target.Find(keyModeNormal, "p"))
	this.Equal(playMsg{}, target.Find(keyModeNormal, "p")())
	this.Require().NotNil(target.Find(keyModeNormal, "e"))
	this.Equal(endMsg{}, target.Find(keyModeNormal, "e")())
	this.Require().NotNil(target.Find(keyModeNormal, "q"))
	this.Equal(tea.QuitMsg{}, target.Find(keyModeNormal, "q")())
	this.Require().NotNil(target.Find(keyModeNormal, "ctrl+c"))
	this.Equal(tea.QuitMsg{}, target.Find(keyModeNormal, "ctrl+c")())
	this.Nil(target.Find(keyModeNormal, "x"))
	this.Nil(target.Find(keyModeModal, "q")) // modal 態 q 無作用
	this.Nil(target.Find(keyModePick, "q"))  // 選取模式無離開鍵(不可取消、必須選出)
}
