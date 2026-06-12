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
// 選取模式且候選在牌堆(M27 R4): 全區非候選 token 變暗(候選堆的落選卡與其餘堆一視同仁)、
// 已選選取色底、游標只在候選間吸附步進。
type panelPile struct {
	curRow int        // 游標堆索引(0 抽 / 1 棄 / 2 流放; 自持 UI 狀態)
	curIdx int        // 游標堆內索引(讀取時夾界)
	pick   *pickState // 選取模式共享狀態(newModel 注入, 唯讀)
}

// View 渲染標題列 + 3 列。
func (this *panelPile) View(game *cores.Game, width int, focus bool) string {
	picking := this.pickSnap(game)
	return strings.Join([]string{
		panelTitle("牌堆", width),
		boxMark(this.pileRow(game, "抽牌堆", game.Deck, 0, width-4, focus, picking), width),
		boxMark(this.pileRow(game, "棄牌堆", game.Drop, 1, width-4, focus, picking), width),
		boxMark(this.pileRow(game, "流放堆", game.Exile, 2, width-4, focus, picking), width),
	}, "\n")
}

// Move 游標移動: 上下換堆、左右堆內移(【營業顯示規格書 | 7、互動規格 | 7.2】); 夾界不迴繞。
// 選取模式: 方向語意保留, 沿方向掃描至下一個候選、無則不動(【7.4】自動略過非候選)。
func (this *panelPile) Move(game *cores.Game, key string) {
	if this.pickSnap(game) {
		this.pickMove(game, key)
		return
	} // if

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

// pickSnap 選取模式吸附判定: 本區含候選時若游標不在候選上, 吸到第一個候選; 回報本區是否處於
// 選取著色狀態(無候選即非本區選取, 照常渲染; 規則同 panelSeat)。
func (this *panelPile) pickSnap(game *cores.Game) bool {
	if this.pick.active() == false {
		return false
	} // if

	spot := this.pickSpot(game)

	if len(spot) == 0 {
		return false
	} // if

	if this.pickAt(game) < 0 {
		this.curRow, this.curIdx = spot[0][0], spot[0][1]
	} // if

	return true
}

// pickSpot 候選位置序列(堆索引 x 堆內索引, 堆序為主序; 候選身分 = 實例指標)。
func (this *panelPile) pickSpot(game *cores.Game) (result [][2]int) {
	for row, member := range [][]*cores.Card{game.Deck, game.Drop, game.Exile} {
		for index, itor := range member {
			if this.pick.cardIndex(itor) >= 0 {
				result = append(result, [2]int{row, index})
			} // if
		} // for
	} // for

	return result
}

// pickAt 游標在候選位置序列的索引(不在候選上回 -1)。
func (this *panelPile) pickAt(game *cores.Game) int {
	for index, itor := range this.pickSpot(game) {
		if itor[0] == clampIndex(this.curRow, 3) && itor[1] == this.curIdx {
			return index
		} // if
	} // for

	return -1
}

// pickMove 候選間方向移動: 左右沿堆內序掃描、上下沿堆序掃描(跨堆落點 = 該堆第一個候選),
// 掃到候選即停、掃不到不動(方向語意與非選取模式一致, 僅略過非候選)。
func (this *panelPile) pickMove(game *cores.Game, key string) {
	member := [][]*cores.Card{game.Deck, game.Drop, game.Exile}
	row := clampIndex(this.curRow, 3)

	switch key {
	case keyLeft:
		for i := this.curIdx - 1; i >= 0; i-- {
			if this.pick.cardIndex(member[row][i]) >= 0 {
				this.curIdx = i
				return
			} // if
		} // for

	case keyRight:
		for i := this.curIdx + 1; i < len(member[row]); i++ {
			if this.pick.cardIndex(member[row][i]) >= 0 {
				this.curIdx = i
				return
			} // if
		} // for

	case keyUp:
		for r := row - 1; r >= 0; r-- {
			if at := this.rowFirst(member[r]); at >= 0 {
				this.curRow, this.curIdx = r, at
				return
			} // if
		} // for

	case keyDown:
		for r := row + 1; r < 3; r++ {
			if at := this.rowFirst(member[r]); at >= 0 {
				this.curRow, this.curIdx = r, at
				return
			} // if
		} // for
	} // switch
}

// rowFirst 單堆第一個候選索引(無候選回 -1)。
func (this *panelPile) rowFirst(member []*cores.Card) int {
	for index, itor := range member {
		if this.pick.cardIndex(itor) >= 0 {
			return index
		} // if
	} // for

	return -1
}

// pileRow 單列: 標籤(N): 識別碼序列(第 1 個 = 堆頂); 聚焦且游標在本堆時 token 反白 + 窗格跟游標捲;
// 選取著色狀態時非候選 token 變暗、已選選取色底(先著色後游標, 樣式不動寬度)。
func (this *panelPile) pileRow(game *cores.Game, label string, member []*cores.Card, index, inner int, focus, picking bool) string {
	token := []string{}
	size := []int{}

	for _, itor := range member {
		text := cores.IdentCard(game.GetSheet(), itor.GetCardID(), cores.NoneID)

		if picking {
			if at := this.pick.cardIndex(itor); at < 0 {
				text = styleDim.Render(text)
			} else if this.pick.chosen(at) {
				text = styleChosen.Render(text)
			} // if
		} // if

		token = append(token, text)
		size = append(size, lipgloss.Width(cores.IdentCard(game.GetSheet(), itor.GetCardID(), cores.NoneID)))
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
