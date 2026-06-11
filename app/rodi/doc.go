// Package rodi 是營業顯示/操作層的應用本體(Bubble Tea): 實作 Operator(鍵盤 → 選取/動作)與
// Presenter(EventData → 事件日誌 + 盤面渲染), 以背景 goroutine 驅動核心、channel 橋接事件流——
// goroutine + channel 只活在本層, 核心維持同步單執行緒(【營業實作規格書 | 五、顯示／操作層（TUI）】)。
//
// M19 殼＋橋接(被動觀看 passiveOperator + 事件純文字 dump)與 M20 渲染地基(組件契約 + 世界鏡像 +
// 渲染工具 + 狀態列 / 鍵位列, inline 模式)已落地; 六區組件(M21)、事件日誌與 alt-screen(M22)、
// 互動(M23)、鍵盤 Operator(M24)、速率(M25)後續疊上。
package rodi
