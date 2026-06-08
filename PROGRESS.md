# 工作進度與接續記錄

本檔是跨 session 的接續工作交接點。每次工作告一段落時更新「里程碑進度」與「接續待辦」,下次開新 session 先讀本檔即可接手。

紀律:本檔只記「從 git 歷史與 `doc/` 規格看不出來的待辦與決策」。規則細節以 `doc/` 規格為準、程式現況以 git 為準,不在此重複。

## 現況

- **架構已定案、不再考慮 C# 移植**(實作規格書 §一/§二/§四/§十)。核心三包:`cores`(純資料模型,不懂規則、不 import `games`)+ `games`(全部營業邏輯,單向 import `cores`/`exprs`,持驅動引擎 struct)+ `exprs`(零遊戲依賴的運算式語言,可單獨測;builtin 與 `Resolver` 由 `games` 注入)。`infra` 為 Go-only 基礎設施。
- **實作重做**:先前在舊結構 `internal/{defines,expr,game}` 下完成的實作(骨架／運算式／屬性＋命令)將於新(細切)里程碑下重建,舊進度與工作盤點作廢。下方「已敲定設計決策」是跨重構仍成立、重建時要沿用的決策(套件名已對齊現行三包結構)。舊實作僅存於 `yilin/m2`(tip `d499db6`),僅供型別形狀參考(`dev` 已 reset 到當前工作路線、不再是舊實作來源)——以新規格 + 規則 SSOT 為準,避開單一 `game` 包耦合、全小寫檔名、C# 痕跡。
- **M0 已落地**:`internal/cores` 純資料模型骨架完成(define/type/instance/runtime + 測試),`internal/infra` 自頂層搬入 `internal/`、註解校正。建置 / vet / gofmt / golangci-lint / 測試全綠。
- **M2 已落地**:`internal/exprs` 詞法層完成——`error.go`(`SyntaxError` Pos+Msg、建構式 `newError(pos, msg)`、`Error()`「第 N 字附近」)、`lexer.go`(`tokenKind` 列舉 + `token` 帶 `pos` + `lex` 掃描器:數字 / 單引號字串含 CJK / 識別子 / 雙字元運算符 / `AND`-`OR`-`true`-`false`-`none` 大小寫不敏特判,尾端 `tokenEOF`)。`token` / `tokenKind` 暫置 `lexer.go`(M3 parser 同包共用,不需搬)。`=` 單字元、未結束字串、未知字元皆吐帶位置中文錯誤。`ParseFloat` 錯誤分支為防禦性(掃描器只組合合法數字串、實際不可達)。建置 / vet / gofmt / golangci-lint / 測試全綠。M1 之 infra 不在 exprs 依賴鏈。

## 里程碑進度

對應 `doc/營業實作規格書.md`【九、里程碑】(細節以該處為準)。全部依新(細切)里程碑重做;M0–M9 firm、M10 起 provisional(到站再細修)。

| 里程碑 | 狀態 | 說明                            |
|:-------|:-----|:--------------------------------|
| M0     | ✅   | cores 型別骨架                  |
| M1     | ✅   | infra.Load（搬入 internal/）    |
| M2     | ✅   | exprs lexer + SyntaxError       |
| M3     | ⬜   | exprs 純語言(parser/AST/eval)  |
| M4     | ⬜   | exprs 接縫(Resolver + builtin) |
| M5     | ⬜   | 命令解析(企劃驗證器地基)        |
| M6     | ⬜   | 屬性讀取側(registry + Resolver) |
| M7     | ⬜   | 屬性修改命令 執行               |
| M8     | ⬜   | 命令對象 selector               |
| M9     | ⬜   | 操作命令 執行(三 seam 切)      |
| M10    | ⬜   | 效果實例 + 佇列 *(prov)*        |
| M11    | ⬜   | 觸發派發 fireTrigger *(prov)*   |
| M12    | ⬜   | 堆疊處理 + 啟動效果列表 *(prov)* |
| M13    | ⬜   | 推進／清理／免疫 + 接回延後命令 *(prov)* |
| M14    | ⬜   | phase 狀態機 *(prov)*           |
| M15    | ⬜   | 啟動技能 *(prov)*               |
| M16    | ⬜   | 執行結算 → 跑通第一局 *(prov)*  |
| M17    | ⬜   | Presenter 顯示 *(prov)*         |
| M18    | ⬜   | Operator + adapter *(prov)*     |
| M19    | ⬜   | 速率(快/慢/步進) *(prov)*      |
| M20    | ⬜   | conformance golden *(prov)*     |

## 接續待辦

