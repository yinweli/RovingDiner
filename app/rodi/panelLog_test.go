package rodi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuitePanelLog(t *testing.T) {
	suite.Run(t, new(SuitePanelLog))
}

// SuitePanelLog 驗證事件日誌組件(panelLog.go): 標題列 / 尾段釘最新 / 底部補空行 / 超寬截斷 / 高度防禦。
// 行內容由引擎組畢(M24 換軌), 本處只驗組件層的收行 / 取尾與裁切。
type SuitePanelLog struct {
	suite.Suite
}

// TestPanelLogAppend 驗證收行: 一拍多行原樣攤平入歷史。
func (this *SuitePanelLog) TestPanelLogAppend() {
	target := newPanelLog()
	target.Append("[R3 玩家行動] 玩家出牌", "* 101@上菜#1")
	target.Append("$ 出牌點數 -= 2 >> 8")
	this.Equal([]string{"[R3 玩家行動] 玩家出牌", "* 101@上菜#1", "$ 出牌點數 -= 2 >> 8"}, target.line)
}

// TestPanelLogView 驗證渲染: 標題 + 行歷史、行少底部補空行、行多取尾段、超寬補右緣 >、高度耗盡回空。
func (this *SuitePanelLog) TestPanelLogView() {
	target := newPanelLog()
	target.Append("$ 出牌點數 -= 2 >> 8")
	target.Append("[R3 玩家行動] 時機:玩家出牌")

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
