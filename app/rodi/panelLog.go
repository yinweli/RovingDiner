package rodi

import (
	"strings"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// panelLog 事件日誌組件(區 8; 【營業顯示規格書 | 6、畫面規格 | 6.10】): 右欄常駐、與左欄同高。
// 行歷史自持於 journal、不讀盤面, 故簽章自立(寬度 + 高度)而非 component 契約——
// 高度歸父層 layout 分配(M22 拍板); 不橫向捲動, 超寬截斷末端補 >(完整值歸 M24 檢視 modal)。
type panelLog struct {
	journal *journal // 行合成器(行歷史 + 歸因狀態)
}

func newPanelLog(sheet *sheeter.Sheeter) *panelLog {
	return &panelLog{journal: newJournal(sheet)}
}

// Append 收一筆事件(轉寫委派 journal)。
func (this *panelLog) Append(eventData cores.EventData) {
	this.journal.Append(eventData)
}

// View 渲染標題列 + 行歷史共 height 行: 行多取尾段釘最新(與步進同一時間軸, 下一行即下一步)、
// 行少底部補空行(版面穩定); 高度耗盡(防禦)回空字串。
func (this *panelLog) View(width, height int) string {
	if height <= 0 {
		return ""
	} // if

	row := []string{panelTitle("事件日誌", width)}
	tail := this.journal.line

	if over := len(tail) - (height - 1); over > 0 {
		tail = tail[over:]
	} // if

	for _, itor := range tail {
		row = append(row, lineStyle(itor).Render(truncMark(itor, width))) // 先截斷後上色, 樣式與排版正交
	} // for

	for len(row) < height {
		row = append(row, "")
	} // for

	return strings.Join(row, "\n")
}
