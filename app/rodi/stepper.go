package rodi

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/games"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// turnRole 交棒輪次角色(M27 拍板: 交棒廣義化——行組 | 輸入請求 | 終局共用單一 channel)。
type turnRole int

const (
	turnLine    turnRole = iota // 行組(一拍 0..n 行日誌)
	turnRequest                 // 輸入請求(引擎暫停等玩家輸入)
	turnOver                    // 終局(成敗)
)

// turn 交棒輪次(tagged): role 辨型、對應欄位有效; 單一 channel 取代 M23 的 event / done 兩條。
type turn struct {
	role turnRole // 輪次角色
	line []string // 行組(turnLine)
	req  *request // 輸入請求(turnRequest; 自帶答覆通道)
	succ bool     // 終局成敗(turnOver)
}

// stepper 暫停機橋接器(【營業實作規格書 | 四、解耦的關鍵：邊界介面 | TUI stepper】): 引擎跑在交棒
// goroutine 上、只在 Next / Answer 期間推進——放行一拍(gate), 引擎跑到下一個事件邊界送出輪次
// (行組 / 輸入請求 / 終局)後再停。返回後引擎必停在事件邊界(Emit 停棒 / Operator 等答覆 / 已終局),
// TUI 直讀 game 盤面(單一真相), 不持第二份規則狀態; 快/慢/步進(M25)只是呼叫 Next 的節奏、引擎端零改動。
// 終局只在最後一拍放行後送出, 單一 channel 天然保序。
type stepper struct {
	game *cores.Game   // 營業實例(唯一真相; 停點間組件唯讀直讀)
	turn chan turn     // 交棒輪次(無緩衝; 行組 / 輸入請求 / 終局共用一條)
	gate chan struct{} // 放行訊號(無緩衝; 嚴格會合, 引擎每拍消費一個)
	over bool          // 終局已收(再 Next 直接回終局輪次, 不再放行)
	succ bool          // 終局成敗(over == true 後有效)
}

// newStepper 組裝營業實例與交棒 goroutine: 引擎建後完全靜止(首拍放行前不推進),
// 建構與首拍之間讀到的是開局建置前的空白盤面。operator 為鍵盤 Operator(暫停點包成輸入請求交棒; M27)。
// 中途離開不做取消: process 結束時停棒中的引擎 goroutine 隨之消滅,
// 核心不吃 context(鐵則 1)、亦無外部資源要清理。
func newStepper(seed int64, stageID int32, sheet *sheeter.Sheeter) *stepper {
	this := &stepper{
		turn: make(chan turn),
		gate: make(chan struct{}),
	}
	this.game = games.Build(seed, stageID, sheet, keyboardOperator{turn: this.turn}, gatePresenter{turn: this.turn, gate: this.gate})
	go func() {
		<-this.gate // 首拍
		this.turn <- turn{role: turnOver, succ: games.Loop(this.game)}
	}()
	return this
}

// Next 推進一拍: 放行引擎跑到下一個事件邊界並收輪次。
// 同步呼叫且只在 Update 內使用: Bubble Tea 的 Update / View 跑在同一條 goroutine,
// 引擎只在 Next / Answer 期間推進 → 「TUI 在看畫面時引擎必停」不需鎖即成立。
func (this *stepper) Next() turn {
	if this.over {
		return turn{role: turnOver, succ: this.succ} // 終局後防呆: goroutine 已消滅, 再放行會永久阻塞
	} // if

	this.gate <- struct{}{}
	return this.recv()
}

// Answer 答覆輸入請求並收下一輪: 引擎自 Operator 返回續跑到下一個事件邊界(行組 / 再請求 / 終局)。
// 請求路徑的恢復訊號 = 答覆本身、不經 gate(M27 拍板); 等待輸入期間 UI 不排拍不 Next, 協定上 gate 恆有人收。
func (this *stepper) Answer(req *request, ans answer) turn {
	req.answer <- ans
	return this.recv()
}

// recv 收一個輪次並記錄終局(Next 的 over 防呆據此)。
func (this *stepper) recv() turn {
	result := <-this.turn

	if result.role == turnOver {
		this.over = true
		this.succ = result.succ
	} // if

	return result
}

// gatePresenter 日誌流輸出 port 的交棒半身: Emit 送出行組輪次(發後不改)後停棒等放行——
// 自消費端收走行組起到下一拍放行前, 引擎停在事件邊界(Next 直讀安全窗的另一半)。
type gatePresenter struct {
	turn chan<- turn     // 與 stepper.turn 同一條
	gate <-chan struct{} // 與 stepper.gate 同一條
}

func (this gatePresenter) Emit(line ...string) {
	this.turn <- turn{role: turnLine, line: line}
	<-this.gate
}
