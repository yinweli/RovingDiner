package rodi

// 游標工具(M26 R2; 【營業顯示規格書 | 7、互動規格 | 7.2】): 游標歸組件自持、存 index、讀取時夾界
// (盤面拍間會變, 指標當游標會懸空)——夾界 / 單列移動 / cursor-follow 窗格起點為各組件共用原語。

// 方向鍵名(對應 tea.KeyMsg.String(); 綁定表與各組件 Move 共用)。
const (
	keyUp    = "up"    // 上
	keyDown  = "down"  // 下
	keyLeft  = "left"  // 左
	keyRight = "right" // 右
)

// 窗格左緣記號與同寬佔位(【營業顯示規格書 | 6、畫面規格 | 6.12】< 貼左緣; 多行區塊其餘行以佔位同縮排對齊)。
const (
	markHead   = "< " // 左緣記號(首行)
	markIndent = "  " // 同寬佔位(其餘行)
)

// clampIndex 游標夾界: 夾進 [0, size-1]; 空列表回 0。
func clampIndex(cursor, size int) int {
	if size <= 0 || cursor < 0 {
		return 0
	} // if

	if cursor >= size {
		return size - 1
	} // if

	return cursor
}

// moveIndex 單列游標移動: 左右增減後夾界(上下與其他鍵不動作, 由呼叫端自理)。
func moveIndex(cursor int, key string, size int) int {
	cursor = clampIndex(cursor, size)

	if key == keyLeft {
		cursor--
	} // if

	if key == keyRight {
		cursor++
	} // if

	return clampIndex(cursor, size)
}

// stripFirst 游標窗格起點(cursor-follow; 【營業顯示規格書 | 6、畫面規格 | 6.12】): 自 0 起找最小起點使
// 游標 cell 完整可視——起點 > 0 時左緣記號佔 2 格、游標非末項時再保 2 格(右緣記號的截斷不吃掉游標尾),
// 一併扣除。cursor 由呼叫端先夾界(clampIndex)。不另存窗格狀態, 起點純由游標導出
// (盤面拍間變動自然重算, 不會殘留過期窗格)。
func stripFirst(width []int, gap, budget, cursor int) int {
	for first := 0; first < cursor; first++ {
		need := 0

		if first > 0 {
			need += 2
		} // if

		for itor := first; itor <= cursor; itor++ {
			if itor > first {
				need += gap
			} // if

			need += width[itor]
		} // for

		if cursor < len(width)-1 {
			need += 2
		} // if

		if need <= budget {
			return first
		} // if
	} // for

	return cursor
}
