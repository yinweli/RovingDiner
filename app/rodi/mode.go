package rodi

import (
	"time"
)

// mode 執行模式(【營業顯示規格書 | 3、日誌流的消費：速率與步進】): 三種速率消費同一條日誌流——
// 快/慢 = 排拍 Cmd 的節奏、步進 = 不排拍等 [N], 引擎端零改動。屬 UI 狀態、由 model 持有(M25 拍板):
// 模式驅動 model 自己的推進迴圈, 且狀態列顯示與 [N] 門控都消費它, 跨組件狀態歸父層
// (組件自持只留「只有該組件自己用」的狀態, 如游標)。
type mode int

const (
	modeFast mode = iota // 快速(0.1 秒一拍)
	modeSlow             // 慢速(1 秒一拍)
	modeStep             // 步進(停止自動消費, 按 [N] 逐拍前進的暫停態)
)

// 排拍間隔(M25 拍板數值)。
const (
	intervalFast = 100 * time.Millisecond // 快速
	intervalSlow = time.Second            // 慢速
)

// name 顯示名稱(狀態列模式欄用; 【營業顯示規格書 | 6、畫面規格 | 6.9】)。
func (this mode) name() string {
	switch this {
	case modeFast:
		return "快速"

	case modeSlow:
		return "慢速"

	case modeStep:
		return "步進"

	default:
		return "-"
	} // switch
}

// interval 排拍間隔(快 0.1 秒 / 慢 1 秒): 步進不排拍、回 0(呼叫端先以模式擋下, 不據此排拍)。
func (this mode) interval() time.Duration {
	switch this {
	case modeFast:
		return intervalFast

	case modeSlow:
		return intervalSlow

	default:
		return 0
	} // switch
}

// next 循環切換(快速 → 慢速 → 步進 → 快速; [Space] 消費, 【營業顯示規格書 | 3、日誌流的消費：速率與步進】)。
func (this mode) next() mode {
	switch this {
	case modeFast:
		return modeSlow

	case modeSlow:
		return modeStep

	default:
		return modeFast
	} // switch
}
