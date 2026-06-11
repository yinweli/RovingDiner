package rodi

import (
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

// alignPair 兩行對齊表: 第 1 行標籤、第 2 行數值逐欄上下對齊, 欄寬 = 該欄標籤 / 數值較寬者(content-fit),
// 相鄰欄間隔 2 空白(【營業顯示規格書 | 4、渲染政策：ASCII + CJK only | 8】); 兩切片等長(呼叫端保證)。
func alignPair(label, value []string) (row1, row2 string) {
	cell1 := []string{}
	cell2 := []string{}

	for itor := range label {
		width := lipgloss.Width(label[itor])

		if w := lipgloss.Width(value[itor]); w > width {
			width = w
		} // if

		cell1 = append(cell1, padTo(label[itor], width))
		cell2 = append(cell2, padTo(value[itor], width))
	} // for

	return strings.TrimRight(strings.Join(cell1, "  "), " "), strings.TrimRight(strings.Join(cell2, "  "), " ")
}
