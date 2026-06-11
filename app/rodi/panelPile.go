package rodi

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// panelPile 牌堆組件(區 6; 【營業顯示規格書 | 6、畫面規格 | 6.8】): 抽 / 棄 / 流放三堆各一行、
// 常駐列出內容(debug viewer 定位)、堆頂在左; 卡牌識別碼以空白分隔(主畫面省實例段)。
// 游標 = 堆 x 張(上下換堆、左右堆內移; M26 R2), 游標態反白整張卡 token(【營業顯示規格書 | 7、互動規格 | 7.4】
// 粒度); 游標堆窗格跟游標捲(左緣 < 緊接標籤後), 其餘堆固定窗、超寬補右緣 >。
type panelPile struct {
	curRow int // 游標堆索引(0 抽 / 1 棄 / 2 流放; 自持 UI 狀態)
	curIdx int // 游標堆內索引(讀取時夾界)
}

// View 渲染標題列 + 3 列。
func (this *panelPile) View(game *cores.Game, width int, focus bool) string {
	return strings.Join([]string{
		panelTitle("牌堆", width),
		boxMark(this.pileRow(game, "抽牌堆", game.Deck, 0, width-4, focus), width),
		boxMark(this.pileRow(game, "棄牌堆", game.Drop, 1, width-4, focus), width),
		boxMark(this.pileRow(game, "流放堆", game.Exile, 2, width-4, focus), width),
	}, "\n")
}

// Move 游標移動: 上下換堆、左右堆內移(【營業顯示規格書 | 7、互動規格 | 7.2】); 夾界不迴繞。
func (this *panelPile) Move(game *cores.Game, key string) {
	if key == keyUp {
		this.curRow--
	} // if

	if key == keyDown {
		this.curRow++
	} // if

	this.curRow = clampIndex(this.curRow, 3)
	this.curIdx = moveIndex(this.curIdx, key, len([][]*cores.Card{game.Deck, game.Drop, game.Exile}[this.curRow]))
}

// Item 回游標堆游標位置的卡牌(空堆回 nil; 牌堆項目為單張卡, 檢視 modal 同手牌)。
func (this *panelPile) Item(game *cores.Game) any {
	member := [][]*cores.Card{game.Deck, game.Drop, game.Exile}[clampIndex(this.curRow, 3)]

	if len(member) == 0 {
		return nil
	} // if

	return member[clampIndex(this.curIdx, len(member))]
}

// pileRow 單列: 標籤(N): 識別碼序列(第 1 個 = 堆頂); 聚焦且游標在本堆時 token 反白 + 窗格跟游標捲。
func (this *panelPile) pileRow(game *cores.Game, label string, member []*cores.Card, index, inner int, focus bool) string {
	token := []string{}
	size := []int{}

	for _, itor := range member {
		text := cores.IdentCard(game.GetSheet(), itor.GetCardID(), cores.NoneID)
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
