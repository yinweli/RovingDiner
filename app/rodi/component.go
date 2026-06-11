package rodi

import (
	"strings"
)

// component 畫面組件契約(【營業實作規格書 | 九、里程碑 | M20】組件化 layout): 各組件獨立自持渲染——
// View 為純函式(唯讀世界鏡像 + 寬度預算 → 渲染字串), 組件間互不伸手、規則狀態一律讀鏡像(M20 拍板);
// 新增介面 = 新掛組件。M20 以狀態列 + 鍵位列驗證框架, 六區(M21)/ 事件日誌(M22)循同契約疊上。
type component interface {
	View(world *mirror, width int) string
}

// composeView 父層組合: 依掛載順序逐組件渲染、換行堆疊; 父層只組合與分配空間——M20 只發寬度預算,
// 高度分配與最低尺寸守門隨 alt-screen 收進 M22(M20 拍板)。
func composeView(world *mirror, width int, comp []component) string {
	view := []string{}

	for _, itor := range comp {
		view = append(view, itor.View(world, width))
	} // for

	return strings.Join(view, "\n")
}
