package rodi

import (
	"fmt"
	"strings"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// pilePanel 牌堆組件(區 6; 【營業顯示規格書 | 6、畫面規格 | 6.8】): 抽 / 棄 / 流放三堆各一行、
// 常駐列出內容(debug viewer 定位)、堆頂在左; 卡牌識別碼以空白分隔(主畫面省實例段);
// 超寬固定窗截斷補右緣 >(M21 拍板⑦)。
type pilePanel struct{}

// View 渲染標題列 + 3 列。
func (this pilePanel) View(world *mirror, width int) string {
	return strings.Join([]string{
		panelTitle("牌堆", width),
		truncMark(pileRow(world, "抽牌堆", cores.ContainerDeck), width),
		truncMark(pileRow(world, "棄牌堆", cores.ContainerDrop), width),
		truncMark(pileRow(world, "流放堆", cores.ContainerExile), width),
	}, "\n")
}

// pileRow 單列: 標籤(N): 識別碼序列(第 1 個 = 堆頂)。
func pileRow(world *mirror, label string, kind cores.ContainerKind) string {
	member := world.zone[kind]
	text := fmt.Sprintf("%v(%v):", label, len(member))

	for _, itor := range member {
		if view := world.card[itor]; view != nil {
			text += " " + identCard(world.sheet, view.dataID, cores.NoneID)
		} // if
	} // for

	return text
}
