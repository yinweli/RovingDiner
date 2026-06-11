package rodi

import (
	"strings"
	"testing"

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
	}, "\n"), panelPool{}.View(game, 60))

	row := strings.Split(panelPool{}.View(game, 12), "\n") // 超寬: 內容寬 8、> 站最後內容格
	this.Equal("| 排隊(2 > |", row[1])
}
