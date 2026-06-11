package rodi

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// panelAction 行動佇列組件(區 3; 【營業顯示規格書 | 6、畫面規格 | 6.5】): 先進先出、隊頭在左;
// 每項 2 行垂直區塊橫向並排——第 1 行技能識別碼(無實例段)、第 2 行顧客識別碼 + 行動類型標記(飽 / 耐);
// 欄寬 content-fit、欄距 2; 超寬固定窗截斷補右緣 >(M21 拍板⑦)。
type panelAction struct{}

// View 渲染標題列(含佇列數) + 2 行; 空佇列兩行留白(高度穩定)。
func (this panelAction) View(game *cores.Game, width int) string {
	row1 := []string{}
	row2 := []string{}

	for _, itor := range game.Action {
		text1 := cores.IdentSkill(game.GetSheet(), itor.GetSkillID())
		text2 := cores.IdentGuest(game.GetSheet(), itor.GetGuest().GetGuestID(), cores.NoneID) + " " + taskName(itor.GetKind())
		size := max(lipgloss.Width(text1), lipgloss.Width(text2))
		row1 = append(row1, padTo(text1, size))
		row2 = append(row2, padTo(text2, size))
	} // for

	return strings.Join([]string{
		panelTitle(fmt.Sprintf("行動佇列(%v)", len(game.Action)), width),
		boxMark(strings.TrimRight(strings.Join(row1, "  "), " "), width),
		boxTrunc(strings.TrimRight(strings.Join(row2, "  "), " "), width),
	}, "\n")
}

// taskName 行動類型標記(飽 / 耐; 越界顯 ?)。
func taskName(task cores.TaskKind) string {
	switch task {
	case cores.TaskSate:
		return "飽"

	case cores.TaskCalm:
		return "耐"

	default:
		return "?"
	} // switch
}
