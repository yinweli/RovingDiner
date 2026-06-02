# 工作進度與接續記錄

本檔是跨 session 的接續工作交接點。每次工作告一段落時更新「里程碑進度」與「接續待辦」,下次開新 session 先讀本檔即可接手。

紀律:本檔只記「從 git 歷史與 `doc/` 規格看不出來的待辦與決策」。規則細節以 `doc/` 規格為準、程式現況以 git 為準,不在此重複。

## 里程碑進度

對應 `doc/營業實作規格書.md`【九、里程碑】。

| 里程碑          | 狀態    | 說明                                                                       |
|:----------------|:--------|:---------------------------------------------------------------------------|
| M0 骨架         | ✅ 完成  | defines + Value 型別 + runtime 容器 + 三個行為邊界介面 + infra.Load         |
| M1 expr         | ✅ 完成  | `internal/expr` 運算式引擎 + 完整單測;規格 BNF 已連帶修正                  |
| M2 屬性 + 命令  | ⬜ 下一步 | property registry + 屬性修改命令 + 操作命令 + 命令對象解析(selector)     |
| M3 效果系統     | ⬜       | 佇列 / 堆疊 / 觸發 / 推進 / 清理                                           |
| M4 流程         | ⬜       | 6 phase + 啟動技能 + 執行結算,跑出第一局可結算營業(scripted Operator)   |
| M5 TUI          | ⬜       | Bubble Tea adapter + 事件日誌 + 步進                                       |

## 接續待辦(留待後續里程碑釐清)

- **`none` 字面值(M2)**:expr 文法目前無 `none` 字面值,守衛式裡的「空物件」需由 Resolver 提供「回傳空 ref 的條件對象」。M2 接 game 實作 Resolver 時,需決定是否把 `none` 納入 expr 文法(納入則要改 lexer / parser / BNF 與規格)。
- **§12 觸發次數的限制由呼叫方套用**:expr 只提供完整文法的 Parse / Eval。【十二、觸發次數】「採算術式、結果 < 0 或評估失敗 → 視為 0、結果 = 0 → 不觸發」的語意,由呼叫方(M3 觸發流程)在 Eval 之後套用,非 expr 內建。

## 已敲定的設計決策(勿重新爭論)

- **expr 優先級 `NOT > AND > OR`**(非規格原 BNF 的同級);BNF 已修正為分層(邏輯或 → 邏輯且 → 否定)。
- **AND / OR 短路求值**:AND 左假不評右、OR 左真不評右(與三元「只評被選中分支」一致);左側本身評估失敗時仍整體失敗。守衛式 `self != none AND self.calm > 5` 因此成立。
- **expr 數值用 float64**:中間值可為小數,expr 內不四捨五入;捨入(half-away-from-zero)由呼叫方寫回屬性時以 `expr.Round` 處理。
- **Resolver 是 expr 與 game 的唯一接縫**,對齊【二十三、屬性清單】表結構:主表(全域屬性 / 查詢函式 / 物件引用)走 `Property`、子表(卡牌 / 顧客引用屬性)走 `Member`。
- **核心邊界介面定為三個**(Operator / Presenter / Rander),靜態資料直接以 `*sheeter.Sheeter` 注入;不再有 per-table Dater 介面。衍生索引(如 Award 依群組聚合)移入核心預建。

## expr 套件研究順序(自學路線)

研究 `internal/expr/` 的建議順序:順著「原始字串 → token → AST → 求值」資料流,每個檔搭配同名 `_test.go` 當可執行規格看。重點在 `ast.go`(求值語意)與 `parser.go`(優先級),值得花最多時間。

先讀規格背景:`doc/營業規格書.md`【二十七、運算式】【二十六、內建函式】。

- [ ] 1. `doc.go` — 套件全貌(零遊戲依賴、Resolver 為唯一接縫、失敗以 `ok=false` 表達)
- [ ] 2. `value.go` + `value_test.go` — `Value` 標籤聯合;`AsBool`(【二十七〇6】省略比較符布林判定)、`Round`(half-away-from-zero,expr 內**不**捨入)
- [ ] 3. `resolver.go` — 兩方法介面;`Property` / `Member` 對應【二十三】主表 / 子表結構(解耦關鍵)
- [ ] 4. `token.go` — token 種類(速覽)
- [ ] 5. `lexer.go` + `lexer_test.go` — 詞法:`!` vs `!=`、單引號字串、關鍵字大小寫不敏感、裸 `=` 報錯
- [ ] 6. `ast.go` 第一遍 — 只看 `node` 介面與各節點 struct,建立「AST 長相」
- [ ] 7. `ast.go` 第二遍 + `eval_test.go` — 各節點 `eval`:`evalLogic` 短路、比較 / 跨型別 / ref 比較、`evalArith` 除零、ternary 惰性、`memberNode` 空物件失敗、max / min
- [ ] 8. `parser.go` + `parser_test.go` — 遞迴下降優先級階梯 `parseExpr→Or→And→Not→Compare→Add→Mul→Factor`;分層即優先級
- [ ] 9. `expr.go` — `Parse`(編譯一次)+ `Expr.Eval`(對不同 Runtime 多次求值)
- [ ] 10. `expr_test.go` + `helper_test.go` — 端到端範例 + `mockResolver`(M2 實作真 Resolver 的雛形)

貫穿練習:拿 `morale > 0 AND self.calm >= 5`,自己跑一遍 lex → parse(畫 AST,頂層為 AND)→ eval(給 `self` 為空物件的 Resolver,觀察短路如何讓它不去存取 `self.calm`)。