- **操作命令分三類、深度耦合效果系統(取代舊盤點)**:不要在操作命令階段(M9)硬切「整族 stub」——關鍵事實:連 `deckToHand` 都觸發 system 時機(cardDraw / drawLast),搬牌一樣碰觸發,真正接縫是三個 seam(見下條),皆屬效果系統。三類:① 實例化類 `*Add` / `*Copy` / `*Clone` / `*Roll` / `guestSpawn` / `waitAdd`(卡牌 / 顧客實例化資料模型待釐清,尤其新卡的實例效果列表來源——Card 表無效果欄)、② 處理流程類 `cardRun` / `*Morph` / `guestExit` / `guestReturn` / `guestSeat` / `cardify` / `restore`(多時機觸發 + 啟動 / 清理效果 + 座位指派)、③ 效果佇列類 `effectClear` / `effectDel` / `effectRun`(直接讀寫效果佇列 + 堆疊規則)。
- **三 seam**:`fireTrigger`(觸發派發)、啟動效果列表 [堆疊處理]、清理效果——皆屬效果系統(M11–M13)建;呼叫者(cardRun / morph / cardify / restore / guestExit)同期建,避免建了沒人呼叫被 unused lint。
- **morale `-=` 特例的來源**:Phase 4 命令路徑取 `self.Guest`(self 非顧客則空物件);guestExit(效果系統階段)走特例時改以「離場顧客」為來源。lock 與特例的交互(morale 鎖定時格擋 / 護盾仍消耗、morale 不動、無實際扣減)為目前解讀,效果／流程階段跑流程時複核。
- **clamp 範圍只做規格明寫者**:屬性修改僅 護盾 / 格擋 夾下限 0;其餘(morale 對 moraleMax 上限、sate / calm 下限等)規格未明寫,不臆測,跑流程時補。
- **§12 觸發次數限制由呼叫方套用**:`exprs` 只提供完整文法 Parse / Eval;【十二、觸發次數】「採算術式、結果 < 0 或評估失敗 → 視為 0、結果 = 0 → 不觸發」由呼叫方(觸發流程,M11)在 Eval 後套用,非 `exprs` 內建。
- **企劃驗證器 = 命令解析(M5)後置支線**:詳見實作規格書【附錄:企劃驗證器】。地基(共用帶位置錯誤型別 `exprs.SyntaxError`)入 M5 驗收;工具本體 M5 後置、可與效果系統並行。
- **[規格待補] §五 顧客實例漏列 `飽食值離場線 SateMax`**:§二十三 顧客引用屬性已把 `sateMax` 列為「寫鎖」(需 Value 儲存),但 §五 顧客實例表未列該欄。`cores.Guest` 已含 `SateMax Value`(被 §二十三 強制);請回頭在 §五 補上該列使兩處自洽。

## 已敲定的設計決策(勿重新爭論;重建時沿用)

> 套件名對照:原 `expr`→`exprs`、`game`→`cores` + `games`、`defines`→`cores/define.go`。

