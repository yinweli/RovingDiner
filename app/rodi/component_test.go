package rodi

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/games"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteComponent(t *testing.T) {
	suite.Run(t, new(SuiteComponent))
}

// SuiteComponent 驗證組件框架(component.go): 父層組合依掛載順序堆疊、寬度預算逐組件下發。
type SuiteComponent struct {
	suite.Suite
}

// TestComposeView 驗證父層組合: 掛載順序 = 堆疊順序、每組件收到同一寬度預算; 空組件列表回空字串;
// focus 命中掛載索引時該組件首行高亮(範圍外索引全不高亮)。
func (this *SuiteComponent) TestComposeView() {
	game := testGame()
	comp := []component{fakeComponent{text: "a"}, fakeComponent{text: "b"}}
	this.Equal("a:80\nb:80", composeView(game, 80, comp, -1))
	this.Equal("", composeView(game, 80, nil, -1))

	lipgloss.SetColorProfile(termenv.ANSI) // 臨時升 profile 使樣式可見(同 TestFocusView)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	this.Equal("a:80\n"+styleFocus.Render("b:80"), composeView(game, 80, comp, 1)) // 替身單行輸出, 首行即整行
	this.Equal("a:80\nb:80", composeView(game, 80, comp, focusStatus))             // 狀態列 / 日誌索引不在掛載範圍: 全不高亮
}

// TestEndGameView 驗證整場跑完的盤面直讀: 六區組件與狀態列對終局盤面渲染不爆、皆有內容(跨組件冒煙;
// 含聚焦游標態渲染與終局盤面上的游標移動夾界)。狀態列另吃 UI 狀態(模式欄; M25),
// 不符 component 契約(唯讀盤面 + 寬度預算), 單獨呼叫。
func (this *SuiteComponent) TestEndGameView() {
	game := games.Build(0, 601, tester.BuildSheet(), tester.FakeOperator{}, nil)
	games.Loop(game)

	for _, itor := range []component{&panelSeat{}, &panelPool{}, &panelAction{}, &panelEffect{}, &panelHand{}, &panelPile{}} {
		this.NotEmpty(itor.View(game, 100, false))
		itor.Move(game, "right")
		itor.Move(game, "down")
		this.NotEmpty(itor.View(game, 100, true))
	} // for

	this.NotEmpty(barStatus{}.View(game, modeFast, 100))
}

// fakeComponent 測試替身: 渲染自身文字與收到的寬度預算, 供 TestComposeView 驗證下發。
type fakeComponent struct {
	text string
}

func (this fakeComponent) View(game *cores.Game, width int, focus bool) string {
	return fmt.Sprintf("%v:%v", this.text, width)
}

func (this fakeComponent) Move(game *cores.Game, key string) {
}
