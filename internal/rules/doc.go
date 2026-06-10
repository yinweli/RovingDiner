// Package rules 是命令語言的詞彙層：全域 / 引用屬性讀寫、命令對象、操作命令、內建函式的詞彙表與行為實作，
// 以及效果流程（effectStack / effectStart / effectTrigger）。每詞條一具名函式，套件層 map 字面值為單一登錄來源，
// 對應【營業規格書 | 二十三、屬性清單】【營業規格書 | 二十四、命令對象清單】【營業規格書 | 二十五、操作命令清單】【營業規格書 | 二十六、內建函式清單】。
//
// Register 把全部詞彙經 Game.Register* 裝備進 cores.Game；Has* 供 games 的 Validate 做名稱合法性查詢。
// rules 單向 import cores（營業引擎本體）→ exprs（運算式語言）、不 import games；依賴鏈 games → rules → cores → exprs。
package rules
