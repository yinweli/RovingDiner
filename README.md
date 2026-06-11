# RovingDiner（流浪食堂）

![lint](https://github.com/yinweli/RovingDiner/actions/workflows/lint.yml/badge.svg)
![test](https://github.com/yinweli/RovingDiner/actions/workflows/test.yml/badge.svg)

《流浪食堂 RovingDiner》「營業」階段的回合制規則引擎，以 Go 撰寫。

一次營業是一場可被結算的遊戲流程：玩家透過手牌與技能處理顧客，系統依照回合推進，過程中處理觸發時機、效果佇列與結算規則。

## 系統概述

- **營業目標**：處理完所有顧客（座位列表、排隊佇列、遊蕩列表、卡牌化列表）、維持餐廳士氣值大於 0、並在回合達到上限前完成營業。
- **核心流程**：營業開始 → 回合開始 → 玩家行動 → 顧客行動 → 回合結束 → 營業成功／失敗。玩家行動以卡牌為核心（補牌、出牌、結束行動），顧客行動以行動佇列為核心。
- **核心機制**：卡牌、顧客、技能、效果、命令五大機制，透過條件對象、命令對象與觸發時機構成資料驅動的規則介面。
- **資料驅動**：靜態表格定義設定、座位、抽獎、卡牌、顧客、技能與效果；卡牌、顧客與效果皆以執行期實例運作，同一份靜態資料可在營業中產生多個獨立狀態。

## 架構重點

營業核心是一個**純邏輯、單執行緒、決定性（deterministic）**的狀態機，所有對外互動都走邊界介面。四條鐵則：

1. **核心零 I/O、零框架依賴**：`game` 核心只依賴介面，不碰終端機、goroutine 或全域時間／亂數。
2. **決定性**：同 `seed` + 同玩家輸入 = 同一局，所有隨機來源共用單一注入的 seeded PRNG，作為回歸測試與 bug 重現的基礎。
3. **yield-per-unit 事件流**：核心每跑一個單位（一個命令／觸發／phase 切換／玩家動作）就吐一個事件；快速／慢速／步進只是同一條事件流的不同消費速率。
4. **單一真相＋停點直讀**：引擎是唯一真相，持有全部實例與容器；前端（TUI）在引擎停點間唯讀直讀盤面渲染，不持有規則狀態的第二份拷貝；事件流供日誌敘事、步進節拍與重現比對。

## 顯示／操作層

以 Go + 終端機介面（TUI）實作，採用 **Bubble Tea + Lip Gloss + Bubbles**（Elm 架構）。定位為 **debug viewer 而非 game UI**，大量曝露內部狀態（效果佇列、行動佇列、觸發時機、命令日誌、各容器即時內容），用以驗證 spec 的可實作性。渲染政策限定 ASCII + CJK，避免歧義寬度符號造成框線歪斜。

## 設計文件（`doc/`）

| 文件                | 角色                                                                                         |
|:--------------------|:---------------------------------------------------------------------------------------------|
| `營業規格書.md`     | **規則 SSOT**：營業階段的完整規則、核心流程、機制、觸發時機、結算與英文詞彙對照。            |
| `營業實作規格書.md` | **專案架構與工程決策**：套件結構、核心引擎（事件流、PRNG、決定性）、邊界介面解耦、測試策略。 |
| `營業顯示規格書.md` | **TUI 顯示／操作層**：Bubble Tea debug viewer、事件流消費、畫面佈局、互動與渲染政策。        |

**網頁版（GitHub Pages）**：

- [營業規格書](https://yinweli.github.io/RovingDiner/doc/營業規格書.html)
- [營業實作規格書](https://yinweli.github.io/RovingDiner/doc/營業實作規格書.html)
- [營業顯示規格書](https://yinweli.github.io/RovingDiner/doc/營業顯示規格書.html)

文件層級：`營業規格書.md` 為規則 SSOT；`營業實作規格書.md` 定義引擎與架構（含事件流定義）；`營業顯示規格書.md` 僅描述顯示層如何消費引擎。實作文件若與規則文件衝突，以規則文件為準。

## 專案結構

| 路徑         | 內容                                                            |
|:-------------|:----------------------------------------------------------------|
| `gamedata/`  | xlsx 來源表格與 Sheeter 建置設定（`sheeter.yaml`、`build.bat`） |
| `sheet/`     | Sheeter 生成的 Go 讀取器（自動生成，請勿手動編輯）              |
| `sheetdata/` | Sheeter 生成的 JSON 資料（自動生成，請勿手動編輯）              |
| `doc/`       | 設計文件                                                        |

## 開發指令

```bash
task lint       # 格式化與 lint（golangci-lint fmt + run、markdownlint、prettier）
task sheet      # 由 gamedata/*.xlsx 重新生成 sheet 程式碼與資料，再 lint
task install    # 安裝開發工具（golangci-lint、sheeter、markdownlint、prettier）
```

### Sheet 資料管線

`sheet/` 與 `sheetdata/` 由 [Sheeter](https://github.com/yinweli/Sheeter) 從 `gamedata/` 的 xlsx 檔生成。若要修改遊戲資料，請編輯 xlsx 檔（Card、Effect、Guest、Seat、Setting、Skill）後執行 `task sheet`，切勿直接修改生成的 `.go` 或 `.json` 檔。

## 開發狀態

營業核心（`exprs` 運算式語言、`cores` 引擎本體、`rules` 詞彙與流程、`games` 對外介面）已完成並可跑通完整一局，事件流（合約＋全引擎發射點）已收口。TUI debug viewer 進行中：殼＋橋接、渲染地基、六區盤面與事件日誌（alt-screen 全畫面、行角色上色）已落地，並完成暫停機重構（引擎逐拍交棒推進、盤面停點直讀、世界鏡像退役）；互動與速率控制後續疊上，其後為企劃驗證器與 conformance golden 測試。里程碑進度詳見 `PROGRESS.md`。
