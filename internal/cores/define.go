package cores

// 這三個介面是核心對外的唯一行為邊界；TUI 與測試各自提供實作。
// 對應【營業實作規格書 | 四、解耦的關鍵：邊界介面】。
//
// 靜態表格不走介面：核心直接吃 Sheeter 雙語言生成的 *sheeter.Sheeter（資料 port 本身），
// 以 reader.Get 查詢；衍生索引由核心（games）預建。
//
// 核心完全同步、單執行緒、無 channel：需要玩家輸入時阻塞呼叫 Operator，
// 每跑一個單位呼叫 Presenter.Emit。goroutine + channel 只活在 TUI adapter。
// 三 port 定義於 cores（純資料模型）；games 驅動引擎與 infra / TUI 各自實作。

// Rander 唯一亂數來源（單一 seeded PRNG）；決定性的基礎。
type Rander interface {
	Intn(n int) int                     // 回傳 [0, n) 的隨機整數
	Shuffle(n int, swap func(i, j int)) // 對 n 個元素洗牌
	Weighted(weight []int32) int        // 抽獎 weighted random；回傳命中索引
}

// Operator 玩家輸入；對應所有「暫停流程」點。
type Operator interface {
	PlayerAction(runtime *Runtime) Action          // 玩家行動：出牌 / 結束
	PickGuest(source []*Guest, count int) []*Guest // 新選顧客 / guestPick / nearPick / samePick
	PickCard(source []*Card, count int) []*Card    // 新選手牌 / *Pick 牌堆類
	PickDiscard(source []*Card, over int) []*Card  // 手牌上限棄牌
}

// Presenter 事件流輸出（yield-per-unit；步進時可阻塞）。
type Presenter interface {
	Emit(eventData EventData) // 一個命令 / 一次觸發 / 一次 phase 切換 / 一個玩家動作 = 一個 EventData
}

// EventData 核心吐給前端的細粒度投影事件。
// 對應【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 四之二】投影事件分類。
// TODO(M17)：各類別的詳細欄位（目標 InstanceID、屬性前後值、表演資訊）於 Presenter 顯示層實作時補齊，
// 依【營業顯示規格書 | 3、事件流的消費：速率與步進】與【營業規格書 | 十八、表演資訊】。
type EventData struct {
	Kind EventKind // 投影事件類別
}

// InstanceID 實例的唯一識別碼；卡牌實例 / 顧客實例 / 效果實例共用。
// 對應【營業規格書 | 五、實例結構】各實例的「實例編號」欄位。
type InstanceID int64

// NoneID 空實例編號；表示尚未配發或不存在的實例。
const NoneID InstanceID = 0

// EventKind 投影事件類別；對應【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 四之二】投影事件分類。
type EventKind int

const (
	EventInstance  EventKind = iota // 實例建立 / 銷毀（卡牌 / 顧客 / 效果）
	EventContainer                  // 卡牌移動牌堆、顧客入座 / 離場 / 遊蕩 / 卡牌化
	EventProperty                   // sate / calm / score / 鎖定計數 變化
	EventTrigger                    // 觸發時機到達、某效果觸發
	EventPhase                      // phase 切換
)

// PhaseKind 營業的執行階段；對應【營業規格書 | 十九、核心流程】。
//
// 值採用規格書使用的中文字面值：全域屬性「下一階段」(NextPhase) 即以此型別儲存，
// 合法跳轉值為 玩家行動 / 顧客行動 / 回合結束 / 空字串（無跳轉）。
type PhaseKind string

const (
	PhaseNone         PhaseKind = ""     // 無階段 / 無跳轉（下一階段清除後的值）
	PhaseGameStart    PhaseKind = "營業開始" // 營業開始階段；啟動前置技能
	PhaseRoundStart   PhaseKind = "回合開始" // 回合開始階段；回合數遞增、入座
	PhasePlayerAction PhaseKind = "玩家行動" // 玩家行動階段；補牌 / 出牌 / 結束
	PhaseGuestAction  PhaseKind = "顧客行動" // 顧客行動階段；彈出行動佇列
	PhaseRoundEnd     PhaseKind = "回合結束" // 回合結束階段；推進效果、執行結算
	PhaseGameSucc     PhaseKind = "營業成功" // 營業成功階段（終止）
	PhaseGameFail     PhaseKind = "營業失敗" // 營業失敗階段（終止）
)

// PhaseJumpLegal phaseJump 命令允許設定的「下一階段」合法值集合（不含空字串）。
// 對應【營業規格書 | 二十五、操作命令清單 | phaseJump】。
var PhaseJumpLegal = map[PhaseKind]bool{
	PhasePlayerAction: true,
	PhaseGuestAction:  true,
	PhaseRoundEnd:     true,
}

// TriggerKind 觸發時機；對應【營業規格書 | 二十二、觸發時機清單】。
//
// 觸發時機僅作為時間訊號、不攜帶資料。值採用規格書的英文名稱字面值。
// 效果靜態表格 Effect.TriggerKind 為 int32 編碼，其與本型別的對照於載入器建立。
type TriggerKind string

