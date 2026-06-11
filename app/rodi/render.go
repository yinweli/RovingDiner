package rodi

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// 渲染工具(【營業顯示規格書 | 4、渲染政策：ASCII + CJK only】): 內容限 ASCII + CJK,
// 寬度一律以顯示寬計算(半形 1 格 / 全形 2 格, 經 lipgloss 寬度感知), 禁止手動空格對齊;
// 每格設寬度預算、超出截斷。各組件共用, 組件內不得自算字元數。

// padTo 右補空白至顯示寬 width; 已達寬度原樣回傳(截斷歸 truncTo, 不混責)。
func padTo(text string, width int) string {
	gap := width - lipgloss.Width(text)

	if gap <= 0 {
		return text
	} // if

	return text + strings.Repeat(" ", gap)
}

// truncTo 截斷至顯示寬 width(全形字不切半字, 容不下整字即捨去); 未超寬原樣回傳。
func truncTo(text string, width int) string {
	if lipgloss.Width(text) <= width {
		return text
	} // if

	return lipgloss.NewStyle().MaxWidth(width).Render(text)
}

// truncMark 截斷並於右緣補 > 記號(【營業顯示規格書 | 6、畫面規格 | 6.12】; M21 固定窗只出右緣 >,
// 左緣 < 與 cursor-follow 隨 M26 游標進場); 未超寬原樣回傳。
func truncMark(text string, width int) string {
	if lipgloss.Width(text) <= width {
		return text
	} // if

	if width < 3 {
		return truncTo(text, width) // 過窄(防禦) → 純截斷
	} // if

	return padTo(truncTo(text, width-2), width-1) + ">"
}

// panelTitle 面板標題列(【營業顯示規格書 | 4、渲染政策：ASCII + CJK only | 7】標題前後各留 1 空白);
// 餘寬補橫線、收尾 +(M26 R1.5 框線補齊: 收尾接點即與下方列框 / 中線的交點)、超寬截斷。
func panelTitle(title string, width int) string {
	text := "+- " + title + " "
	gap := width - 1 - lipgloss.Width(text)

	if gap > 0 {
		text += strings.Repeat("-", gap)
	} // if

	return truncTo(text, width-1) + "+"
}

// panelTitleSeam 接縫版標題列(無左端 +; 該接點由左欄各行收尾的中線字元供應——M26 R1.5 拍板
// 「中線歸左欄」, 右欄日誌專用)。
func panelTitleSeam(title string, width int) string {
	return panelTitle(title, width+1)[1:] // 首字必為 ASCII '+', 裁 1 byte 安全
}

// boxRow 帶框內容行: 內容(呼叫端已截至內容寬)右補空白至內容寬後包「| 」與「 |」——內容寬 = 區寬 - 4
// (左右框與 cell padding 各 1; 【營業顯示規格書 | 4、渲染政策：ASCII + CJK only | 6】)。
// 左欄各行收尾的框字元即左欄與日誌的共用中線(M26 R1.5 拍板「中線歸左欄」)。
func boxRow(text string, width int) string {
	return "| " + padTo(text, width-4) + " |"
}

// boxMark 帶框內容行(截斷補右緣 > 版): 超寬截至內容寬並於最後內容格補 >, padding 與框保留(即「 > |」,
// 【營業顯示規格書 | 6、畫面規格 | 6.12】貼右緣 = 緊鄰面板右框前 1 格)。
func boxMark(text string, width int) string {
	return boxRow(truncMark(text, width-4), width)
}

// boxTrunc 帶框內容行(純截斷版): 超寬截至內容寬、不補記號(非捲動行如旗標列 / 狀態列)。
func boxTrunc(text string, width int) string {
	return boxRow(truncTo(text, width-4), width)
}

// boxRowSeam 接縫版帶框內容行(無左框; 左緣由左欄中線供應, 右欄日誌專用): 左 padding 1 +
// 內容補白至內容寬(= 區寬 - 3)+ 右 padding 1 + 右框。
func boxRowSeam(text string, width int) string {
	return " " + padTo(text, width-3) + " |"
}

// num 整數屬性值轉顯示字串(屬性容器為整數, 直讀 GetValue 即 int32)。
func num(value int32) string {
	return strconv.FormatInt(int64(value), 10)
}

// numFloor 護盾 / 格擋專用: <= 0 顯 0(【營業顯示規格書 | 6、畫面規格 | 6.9】固定欄)。
func numFloor(value int32) string {
	if value <= 0 {
		return "0"
	} // if

	return num(value)
}

// alignTable 多行對齊表: 各欄寬 = 該欄各列最大顯示寬(content-fit)、相鄰欄間隔 2 空白、行尾不留補白
// (【營業顯示規格書 | 4、渲染政策：ASCII + CJK only | 8】); 各列欄數一致由呼叫端保證。
func alignTable(table [][]string) (result []string) {
	if len(table) == 0 {
		return nil
	} // if

	width := make([]int, len(table[0]))

	for _, itor := range table {
		for index, text := range itor {
			if w := lipgloss.Width(text); w > width[index] {
				width[index] = w
			} // if
		} // for
	} // for

	for _, itor := range table {
		cell := []string{}

		for index, text := range itor {
			cell = append(cell, padTo(text, width[index]))
		} // for

		result = append(result, strings.TrimRight(strings.Join(cell, "  "), " "))
	} // for

	return result
}

// alignPair 兩行對齊表(alignTable 的 2 列特化): 第 1 行標籤、第 2 行數值逐欄上下對齊;
// 兩切片等長(呼叫端保證)。
func alignPair(label, value []string) (row1, row2 string) {
	row := alignTable([][]string{label, value})
	return row[0], row[1]
}
