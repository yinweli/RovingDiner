// Package games 是營業核心的對外介面層：命令文法解析（Parse）、命令名稱驗證（Validate），
// 以及未來的營業執行入口（Run）。引擎本體與命令語言行為（engine / attr / attrRef / command / selector）歸 cores。
//
// games 單向 import cores（營業引擎本體）與 exprs（運算式語言）：Parse 為純文法解析、
// Validate 逐名查 cores 詞彙表（HasCommand / HasSelector）；cores 不反向 import games，分層無環。
// 對應【營業實作規格書 | 二、套件結構】與【營業實作規格書 | 三、核心引擎的內部分層】。
package games