const (
	TriggerGameStart  TriggerKind = "gameStart"  // 營業開始；前置技能啟動後
	TriggerRoundReady TriggerKind = "roundReady" // 回合準備；回合數遞增後、回合開始設置前
	TriggerRoundStart TriggerKind = "roundStart" // 回合開始；回合開始設置後
	TriggerGuestSeat  TriggerKind = "guestSeat"  // 顧客入座；每彈出 1 位顧客入座後
	TriggerCardDraw   TriggerKind = "cardDraw"   // 手牌補充；每張卡牌進入手牌後
	TriggerUserStart  TriggerKind = "userStart"  // 玩家開始；補牌完成後
	TriggerUserEnd    TriggerKind = "userEnd"    // 玩家結束；進入玩家行動結束流程後、處理剩餘手牌前
	TriggerCardDrop   TriggerKind = "cardDrop"   // 卡牌棄置；每張卡牌進入棄牌牌堆後
	TriggerCardExile  TriggerKind = "cardExile"  // 卡牌流放；每張卡牌進入流放牌堆後
	TriggerCardMorph  TriggerKind = "cardMorph"  // 卡牌變身；卡牌完成變身處理流程後
	TriggerCardPlay   TriggerKind = "cardPlay"   // 玩家出牌；實例效果列表啟動、卡牌移至棄牌牌堆後
	TriggerGuestStart TriggerKind = "guestStart" // 顧客開始；進入顧客行動階段後
	TriggerGuestTask  TriggerKind = "guestTask"  // 顧客行動；每彈出 1 個行動、啟動顧客行動技能後
	TriggerGuestEnd   TriggerKind = "guestEnd"   // 顧客結束；行動佇列清空後
	TriggerRoundEnd   TriggerKind = "roundEnd"   // 回合結束；進入回合結束階段後
	TriggerExitAny    TriggerKind = "exitAny"    // 顧客離場；任何原因離場前
	TriggerExitSate   TriggerKind = "exitSate"   // 飽食離場；提供滿意值離場前
	TriggerExitCalm   TriggerKind = "exitCalm"   // 生氣離場；扣士氣離場前
	TriggerExitDone   TriggerKind = "exitDone"   // 顧客離場後；自所在容器移除後、清理效果前
	TriggerDamage     TriggerKind = "damage"     // 士氣受損；餐廳士氣值因 -= 實際扣減後
	TriggerGameSucc   TriggerKind = "gameSucc"   // 營業成功；通關結算判定後
	TriggerGameFail   TriggerKind = "gameFail"   // 營業失敗；失敗結算判定後
)

// EffectKind 效果類型；對應【營業規格書 | 七、效果類型】。
// （效果實例型別見 instance.go 的 Effect；此 enum 表達其靜態類型，故以 Kind 為後綴避免撞名。）
type EffectKind int32

const (
	EffectImmed   EffectKind = iota // 立即；不進佇列，建立當下檢查觸發條件
	EffectTrigger                   // 觸發；進佇列，觸發時機 + 條件符合時執行觸發命令
	EffectPersist                   // 常駐；進佇列，以生命週期語意運作（啟動命令 / 結束命令）
)

// TargetKind 目標類型；對應【營業規格書 | 八、目標類型】。
type TargetKind int32

const (
	TargetNone      TargetKind = iota // 無目標；self 為空物件
	TargetGuestPick                   // 新選顧客；暫停流程由玩家選取
	TargetGuestSame                   // 沿用顧客；繼承前一效果的 self
	TargetGuestRand                   // 隨機顧客；系統隨機選取
	TargetCardPick                    // 新選手牌；暫停流程由玩家選取
	TargetCardSame                    // 沿用手牌；繼承前一效果的 self
	TargetCardRand                    // 隨機手牌；系統隨機選取
)

// TriggerAfter 觸發後行為；對應【營業規格書 | 十三、觸發後行為】（僅觸發類型適用）。
type TriggerAfter int32

const (
	TriggerAfterKeep   TriggerAfter = iota // 保留；觸發後保留於效果佇列
	TriggerAfterRemove                     // 移除；觸發後自效果佇列移除
)

// StackTime 堆疊時間；對應【營業規格書 | 十六、堆疊規則】。
type StackTime int32

const (
	StackTimeStay    StackTime = iota // 不變；堆疊時不刷新作用回合
	StackTimeRefresh                  // 刷新；堆疊時刷新作用回合
)

// TaskKind 行動類型；對應【營業規格書 | 五、實例結構 | 行動（Action）實例】。
// 於顧客行動階段決定封印閘門查詢對象（飽食 / 耐心）。
type TaskKind int32

const (
	TaskSate TaskKind = iota // 飽食；查 封印飽食技能 閘門
	TaskCalm                 // 耐心；查 封印耐心技能 閘門
)
