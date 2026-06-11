// Package rodi 是營業顯示/操作層的應用本體(Bubble Tea): 引擎跑在 stepper 的交棒 goroutine 上、
// 只在 Update 內同步 Next 逐拍推進(暫停機)——停點間引擎必停在事件邊界, 盤面組件直讀 *cores.Game
// (單一真相、不持拷貝), 日誌流只供敘事與步進節拍(引擎直出最終行, 本層零合成); goroutine + channel
// 只活在本層, 核心維持同步單執行緒(【營業實作規格書 | 五、顯示／操作層（TUI）】)。
//
// M19 殼＋橋接(被動觀看 passiveOperator)、M20 渲染地基(組件契約 + 渲染工具 + 狀態列 / 鍵位列)、
// 六區組件(M21)、事件日誌與 alt-screen(M22)、暫停機重構(M23: 鏡像投影退役)、日誌流換軌
// (M24: journal 行合成退役)已落地; 速率(M25)、互動(M26)、鍵盤 Operator(M27)後續疊上。
package rodi
