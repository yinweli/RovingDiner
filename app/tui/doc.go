// Package tui 是營業顯示／操作層的應用本體（Bubble Tea）：實作 Operator（鍵盤 → 選取／動作）與
// Presenter（EventData → 事件日誌 + 盤面渲染），以背景 goroutine 驅動核心、channel 橋接事件流——
// goroutine + channel 只活在本層，核心維持同步單執行緒（【營業實作規格書 | 五、顯示／操作層（TUI，Go-only）】）。
//
// M19 為殼＋橋接站：被動觀看（passiveOperator 永不出牌）+ 事件純文字 dump 證明端到端通；
// 渲染組件（M20+）、事件日誌格式（M22）、鍵盤 Operator（M24）、速率（M25）後續疊上。
package tui
