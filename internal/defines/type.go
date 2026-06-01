// Package defines 定義營業核心的共用常數、列舉與型別別名。
//
// 本套件零遊戲邏輯、零外部依賴（僅標準庫），是可移植核心邊界的最底層；
// 內容對應規格書【二、英文詞彙對照】與各機制清單，作為事件命名、屬性引用、
// 命令、命令對象、欄位與程式碼識別碼的單一命名來源。
//
// 檔案分工：型別定義集中於 type.go，常數 / 資料定義集中於 define.go。
package defines

// InstanceID 實例的唯一識別碼；卡牌實例 / 顧客實例 / 效果實例共用。
// 對應規格書【五、實例結構】各實例的「實例編號」欄位。
type InstanceID int64

// Event 投影事件類別；對應營業實作規格書【四之二、投影事件分類】。
type Event int

// Phase 營業的執行階段；對應規格書【十九、核心流程】。
//
// 值採用規格書使用的中文字面值：全域屬性「下一階段」(NextPhase) 即以此型別儲存，
// 合法跳轉值為 玩家行動 / 顧客行動 / 回合結束 / 空字串（無跳轉）。
type Phase string

// Trigger 觸發時機；對應規格書【二十二、觸發時機清單】。
//
// 觸發時機僅作為時間訊號、不攜帶資料。值採用規格書的英文名稱字面值。
// 效果靜態表格 Effect.TriggerType 為 int32 編碼，其與本型別的對照於載入器（M3）建立。
type Trigger string

// Effect 效果類型；對應規格書【七、效果類型】。
type Effect int32

// TargetType 目標類型；對應規格書【八、目標類型】。
type TargetType int32

// TriggerAfter 觸發後行為；對應規格書【十三、觸發後行為】（僅觸發類型適用）。
type TriggerAfter int32

// StackTime 堆疊時間；對應規格書【十六、堆疊規則】。
type StackTime int32

// Task 行動類型；對應規格書【五、實例結構｜行動（Action）實例】。
// 於顧客行動階段決定封印閘門查詢對象（飽食 / 耐心）。
type Task int32

// Selector 命令對象的英文名稱；對應規格書【二十四、命令對象清單】。
//
// 命令對象解析「這次命令作用到誰」，恆解析為集合（可能為空 / 多個）。
// 刻意不命名為 target，以免與【八、目標類型】(TargetType) 撞名。
type Selector string

// Property 全域屬性名稱；對應規格書【二十三、屬性清單】主表。
//
// 本表雙用途：作為【十一、觸發條件】的條件對象讀取來源，
// 亦作為【十七、命令｜屬性修改命令】的可寫屬性清單。
// 名稱帶 (...) 的查詢函式（deckSize / drawTotal / tableCount …）此處僅登記其字根名稱，
// 參數解析由 expr 引擎（M1）/ 屬性註冊表（M2）處理。
type Property string

// Attr dot-syntax 引用屬性名稱；對應規格書【二十三、屬性清單｜卡牌引用屬性 / 顧客引用屬性】。
// 以 <引用>.<屬性> 形式存取（如 self.sate、drawLast.cardID、taskGuest.calm）。
type Attr string

// Command 動作類命令的英文名稱；對應規格書【二十五、動作類命令清單】。
//
// 屬性修改命令（運算符語法 = += -= *= /= %= @ #）不在此清單，
// 其可寫屬性見【二十三、屬性清單】與 Property。
type Command string
