package rodi

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// scopeText 範圍事件名(標題的 {事件}; 【營業顯示規格書 | 6、畫面規格 | 6.10】範圍事件表)。
// 詞彙對照本體已下沉 cores(發射端 / 顯示端共用), 本檔僅餘事件流消費端專用的範圍轉換。
func scopeText(scope cores.ScopeKind, trigger cores.TriggerKind) string {
	switch scope {
	case cores.ScopePlay:
		return "玩家出牌"

	case cores.ScopeGuest:
		return "顧客行動"

	case cores.ScopePrefix:
		return "前置技能"

	case cores.ScopeTrigger:
		return "時機:" + cores.TriggerText(trigger)

	case cores.ScopeSettle:
		return "執行結算"

	case cores.ScopeManual:
		return "手動結束"

	default:
		return "?"
	} // switch
}
