package rodi

import (
	"fmt"
	"strings"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// panelPile 牌堆組件(區 6; 【營業顯示規格書 | 6、畫面規格 | 6.8】): 抽 / 棄 / 流放三堆各一行、
// 常駐列出內容(debug viewer 定位)、堆頂在左; 卡牌識別碼以空白分隔(主畫面省實例段);
// 超寬固定窗截斷補右緣 >(M21 拍板⑦)。
type panelPile struct{}

// View 渲染標題列 + 3 列。
func (this panelPile) View(game *cores.Game, width int) string {
	return strings.Join([]string{
		panelTitle("牌堆", width),
		truncMark(pileRow(game.GetSheet(), "抽牌堆", game.Deck), width),
		truncMark(pileRow(game.GetSheet(), "棄牌堆", game.Drop), width),
		truncMark(pileRow(game.GetSheet(), "流放堆", game.Exile), width),
	}, "\n")
}

// pileRow 單列: 標籤(N): 識別碼序列(第 1 個 = 堆頂)。
func pileRow(sheet *sheeter.Sheeter, label string, member []*cores.Card) string {
	text := fmt.Sprintf("%v(%v):", label, len(member))

	for _, itor := range member {
		text += " " + cores.IdentCard(sheet, itor.GetCardID(), cores.NoneID)
	} // for

	return text
}
