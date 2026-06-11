package rodi

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// 樣式層(M22 上色): 顏色與排版正交——排版(寬度 / 截斷 / 對齊)先完成、樣式最後上, 不影響任何字元寬度。
// 配色用 ANSI 基本色 + Faint(授權實作自定、不進規格; M22 拍板); 無 TTY 時(測試)termenv 降 Ascii、
// 樣式渲染為原文, 釘字串測試不受擾。

// 日誌行角色樣式(行角色看首字; 【營業顯示規格書 | 6、畫面規格 | 6.10】)與手牌暗色標記。
var (
	styleTitle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // 範圍標題 [: cyan
	styleOperand = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // 標題操作元 *: yellow
	styleEffect  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // 效果 -: green
	styleFlow    = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // 流程直屬 $: magenta
	styleField   = lipgloss.NewStyle()                                 // 效果欄位(欄 2): 原色
	styleDim     = lipgloss.NewStyle().Faint(true)                     // 手牌暗色標記(出不起 / 封印)
)

// lineStyle 依日誌行角色取樣式(首字即角色, 欄 2 為效果欄位)。
func lineStyle(line string) lipgloss.Style {
	switch {
	case strings.HasPrefix(line, "["):
		return styleTitle

	case strings.HasPrefix(line, "* "):
		return styleOperand

	case strings.HasPrefix(line, "- "):
		return styleEffect

	case strings.HasPrefix(line, "$ "):
		return styleFlow

	default:
		return styleField
	} // switch
}
