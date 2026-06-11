package rodi

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/games"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// stepper 暫停機橋接器(【營業實作規格書 | 四、解耦的關鍵：邊界介面 | TUI stepper】): 引擎跑在交棒
// goroutine 上、只在 Next 期間推進——放行一拍(gate), 引擎跑到下一個事件邊界送出行組(line)後再停。
// Next 返回後引擎必停在事件邊界, TUI 直讀 game 盤面(單一真相), 不持第二份規則狀態;
// 快/慢/步進(M27)只是呼叫 Next 的節奏、引擎端零改動。
// done 緩衝 1: 終局只在最後一拍放行後送出, 不會搶在行組之前到達。
type stepper struct {
	game  *cores.Game   // 營業實例(唯一真相; 停點間組件唯讀直讀)
	event chan []string // 日誌流(無緩衝; 一拍一組行)
	gate  chan struct{} // 放行訊號(無緩衝; 嚴格會合, 引擎每拍消費一個)
	done  chan bool     // 終局成敗(緩衝 1)
	over  bool          // 終局已收(再 Next 直接回 more == false, 不再放行)
	succ  bool          // 終局成敗(over == true 後有效)
}

// newStepper 組裝營業實例與交棒 goroutine: 引擎建後完全靜止(首拍放行前不推進),
// 建構與首拍之間讀到的是開局建置前的空白盤面。operator 固定被動觀看(鍵盤 Operator 到 M26 再開參數)。
// 中途離開不做取消: process 結束時停棒中的引擎 goroutine 隨之消滅,
// 核心不吃 context(鐵則 1)、亦無外部資源要清理。
func newStepper(seed int64, stageID int32, sheet *sheeter.Sheeter) *stepper {
	this := &stepper{
		event: make(chan []string),
		gate:  make(chan struct{}),
		done:  make(chan bool, 1),
	}
	this.game = games.Build(seed, stageID, sheet, passiveOperator{}, gatePresenter{event: this.event, gate: this.gate})
	go func() {
		<-this.gate // 首拍
		this.done <- games.Loop(this.game)
	}()
	return this
}

// Next 推進一拍: 放行引擎跑到下一個事件邊界並收回行組(more == true), 引擎停機則收終局成敗(more == false)。
// 同步呼叫且只在 Update 內使用: Bubble Tea 的 Update / View 跑在同一條 goroutine,
// 引擎只在 Next 期間推進 → 「TUI 在看畫面時引擎必停」不需鎖即成立。
func (this *stepper) Next() (line []string, more bool) {
	if this.over {
		return nil, false // 終局後防呆: goroutine 已消滅, 再放行會永久阻塞
	} // if

	this.gate <- struct{}{}

	select {
	case line = <-this.event:
		return line, true

	case succ := <-this.done:
		this.over = true
		this.succ = succ
		return nil, false
	}
}

// gatePresenter 日誌流輸出 port 的交棒半身: Emit 送出行組(發後不改)後停棒等放行——
// 自消費端收走行組起到下一拍放行前, 引擎停在事件邊界(Next 直讀安全窗的另一半)。
type gatePresenter struct {
	event chan<- []string // 與 stepper.event 同一條
	gate  <-chan struct{} // 與 stepper.gate 同一條
}

func (this gatePresenter) Emit(line ...string) {
	this.event <- line
	<-this.gate
}
