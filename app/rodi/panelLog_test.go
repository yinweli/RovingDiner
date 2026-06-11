package rodi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuitePanelLog(t *testing.T) {
	suite.Run(t, new(SuitePanelLog))
}

// SuitePanelLog 驗證事件日誌組件(panelLog.go): 標題列 / 尾段釘最新 / 底部補空行 / 超寬截斷 / 高度防禦。
// 行內容形狀歸 journal_test, 本處只驗組件層的取尾與裁切。
type SuitePanelLog struct {
	suite.Suite
}

// TestPanelLogView 驗證渲染: 標題 + 行歷史、行少底部補空行、行多取尾段、超寬補右緣 >、高度耗盡回空。
func (this *SuitePanelLog) TestPanelLogView() {
	target := newPanelLog(testSheet())
	target.Append(cores.EventData{Kind: cores.EventProperty, Attr: "energy", Op: cores.AssignSub, Operand: 2, Before: 10, After: 8})
	target.Append(cores.EventData{Kind: cores.EventScope, Round: 3, Phase: cores.PhasePlayerAction, Scope: cores.ScopeTrigger, Trigger: cores.TriggerCardPlay})

	this.Equal(strings.Join([]string{ // 行少: 底部補空行至滿高
		"+- 事件日誌 " + strings.Repeat("-", 18),
		"$ 出牌點數 -= 2 >> 8",
		"[R3 玩家行動] 時機:玩家出牌",
		"",
		"",
	}, "\n"), target.View(30, 5))

	this.Equal("+- 事件日誌 "+strings.Repeat("-", 18)+"\n[R3 玩家行動] 時機:玩家出牌", target.View(30, 2)) // 行多: 取尾段釘最新

	row := strings.Split(target.View(12, 3), "\n") // 超寬: 截斷補右緣 >
	this.Equal("$ 出牌點數 >", row[1])

	this.Equal("", target.View(30, 0)) // 高度耗盡(防禦)
}
