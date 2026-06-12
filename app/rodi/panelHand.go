package rodi

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// panelHand 手牌組件(區 5; 【營業顯示規格書 | 6、畫面規格 | 6.7】): 每卡 3 行垂直區塊橫向並排——
// 第 1 行 卡牌識別碼[卡牌化來源] (出牌費用)、第 2 行 flag A(不棄 / 封印)、第 3 行 flag B(出放 / 未放),
// 命中才顯、未命中留白; 欄寬 content-fit、欄距 2。出不起 / 封印的卡名行上暗色標記(顏色與排版正交; M22 拍板)。
// 游標單列左右移(M26 R2), 游標態反白整張卡區塊(3 行; 【營業顯示規格書 | 7、互動規格 | 7.4】粒度);
// 窗格跟游標捲、左緣 < 三行同縮排、超寬補右緣 >。
// 選取模式且候選在手牌(M27 R4): 非候選整卡變暗(取代出不起暗標——候選可選與否以候選身分為準)、
// 已選選取色底、游標只在候選間吸附步進、標題加註 (請選卡牌); 出牌等待標題加註 (請出牌)。
type panelHand struct {
	cursor int        // 游標索引(自持 UI 狀態; 讀取時夾界)
	pick   *pickState // 選取模式共享狀態(newModel 注入, 唯讀)
}

// View 渲染標題列(手牌(N/上限)) + 3 行; 空手牌三行留白(高度穩定)。
func (this *panelHand) View(game *cores.Game, width int, focus bool) string {
	picking := this.pickSnap(game)
	cursor := clampIndex(this.cursor, len(game.Hand))
	row1 := []string{}
	row2 := []string{}
	row3 := []string{}
	size := []int{}

	for index, itor := range game.Hand {
		text1 := handCard(game.GetSheet(), itor)
		text2 := handFlagA(itor)
		text3 := handFlagB(itor)
		w := max(lipgloss.Width(text1), lipgloss.Width(text2), lipgloss.Width(text3))
		cell1 := padTo(text1, w)
		cell2 := padTo(text2, w)
		cell3 := padTo(text3, w)

		switch {
		case picking: // 三視覺態: 非候選整卡變暗、已選選取色底(先著色後游標, 樣式不動寬度)
			if at := this.pick.cardIndex(itor); at < 0 {
				cell1 = styleDim.Render(cell1)
				cell2 = styleDim.Render(cell2)
				cell3 = styleDim.Render(cell3)
			} else if this.pick.chosen(at) {
				cell1 = styleChosen.Render(cell1)
				cell2 = styleChosen.Render(cell2)
				cell3 = styleChosen.Render(cell3)
			} // if

		case handDim(game, itor):
			cell1 = styleDim.Render(cell1) // 排版先完成、樣式最後上(寬度不受擾)
		} // switch

		if focus && index == cursor {
			cell1 = styleCursor.Render(cell1)
			cell2 = styleCursor.Render(cell2)
			cell3 = styleCursor.Render(cell3)
		} // if

		row1 = append(row1, cell1)
		row2 = append(row2, cell2)
		row3 = append(row3, cell3)
		size = append(size, w)
	} // for

	first := stripFirst(size, 2, width-4, cursor)
	head1, head2 := "", ""

	if first > 0 {
		head1, head2 = markHead, markIndent // 左緣記號佔位: 三行同縮排, 卡區塊上下對齊
	} // if

	title := fmt.Sprintf("手牌(%v/%v)", len(game.Hand), num(game.GetHandMax().GetValue()))

	switch { // 標題提醒: 候選 / 出牌等待在本區, Tab 切走仍見等待落點
	case picking:
		title += noteCard

	case this.pick.playing():
		title += notePlay
	} // switch

	return strings.Join([]string{
		panelTitle(title, width),
		boxMark(head1+strings.Join(row1[first:], "  "), width),
		boxTrunc(head2+strings.Join(row2[first:], "  "), width),
		boxTrunc(head2+strings.Join(row3[first:], "  "), width),
	}, "\n")
}

