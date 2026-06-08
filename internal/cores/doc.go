// Package cores 是營業引擎本體：資料模型、驅動引擎與命令語言的執行行為。
// 資料模型為行為邊界介面 / 投影事件 / 列舉（define.go）、實例與數值原語（instance.go）、一場營業的聚合狀態（runtime.go）；
// 驅動引擎為 engine.go（委派實作 exprs.Resolver）；命令語言行為為全域 / 引用屬性、操作命令、命令對象的詞彙表與執行行為（attr.go / attrRef.go / command.go / selector.go）。
//
// cores 單向 import exprs（零遊戲依賴的運算式語言）、不 import games；對外的營業執行 / Parse / Validate 呼叫面由 games 收斂。
// 對應【營業實作規格書 | 二、套件結構】與【營業實作規格書 | 三、核心引擎的內部分層】。
package cores
