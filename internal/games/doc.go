// Package games 是營業核心的對外介面層：命令文法解析（Parse）、命令名稱驗證（Validate），
// 以及未來的營業執行入口（Run）。引擎本體歸 cores、命令語言行為（attr* / command* / selector / effect 流程）歸 rules。
//
// games 單向 import rules（詞彙與行為）→ cores（營業引擎本體）→ exprs（運算式語言）：Parse 為純文法解析、
// Validate 逐名查 rules 詞彙表（HasCommand / HasSelector）；執行入口建構 Game 後以 rules.Register 裝備詞彙。
// 對應【營業實作規格書 | 二、套件結構】與【營業實作規格書 | 三、核心引擎的內部分層】。
package games
