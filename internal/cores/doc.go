// Package cores 是營業引擎本體: 資料模型與驅動引擎。
// 資料模型為行為邊界介面 / 投影事件 / 列舉 / 詞彙行為簽名(define.go)、數值與集合原語(value.go / help.go)、
// 引用(ref.go)、實例與容器(card.go / guest.go / effect.go / action.go)、遊戲資料聚合(data.go);
// 驅動引擎為 game.go: 一場營業的聚合狀態, 委派實作 exprs.Resolver, 持有詞彙表(由 rules 經 Register* 裝備)。
//
// cores 單向 import exprs(零遊戲依賴的運算式語言)、不 import rules / games; 對外的營業執行 / Parse / Validate 呼叫面由 games 收斂。
// 對應【營業實作規格書 | 二、套件結構】與【營業實作規格書 | 三、核心引擎的內部分層】。
package cores
