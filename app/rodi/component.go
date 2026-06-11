package rodi

import (
	"strings"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// component 畫面組件契約(【營業實作規格書 | 九、里程碑 | M20】組件化 layout): 各組件獨立自持渲染——
// View 唯讀引擎盤面 + 寬度預算 + 是否聚焦(游標態只在聚焦區顯示; M26 R2)→ 渲染字串, 組件間互不伸手;
// 規則狀態一律直讀 *cores.Game, 引擎停點間必停在事件邊界、唯讀不需鎖(M23 拍板, 原 M20 讀鏡像作廢)。
// Move 收方向鍵移游標(語意隨區, 【營業顯示規格書 | 7、互動規格 | 7.2】)——游標為組件自持 UI 狀態
// (M20 拍板), 故組件以指標掛載; 夾界一律在讀取時(cursor.go 原語)。
// Item 回游標下的實例指標(M26 R4; 候選身分 = 實例指標拍板, [Enter] 據此開對應檢視 modal——
// 顧客 / 行動 / 效果 / 卡牌, 父層辨型分派; 空格 / 空列回 nil 不動作)。新增介面 = 新掛組件。
type component interface {
	View(game *cores.Game, width int, focus bool) string
	Move(game *cores.Game, key string)
	Item(game *cores.Game) any
}

// composeView 父層組合: 依掛載順序逐組件渲染、換行堆疊; 父層只組合與分配空間——M20 只發寬度預算,
// 高度分配與最低尺寸守門隨 alt-screen 收進 M22(M20 拍板); M26 R1 加聚焦高亮——focus 命中掛載索引時
// 該組件標題列反白(聚焦屬區層級歸父層上色, 組件不知聚焦; 範圍外如 -1 或狀態列 / 日誌索引即全不高亮)。
func composeView(game *cores.Game, width int, comp []component, focus int) string {
	view := []string{}

	for index, itor := range comp {
		render := itor.View(game, width, index == focus)

		if index == focus {
			render = focusView(render)
		} // if

		view = append(view, render)
	} // for

	return strings.Join(view, "\n")
}
