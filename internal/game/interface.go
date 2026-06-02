package game

import (
	"github.com/yinweli/RovingDiner/internal/defines"
)

// 這三個介面是核心對外的唯一行為邊界；TUI 與測試各自提供實作。
// 對應【營業實作規格書 | 四、解耦的關鍵：邊界介面】。
//
// 靜態表格不走介面：核心直接吃 Sheeter 雙語言生成的 *sheeter.Sheeter（資料 port 本身），
// 以 reader.Get 查詢；衍生索引由核心預建（見 helper.go）。
//
// 核心完全同步、單執行緒、無 channel：需要玩家輸入時阻塞呼叫 Operator，
// 每跑一個單位呼叫 Presenter.Emit。goroutine + channel 只活在 TUI adapter。

// Rander 唯一亂數來源（單一 seeded PRNG）；決定性的基礎。
type Rander interface {
	Intn(n int) int                     // 回傳 [0, n) 的隨機整數
	Shuffle(n int, swap func(i, j int)) // 對 n 個元素洗牌
	Weighted(weight []int32) int        // 抽獎 weighted random；回傳命中索引
}

// Operator 玩家輸入；對應所有「暫停流程」點。
type Operator interface {
	PlayerAction(runtime *Runtime) Action          // 玩家行動：出牌 / 結束
	PickGuest(source []*Guest, count int) []*Guest // 新選顧客 / guestPick / nearPick / samePick
	PickCard(source []*Card, count int) []*Card    // 新選手牌 / *Pick 牌堆類
	PickDiscard(source []*Card, over int) []*Card  // 手牌上限棄牌
}

// Presenter 事件流輸出（yield-per-unit；步進時可阻塞）。
type Presenter interface {
	Emit(eventData EventData) // 一個命令 / 一次觸發 / 一次 phase 切換 / 一個玩家動作 = 一個 EventData
}

// EventData 核心吐給前端的細粒度投影事件。
// TODO(M5): 各類別的詳細欄位（目標 InstanceID、屬性前後值、表演資訊）於顯示層實作時補齊，
// 依據【營業顯示規格書 | 3、事件流的消費：速率與步進】與【營業規格書 | 十八、表演資訊】。
type EventData struct {
	Event defines.Event
}
