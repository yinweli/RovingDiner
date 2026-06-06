# 工作進度與接續記錄

本檔是跨 session 的接續工作交接點。每次工作告一段落時更新「里程碑進度」與「接續待辦」,下次開新 session 先讀本檔即可接手。

紀律:本檔只記「從 git 歷史與 `doc/` 規格看不出來的待辦與決策」。規則細節以 `doc/` 規格為準、程式現況以 git 為準,不在此重複。

## 里程碑進度

對應 `doc/營業實作規格書.md`【九、里程碑】。

| 里程碑          | 狀態    | 說明                                                                       |
|:----------------|:--------|:---------------------------------------------------------------------------|
| M0 骨架         | ✅ 完成  | defines + Value 型別 + runtime 容器 + 三個行為邊界介面 + infra.Load         |
| M1 expr         | ✅ 完成  | `internal/expr` 運算式引擎 + 完整單測;規格 BNF 已連帶修正                  |
| M2 屬性 + 命令  | ✅ 完成  | 命令解析 + 屬性註冊表 / Resolver + 屬性修改命令 + selector + 操作命令(純容器 / 屬性類);部分操作命令隨 M3,見接續待辦 |
| M3 效果系統     | ⬜ 下一步 | 佇列 / 堆疊 / 觸發 / 推進 / 清理;併入 M2 延後的操作命令(見接續待辦)      |
| M4 流程         | ⬜       | 6 phase + 啟動技能 + 執行結算,跑出第一局可結算營業(scripted Operator)   |
| M5 TUI          | ⬜       | Bubble Tea adapter + 事件日誌 + 步進                                       |

## 接續待辦(留待後續里程碑釐清)

- **M2 延後至 M3 的操作命令**:解析全部合法(parser 認 59 個動詞),但執行只登記了純容器 / 屬性 / 行動 / 階段類(見 `commandExec` 表)。**未登記 = 解析合法、執行 no-op**,留 M3 一起做,分三類:① 實例化類 `*Add`/`*Copy`/`*Clone`/`*Roll`/`guestSpawn`/`waitAdd`(卡牌 / 顧客實例化資料模型尚待釐清,尤其新卡的實例效果列表來源——Card 表無效果欄)、② 處理流程類 `cardRun`/`*Morph`/`guestExit`/`guestReturn`/`guestSeat`/`cardify`/`restore`(需觸發多時機 + 啟動 / 清理效果 + 座位指派)、③ 效果佇列類 `effectClear`/`effectDel`/`effectRun`(直接讀寫效果佇列 + 堆疊規則)。**此處是對「按 seam 切」決策的務實收斂**:原計畫 E 段「M2 寫完所有命令本體」,實作時發現①②的資料模型與處理流程深度耦合 M3,故一併延後(非僅效果佇列類)。
- **啟動效果列表 / 清理效果 seam 尚未建**:三 seam 中只先建了 `fireTrigger`(Phase 4,記錄至 `Runtime.pending`)。另兩個 seam 的呼叫者(cardRun / morph / cardify / restore)既已延後,seam 本體待 M3 與呼叫者一起建(避免建了沒人呼叫被 `unused` lint)。
- **`Guest.SateMax` 為本次新增欄位**:依【二十三、屬性清單】sateMax(飽食值離場線,寫鎖)補上;M0 結構原漏列。離場判定(sate ≥ sateMax)等用到處在 M3/M4。
- **`morale -=` 特例的 damageGuest 來源**:Phase 4 命令路徑取 `self.Guest`(self 非顧客則 nil → 空物件);guestExit(M3)走特例時須改以「離場顧客」為來源,故 `writeMorale` 已留 `source *Guest` 參數。lock 與特例的交互(morale 鎖定時格擋 / 護盾仍消耗、morale 不動、無實際扣減)為目前實作的解讀,待 M3/M4 跑流程時複核。
- **clamp 範圍只實作規格明寫者**:屬性修改命令目前僅 護盾 / 格擋 夾下限 0;其餘屬性(morale 對 moraleMax 上限、sate / calm 下限等)規格未明寫,未臆測,待跑流程時補。
- **§12 觸發次數的限制由呼叫方套用**:expr 只提供完整文法的 Parse / Eval。【十二、觸發次數】「採算術式、結果 < 0 或評估失敗 → 視為 0、結果 = 0 → 不觸發」的語意,由呼叫方(M3 觸發流程)在 Eval 之後套用,非 expr 內建。
- **企劃驗證器 = M2 後置支線**:運算式／命令語法驗證器(單筆輸入 + Effect 表掃描,同一引擎)。6 欄位中 4 個命令欄位(立即／觸發／啟動／結束命令)是命令文法、需 M2 命令解析器。**地基(共用帶位置錯誤型別)入 M2 驗收;工具本體為 M2 後置、非 gating 支線**(可與 M3 並行),不先出只驗運算式的半套。原「帶參數求值 + 運算過程」工具不做。詳見實作規格書【附錄：企劃驗證器】。錯誤定位地基已完成(見下方 `SyntaxError` 決策),M2 做命令解析器時直接重用。