- **cores enum 一律 `XxxKind` 命名(比照 `exprs.tokenKind`)**:六個列舉 `EventKind` / `PhaseKind` / `TriggerKind` / `EffectKind` / `TargetKind` / `TaskKind`;**常數去掉中間字**(如 `EventInstance` / `PhaseNone` / `EffectImmed` / `TargetNone` / `TaskSate`,非 `EventKindInstance`),**struct 列舉欄位用裸 `Kind`**(`EventData.Kind` / `Action.Kind`,比照 `token.kind`)。其中 `EffectKind` 以 `Kind` 後綴避開效果實例 struct `Effect` 撞名(效果實例維持 `cores.Effect`)。**刻意與規格 / 生成碼分離**:規格書 §二 詞彙表與 `sheeter.Effect` 欄位仍用 `Type`(如 `Effect.Type` / `Effect.TriggerType` int32 原始欄位),cores 解碼後的列舉用 `Kind` 區隔;註解引用生成碼欄位(`Effect.Type` / `Effect.TriggerType`)與規格術語(`(TargetType)`)時保留原 `Type` 名。
- **`infra` 置於 `internal/`**:以實作規格書【二】為準,`internal/infra` 與 `internal/{exprs,cores,games}` 同層;原頂層 `infra/` + doc.go「非 internal」的舊決策作廢。`sheet`/`sheetdata`/`gamedata` 維持頂層(生成物)。
- **衍生索引歸 `games`**:award 等載入時衍生索引(未來 Seat 鄰桌、Skill→Effect 展開)一律歸 `games`(遊戲表衍生＝遊戲知識),`cores` 不放衍生索引、維持純資料模型;檔名 `games/help.go`。
- **exprs 注入機制**:builtin 由 `games` 注入、registry 掛 per-engine 實例,eval 時與 `Resolver` 一起帶入(eval-time env),零全域可變狀態;`exprs` 自身不內建函式、保持 game-agnostic。registry map vs 單一 `Env.Call` 委派、call-node 如何分流 builtin / 查詢函式,屬 exprs 接縫(M4)實作細節。
- **package 命名維持複數**:`exprs` / `cores` / `games`——避免 `expr` / `game` / `core` 遮蔽常見區域變數;為 package 名對單數鐵則(針對識別碼)的刻意例外。
- **exprs 優先級 `NOT > AND > OR`**(非規格原 BNF 的同級);BNF 已修正為分層(邏輯或 → 邏輯且 → 否定)。
- **AND / OR 短路求值**:AND 左假不評右、OR 左真不評右(與三元「只評被選中分支」一致);左側本身評估失敗時仍整體失敗。守衛式 `self != none AND self.calm > 5` 因此成立。
- **exprs 數值用 float64**:中間值可為小數,`exprs` 內不四捨五入;捨入(half-away-from-zero)由呼叫方寫回屬性時以 `exprs.Round` 處理。
- **`none` 作 exprs 字面值關鍵字**(非走 Resolver):與 `true` / `false` 並列,於 lexer `identToken` 特判、parser 出 `nodeLiteral{NewNone()}`。理由:空物件是值系統常數(永不失敗),該與 true / false 同層;`self` 走 Resolver 是「會失敗的綁定」,語意不同。`self != none` 守衛式因此可寫;與既有 eval 規則自動相容(`none == self` 走 ref 比較、`none + 1` 走非數值算術失敗)。
- **語法錯誤型別 `exprs.SyntaxError`**(`error.go`):`Pos`(rune 索引,0 起算,供工具標位置)+ `Msg`(企劃白話,無內部前綴);`Error()` 顯示「第 N 字附近:Msg」(N = Pos+1)。`token` 帶 `pos`,lexer / parser 全部錯誤走 `newError(pos, msg)`。**命令解析器重用同型別**,使命令與運算式的錯誤格式對企劃一致。
- **Resolver 是 `exprs` 與 `games` 的唯一接縫**,對齊【二十三、屬性清單】表結構:主表(全域屬性 / 查詢函式 / 物件引用)走 `Attr`、子表(卡牌 / 顧客引用屬性)走 `AttrRef`(識別碼立於 `cores/define`)。`Resolver` 由 `games` 實作。
- **內建函式註冊表化**:以 `map[string]builtinFunc` 登記(max / min),新增函式只需註冊一筆,`nodeCall` 與 parser 不動;由 `games` 注入 `exprs`。對齊【二十六、內建函式清單】。
- **核心邊界介面三個**(Operator / Presenter / Rander,定義於 `cores`),靜態資料直接以 `*sheeter.Sheeter` 注入;不設 per-table Dater 介面。衍生索引(如 Award 依群組聚合)移入核心預建。
- **命令執行統一走 `Command.Execute`**:屬性修改命令(`commandAssign`)與操作命令(`commandOperate`)同一介面方法,供流程階段一致分派。執行依賴聚合在 **`games` 驅動引擎 struct**(`rt` / `data` / `self` / `Operator` / `Rander`),命令對象解析(`resolveSelector`)與操作命令共用;屬性修改命令不需 Operator / Rander 但仍收引擎(忽略)。
- **屬性讀寫註冊表分讀 / 寫兩組**(非單表雙用):讀側 `attrRead` / `attrRefRead`(+ `attrLockRead` / `attrRefLockRead`)服務 Resolver,鍵集含查詢函式 / 物件引用 / 衍生 / Lock 後綴;寫側 `attrWrite` / `attrRefWrite` 服務屬性修改命令,鍵集僅可寫屬性。存取等級以寫入 helper(`writeValue` 寫鎖 / `writeLockOnly` 鎖 / `writeInt` 寫)編碼。屬性層歸 `games`。
- **命令 AST 架構**:比照 `exprs.Parse` 出 parse-once 編譯型 `Command`(靜態載入時 parse 一次、多次執行);單一入口 `command.Parse(source) (Command, error)` 涵蓋兩種命令 + 內嵌 selector `[...]` + 內嵌 `*exprs.Expr` 參數,錯誤走 `exprs.SyntaxError`。命令**非遞迴**:depth-1 tagged 結構(`commandAssign` / `commandOperate`),非樹(每個參數本身才是 `*exprs.Expr` tree)。AST 為純資料(只引 `cores` enum + `*exprs.Expr`,不含 runtime),executor 在 `games` 用 `switch` 直吃 `*cores.Runtime`。parser 放 `games`(命令 token 本質是遊戲詞彙——動詞 / 命令對象 / 屬性,皆 `cores/define` enum)。
- **操作命令／效果系統 執行界線按 seam 切**(非按命令整族 stub):真正接縫是三 seam(`fireTrigger` / 啟動效果列表 [堆疊處理] / 清理效果),皆屬效果系統;effect orchestration 整包在 `games`。操作命令階段(M9)把三 seam 定為穩定簽章,所有命令本體在 M9 寫完並可測(容器搬移 + 屬性 / 事件更新 + N 規則 / filter / deckTop auto-shuffle / 型別位置不符 no-op),只剩 核心即效果佇列的命令(`effectClear` / `effectDel` / `effectRun`,及 `cardRun` / `*Morph` / `cardify` / `restore` 的「啟動 / 清理效果」步)留效果系統階段。
