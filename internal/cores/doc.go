// Package cores 是營業核心的純資料模型 substrate：
// 行為邊界介面 / 投影事件 / 型別別名 / 列舉（define.go）、
// 實例與數值原語（instance.go）、一場營業的聚合狀態（runtime.go）。
//
// cores 不懂遊戲規則、不 import games / exprs，永遠不反向依賴遊戲邏輯；
// 全部營業邏輯（屬性 / 命令 / 命令對象 / 效果 / 流程）歸 games，單向 import cores。
// 對應【營業實作規格書 | 二、套件結構】與【營業實作規格書 | 十、決策點】核心分包粒度。
package cores
