package rodi

import (
	"strings"
	"testing"

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
		"+- 牌堆 " + strings.Repeat("-", 52),
		"抽牌堆(2): 103@結帳 101@上菜",
		"棄牌堆(1): 101@上菜",
		"流放堆(0):",
	}, "\n"), panelPile{}.View(game, 60))

	row := strings.Split(panelPile{}.View(game, 14), "\n") // 超寬: 補右緣 >
	this.Equal("抽牌堆(2): 1 >", row[1])
}
