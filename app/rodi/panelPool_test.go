package rodi

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuitePanelPool(t *testing.T) {
	suite.Run(t, new(SuitePanelPool))
}

// SuitePanelPool 驗證場外組件(panelPool.go): 三列格式 / 空列留白 / 截斷記號。
type SuitePanelPool struct {
	suite.Suite
}

// TestPanelPoolView 驗證渲染: 排隊 / 遊蕩 / 卡牌化三列(隊頭在左)、空列冒號後留空、超寬補右緣 >。
func (this *SuitePanelPool) TestPanelPoolView() {
	game := testGame()
	game.Wait.Insert(cores.NewGuest(game, 501))
	game.Wait.Insert(cores.NewGuest(game, 501))
	game.Cardify.Push(cores.NewGuest(game, 501))

	this.Equal(strings.Join([]string{
		"+- 場外 " + strings.Repeat("-", 51) + "+",
		"| " + padTo("排隊(2): 501@老饕 501@老饕", 56) + " |",
		"| " + padTo("遊蕩(0):", 56) + " |",
		"| " + padTo("卡牌化(1): 501@老饕", 56) + " |",
	}, "\n"), (&panelPool{}).View(game, 60, false))

	row := strings.Split((&panelPool{}).View(game, 12, false), "\n") // 超寬: 內容寬 8、> 站最後內容格
	this.Equal("| 排隊(2 > |", row[1])
}

// TestPanelPoolMove 驗證游標移動: 上下換列(換到空列游標歸 0)、左右列內移、夾界不迴繞;
// 聚焦時游標 token 反白、游標列窗格跟游標捲(左緣 < 緊接標籤後)。
func (this *SuitePanelPool) TestPanelPoolMove() {
	game := testGame()
	game.Wait.Insert(cores.NewGuest(game, 501))
	game.Wait.Insert(cores.NewGuest(game, 501))
	target := &panelPool{}
	target.Move(game, "right")
	this.Equal(1, target.curIdx)
	target.Move(game, "down") // 換到空列: 游標歸 0
	this.Equal(1, target.curRow)
	this.Equal(0, target.curIdx)
	target.Move(game, "down")
	target.Move(game, "down") // 下端夾住
	this.Equal(2, target.curRow)
	target.Move(game, "up")
	target.Move(game, "up")
	target.Move(game, "up") // 上端夾住
	this.Equal(0, target.curRow)

	lipgloss.SetColorProfile(termenv.ANSI) // 臨時升 profile 使樣式可見(同 TestFocusView)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	target = &panelPool{curIdx: 1}
	row := strings.Split(target.View(game, 60, true), "\n")
	this.Contains(row[1], "501@老饕 "+styleCursor.Render("501@老饕")) // 游標 token 反白

	row = strings.Split(target.View(game, 26, true), "\n") // 窄寬: 窗格跟游標捲、左緣 < 緊接標籤後
	this.Contains(row[1], "排隊(2): < "+styleCursor.Render("501@老饕"))
}