## M2 工作盤點(屬性 + 命令)

**已落地**(分 6 個 commit:expr `none` → 命令解析 → 屬性註冊表 / Resolver → 屬性修改命令 → selector → 操作命令)。檔案:`internal/game/` 的 `command.go`(解析)/ `attr.go` `resolver.go`(讀)/ `commandassign.go`(屬性修改寫)/ `selector.go`(命令對象)/ `commandoperate.go`(操作命令)/ `trigger.go`(fireTrigger seam),`internal/expr/` 補 `none` 字面值。延後項見上方接續待辦。以下為原規劃分工(供對照)。

`defines` 已把 `Attr` / `AttrRef` / `Selector` / `Command` 全部識別碼立名完畢——M2 是「填行為」不是「立名」。規則細節一律指回規格【十七、命令】【二十三、屬性清單】【二十四、命令對象清單】【二十五、操作命令清單】,以下只記分工與界線。

工作包(五塊,依賴由淺到深):

- **A. 命令解析層(`command*.go` 前端)**:比照 `expr.Parse` 出可重複執行的靜態 AST,純語法、不碰 runtime。屬性修改命令文法(`<左值> <賦值符> [<算術式>]`,左值不帶 `()`)+ 操作命令文法(`<命令>(命令對象, 參數...)`,命令對象帶參數用 `[...]`,數值參數走 `expr`,varargs)。錯誤走 `newSyntaxError`。**此即企劃驗證器地基,M2 驗收條件**。
- **B. 屬性註冊表 + Resolver 讀取側**:game 端實作 `expr.Resolver`(目前尚無實作)。table-driven metadata(存取等級 / 類別 / clamp / 衍生讀寫 hook);`Attr` 蓋裸屬性 / 物件引用 / 查詢函式 / 衍生屬性 / `Lock` 後綴;`AttrRef` 蓋 dot-syntax 引用屬性。self 綁定來源 = 效果建立時固定的 `Self`。
- **C. 屬性修改命令 執行(寫入側)**:八賦值符語意、`@`/`#` 改 `Value.Lock`、`Lock>0` 帶值賦值 no-op、引用左值解析、衍生寫入轉譯(`roundLeft = N` → `roundMax`)、`morale -=` 特例(格擋→護盾→morale;最後「士氣受損」觸發點留 hook 給 M3)。
- **D. 命令對象解析(`selector*.go`)**:每個 `Selector` → `[]*Card` / `[]*Guest` 集合;容器來源 / filter / N 規則;`*Pick` 呼叫 `Operator`、`*Rand` 呼叫 `Rander`;`deckTop` auto-shuffle(唯一解析時改狀態)。self 系列效果建立時固定。
- **E. 操作命令 執行(`command*.go`)**:按 seam 切(見已敲定區「M2/M3 執行界線」)。M2 定義三個 no-op seam(`fireTrigger` / 啟動效果列表 / 清理效果),**所有命令本體在 M2 寫完**(容器搬移 + 屬性/事件更新如 drawLast,搬牌照呼叫 `fireTrigger`)。**只剩** 核心即效果佇列的命令 stub(`effectClear`/`effectDel`/`effectRun`,及 `cardRun`/`*Morph`/`cardify`/`restore` 的「啟動/清理效果」那一步)留給 M3。

決策點(均已收斂,見下方已敲定區):

1. ~~**`none` 字面值**~~ → A(字面值關鍵字)。
2. ~~**M2/M3 執行界線**~~ → A(按 seam 切)。
3. ~~**命令 AST 架構**~~ → A(編譯型 `Command`、parser 放 game)。

## 已敲定的設計決策(勿重新爭論)

