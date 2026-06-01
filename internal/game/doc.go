// Package game 是流浪食堂「營業」階段的純邏輯規則引擎。
//
// 核心零 I/O、零框架依賴、單執行緒、決定性；所有對外互動都走邊界介面
// （Operator / Presenter / Rander / Dater）。本套件依關注點分檔，
// 但同屬單一 package：flow / command / effect / settlement 共讀寫同一份 Runtime 狀態。
// 設計依據見 doc/營業實作規格書.md，規則 SSOT 見 doc/營業規格書.md。
package game
