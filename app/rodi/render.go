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
// 餘寬補橫線、超寬截斷。
func panelTitle(title string, width int) string {
	text := "+- " + title + " "
	gap := width - lipgloss.Width(text)

	if gap > 0 {
		text += strings.Repeat("-", gap)
	} // if

	return truncTo(text, width)
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
