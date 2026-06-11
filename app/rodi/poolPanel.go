package rodi

import (
	"fmt"
	"strings"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// poolPanel 場外組件(區 2; 【營業顯示規格書 | 6、畫面規格 | 6.4】): 不在座的顧客池三列——
// 排隊(隊頭在左)→ 遊蕩 → 卡牌化, 顧客識別碼以空白分隔(主畫面省實例段); 空列冒號後留空;
// 超寬固定窗截斷補右緣 >(M21 拍板⑦)。
type poolPanel struct{}

// View 渲染標題列 + 3 列。
func (this poolPanel) View(world *mirror, width int) string {
	return strings.Join([]string{
		panelTitle("場外", width),
		truncMark(poolRow(world, "排隊", cores.ContainerWait), width),
		truncMark(poolRow(world, "遊蕩", cores.ContainerRoam), width),
		truncMark(poolRow(world, "卡牌化", cores.ContainerCardify), width),
	}, "\n")
}

// poolRow 單列: 標籤(N): 識別碼序列(第 1 個 = 隊頭 / 首位)。
func poolRow(world *mirror, label string, kind cores.ContainerKind) string {
	member := world.zone[kind]
	text := fmt.Sprintf("%v(%v):", label, len(member))

	for _, itor := range member {
		if view := world.guest[itor]; view != nil {
			text += " " + identGuest(world.sheet, view.dataID, cores.NoneID)
		} // if
	} // for

	return text
}
