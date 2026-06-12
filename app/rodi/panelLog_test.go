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

// TestPanelLogMove 驗證 viewport 捲動: 上捲 + / 下捲 -、下限 0 釘最新、上限過衝由 View 自校正(無死按鍵)。
func (this *SuitePanelLog) TestPanelLogMove() {
	target := newPanelLog()
	target.Append("a", "b", "c")
	target.Move("down")
	this.Equal(0, target.offset) // 下限夾住(釘最新)
	target.Move("up")
	target.Move("up")
	target.Move("up")
	this.Equal(3, target.offset) // Move 不知可視高, 先收著

	row := strings.Split(target.View(30, 3), "\n") // 可視 2 行: 上限 = 3 - 2 = 1, 過衝拉回
	this.Equal(1, target.offset)
	this.Contains(row[1], "a")
	this.Contains(row[2], "b")

	target.Move("down")
	row = strings.Split(target.View(30, 3), "\n") // 回釘最新
	this.Contains(row[1], "b")
	this.Contains(row[2], "c")
}

// TestPanelLogView 驗證渲染: 標題 + 行歷史、行少底部補空行、行多取尾段、超寬補右緣 >、高度耗盡回空。
func (this *SuitePanelLog) TestPanelLogView() {
	target := newPanelLog()
	target.Append("$ 出牌點數 -= 2 >> 8")
	target.Append("[R3 玩家行動] 時機:玩家出牌")

	this.Equal(strings.Join([]string{ // 行少時底部補帶框空行至滿高, 接縫版無左框(左緣由左欄中線供應)
		"- 事件日誌 " + strings.Repeat("-", 18) + "+",
		" " + padTo("$ 出牌點數 -= 2 >> 8", 27) + " |",
		" " + padTo("[R3 玩家行動] 時機:玩家出牌", 27) + " |",
		" " + padTo("", 27) + " |",
		" " + padTo("", 27) + " |",
	}, "\n"), target.View(30, 5))

	this.Equal("- 事件日誌 "+strings.Repeat("-", 18)+"+\n"+" "+padTo("[R3 玩家行動] 時機:玩家出牌", 27)+" |", target.View(30, 2)) // 行多: 取尾段釘最新

	row := strings.Split(target.View(12, 3), "\n") // 超寬: 內容寬 9、> 站最後內容格
	this.Equal(" $ 出牌  > |", row[1])

	this.Equal("", target.View(30, 0)) // 高度耗盡(防禦)
}
