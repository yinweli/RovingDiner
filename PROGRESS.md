# 工作進度與接續記錄

本檔是跨 session 的接續工作交接點。每次工作告一段落時更新「里程碑進度」與「接續待辦」,下次開新 session 先讀本檔即可接手。

紀律:本檔只記「從 git 歷史與 `doc/` 規格看不出來的待辦與決策」。規則細節以 `doc/` 規格為準、程式現況以 git 為準,不在此重複。

## 現況

- **架構已定案、不再考慮 C# 移植**(實作規格書 §一/§二/§四/§十)。核心三包:`cores`(營業引擎本體＝資料模型 + 驅動引擎 + 命令語言行為,單向 import `exprs`、不 import `games`)+ `games`(對外介面層＝`Parse`/`Validate`/`Run` 三函式,單向 import `cores`/`exprs`)+ `exprs`(零遊戲依賴的運算式語言,可單獨測;`Resolver` 與 builtin 由 `cores` 實作/注入)。依賴呈 `games → cores → exprs` 一條直線。`infra` 為 Go-only 基礎設施。
- **實作重做**:先前在舊結構 `internal/{defines,expr,game}` 下完成的實作(骨架／運算式／屬性＋命令)將於新(細切)里程碑下重建,舊進度與工作盤點作廢。下方「已敲定設計決策」是跨重構仍成立、重建時要沿用的決策(套件名已對齊現行三包結構)。舊實作僅存於 `yilin/m2`(tip `d499db6`),僅供型別形狀參考(`dev` 已 reset 到當前工作路線、不再是舊實作來源)——以新規格 + 規則 SSOT 為準,避開單一 `game` 包耦合、全小寫檔名、C# 痕跡。
- **M0 已落地**:`internal/cores` 純資料模型骨架完成(define/type/instance/runtime + 測試),`internal/infra` 自頂層搬入 `internal/`、註解校正。建置 / vet / gofmt / golangci-lint / 測試全綠。
- **M2 已落地**:`internal/exprs` 詞法層完成——`error.go`(`SyntaxError` Pos+Msg、建構式 `newError(pos, msg)`、`Error()`「第 N 字附近」)、`lexer.go`(`tokenKind` 列舉 + `token` 帶 `pos` + `lex` 掃描器:數字 / 單引號字串含 CJK / 識別子 / 雙字元運算符 / `AND`-`OR`-`true`-`false`-`none` 大小寫不敏特判,尾端 `tokenEOF`)。`token` / `tokenKind` 暫置 `lexer.go`(M3 parser 同包共用,不需搬)。`=` 單字元、未結束字串、未知字元皆吐帶位置中文錯誤。`ParseFloat` 錯誤分支為防禦性(掃描器只組合合法數字串、實際不可達)。建置 / vet / gofmt / golangci-lint / 測試全綠。M1 之 infra 不在 exprs 依賴鏈。
- **M3 已落地**:`internal/exprs` 純語言層完成(零遊戲接縫、可獨立全測)——`value.go`(求值結果型別 `Value`:`valueKind` num / bool / text / none + 建構 `NewNum` / `NewBool` / `NewText` / `NewNone` + 取值 + `Truthy` §6 真假判定 + `Round` half-away-from-zero)、`node.go`(AST:`nodeLiteral` / `nodeUnary` / `nodeBinary` / `nodeTernary`,marker 介面 `node`)、`parser.go`(遞迴下降 `Parse(source) (*Expr, error)`;優先序 括號 > 乘除餘 > 加減 > 比較 > 否定 > AND > OR > 三元;三元右結合、其餘左結合;比較不串接)、`eval.go`(`(*Expr).Eval() (Value, ok)`、`evaluator` 走訪器、短路 AND / OR、三元惰性求值、算術 / 比較 / 邏輯)。建置 / vet / golangci-lint / 測試全綠。
- **M4 已落地**:`internal/exprs` 接縫層完成——`resolver.go`(`Resolver` 介面:主表 `Attr(name, arg)` / 子表 `AttrRef(ref, name, arg)`;`Ref` 介面 `Same`(比實例編號);`Builtin = func([]Value)(Value, bool)`;求值期環境 `Env{Resolver, Builtin}`)、`value.go` 新增 `valueRef` + `NewRef` / `IsRef` / `Ref` + 相等比較把 none / ref 歸為同一「物件」家族(`objectEqual`)、`node.go` 新增 `nodeIdent` / `nodeCall` / `nodeRef`、`parser.go` 的 `parsePrimary` 對 `tokenIdent` 擴充(屬性 / 物件引用 / 函式 / 引用屬性 / 引用查詢函式 + `parseArgs`)、`expr.go` 的 `Eval` 簽章改為 `Eval(env Env)`。內建函式分派優先於查詢函式;exprs 自身不內建任何函式(min / max 由 games 注入)。建置 / vet / golangci-lint / 測試全綠,覆蓋率 98%(其餘為防禦性死碼與 sealed marker)。
- **M5 已落地(全域詞彙表重做)**:`internal/games` 命令解析層採全域詞彙表架構——`parse.go`(純文法 `Parse(source) (Command, error)`:`Command` 封閉介面 + `commandAssign`(全域 / 引用左值 + 名稱位置、`assignKind`、右值 `*exprs.Expr`)/ `commandOperate`(verb + 位置、`selectorArg` + `[...]` 參數、varargs);讀首個識別子後接 `(` → 操作命令、否則屬性修改;**名稱以原始字串擷取、parser 不驗成員**;內嵌算術式深度掃描(計 `()` `[]`、跳單引號字串)定界後委由 `exprs.Parse`、`offsetError` 回算位置、共用 `exprs.SyntaxError`)。`exec{Attr,AttrRef,Selector,Command}.go`(各 concern 檔自持套件層**全域詞彙表** `tableAttr` / `tableAttrRef` / `tableSelector` / `tableCommand` 四張 map[常數、只裝純函式、M5 空]與其行為)+ `validate.go`(自由函式 `Validate` 逐名查表[M5 空表故一律未知、M6+ 生效];**無 Lexicon struct / 無建構**)。`engine.go`(engine 骨架,委派實作 `exprs.Resolver`、讀全域詞彙表)。**移除** `cores` 的 `Attr` / `AttrRef` / `Command` / `Selector` 列舉與 `lookup.go`。建置 / vet / golangci-lint(0 issues)/ 測試全綠,games 覆蓋率 97.3%(其餘為 sealed marker、engine 委派 happy-path[M5 空表未達]、防禦性後備)。
- **架構調整:引擎本體下沉 `cores`、`games` 縮為對外層(本 session)**:`engine` 與命令語言四詞彙行為檔自 `games` 下沉 `cores`、`exec*.go` 更名 `attr.go` / `attrRef.go` / `command.go` / `selector.go`(各出 `HasAttr` / `HasAttrRef` / `HasCommand` / `HasSelector` 供查名),與 `engine` 同包故行為直接吃 `*engine`、免狀態存取 port;`games` 縮為 `Parse` / `Validate` / `Run` 三對外函式(`Validate` 改查 `cores.HasCommand` / `HasSelector`)。依賴改 `games → cores → exprs` 單一直線,`cores` 不再純資料、為引擎本體;取代原『模型 B:cores 純資料 + games 全部邏輯』。未來 `flow*.go` / `effect*.go` / `help.go` / `builtin*.go` 一併落 `cores`。詳見實作規格書 §二/§三/§十。

