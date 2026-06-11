package rodi

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuitePanelPile(t *testing.T) {
	suite.Run(t, new(SuitePanelPile))
}

// SuitePanelPile 驗證牌堆組件(panelPile.go): 三堆列格式(堆頂在左)/ 空堆留白 / 截斷記號。
type SuitePanelPile struct {
	suite.Suite
}

// TestPanelPileView 驗證渲染: 抽 / 棄 / 流放三列、堆頂在左、空堆冒號後留空、超寬補右緣 >。
func (this *SuitePanelPile) TestPanelPileView() {
	game := testGame()
	game.Deck.Push(cores.NewCard(game, 101))
	game.Deck.Push(cores.NewCard(game, 103))
	game.Drop.Push(cores.NewCard(game, 101))

	this.Equal(strings.Join([]string{ // 牌堆序 = 新進入者置頂: 抽 [103 101]
		"+- 牌堆 " + strings.Repeat("-", 51) + "+",
		"| " + padTo("抽牌堆(2): 103@結帳 101@上菜", 56) + " |",
		"| " + padTo("棄牌堆(1): 101@上菜", 56) + " |",
		"| " + padTo("流放堆(0):", 56) + " |",
	}, "\n"), (&panelPile{}).View(game, 60, false))

	row := strings.Split((&panelPile{}).View(game, 14, false), "\n") // 超寬: 內容寬 10、> 站最後內容格
	this.Equal("| 抽牌堆(2 > |", row[1])
}

// TestPanelPileMove 驗證游標移動: 上下換堆、左右堆內移、夾界不迴繞; 聚焦時游標卡 token 反白。
func (this *SuitePanelPile) TestPanelPileMove() {
	game := testGame()
	game.Deck.Push(cores.NewCard(game, 101))
	game.Deck.Push(cores.NewCard(game, 103))
	target := &panelPile{}
	target.Move(game, "right")
	this.Equal(1, target.curIdx)
	target.Move(game, "down") // 換到棄牌堆(空): 游標歸 0
	this.Equal(1, target.curRow)
	this.Equal(0, target.curIdx)
	target.Move(game, "up")
	target.Move(game, "up") // 上端夾住
	this.Equal(0, target.curRow)

	lipgloss.SetColorProfile(termenv.ANSI) // 臨時升 profile 使樣式可見(同 TestFocusView)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	row := strings.Split((&panelPile{curIdx: 1}).View(game, 60, true), "\n")
	this.Contains(row[1], "103@結帳 "+styleCursor.Render("101@上菜")) // 游標 token 反白(堆頂在左, 索引 1 = 101)

	row = strings.Split((&panelPile{curIdx: 1}).View(game, 24, true), "\n") // 窄寬: 窗格捲到游標、左緣 < 緊接標籤後
	this.Contains(row[1], "抽牌堆(2): < ")
}
