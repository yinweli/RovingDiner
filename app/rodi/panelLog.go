package rodi

import (
	"strings"
)

// panelLog 事件日誌組件(區 8; 【營業顯示規格書 | 6、畫面規格 | 6.10】): 右欄常駐、與左欄同高。
// 行歷史自持(引擎直出最終行, 本組件零合成零解析; M24 換軌)、不讀盤面, 故簽章自立(寬度 + 高度)
// 而非 component 契約——高度歸父層 layout 分配(M22 拍板); 不橫向捲動, 超寬截斷末端補 >
// (完整值歸 M26 檢視 modal)。框線用接縫版(無左框, 左緣由左欄中線供應——M26 R1.5 拍板「中線歸左欄」)。
// 游標 = viewport 上下捲動(非逐項; 【營業顯示規格書 | 7、互動規格 | 7.2】, M26 R2)。
type panelLog struct {
	line   []string // 全量行歷史(渲染端取 viewport 窗)
	offset int      // viewport 自底端的上捲偏移(0 = 釘最新; View 依可視高自校正)
}

func newPanelLog() *panelLog {
	return &panelLog{}
}

// Append 收一拍行組(引擎組畢的最終行, 原樣入歷史); 不動偏移(偏移 0 時自然跟最新)。
func (this *panelLog) Append(line ...string) {
	this.line = append(this.line, line...)
}

// Move viewport 捲動: 上 +1 / 下 -1(下限 0 = 釘最新; 上限依可視高, 留待 View 自校正——
// Move 不知高度, 偏移過衝由 View 寫回修正, 不產生死按鍵)。
func (this *panelLog) Move(key string) {
	if key == keyUp {
		this.offset++
	} // if

	if key == keyDown {
		this.offset--
	} // if

	if this.offset < 0 {
		this.offset = 0
	} // if
}

// View 渲染標題列 + 行歷史共 height 行: 取 viewport 窗(偏移 0 釘最新, 與步進同一時間軸)、
// 行少底部補帶框空行(版面與右框穩定); 高度耗盡(防禦)回空字串。
func (this *panelLog) View(width, height int) string {
	if height <= 0 {
		return ""
	} // if

	visible := height - 1
	limit := len(this.line) - visible

	if limit < 0 {
		limit = 0
	} // if

	if this.offset > limit {
		this.offset = limit // 偏移過衝拉回上限, 上捲到頂即停(行數 / 高度變動後的自校正)
	} // if

	row := []string{panelTitleSeam("事件日誌", width)}
	tail := this.line

	if over := len(tail) - visible - this.offset; over >= 0 {
		tail = tail[over : len(tail)-this.offset]
	} // if

	for _, itor := range tail {
		row = append(row, boxRowSeam(lineStyle(itor).Render(truncMark(itor, width-3)), width)) // 先截斷後上色, 樣式與排版正交
	} // for

	for len(row) < height {
		row = append(row, boxRowSeam("", width))
	} // for

	return strings.Join(row, "\n")
}