- **expr 優先級 `NOT > AND > OR`**(非規格原 BNF 的同級);BNF 已修正為分層(邏輯或 → 邏輯且 → 否定)。
- **AND / OR 短路求值**:AND 左假不評右、OR 左真不評右(與三元「只評被選中分支」一致);左側本身評估失敗時仍整體失敗。守衛式 `self != none AND self.calm > 5` 因此成立。
- **expr 數值用 float64**:中間值可為小數,expr 內不四捨五入;捨入(half-away-from-zero)由呼叫方寫回屬性時以 `expr.Round` 處理。
- **Resolver 是 expr 與 game 的唯一接縫**,對齊【二十三、屬性清單】表結構:主表(全域屬性 / 查詢函式 / 物件引用)走 `Attr`、子表(卡牌 / 顧客引用屬性)走 `AttrRef`。`defines` 端的 `Property` / `Attr` 亦同步改名為 `Attr` / `AttrRef`。
- **內建函式註冊表化**:`builtin.go` 以 `map[string]builtinFunc` 登記內建函式(max / min),新增函式只需註冊一筆,`nodeCall` 與 parser 不動。對齊【二十六、內建函式清單】。
- **語法錯誤型別 `expr.SyntaxError`**(放 `expr` 套件,`error.go`):欄位 `Pos`(rune 索引,0 起算,供工具標位置)+ `Msg`(企劃白話,無 `expr:` 前綴);`Error()` 顯示「第 N 字附近:Msg」(N = Pos+1)。`token` 帶 `pos`,lexer/parser 全部錯誤皆走 `newError(pos, msg)`。**M2 命令解析器應重用同型別**,使命令與運算式的錯誤格式對企劃一致。
- **核心邊界介面定為三個**(Operator / Presenter / Rander),靜態資料直接以 `*sheeter.Sheeter` 注入;不再有 per-table Dater 介面。衍生索引(如 Award 依群組聚合)移入核心預建。
- **命令執行統一走 `Command.Execute(run *runner)`**:`commandAssign` / `commandOperate` 同一介面方法,供 M4 流程一致分派(不必按命令型別分簽章)。`runner` 聚合一次執行所需依賴(`rt` / `data` / `self` / `Operator` / `Rander`),命令對象解析(`resolveSelector`)與操作命令共用。屬性修改命令不需 Operator / Rander 但仍收 `runner`(忽略)。
- **屬性讀寫註冊表分讀 / 寫兩組**(非單一表雙用):讀側 `attrRead` / `attrRefRead`(+ `attrLockRead` / `attrRefLockRead`)服務 Resolver,鍵集含查詢函式 / 物件引用 / 衍生 / Lock 後綴;寫側 `attrWrite` / `attrRefWrite` 服務屬性修改命令,鍵集僅可寫屬性。理由:讀寫鍵集本就不同(唯讀屬性可讀不可寫),且存取等級以寫入 helper(`writeValue` 寫鎖 / `writeLockOnly` 鎖 / `writeInt` 寫)編碼,分表免去每筆雙寫。
- **命令 AST 架構**:比照 `expr.Parse` 出 parse-once 編譯型 `Command`(靜態資料載入時 parse 一次、多次執行);單一入口 `command.Parse(source) (Command, error)` 涵蓋兩種命令 + 內嵌 selector `[...]` + 內嵌 `*expr.Expr` 參數,錯誤走 `expr.SyntaxError`。命令**非遞迴**:depth-1 tagged 結構(`commandAssign` / `commandOperate`),非樹(每個參數本身才是 `*expr.Expr` tree)。**parser 放 `game`**(`command.go`,非新增第 4 個套件):理由——命令 token 本質是遊戲詞彙(動詞 / 命令對象 / 屬性,皆 `defines` enum),不像 expr 是領域中立語言,該緊鄰 game 且遵 §十「單一 game package、不拆子 package」。AST 為純資料(只引 defines enum + `*expr.Expr`,不含 runtime),executor 用 `switch` 直吃 `*Runtime`。驗證器 import `game`、僅呼叫純 `command.Parse`(不建 Runtime;link 進的執行碼是 dead weight,不影響純語法檢查)。
- **M2/M3 執行界線按 seam 切**(非按命令整族 stub):關鍵事實——連 `deckToHand` 都會觸發 system 時機(cardDraw / drawLast 更新),故搬牌命令一樣碰觸發,「按命令切」是假乾淨。真正接縫是三個 seam:`fireTrigger`(觸發派發)/ 啟動效果列表 [堆疊處理] / 清理效果,皆 **M3**。M2 把三 seam 定為 no-op stub(簽章穩定、M3 才填、呼叫點不回頭改),**所有命令本體在 M2 寫完並可測**(容器搬移 + 屬性/事件更新 + N 規則 / filter / deckTop auto-shuffle / 型別位置不符 no-op)。**只剩** 核心即效果佇列的命令留 M3:`effectClear`/`effectDel`/`effectRun`,及 `cardRun`/`*Morph`/`cardify`/`restore` 的「啟動 / 清理效果」步。M2 因此可獨立交付,唯一未測面是「正確效果有無被觸發」(本就 M3 職責)。
- **`none` 作 expr 字面值關鍵字**(非走 Resolver):`none` 與 `true`／`false` 並列為字面值,於 lexer `identToken` 特判、parser 出 `nodeLiteral{NewNone()}`。理由:空物件是值系統常數(永不失敗),該與 true/false 同層;`self` 走 Resolver 是因它是「會失敗的綁定」,語意不同。value 層(`NewNone` / `IsNone` / ref 比較「空物件只與空物件相等」)早已就位,只補文法入口。`self != none` 守衛式因此可寫;與既有 eval 規則自動相容(`none==self` 走 ref 比較、`none+1` 走非數值算術失敗)。待落地:lexer/parser/§27 BNF 與 §27.4 字面值表加 none、補測試。
