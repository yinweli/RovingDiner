package rodi

import (
	"fmt"
	"strings"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// panelPool 場外組件(區 2; 【營業顯示規格書 | 6、畫面規格 | 6.4】): 不在座的顧客池三列——
// 排隊(隊頭在左)→ 遊蕩 → 卡牌化, 顧客識別碼以空白分隔(主畫面省實例段); 空列冒號後留空;
// 超寬固定窗截斷補右緣 >(M21 拍板⑦)。
type panelPool struct{}

// View 渲染標題列 + 3 列。
func (this panelPool) View(game *cores.Game, width int) string {
	return strings.Join([]string{
		panelTitle("場外", width),
		truncMark(poolRow(game.GetSheet(), "排隊", game.Wait), width),
		truncMark(poolRow(game.GetSheet(), "遊蕩", game.Roam), width),
		truncMark(poolRow(game.GetSheet(), "卡牌化", game.Cardify), width),
	}, "\n")
}

// poolRow 單列: 標籤(N): 識別碼序列(第 1 個 = 隊頭 / 首位)。
func poolRow(sheet *sheeter.Sheeter, label string, member []*cores.Guest) string {
	text := fmt.Sprintf("%v(%v):", label, len(member))

	for _, itor := range member {
		text += " " + cores.IdentGuest(sheet, itor.GetGuestID(), cores.NoneID)
	} // for

	return text
}
