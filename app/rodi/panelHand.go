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
// 命中才顯、未命中留白; 欄寬 content-fit、欄距 2; 超寬固定窗截斷補右緣 >(M21 拍板⑦)。
// 出不起 / 封印的卡名行上暗色標記(顏色與排版正交; M22 拍板)。
type panelHand struct{}

// View 渲染標題列(手牌(N/上限)) + 3 行; 空手牌三行留白(高度穩定)。
func (this panelHand) View(game *cores.Game, width int) string {
	row1 := []string{}
	row2 := []string{}
	row3 := []string{}

	for _, itor := range game.Hand {
		text1 := handCard(game.GetSheet(), itor)
		text2 := handFlagA(itor)
		text3 := handFlagB(itor)
		size := max(lipgloss.Width(text1), lipgloss.Width(text2), lipgloss.Width(text3))
		cell1 := padTo(text1, size)

		if handDim(game, itor) {
			cell1 = styleDim.Render(cell1) // 排版先完成、樣式最後上(寬度不受擾)
		} // if

		row1 = append(row1, cell1)
		row2 = append(row2, padTo(text2, size))
		row3 = append(row3, padTo(text3, size))
	} // for

	return strings.Join([]string{
		panelTitle(fmt.Sprintf("手牌(%v/%v)", len(game.Hand), num(game.GetHandMax().GetValue())), width),
		truncMark(strings.TrimRight(strings.Join(row1, "  "), " "), width),
		truncTo(strings.TrimRight(strings.Join(row2, "  "), " "), width),
		truncTo(strings.TrimRight(strings.Join(row3, "  "), " "), width),
	}, "\n")
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