// Move 游標移動: 單列左右(【營業顯示規格書 | 7、互動規格 | 7.2】); 夾界不迴繞。
// 選取模式: 方向語意保留(左右跳下一個候選、自動略過非候選, 上下不動作; 【7.4】)。
func (this *panelHand) Move(game *cores.Game, key string) {
	if this.pickSnap(game) {
		spot := this.pickSpot(game)
		at := this.pickAt(game)

		if key == keyLeft && at > 0 {
			this.cursor = spot[at-1]
		} // if

		if key == keyRight && at < len(spot)-1 {
			this.cursor = spot[at+1]
		} // if

		return
	} // if

	this.cursor = moveIndex(this.cursor, key, len(game.Hand))
}

// Item 回游標下的卡牌(空手牌回 nil; 檢視 modal 用)。
func (this *panelHand) Item(game *cores.Game) any {
	if len(game.Hand) == 0 {
		return nil
	} // if

	return game.Hand[clampIndex(this.cursor, len(game.Hand))]
}

// pickSnap 選取模式吸附判定: 本區含候選時若游標不在候選上, 吸到第一個候選; 回報本區是否處於
// 選取著色狀態(無候選即非本區選取, 照常渲染; 規則同 panelSeat)。
func (this *panelHand) pickSnap(game *cores.Game) bool {
	if this.pick.active() == false {
		return false
	} // if

	spot := this.pickSpot(game)

	if len(spot) == 0 {
		return false
	} // if

	if this.pickAt(game) < 0 {
		this.cursor = spot[0]
	} // if

	return true
}

// pickSpot 候選索引序列(手牌序; 候選身分 = 實例指標)。
func (this *panelHand) pickSpot(game *cores.Game) (result []int) {
	for index, itor := range game.Hand {
		if this.pick.cardIndex(itor) >= 0 {
			result = append(result, index)
		} // if
	} // for

	return result
}

// pickAt 游標在候選索引序列的索引(不在候選上回 -1)。
func (this *panelHand) pickAt(game *cores.Game) int {
	for index, itor := range this.pickSpot(game) {
		if itor == this.cursor {
			return index
		} // if
	} // for

	return -1
}

// handCard 卡名行: 識別碼(主畫面省實例段) + cardify 來源段(已綁才有) + (出牌費用)。
func handCard(sheet *sheeter.Sheeter, card *cores.Card) string {
	text := cores.IdentCard(sheet, card.GetCardID(), cores.NoneID)

	if bind := card.GetCardify(); bind != nil {
		text += "[" + cores.IdentGuest(sheet, bind.GetGuestID(), cores.NoneID) + "]"
	} // if

	return text + " (" + num(card.GetCost().GetValue()) + ")"
}

// handDim 暗色標記判定(M22 拍板): 出不起(出牌費用 > 出牌點數)或封印(cardSeal 鎖定中)的卡, 卡名行轉暗。
func handDim(game *cores.Game, card *cores.Card) bool {
	return card.GetCost().GetValue() > game.GetEnergy().GetValue() || card.GetSeal().IsLock()
}

// handFlagA flag A 行: 不棄(keep 鎖定中; 綁卡牌化來源的 +1 由 CardifyBind 入鎖, 直讀即涵蓋)/
// 封印(cardSeal 鎖定中); 命中以空白並列、全空回空字串。
func handFlagA(card *cores.Card) string {
	flag := []string{}

	if card.GetKeep().IsLock() {
		flag = append(flag, "不棄")
	} // if

	if card.GetSeal().IsLock() {
		flag = append(flag, "封印")
	} // if

	return strings.Join(flag, " ")
}

// handFlagB flag B 行: 出放(playExile 鎖定中)/ 未放(unplayExile 鎖定中); 規則同 flag A。
func handFlagB(card *cores.Card) string {
	flag := []string{}

	if card.GetPlayExile().IsLock() {
		flag = append(flag, "出放")
	} // if

	if card.GetUnplayExile().IsLock() {
		flag = append(flag, "未放")
	} // if

	return strings.Join(flag, " ")
}
