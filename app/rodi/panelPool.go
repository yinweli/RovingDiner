package rodi

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// panelPool 場外組件(區 2; 【營業顯示規格書 | 6、畫面規格 | 6.4】): 不在座的顧客池三列——
// 排隊(隊頭在左)→ 遊蕩 → 卡牌化, 顧客識別碼以空白分隔(主畫面省實例段); 空列冒號後留空。
// 游標 = 列 x 項(上下換列、左右列內移; M26 R2), 游標態反白該 token; 游標列窗格跟游標捲
// (左緣 < 緊接標籤後), 其餘列固定窗、超寬補右緣 >。
type panelPool struct {
	curRow int // 游標列索引(0 排隊 / 1 遊蕩 / 2 卡牌化; 自持 UI 狀態)
	curIdx int // 游標列內索引(讀取時夾界)
}

// View 渲染標題列 + 3 列。
func (this *panelPool) View(game *cores.Game, width int, focus bool) string {
	return strings.Join([]string{
		panelTitle("場外", width),
		boxMark(this.poolRow(game, "排隊", game.Wait, 0, width-4, focus), width),
		boxMark(this.poolRow(game, "遊蕩", game.Roam, 1, width-4, focus), width),
		boxMark(this.poolRow(game, "卡牌化", game.Cardify, 2, width-4, focus), width),
	}, "\n")
}

// Move 游標移動: 上下換列、左右列內移(【營業顯示規格書 | 7、互動規格 | 7.2】); 夾界不迴繞。
func (this *panelPool) Move(game *cores.Game, key string) {
	if key == keyUp {
		this.curRow--
	} // if

	if key == keyDown {
		this.curRow++
	} // if

	this.curRow = clampIndex(this.curRow, 3)
	this.curIdx = moveIndex(this.curIdx, key, len([][]*cores.Guest{game.Wait, game.Roam, game.Cardify}[this.curRow]))
}

// poolRow 單列: 標籤(N): 識別碼序列(第 1 個 = 隊頭 / 首位); 聚焦且游標在本列時 token 反白 + 窗格跟游標捲。
func (this *panelPool) poolRow(game *cores.Game, label string, member []*cores.Guest, index, inner int, focus bool) string {
	token := []string{}
	size := []int{}

	for _, itor := range member {
		text := cores.IdentGuest(game.GetSheet(), itor.GetGuestID(), cores.NoneID)
		token = append(token, text)
		size = append(size, lipgloss.Width(text))
	} // for

	prefix := fmt.Sprintf("%v(%v):", label, len(member))

	if len(token) == 0 {
		return prefix
	} // if

	first := 0

	if focus && index == this.curRow {
		cursor := clampIndex(this.curIdx, len(token))
		first = stripFirst(size, 1, inner-lipgloss.Width(prefix)-1, cursor)
		token[cursor] = styleCursor.Render(token[cursor])
	} // if

	if first > 0 {
		return prefix + " < " + strings.Join(token[first:], " ")
	} // if

	return prefix + " " + strings.Join(token, " ")
}
