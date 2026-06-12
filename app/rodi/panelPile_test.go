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

// TestPanelPileItem 驗證游標項目: 游標堆游標位置的卡牌; 空堆回 nil。
func (this *SuitePanelPile) TestPanelPileItem() {
	game := testGame()
	game.Deck.Push(cores.NewCard(game, 101))
	target := &panelPile{}
	this.Equal(game.Deck[0], target.Item(game))

	target.Move(game, "down") // 棄牌堆空
	this.Nil(target.Item(game))
}

// TestPanelPilePick 驗證選取模式(M27 R4): 游標吸附跳到含候選堆、候選間步進略過非候選、夾界不迴繞、
// 已選 / 非候選著色路徑、非本區候選(顧客選取)照常渲染。
func (this *SuitePanelPile) TestPanelPilePick() {
	game := testGame()
	c1 := cores.NewCard(game, 101)
	c2 := cores.NewCard(game, 103)
	c3 := cores.NewCard(game, 103)
	game.Drop = cores.CardList{c1, c2, c3}
	pick := &pickState{}
	pick.start(&request{card: []*cores.Card{c1, c3}, count: 1})
	target := &panelPile{pick: pick} // 游標在抽牌堆(無候選)→ 吸附棄牌堆第一個候選
	target.View(game, 60, true)
	this.Equal(1, target.curRow)
	this.Equal(0, target.curIdx)

	target.Move(game, "right") // 下一個候選 = c3(略過 c2)
	this.Equal(2, target.curIdx)

	target.Move(game, "left") // 前一個候選 = c1
	this.Equal(0, target.curIdx)

	target.Move(game, "up") // 前端夾住
	this.Equal(0, target.curIdx)
	this.Equal(1, target.curRow)

	pick.toggle(0) // c1 已選 → 已選色底路徑(無 TTY 樣式渲原文, 內容不變)
	this.Contains(target.View(game, 60, false), "棄牌堆(3): 101@上菜")

	pick.start(&request{guest: []*cores.Guest{{}}, count: 1}) // 非本區候選: 照常渲染
	this.Equal((&panelPile{curRow: 1}).View(game, 60, false), target.View(game, 60, false))
}
