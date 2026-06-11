package rodi

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuitePanelAction(t *testing.T) {
	suite.Run(t, new(SuitePanelAction))
}

// SuitePanelAction 驗證行動佇列組件(panelAction.go): 2 行項橫向並排 / 行動類型標記 / 截斷記號。
type SuitePanelAction struct {
	suite.Suite
}

// TestPanelActionView 驗證渲染: 隊頭在左、技能 / 顧客兩行 content-fit 對齊、空佇列兩行留白、超寬補右緣 >。
func (this *SuitePanelAction) TestPanelActionView() {
	game := testGame()
	this.Equal(strings.Join([]string{ // 空佇列留白(帶框)
		"+- 行動佇列(0) " + strings.Repeat("-", 44) + "+",
		"| " + padTo("", 56) + " |",
		"| " + padTo("", 56) + " |",
	}, "\n"), (&panelAction{}).View(game, 60, false))

	guest := cores.NewGuest(game, 501)
	game.Action.Push(cores.NewAction(guest, cores.TaskCalm, 301))
	game.Action.Push(cores.NewAction(guest, cores.TaskSate, 301))

	this.Equal(strings.Join([]string{
		"+- 行動佇列(2) " + strings.Repeat("-", 44) + "+",
		"| " + padTo("301@開朗     301@開朗", 56) + " |",
		"| " + padTo("501@老饕 耐  501@老饕 飽", 56) + " |",
	}, "\n"), (&panelAction{}).View(game, 60, false))

	row := strings.Split((&panelAction{}).View(game, 10, false), "\n") // 超寬: 內容寬 6、> 站最後內容格
	this.Equal("| 301@ > |", row[1])
}

// TestPanelActionMove 驗證游標移動: 單列左右、夾界不迴繞; 聚焦時游標項目區塊 2 行反白、
// 窗格跟游標捲(左緣 < 兩行同縮排)。
func (this *SuitePanelAction) TestPanelActionMove() {
	game := testGame()
	guest := cores.NewGuest(game, 501)
	game.Action.Push(cores.NewAction(guest, cores.TaskCalm, 301))
	game.Action.Push(cores.NewAction(guest, cores.TaskSate, 301))
	target := &panelAction{}
	target.Move(game, "right")
	this.Equal(1, target.cursor)
	target.Move(game, "right") // 右端夾住
	this.Equal(1, target.cursor)
	target.Move(game, "up") // 上下不動作
	this.Equal(1, target.cursor)

	lipgloss.SetColorProfile(termenv.ANSI) // 臨時升 profile 使樣式可見(同 TestFocusView)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	row := strings.Split(target.View(game, 17, true), "\n") // 內容寬 13 = 左緣 2 + 欄寬 11: 窗格捲到游標
	this.Equal("| < "+styleCursor.Render(padTo("301@開朗", 11))+" |", row[1])
	this.Equal("|   "+styleCursor.Render(padTo("501@老饕 飽", 11))+" |", row[2]) // 第 2 行同縮排對齊
}

// TestPanelActionItem 驗證游標項目: 游標下的行動項; 空佇列回 nil。
func (this *SuitePanelAction) TestPanelActionItem() {
	game := testGame()
	target := &panelAction{}
	this.Nil(target.Item(game)) // 空佇列

	guest := cores.NewGuest(game, 501)
	game.Action.Push(cores.NewAction(guest, cores.TaskCalm, 301))
	game.Action.Push(cores.NewAction(guest, cores.TaskSate, 301))
	target.Move(game, "right")
	this.Equal(game.Action[1], target.Item(game))
}

// TestTaskName 驗證行動類型標記; 越界顯 ?。
func (this *SuitePanelAction) TestTaskName() {
	this.Equal("飽", taskName(cores.TaskSate))
	this.Equal("耐", taskName(cores.TaskCalm))
	this.Equal("?", taskName(cores.TaskKind(9)))
}
