// Package features 是 Go-only 的基礎設施層（不移植 C#）：
// sheet 載入、Dater 組裝、PRNG 來源。與 sheet/ 同層、非 internal。
//
// 載入流程：Loader 讀 sheetdata/*.json → sheeter.NewSheeter(loader).FromData()
// 建立 *sheeter.Sheeter → NewDater(...) 包成 game.Dater（含衍生索引）。
package features
