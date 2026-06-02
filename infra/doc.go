// Package infra 是 Go-only 的基礎設施層（不移植 C#）：
// sheet 載入、PRNG 來源。與 sheet/ 同層、非 internal。
//
// 載入流程：Load 以 Loader 讀 sheetdata/*.json → sheeter.NewSheeter(loader).FromData()
// 建立 *sheeter.Sheeter，核心直接吃這份生成的跨語言資料 port；
// 衍生索引（如抽獎依群組聚合）由核心自行預建，不在本層。
package infra
