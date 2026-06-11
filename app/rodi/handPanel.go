package rodi

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// handPanel 手牌組件(區 5; 【營業顯示規格書 | 6、畫面規格 | 6.7】): 每卡 3 行垂直區塊橫向並排——
// 第 1 行 卡牌識別碼[卡牌化來源] (出牌費用)、第 2 行 flag A(不棄 / 封印)、第 3 行 flag B(出放 / 未放),
// 命中才顯、未命中留白; 欄寬 content-fit、欄距 2; 超寬固定窗截斷補右緣 >(M21 拍板⑦)。
// 出不起 / 封印的卡名行上暗色標記(顏色與排版正交; M22 拍板)。
type handPanel struct{}

// View 渲染標題列(手牌(N/上限)) + 3 行; 空手牌三行留白(高度穩定)。
func (this handPanel) View(world *mirror, width int) string {
	row1 := []string{}
	row2 := []string{}
	row3 := []string{}

	for _, itor := range world.zone[cores.ContainerHand] {
		view := world.card[itor]

		if view == nil {
			continue // 防禦: 容器有編號但視圖缺
		} // if

		text1 := handCard(world, view)
		text2 := handFlagA(view)
		text3 := handFlagB(view)
		size := max(lipgloss.Width(text1), lipgloss.Width(text2), lipgloss.Width(text3))
		cell1 := padTo(text1, size)

		if handDim(world, view) {
			cell1 = styleDim.Render(cell1) // 排版先完成、樣式最後上(寬度不受擾)
		} // if

		row1 = append(row1, cell1)
		row2 = append(row2, padTo(text2, size))
		row3 = append(row3, padTo(text3, size))
	} // for

	return strings.Join([]string{
		panelTitle(fmt.Sprintf("手牌(%v/%v)", len(world.zone[cores.ContainerHand]), num(world.attr["handMax"])), width),
		truncMark(strings.TrimRight(strings.Join(row1, "  "), " "), width),
		truncTo(strings.TrimRight(strings.Join(row2, "  "), " "), width),
		truncTo(strings.TrimRight(strings.Join(row3, "  "), " "), width),
	}, "\n")
}

// handCard 卡名行: 識別碼(主畫面省實例段) + cardify 來源段(已綁才有) + (出牌費用)。
func handCard(world *mirror, view *cardView) string {
	text := identCard(world.sheet, view.dataID, cores.NoneID)

	if view.bindID != 0 {
		text += "[" + identGuest(world.sheet, view.bindID, cores.NoneID) + "]"
	} // if

	return text + " (" + num(view.attr["cost"]) + ")"
}

// handDim 暗色標記判定(M22 拍板): 出不起(出牌費用 > 出牌點數)或封印(cardSeal 鎖 > 0)的卡, 卡名行轉暗。
func handDim(world *mirror, view *cardView) bool {
	return view.attr["cost"] > world.attr["energy"] || view.lock["cardSeal"] > 0
}

// handFlagA flag A 行: 不棄(keep 鎖 > 0 或已綁卡牌化來源——綁定即不棄 +1, 該鎖定變更無事件) / 封印(cardSeal 鎖 > 0);
// 命中以空白並列、全空回空字串。
func handFlagA(view *cardView) string {
	flag := []string{}

	if view.lock["keep"] > 0 || view.bindID != 0 {
		flag = append(flag, "不棄")
	} // if

	if view.lock["cardSeal"] > 0 {
		flag = append(flag, "封印")
	} // if

	return strings.Join(flag, " ")
}

// handFlagB flag B 行: 出放(playExile 鎖 > 0)/ 未放(unplayExile 鎖 > 0); 規則同 flag A。
func handFlagB(view *cardView) string {
	flag := []string{}

	if view.lock["playExile"] > 0 {
		flag = append(flag, "出放")
	} // if

	if view.lock["unplayExile"] > 0 {
		flag = append(flag, "未放")
	} // if

	return strings.Join(flag, " ")
}
