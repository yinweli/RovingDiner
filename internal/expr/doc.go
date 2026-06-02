// Package expr 是流浪食堂的運算式引擎(lexer → parser → AST → evaluator)。
//
// 本套件零遊戲依賴、僅依賴標準庫,因此可單獨測試、可整段移植到 C#。
// 它是【營業規格書 | 十一、觸發條件】【營業規格書 | 十二、觸發次數】【營業規格書 | 十七、命令】(RHS 算術式)
// 共用的運算式 SSOT,語法定義見【營業規格書 | 二十六、內建函式清單】【營業規格書 | 二十七、運算式】。
//
// 與 game 的唯一接縫是 Resolver:expr 只懂語法與運算,遇到「條件對象」
// (morale、self.calm、drawLast.cardID、tableCount('>=',2)…)時呼叫 Resolver 取值。
// game(M2)實作 Resolver,把【營業規格書 | 二十三、屬性清單】的綁定塞進去。
//
// 評估不拋錯:語意失敗以 Eval 的第二回傳值 ok=false 表達(即【營業規格書 | 二十七、運算式 | 8】評估失敗狀態);
// 語法錯誤則由 Parse 回傳 error。中間值可為小數,expr 不四捨五入——捨入由呼叫方寫回屬性時以 Round 處理。
package expr
