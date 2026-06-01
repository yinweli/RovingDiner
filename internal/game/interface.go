// Package game 是流浪食堂「營業」階段的純邏輯規則引擎。
//
// 核心零 I/O、零框架依賴、單執行緒、決定性；所有對外互動都走邊界介面
// （Operator / Presenter / Rander / Dater）。本套件依關注點分檔，
// 但同屬單一 package：flow / command / effect / settlement 共讀寫同一份 Runtime 狀態。
// 設計依據見 doc/營業實作規格書.md，規則 SSOT 見 doc/營業規格書.md。
package game

import (
	"github.com/yinweli/Project_005/internal/defines"
	sheeter "github.com/yinweli/Project_005/sheet"
)

// 這四個介面是核心對外的唯一邊界；TUI 與測試各自提供實作。
// 對應營業實作規格書【四、解耦的關鍵：四個邊界介面】。
//
// 核心完全同步、單執行緒、無 channel：需要玩家輸入時阻塞呼叫 Operator，
// 每跑一個單位呼叫 Presenter.Emit。goroutine + channel 只活在 TUI adapter。

// Dater 唯讀靜態表格（包既有 sheet readers）。
type Dater interface {
	Card(id int32) *sheeter.Card
	Guest(id int32) *sheeter.Guest
	Skill(id int32) *sheeter.Skill
	Effect(id int32) *sheeter.Effect
	Seat(id int32) *sheeter.Seat
	Setting(id string) *sheeter.Setting
	Award(group int32) []*sheeter.Award // 依抽獎群組聚合（衍生索引在 features 預建）
}

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
// 依據營業顯示規格書 §3 / spec【十八】。
type EventData struct {
	Event defines.Event
}
