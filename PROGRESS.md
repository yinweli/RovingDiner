# 工作進度與接續記錄

本檔是跨 session 的接續工作交接點。每次工作告一段落時更新「里程碑進度」與「接續待辦」,下次開新 session 先讀本檔即可接手。

紀律:本檔只記「從 git 歷史與 `doc/` 規格看不出來的待辦與決策」。規則細節以 `doc/` 規格為準、程式現況以 git 為準,不在此重複;站內拍板一經成文於規格即自本檔刪除(過程細節 git log 可查;2026-06-11 大砍前的歷史拍板全文見 git 歷史)。

## 現況

- **架構已定案**(實作規格書 §一~§四):核心四包依賴 `games → rules → cores → exprs` 一條直線,`internal/infra` 基礎設施、`internal/tester` 跨包測試基建;TUI = `app/rodi`(停點直讀 + 交棒 stepper)+ `cmd/rodi` 瘦進入點。
- **M0–M27 已落地,M28(企劃驗證器)進行中(R1 完成)**:核心線(M0–M16)→ 事件流收口(M17–M18)→ TUI 被動觀看(M19–M22)→ 暫停機重構(M23)→ 日誌流換軌(M24)→ 速率(M25)→ 互動與 modal(M26)→ 格式說明 modal(M26A 插站)→ 選取＋Operator(M27,可全鍵盤跑完整一局)。建置 / golangci-lint(0 issues)/ 測試全綠,cores 與 rules 覆蓋率 100%(app/rodi 除 TTY 組裝入口 Run 外 100%)。
- **剩餘站序已凍結**(2026-06-11 拍板):M28 企劃驗證器 → M29 conformance golden;細目以實作規格書【九】為準,再插站不改號(M26A 格式說明 modal 即依此插站,後續編號未動)。

## 里程碑進度

對應 `doc/營業實作規格書.md`【九、里程碑】(細節以該處為準)。

| 里程碑  | 狀態 | 說明                                       |
|:--------|:-----|:-------------------------------------------|
| M0–M16  | ✅   | 核心線(exprs/cores/rules/games,跑通第一局) |
| M17–M18 | ✅   | 事件流收口(合約+全引擎發射點)              |
| M19–M22 | ✅   | TUI 被動觀看(殼/渲染地基/六區/事件日誌)    |
| M23     | ✅   | 暫停機重構(交棒 stepper+盤面直讀)          |
| M24     | ✅   | 日誌流換軌(行組 Emit+發射台+golden 常駐)   |
| M25     | ✅   | 速率(快/慢/步進排拍+世代驗章)              |
| M26     | ✅   | 互動與 modal(切區游標+框線網格+五 modal)   |
| M26A    | ✅   | 格式說明 modal(插站; F1 說明/F2 計數)      |
| M27     | ✅   | 選取＋Operator(交棒廣義化+選取/出牌模式)   |
| M28     | 🔶   | 企劃驗證器(R1~R3 完成,R4 表單掃描器起未動)  |
| M29     | ⬜   | conformance golden(擴充覆蓋) *(prov)*      |

## 接續待辦

- **clamp 範圍只做規格明寫者**:屬性修改僅 護盾 / 格擋 夾下限 0(`Value.Clamp`);其餘(morale 對 moraleMax 上限、sate / calm 下限等)規格未明寫,不臆測,跑流程時補。
- **[企劃驗證器 M28] 切站序(2026-06-12 拍板)**:R1 詞彙 metadata 形狀 + 命令對象 arity 校驗 → R2 `HasObjectRef` 引用基底補強 → R3 門檻配對共用 parse → R4 表單掃描器 → R5 `cmd/roditool` CLI + `task`/CI 接線。**R1、R2 完成**:metadata 形狀拍板為 map 值升級小 struct(具名函式 + metadata 欄,單一定義點;每詞條一具名函式不變)——R1 `selectorEntry`(arity 欄)+ `rules.SelectorArity`,`games.Validate` 依【二十四|參數規則】報數量錯誤;R2 `attrReadEntry`(ref 欄)+ `rules.HasObjectRef`,Validate 對引用左值先驗基底再驗屬性可寫性(`morale.cost` / 未知基底不再漏過)。**R3 完成**:cores 抽單筆 `ParseThreshold` 匯出為門檻配對格式單一定義點(錯誤用 `exprs.SyntaxError` 與兩文法同型、位置指向壞段起點),`prepareGuest` 寬鬆面改呼叫共用 parse 跳筆;驗證器嚴格面(R4 起)同一來源。執行期仍寬鬆 no-op 不變。參數**型別**靜態校驗不做:參數為算術式、值型別屬執行期知識(數值檢查仍由 oneInt / twoInt 執行期把關)。收站時 entry 機制成文於實作規格書。
- **[企劃驗證器 M28] 其餘**:門檻配對(`門檻值^技能編號`)的嚴格驗證要與 cores 寬鬆解析共用單一格式來源(抽共用 parse,勿雙寫格式知識);緩議:適用類型矩陣要不要驗(如立即類型填了 TriggerKind 報不報),開站時拍板。
- **[緩議] -race 未跑**:無 gcc 環境;嚴格交棒按構造無並行存取,雙跑同序測試為行為釘。

## 已敲定的設計決策(勿重新爭論;只列規格與程式看不出來的)

- **牌堆順序 = slice 前端為頂端**(index 0、最新進入者):`deckTop` / `dropTop` 取 `slice[:N]`、新進入者 prepend、auto-shuffle 洗後棄牌 append 至尾端(底部)。規格只定「新進入者置頂」未綁 slice 哪端;M8 起多站依賴,已鎖定不再改向。
- **rules 盤點全數否決**:維持每詞條一具名函式;詞條工廠化 / 合併 / 預建索引等十項盤點建議全數不採、不再重提。
- **C#/Unity 移植需求已撤銷**:架構不受可攜性約束,Go 慣用法自由採用。
- **規格未明處以實作現況為準**(exprs 比較 / 相等 / Truthy 語意、寫入語意走 Value 守衛方法等):規格補強時再對齊,不反向遷就臆測。