## 里程碑進度

對應 `doc/營業實作規格書.md`【九、里程碑】(細節以該處為準)。全部依新(細切)里程碑重做;M0–M9 firm、M10 起 provisional(到站再細修)。

| 里程碑 | 狀態 | 說明                            |
|:-------|:-----|:--------------------------------|
| M0     | ✅   | cores 型別骨架                  |
| M1     | ✅   | infra.Load（搬入 internal/）    |
| M2     | ✅   | exprs lexer + SyntaxError       |
| M3     | ✅   | exprs 純語言(parser/AST/eval)  |
| M4     | ✅   | exprs 接縫(Resolver + builtin) |
| M5     | ✅   | 命令解析(企劃驗證器地基)        |
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
- **企劃驗證器工具本體 = M5 後置支線**:命令解析地基已隨 M5 落地(`games` 的 `parse.go` 重用 `exprs.SyntaxError`、吐帶位置中文錯誤;名稱合法性由自由 `Validate` 查詞彙表 keys、M6+ 詞條到位後生效);工具本體(CLI＋Effect 表掃描器＋`task`／CI)依賴 M5、可與效果系統並行,尚未做。詳見實作規格書【附錄:企劃驗證器】。
- **[規格待補] §五 顧客實例漏列 `飽食值離場線 SateMax`**:§二十三 顧客引用屬性已把 `sateMax` 列為「寫鎖」(需 Value 儲存),但 §五 顧客實例表未列該欄。`cores.Guest` 已含 `SateMax Value`(被 §二十三 強制);請回頭在 §五 補上該列使兩處自洽。

