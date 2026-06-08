// Package exprs 是流浪食堂營業系統的運算式引擎（lexer → parser → AST → evaluator）。
//
// 本套件零遊戲依賴、僅依賴標準庫,因此可單獨測試。它是【營業規格書 | 十一、觸發條件】
// 【營業規格書 | 十二、觸發次數】【營業規格書 | 十七、命令】(RHS 算術式)共用的運算式 SSOT,
// 語法定義見【營業規格書 | 二十六、內建函式清單】【營業規格書 | 二十七、運算式】。
//
// 與 games 的唯一接縫是 Resolver:exprs 只懂語法與運算,遇到「條件對象」
// (morale、self.calm、drawLast.cardID、tableCount('>=',2)…)時呼叫 Resolver 取值,
// 內建函式(min / max…)亦由 games 注入;exprs 自身零遊戲知識。
//
// 評估不拋錯:語意失敗以評估失敗狀態表達(即【營業規格書 | 二十七、運算式 | 8】);
// 語法錯誤則回傳帶位置的 SyntaxError。中間值可為小數,exprs 不四捨五入——
// 捨入(half-away-from-zero)由呼叫方寫回屬性時處理。
package exprs
