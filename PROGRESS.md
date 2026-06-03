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
- **企劃驗證器 = M2 後置支線**:運算式／命令語法驗證器(單筆輸入 + Effect 表掃描,同一引擎)。6 欄位中 4 個命令欄位(立即／觸發／啟動／結束命令)是命令文法、需 M2 命令解析器。**地基(共用帶位置錯誤型別)入 M2 驗收;工具本體為 M2 後置、非 gating 支線**(可與 M3 並行),不先出只驗運算式的半套。原「帶參數求值 + 運算過程」工具不做。詳見實作規格書【附錄：企劃驗證器】。錯誤定位地基已完成(見下方 `SyntaxError` 決策),M2 做命令解析器時直接重用。

## 已敲定的設計決策(勿重新爭論)

- **expr 優先級 `NOT > AND > OR`**(非規格原 BNF 的同級);BNF 已修正為分層(邏輯或 → 邏輯且 → 否定)。
- **AND / OR 短路求值**:AND 左假不評右、OR 左真不評右(與三元「只評被選中分支」一致);左側本身評估失敗時仍整體失敗。守衛式 `self != none AND self.calm > 5` 因此成立。
- **expr 數值用 float64**:中間值可為小數,expr 內不四捨五入;捨入(half-away-from-zero)由呼叫方寫回屬性時以 `expr.Round` 處理。
- **Resolver 是 expr 與 game 的唯一接縫**,對齊【二十三、屬性清單】表結構:主表(全域屬性 / 查詢函式 / 物件引用)走 `Attr`、子表(卡牌 / 顧客引用屬性)走 `AttrRef`。`defines` 端的 `Property` / `Attr` 亦同步改名為 `Attr` / `AttrRef`。
- **內建函式註冊表化**:`builtin.go` 以 `map[string]builtinFunc` 登記內建函式(max / min),新增函式只需註冊一筆,`nodeCall` 與 parser 不動。對齊【二十六、內建函式清單】。
- **語法錯誤型別 `expr.SyntaxError`**(放 `expr` 套件,`error.go`):欄位 `Pos`(rune 索引,0 起算,供工具標位置)+ `Msg`(企劃白話,無 `expr:` 前綴);`Error()` 顯示「第 N 字附近:Msg」(N = Pos+1)。`token` 帶 `pos`,lexer/parser 全部錯誤皆走 `newError(pos, msg)`。**M2 命令解析器應重用同型別**,使命令與運算式的錯誤格式對企劃一致。
- **核心邊界介面定為三個**(Operator / Presenter / Rander),靜態資料直接以 `*sheeter.Sheeter` 注入;不再有 per-table Dater 介面。衍生索引(如 Award 依群組聚合)移入核心預建。