## 已敲定的設計決策(勿重新爭論;重建時沿用)

> 套件名對照:原 `expr`→`exprs`、`game`→`cores` + `games`、`defines`→`cores/define.go`。

- **全面統一 `Kind` 命名(比照 `exprs.tokenKind`)— 三層皆已落地**:cores 六個列舉 `EventKind` / `PhaseKind` / `TriggerKind` / `EffectKind` / `TargetKind` / `TaskKind`;**常數去掉中間字**(如 `EventInstance` / `PhaseNone` / `EffectImmed` / `TargetNone` / `TaskSate`,非 `EventKindInstance`),**struct 列舉欄位用裸 `Kind`**(`EventData.Kind` / `Action.Kind`,比照 `token.kind`)。其中 `EffectKind` 以 `Kind` 後綴避開效果實例 struct `Effect` 撞名(效果實例維持 `cores.Effect`)。①**生成碼**:Effect 表(`gamedata/Effect.xlsx` row3)三個類型欄位 効果類型 / 目標類型 / 觸發類型 全改 `Kind` / `TargetKind` / `TriggerKind`,`sheeter.Effect` 同步(`Effect.Kind` 依字母序排在 `ID`/`Name` 間);`effect.json` 為空、不受影響。②**規格三書**:`營業規格書` §二 詞彙表 + §五、`營業實作規格書` prose、`營業顯示規格書` 欄位表全部 Type→Kind,已 `task lint` + `task doc` 重建 HTML。**唯一保留的 `Type`**:`營業顯示規格書` §「適用類型矩陣」泛指的 `Type`(實例適用類型概念、非欄位),刻意不動。
- **`infra` 置於 `internal/`**:以實作規格書【二】為準,`internal/infra` 與 `internal/{exprs,cores,games}` 同層;原頂層 `infra/` + doc.go「非 internal」的舊決策作廢。`sheet`/`sheetdata`/`gamedata` 維持頂層(生成物)。
- **衍生索引歸 `cores`**:award 等載入時衍生索引(未來 Seat 鄰桌、Skill→Effect 展開)隨引擎本體歸 `cores`(引擎初始化時建一次快取);檔名 `cores/help.go`。
- **exprs 注入機制**:builtin 由 `cores` 注入;屬性詞彙綁定收在套件層**全域詞彙表**(常數;name→行為、runtime 經 `engine` 當參數帶入),`engine` 委派實作 `Resolver`、eval 時與其一起帶入(eval-time env);全域表只裝純函式、零 run-state、建後唯讀、等同常數,不破壞決定性;`exprs` 自身不內建函式、保持 game-agnostic。registry map vs 單一 `Env.Call` 委派、call-node 如何分流 builtin / 查詢函式,屬 exprs 接縫(M4)實作細節。
- **package 命名維持複數**:`exprs` / `cores` / `games`——避免 `expr` / `game` / `core` 遮蔽常見區域變數;為 package 名對單數鐵則(針對識別碼)的刻意例外。
- **exprs 優先級 `NOT > AND > OR`**(非規格原 BNF 的同級);BNF 已修正為分層(邏輯或 → 邏輯且 → 否定)。
- **AND / OR 短路求值**:AND 左假不評右、OR 左真不評右(與三元「只評被選中分支」一致);左側本身評估失敗時仍整體失敗。守衛式 `self != none AND self.calm > 5` 因此成立。
- **exprs 數值用 float64**:中間值可為小數,`exprs` 內不四捨五入;捨入(half-away-from-zero)由呼叫方寫回屬性時以 `exprs.Round` 處理。
- **`none` 作 exprs 字面值關鍵字**(非走 Resolver):與 `true` / `false` 並列,於 lexer `identToken` 特判、parser 出 `nodeLiteral{NewNone()}`。理由:空物件是值系統常數(永不失敗),該與 true / false 同層;`self` 走 Resolver 是「會失敗的綁定」,語意不同。`self != none` 守衛式因此可寫;與既有 eval 規則自動相容(`none == self` 走 ref 比較、`none + 1` 走非數值算術失敗)。
- **語法錯誤型別 `exprs.SyntaxError`**(`error.go`):`Pos`(rune 索引,0 起算,供工具標位置)+ `Msg`(企劃白話,無內部前綴);`Error()` 顯示「第 N 字附近:Msg」(N = Pos+1)。`token` 帶 `pos`,lexer / parser 全部錯誤走 `newError(pos, msg)`。**命令解析器重用同型別**,使命令與運算式的錯誤格式對企劃一致。
- **exprs 純語言 API 與失敗分層(M3)**:`Parse(source) (*Expr, error)` 載入時編譯一次、`(*Expr).Eval() (result Value, ok bool)` 執行期多次求值。**兩種失敗分層**:語法錯誤於 Parse 期回 `SyntaxError`、評估失敗於 Eval 期以 `ok == false` 表達且不拋錯(對齊【二十七、運算式｜7、8】)。求值結果型別 `exprs.Value`(num / bool / text / none;與屬性容器 `cores.Value` 為不同概念,同名不同包)。`Round`(half-away-from-zero)由呼叫方寫回屬性時使用,exprs 求值過程本身不捨入。
- **exprs 規格未明處的求值語意(敲定;規格若補強再對齊)**:① 大小比較 `< > <= >=` 僅 num 對 num,字串 / 布林 / 空物件無順序 → 評估失敗;② 相等 `== !=` 僅同型別(num / text / bool / none)可比,跨型別 → 失敗,空物件只與空物件相等;③ §6「省略比較符」真假判定(布林取用、非 0 數值為真、0 為假、其餘失敗)集中於 `Value.Truthy`,邏輯運算元 / 三元條件 / 呼叫方頂層判定共用同一規則。
- **M3→M4 接縫邊界**:操作數解析(識別子屬性 / 引用 / 查詢函式 / self)屬 M4(里程碑「操作數解析」),M3 parser 對 `tokenIdent` 直接吐「非法運算元」錯。`evaluator` 目前為空 struct,M4 加 `resolver` / `builtin` 欄位、遞迴 eval 方法簽章不動(走訪器收斂改動面);`parsePrimary` 對 `tokenIdent` 擴充出 nodeIdent / nodeCall / nodeRef;`none` 為 M3 唯一 ref-like 值,M4 加 `valueRef` 與「none vs 物件引用比實例編號」。
- **lint 政策:`exhaustive` 帶 default 即窮舉**:`.golangci.yml` 新增 `exhaustive.default-signifies-exhaustive: true`——parser / evaluator 對 `tokenKind` 只 switch 相關子集 + `default`,不逐一列 25 個 token;M4+ 命令 verb / selector / cores 列舉 switch 沿用。(M2 lexer 是 switch rune / string 故未觸發,M3 首度 switch 列舉型別。)
- **Resolver 是 `exprs` 與 `games` 的唯一接縫**,對齊【二十三、屬性清單】表結構:主表(全域屬性 / 查詢函式 / 物件引用)走 `Attr`、子表(卡牌 / 顧客引用屬性)走 `AttrRef`(識別碼立於 `cores/define`)。`Resolver` 由 `cores` 實作。
- **exprs 接縫實作定案(M4)**:`Resolver` 兩方法 `Attr(name string, arg []Value)`(主表:全域屬性 arg 空 / 查詢函式 arg 帶參 / 物件引用回 ref)、`AttrRef(ref Ref, name string, arg []Value)`(子表:引用屬性 arg 空 / 引用查詢函式 arg 帶參);`Ref` 介面僅 `Same(other Ref) bool`——exprs 不解讀引用內容、只在 `==` / `!=` 比實例編號;`Builtin = func([]Value)(Value, bool)`,以 `map[string]Builtin` 注入。求值期環境 `Env{Resolver, Builtin}` 由 `Eval(env)` 帶入、零全域可變狀態,`Env{}` 即純語言求值(遇條件對象 / 函式即失敗)。**函式分流**:`name(args)` 先查 builtin 註冊表、未命中才走 `Resolver.Attr`(查詢函式)——builtin 優先、同名 builtin 勝。**引用存取**:`name.attr` / `name.attr(args)` 先 `Attr(name)` 取主體,主體須為 ref(空物件 / 型別不符即失敗),再 `AttrRef`。**`self`** 走 `Attr("self")`:綁定顧客 / 卡牌回 ref、綁定空物件回 none、未綁定回失敗;守衛式 `self != none AND self.calm > 5` 因短路而成立。**參數**為算術式層級(`parseAdd`);`name()` 空參數結構上可解析,數量 / 型別由 builtin / resolver 於求值期判定(非 parser)。**比較**:`none` 字面值、解析所得空物件、物件引用同屬「物件」家族(`objectEqual`:空只等空、兩 ref 比 `Same`、一空一非空為不等)。
- **內建函式註冊表化**:以 `map[string]builtinFunc` 登記(max / min),新增函式只需註冊一筆,`nodeCall` 與 parser 不動;由 `cores` 注入 `exprs`。對齊【二十六、內建函式清單】。
- **核心邊界介面三個**(Operator / Presenter / Rander,定義於 `cores`),靜態資料直接以 `*sheeter.Sheeter` 注入;不設 per-table Dater 介面。衍生索引(如 Award 依群組聚合)移入核心預建。
- **命令執行 = engine 對 AST 型別 switch + 全域詞彙表分派**(非 AST 上的 `Command.Execute` 方法):`commandAssign` 走 `tableAttrWrite` / `tableAttrRefWrite`、`commandOperate` 走 `tableSelector`(命令對象)+ `tableCommand`(verb)。詞條行為函式以 `engine` 為 context(持 `runtime` / `data` / `self` / `Operator` / `Rander`,M6+ 按需補);屬性修改命令不需 Operator / Rander 但仍收 engine(忽略)。
- **屬性讀寫註冊表分讀 / 寫兩組**(非單表雙用):讀側 `attrRead` / `attrRefRead`(+ `attrLockRead` / `attrRefLockRead`)服務 Resolver,鍵集含查詢函式 / 物件引用 / 衍生 / Lock 後綴;寫側 `attrWrite` / `attrRefWrite` 服務屬性修改命令,鍵集僅可寫屬性。存取等級以寫入 helper(`writeValue` 寫鎖 / `writeLockOnly` 鎖 / `writeInt` 寫)編碼。屬性層歸 `cores`。
- **命令 AST 架構**:比照 `exprs.Parse` 出 parse-once 編譯型 `Command`(靜態載入時 parse 一次、多次執行);純文法入口 `Parse(source) (Command, error)`,涵蓋兩種命令 + 內嵌 selector `[...]` + 內嵌 `*exprs.Expr` 參數,錯誤走 `exprs.SyntaxError`。命令**非遞迴**:depth-1 tagged 結構(`commandAssign` / `commandOperate`),非樹(每個參數本身才是 `*exprs.Expr` tree)。AST 為純資料(名稱為原始字串 + 位置 + `*exprs.Expr`,不含 runtime),executor 在 `cores` 用 `switch` 經全域詞彙表 map 分派。parser(`parse.go`)放 `games`、純文法不驗成員;命令詞彙(動詞 / 命令對象 / 屬性名)由 `cores` 套件層**全域詞彙表**持有,名稱合法性由 `games` 的自由 `Validate` 查 `cores` 表。
- **操作命令／效果系統 執行界線按 seam 切**(非按命令整族 stub):真正接縫是三 seam(`fireTrigger` / 啟動效果列表 [堆疊處理] / 清理效果),皆屬效果系統;effect orchestration 整包在 `cores`。操作命令階段(M9)把三 seam 定為穩定簽章,所有命令本體在 M9 寫完並可測(容器搬移 + 屬性 / 事件更新 + N 規則 / filter / deckTop auto-shuffle / 型別位置不符 no-op),只剩 核心即效果佇列的命令(`effectClear` / `effectDel` / `effectRun`,及 `cardRun` / `*Morph` / `cardify` / `restore` 的「啟動 / 清理效果」步)留效果系統階段。
- **命令解析定案(M5,全域詞彙表重做)**:純文法入口 `Parse(source) (Command, error)`(已匯出);名稱合法性由自由函式 `Validate(Command) error` 查全域詞彙表 keys。`Command` 封閉介面、實作 `commandAssign` / `commandOperate`(depth-1 tagged、非樹;內嵌參數為 `*exprs.Expr`);**名稱以原始字串擷取(含位置 `basePos` / `refAttrPos` / `verbPos` / `selectorPos` 供 Validate 報位置)、parser 不驗成員**;執行於 M7 / M9 補。**兩類命令以語法區分**:首個識別子後接 `(` → 操作命令、否則屬性修改;左值引用 vs 全域以有無 `.<屬性>` 區分。**賦值符** `assignKind`(set/add/sub/mul/div/mod/lock/unlock);`@` / `#` 不帶右值。**內嵌算術式**深度掃描(`()` `[]` 計深度、單引號字串內標點不計,使字串參數與巢狀函式逗號不被誤判)定界後委由 `exprs.Parse`,`offsetError` 回算位置。**名稱合法性移出 parser**、改由自由 `Validate` 查**套件層全域詞彙表** keys(取代已刪的 cores 列舉 + `lookup.go`);三層分離 parse(純)/ validate(查 keys)/ execute(engine+行為)。**不設 Lexicon struct**:詞彙固定、不需可注入,全域常數表足矣,省掉 struct / 建構 / 擁有權。**M5 全域詞彙表為空骨架**:`tableAttr` / `tableAttrRef`(engine 委派 `Resolver` 用)/ `tableSelector` / `tableCommand`(`Validate` 用)四張 map(`cores` 的 `attr.go` / `attrRef.go` / `command.go` / `selector.go` 各持其一),詞條於 M6(讀)/ M7(寫)/ M8(selector)/ M9(command)以 concern 檔 map 字面值填入;寫側留 M7。**Validate M5 邊界**:只驗 verb / 命令對象存在(空表故 M5 一律未知、M6+ 生效);屬性可寫性留 M7、`[N]` / 參數數量留 M8 / M9——故 `self = 5`、`tableCount = 1`、裸 `guestPick`、空 `deckTop[]` 於 M5 皆「結構合法、語意檢查後置」。
