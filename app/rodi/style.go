package rodi

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// 樣式層(M22 上色): 顏色與排版正交——排版(寬度 / 截斷 / 對齊)先完成、樣式最後上, 不影響任何字元寬度。
// 配色用 ANSI 基本色 + Faint(授權實作自定、不進規格; M22 拍板); 無 TTY 時(測試)termenv 降 Ascii、
// 樣式渲染為原文, 釘字串測試不受擾。

// 日誌行角色樣式(行角色看首字; 【營業顯示規格書 | 6、畫面規格 | 6.10】)、結構樣式(框線退後 / 標題)、
// 狀態語意色(負面紅 / 保護青 / 正面綠; M27 全畫面配色拍板)與聚焦 / 游標 / 選取態。
var (
	styleTitle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // 範圍標題 [ 與面板 / modal 標題: cyan(同色跨日誌與盤面)
	styleOperand = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // 標題操作元 *: yellow
	styleEffect  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // 效果 -: green
	styleFlow    = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // 流程直屬 $: magenta
	styleField   = lipgloss.NewStyle()                                 // 效果欄位(欄 2): 原色
	styleLine    = lipgloss.NewStyle().Faint(true)                     // 框線 / 標題列橫線退後(內容浮起; M27 拍板)
	styleDim     = lipgloss.NewStyle().Faint(true)                     // 手牌暗色標記(出不起 / 封印)+ 選取模式非候選態(M27 R4)
	styleFocus   = lipgloss.NewStyle().Reverse(true)                   // 聚焦區標題列反白(M26 R1)
	styleCursor  = lipgloss.NewStyle().Reverse(true)                   // 游標態反白(聚焦區內游標停駐項目; M26 R2)
	styleChosen  = lipgloss.NewStyle().Background(lipgloss.Color("4")) // 選取模式已選態選取色底(M27 R4)
	styleNote    = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // 等待落點標題提醒 + 鍵位列行 2 提示(M27): yellow 獨佔「輪到你」
	styleBad     = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // 負面狀態(封 / 封印 / 營業失敗): red
	styleWard    = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // 保護狀態(免疫): cyan(與標題同色, 語意各自獨立)
	styleGood    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // 正面狀態(有效果 / 營業成功): green(同日誌效果綠)
)

// focusView 聚焦高亮: 區輸出首行(面板標題列 / 狀態列標籤行)整列反白, 其餘行原樣——
// 【營業顯示規格書 | 7、互動規格 | 7.1】「邊框高亮」於本 layout 即標題列(M26 拍板, 不補滿框、不動行數預算);
// 聚焦屬區層級、由父層上色(游標態屬區內項目層級、歸組件; M26 拍板), 先排版後上色、不影響字元寬度。
// 經 restyle 剝除列內既有樣式(標題青 / 提醒黃 / 框線暗)再反白。
func focusView(view string) string {
	row := strings.SplitN(view, "\n", 2)
	row[0] = restyle(&styleFocus, row[0])
	return strings.Join(row, "\n")
}

// restyle 整段重上樣式: 先剝除段內既有樣式再上——內嵌樣式的 reset 會切斷外層, 整段態(反白 / 變暗 /
// 選取色底)直接蓋過段內語意色(讓位; M27 拍板)。所有整段包覆一律經此, 段內可自由上色不互踩。
// 樣式傳指標(Style 體積大, 避免逐呼叫拷貝); 引數恆為上方套件級樣式變數。
func restyle(style *lipgloss.Style, text string) string {
	return style.Render(ansi.Strip(text))
}

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
