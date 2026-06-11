package rodi

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// modal 檢視疊層契約(M26 R3; 【營業顯示規格書 | 7、互動規格 | 7.6】): Body 回標題與全部內容行(未框,
// 不裁不捲)——框線 / 寬高上界 / 置中 / 捲動與位置指示歸框架(modalView + overlay)。內容停點直讀
// (開著時引擎暫停消費, 指標穩定免失效防呆; M26 拍板); 群組分隔線以 modalGroup 前綴行表達,
// 框架代換成具名分隔線(寬度成形後才畫得出來)。
type modal interface {
	Body(game *cores.Game) (title string, row []string)
}

// modal 尺寸上界(【營業顯示規格書 | 7、互動規格 | 7.6】通則: 含框 80 欄 x 22 行; 寬可窄於上界、
// 高隨內容增長, 超過上界改捲動)。
const (
	modalWidthMax  = 80 // 寬上界(含框)
	modalHeightMax = 22 // 高上界(含框)
)

// modalGroup 群組分隔線前綴(Body 內容行以「+- 群組名」表達分隔線; 內容行不會以此開頭, 衝突免疫)。
const modalGroup = "+- "

// modalView 渲染 modal 盒: 寬 = 內容 content-fit 夾上界、高 = 內容 + 框夾上界(超出依 offset 開窗,
// 底框右下嵌位置指示——捲到底 END、其餘百分比、未溢出不顯; 【營業顯示規格書 | 7、互動規格 | 7.6】通則)。
// offset 由呼叫端先夾界(model.Update 捲動時夾)。
func modalView(game *cores.Game, top modal, offset int) []string {
	title, row := top.Body(game)
	width := lipgloss.Width(modalGroup+title+" ") + 1

	for _, itor := range row {
		need := lipgloss.Width(itor) + 4 // 內容行: 左右框與 cell padding 各 1

		if strings.HasPrefix(itor, modalGroup) {
			need = lipgloss.Width(itor+" ") + 1 // 分隔線: 名 + 留白 + 收尾 +
		} // if

		if need > width {
			width = need
		} // if
	} // for

	if width > modalWidthMax {
		width = modalWidthMax
	} // if

	visible := modalHeightMax - 2
	label := ""

	if limit := len(row) - visible; limit > 0 {
		label = num(int32(offset*100/limit)) + "%"

		if offset >= limit {
			label = "END"
		} // if
	} else {
		visible = len(row)
	} // if

	box := []string{panelTitle(title, width)}

	for _, itor := range row[offset : offset+visible] {
		if strings.HasPrefix(itor, modalGroup) {
			box = append(box, panelTitle(strings.TrimPrefix(itor, modalGroup), width))
			continue
		} // if

		box = append(box, boxTrunc(itor, width))
	} // for

	return append(box, modalBottom(width, label))
}

// modalBottom modal 底框列: 無捲動素線收尾; 有捲動於右下嵌位置指示(仿 less; 格式「- 標示 -+」貼右緣)。
func modalBottom(width int, label string) string {
	if label == "" {
		return "+" + strings.Repeat("-", width-2) + "+"
	} // if

	return "+" + strings.Repeat("-", width-5-lipgloss.Width(label)) + " " + label + " -+"
}

// overlay 置中疊層: 把盒列疊在底圖正中(相對整個畫面; 【營業顯示規格書 | 7、互動規格 | 7.6】通則),
// 底圖被遮列以 ANSI 感知裁切保留左右兩側可見(樣式碼不算寬度、不被切壞)。
func overlay(base string, box []string, width, height int) string {
	row := strings.Split(base, "\n")
	x := (width - lipgloss.Width(box[0])) / 2
	y := (height - len(box)) / 2

	for index, itor := range box {
		at := y + index

		if at < 0 || at >= len(row) {
			continue // 過高(防禦): 超出底圖的盒列捨去
		} // if

		row[at] = padTo(ansi.Truncate(row[at], x, ""), x) + itor + ansi.TruncateLeft(row[at], x+lipgloss.Width(itor), "")
	} // for

	return strings.Join(row, "\n")
}
