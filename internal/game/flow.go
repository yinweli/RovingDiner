package game

import (
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// Run 驅動一場營業：內部即規格書【十九】【二十】的同步流程。
// 靜態資料以 *sheeter.Sheeter 直接注入（核心的資料 port），行為互動走三個邊界介面。
// TODO(M4): 尚未實作；M0 僅鎖定整合點的簽章。
func Run(runtime *Runtime, data *sheeter.Sheeter, rander Rander, operator Operator, presenter Presenter) {
	panic("game.Run: not implemented (M4)")
}
