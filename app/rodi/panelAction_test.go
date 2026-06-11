package rodi

import (
	"strings"
	"testing"

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
	this.Equal("+- 行動佇列(0) "+strings.Repeat("-", 45)+"\n\n", panelAction{}.View(game, 60)) // 空佇列留白

	guest := cores.NewGuest(game, 501)
	game.Action.Push(cores.NewAction(guest, cores.TaskCalm, 301))
	game.Action.Push(cores.NewAction(guest, cores.TaskSate, 301))

	this.Equal(strings.Join([]string{
		"+- 行動佇列(2) " + strings.Repeat("-", 45),
		"301@開朗     301@開朗",
		"501@老饕 耐  501@老饕 飽",
	}, "\n"), panelAction{}.View(game, 60))

	row := strings.Split(panelAction{}.View(game, 10), "\n") // 超寬: 補右緣 >
	this.Equal("301@開朗 >", row[1])
}

// TestTaskName 驗證行動類型標記; 越界顯 ?。
func (this *SuitePanelAction) TestTaskName() {
	this.Equal("飽", taskName(cores.TaskSate))
	this.Equal("耐", taskName(cores.TaskCalm))
	this.Equal("?", taskName(cores.TaskKind(9)))
}
