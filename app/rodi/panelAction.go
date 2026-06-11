package rodi

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// panelAction 行動佇列組件(區 3; 【營業顯示規格書 | 6、畫面規格 | 6.5】): 先進先出、隊頭在左;
// 每項 2 行垂直區塊橫向並排——第 1 行技能識別碼(無實例段)、第 2 行顧客識別碼 + 行動類型標記(飽 / 耐);
// 欄寬 content-fit、欄距 2。游標單列左右移(M26 R2), 游標態反白整個項目區塊(2 行);
// 窗格跟游標捲、左緣 < 兩行同縮排、超寬補右緣 >。
type panelAction struct {
	cursor int // 游標索引(自持 UI 狀態; 讀取時夾界)
}

// View 渲染標題列(含佇列數) + 2 行; 空佇列兩行留白(高度穩定)。
func (this *panelAction) View(game *cores.Game, width int, focus bool) string {
	cursor := clampIndex(this.cursor, len(game.Action))
	row1 := []string{}
	row2 := []string{}
	size := []int{}

	for index, itor := range game.Action {
		text1 := cores.IdentSkill(game.GetSheet(), itor.GetSkillID())
		text2 := cores.IdentGuest(game.GetSheet(), itor.GetGuest().GetGuestID(), cores.NoneID) + " " + taskName(itor.GetKind())
		w := max(lipgloss.Width(text1), lipgloss.Width(text2))
		cell1 := padTo(text1, w)
		cell2 := padTo(text2, w)

		if focus && index == cursor {
			cell1 = styleCursor.Render(cell1)
			cell2 = styleCursor.Render(cell2)
		} // if

		row1 = append(row1, cell1)
		row2 = append(row2, cell2)
		size = append(size, w)
	} // for

	first := stripFirst(size, 2, width-4, cursor)
	head1, head2 := "", ""

	if first > 0 {
		head1, head2 = markHead, markIndent // 左緣記號佔位: 兩行同縮排, 項目區塊上下對齊
	} // if

	return strings.Join([]string{
		panelTitle(fmt.Sprintf("行動佇列(%v)", len(game.Action)), width),
		boxMark(head1+strings.Join(row1[first:], "  "), width),
		boxTrunc(head2+strings.Join(row2[first:], "  "), width),
	}, "\n")
}

// Move 游標移動: 單列左右(【營業顯示規格書 | 7、互動規格 | 7.2】); 夾界不迴繞。
func (this *panelAction) Move(game *cores.Game, key string) {
	this.cursor = moveIndex(this.cursor, key, len(game.Action))
}

// Item 回游標下的行動項(空佇列回 nil; 檢視 modal 用)。
func (this *panelAction) Item(game *cores.Game) any {
	if len(game.Action) == 0 {
		return nil
	} // if

	return game.Action[clampIndex(this.cursor, len(game.Action))]
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
