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
	EventType EventType // 投影事件類別
}

// InstanceID 實例的唯一識別碼；卡牌實例 / 顧客實例 / 效果實例共用。
// 對應【營業規格書 | 五、實例結構】各實例的「實例編號」欄位。
type InstanceID int64

// NoneID 空實例編號；表示尚未配發或不存在的實例。
const NoneID InstanceID = 0

// EventType 投影事件類別；對應【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 四之二】投影事件分類。
type EventType int

const (
	EventTypeInstance  EventType = iota // 實例建立 / 銷毀（卡牌 / 顧客 / 效果）
	EventTypeContainer                  // 卡牌移動牌堆、顧客入座 / 離場 / 遊蕩 / 卡牌化
	EventTypeProperty                   // sate / calm / score / 鎖定計數 變化
	EventTypeTrigger                    // 觸發時機到達、某效果觸發
	EventTypePhase                      // phase 切換
)

// PhaseType 營業的執行階段；對應【營業規格書 | 十九、核心流程】。
//
// 值採用規格書使用的中文字面值：全域屬性「下一階段」(NextPhase) 即以此型別儲存，
// 合法跳轉值為 玩家行動 / 顧客行動 / 回合結束 / 空字串（無跳轉）。
type PhaseType string

const (
	PhaseTypeNone         PhaseType = ""     // 無階段 / 無跳轉（下一階段清除後的值）
	PhaseTypeGameStart    PhaseType = "營業開始" // 營業開始階段；啟動前置技能
	PhaseTypeRoundStart   PhaseType = "回合開始" // 回合開始階段；回合數遞增、入座
	PhaseTypePlayerAction PhaseType = "玩家行動" // 玩家行動階段；補牌 / 出牌 / 結束
	PhaseTypeGuestAction  PhaseType = "顧客行動" // 顧客行動階段；彈出行動佇列
	PhaseTypeRoundEnd     PhaseType = "回合結束" // 回合結束階段；推進效果、執行結算
	PhaseTypeGameSucc     PhaseType = "營業成功" // 營業成功階段（終止）
	PhaseTypeGameFail     PhaseType = "營業失敗" // 營業失敗階段（終止）
)

// PhaseJumpLegal phaseJump 命令允許設定的「下一階段」合法值集合（不含空字串）。
// 對應【營業規格書 | 二十五、操作命令清單 | phaseJump】。
var PhaseJumpLegal = map[PhaseType]bool{
	PhaseTypePlayerAction: true,
	PhaseTypeGuestAction:  true,
	PhaseTypeRoundEnd:     true,
}

// TriggerType 觸發時機；對應【營業規格書 | 二十二、觸發時機清單】。
//
// 觸發時機僅作為時間訊號、不攜帶資料。值採用規格書的英文名稱字面值。
// 效果靜態表格 Effect.TriggerType 為 int32 編碼，其與本型別的對照於載入器建立。
type TriggerType string

const (
	TriggerTypeGameStart  TriggerType = "gameStart"  // 營業開始；前置技能啟動後
	TriggerTypeRoundReady TriggerType = "roundReady" // 回合準備；回合數遞增後、回合開始設置前
	TriggerTypeRoundStart TriggerType = "roundStart" // 回合開始；回合開始設置後
	TriggerTypeGuestSeat  TriggerType = "guestSeat"  // 顧客入座；每彈出 1 位顧客入座後
	TriggerTypeCardDraw   TriggerType = "cardDraw"   // 手牌補充；每張卡牌進入手牌後
	TriggerTypeUserStart  TriggerType = "userStart"  // 玩家開始；補牌完成後
	TriggerTypeUserEnd    TriggerType = "userEnd"    // 玩家結束；進入玩家行動結束流程後、處理剩餘手牌前
	TriggerTypeCardDrop   TriggerType = "cardDrop"   // 卡牌棄置；每張卡牌進入棄牌牌堆後
	TriggerTypeCardExile  TriggerType = "cardExile"  // 卡牌流放；每張卡牌進入流放牌堆後
	TriggerTypeCardMorph  TriggerType = "cardMorph"  // 卡牌變身；卡牌完成變身處理流程後
	TriggerTypeCardPlay   TriggerType = "cardPlay"   // 玩家出牌；實例效果列表啟動、卡牌移至棄牌牌堆後
	TriggerTypeGuestStart TriggerType = "guestStart" // 顧客開始；進入顧客行動階段後
	TriggerTypeGuestTask  TriggerType = "guestTask"  // 顧客行動；每彈出 1 個行動、啟動顧客行動技能後
	TriggerTypeGuestEnd   TriggerType = "guestEnd"   // 顧客結束；行動佇列清空後
	TriggerTypeRoundEnd   TriggerType = "roundEnd"   // 回合結束；進入回合結束階段後
	TriggerTypeExitAny    TriggerType = "exitAny"    // 顧客離場；任何原因離場前
	TriggerTypeExitSate   TriggerType = "exitSate"   // 飽食離場；提供滿意值離場前
	TriggerTypeExitCalm   TriggerType = "exitCalm"   // 生氣離場；扣士氣離場前
	TriggerTypeExitDone   TriggerType = "exitDone"   // 顧客離場後；自所在容器移除後、清理效果前
	TriggerTypeDamage     TriggerType = "damage"     // 士氣受損；餐廳士氣值因 -= 實際扣減後
	TriggerTypeGameSucc   TriggerType = "gameSucc"   // 營業成功；通關結算判定後
	TriggerTypeGameFail   TriggerType = "gameFail"   // 營業失敗；失敗結算判定後
)

// EffectType 效果類型；對應【營業規格書 | 七、效果類型】。
// （效果實例型別見 instance.go 的 Effect；此 enum 表達其靜態類型，故以 Type 為後綴避免撞名。）
type EffectType int32

const (
	EffectTypeImmed   EffectType = iota // 立即；不進佇列，建立當下檢查觸發條件
	EffectTypeTrigger                   // 觸發；進佇列，觸發時機 + 條件符合時執行觸發命令
	EffectTypePersist                   // 常駐；進佇列，以生命週期語意運作（啟動命令 / 結束命令）
)

// TargetType 目標類型；對應【營業規格書 | 八、目標類型】。
type TargetType int32

const (
	TargetTypeNone      TargetType = iota // 無目標；self 為空物件
	TargetTypeGuestPick                   // 新選顧客；暫停流程由玩家選取
	TargetTypeGuestSame                   // 沿用顧客；繼承前一效果的 self
	TargetTypeGuestRand                   // 隨機顧客；系統隨機選取
	TargetTypeCardPick                    // 新選手牌；暫停流程由玩家選取
	TargetTypeCardSame                    // 沿用手牌；繼承前一效果的 self
	TargetTypeCardRand                    // 隨機手牌；系統隨機選取
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

// TaskType 行動類型；對應【營業規格書 | 五、實例結構 | 行動（Action）實例】。
// 於顧客行動階段決定封印閘門查詢對象（飽食 / 耐心）。
type TaskType int32

const (
	TaskTypeSate TaskType = iota // 飽食；查 封印飽食技能 閘門
	TaskTypeCalm                 // 耐心；查 封印耐心技能 閘門
)

// Selector 命令對象的英文名稱；對應【營業規格書 | 二十四、命令對象清單】。
//
// 命令對象解析「這次命令作用到誰」，恆解析為集合（可能為空 / 多個）。
// 刻意不命名為 target，以免與【營業規格書 | 八、目標類型】(TargetType) 撞名。
type Selector string

const (
	SelectorNone Selector = "none" // 無命令對象

	// self 系列（以目標類型綁定的 self 為起點）

	SelectorSelf     Selector = "self"     // 自身
	SelectorSelfNear Selector = "selfNear" // 鄰桌（不含自身與同桌）
	SelectorSelfSame Selector = "selfSame" // 同桌（含自身）

	// 事件型引用（取最近一次事件對象，無則空集合）

	SelectorDamageGuest Selector = "damageGuest" // 士氣受損顧客
	SelectorDrawLast    Selector = "drawLast"    // 最後抽出卡牌
	SelectorDropLast    Selector = "dropLast"    // 最後棄置卡牌
	SelectorExileLast   Selector = "exileLast"   // 最後流放卡牌
	SelectorExitLast    Selector = "exitLast"    // 最後離場顧客
	SelectorMorphLast   Selector = "morphLast"   // 最後變身卡牌
	SelectorPlayLast    Selector = "playLast"    // 最後出牌卡牌
	SelectorSeatLast    Selector = "seatLast"    // 最後入座顧客
	SelectorTaskGuest   Selector = "taskGuest"   // 最後行動顧客

	// 顧客（僅取座位列表 / 排隊佇列）

	SelectorGuestAll  Selector = "guestAll"  // 全部顧客
	SelectorGuestPick Selector = "guestPick" // 指定顧客（玩家挑最多 N 位）
	SelectorGuestRand Selector = "guestRand" // 隨機顧客
	SelectorGuestWait Selector = "guestWait" // 排隊顧客（排隊前 N 位）
	SelectorNearPick  Selector = "nearPick"  // 指定顧客鄰桌
	SelectorNearRand  Selector = "nearRand"  // 隨機顧客鄰桌
	SelectorSamePick  Selector = "samePick"  // 指定顧客桌
	SelectorSameRand  Selector = "sameRand"  // 隨機顧客桌

	// 手牌

	SelectorHandAll  Selector = "handAll"  // 全部手牌
	SelectorHandPick Selector = "handPick" // 指定手牌
	SelectorHandRand Selector = "handRand" // 隨機手牌

	// 抽牌牌堆（deckTop 為唯一會修改狀態的命令對象）

	SelectorDeckAll  Selector = "deckAll"  // 全部抽牌牌堆
	SelectorDeckPick Selector = "deckPick" // 指定抽牌牌堆
	SelectorDeckRand Selector = "deckRand" // 隨機抽牌牌堆
	SelectorDeckTop  Selector = "deckTop"  // 抽牌牌堆頂端（不足時 auto-shuffle）

	// 棄牌牌堆

	SelectorDropAll  Selector = "dropAll"  // 全部棄牌牌堆
	SelectorDropPick Selector = "dropPick" // 指定棄牌牌堆
	SelectorDropRand Selector = "dropRand" // 隨機棄牌牌堆
	SelectorDropTop  Selector = "dropTop"  // 棄牌牌堆頂端

	// 流放牌堆

	SelectorExileAll  Selector = "exileAll"  // 全部流放牌堆
	SelectorExilePick Selector = "exilePick" // 指定流放牌堆
	SelectorExileRand Selector = "exileRand" // 隨機流放牌堆
)

// Attr 全域屬性名稱；對應【營業規格書 | 二十三、屬性清單】主表。
//
// 本表雙用途：作為【營業規格書 | 十一、觸發條件】的條件對象讀取來源，
// 亦作為【營業規格書 | 十七、命令 | 屬性修改命令】的可寫屬性清單。
// 名稱帶 (...) 的查詢函式（deckSize / drawTotal / tableCount …）此處僅登記其字根名稱，
// 參數解析由 exprs 引擎 / 屬性註冊表處理。
type Attr string

const (
	AttrCardifySize  Attr = "cardifySize"  // 卡牌化列表當下大小
	AttrDamageGuest  Attr = "damageGuest"  // 士氣受損顧客（顧客引用）
	AttrDamageValue  Attr = "damageValue"  // 士氣受損值
	AttrDeckSize     Attr = "deckSize"     // 抽牌牌堆卡牌張數（查詢函式）
	AttrDrawCount    Attr = "drawCount"    // 回合抽牌張數
	AttrDrawLast     Attr = "drawLast"     // 最後抽出卡牌（卡牌引用）
	AttrDrawMax      Attr = "drawMax"      // 補牌張數上限（寫鎖）
	AttrDrawTotal    Attr = "drawTotal"    // 累積抽牌張數（查詢函式）
	AttrDropCount    Attr = "dropCount"    // 回合棄牌張數
	AttrDropLast     Attr = "dropLast"     // 最後棄置卡牌（卡牌引用）
	AttrDropSize     Attr = "dropSize"     // 棄牌牌堆卡牌張數（查詢函式）
	AttrDropTotal    Attr = "dropTotal"    // 累積棄牌張數（查詢函式）
	AttrEnergy       Attr = "energy"       // 出牌點數（寫鎖）
	AttrEnergyMax    Attr = "energyMax"    // 出牌點數上限（寫鎖）
	AttrEnergyKeep   Attr = "energyKeep"   // 出牌點數保留（鎖）
	AttrExileCount   Attr = "exileCount"   // 回合流放張數
	AttrExileLast    Attr = "exileLast"    // 最後流放卡牌（卡牌引用）
	AttrExileSize    Attr = "exileSize"    // 流放牌堆卡牌張數（查詢函式）
	AttrExileTotal   Attr = "exileTotal"   // 累積流放張數（查詢函式）
	AttrExitCount    Attr = "exitCount"    // 回合離場人數
	AttrExitLast     Attr = "exitLast"     // 最後離場顧客（顧客引用）
	AttrExitLastSeat Attr = "exitLastSeat" // 最後離場座位
	AttrGuestSize    Attr = "guestSize"    // 店內顧客當下總數（衍生）
	AttrHandMax      Attr = "handMax"      // 手牌張數上限（寫鎖）
	AttrHandSize     Attr = "handSize"     // 手牌卡牌張數（查詢函式）
	AttrMorale       Attr = "morale"       // 餐廳士氣值（寫鎖）
	AttrMoraleBlock  Attr = "moraleBlock"  // 餐廳士氣值格擋（寫鎖）
	AttrMoraleMax    Attr = "moraleMax"    // 餐廳士氣值上限（寫鎖）
	AttrMoraleShield Attr = "moraleShield" // 餐廳士氣值護盾（寫鎖）
	AttrMorphCount   Attr = "morphCount"   // 回合變身次數
	AttrMorphLast    Attr = "morphLast"    // 最後變身卡牌（卡牌引用）
	AttrMorphNewID   Attr = "morphNewID"   // 變身後卡牌編號
	AttrMorphOldID   Attr = "morphOldID"   // 變身前卡牌編號
	AttrNextPhase    Attr = "nextPhase"    // 下一階段（唯讀；僅 phaseJump 可寫）
	AttrPlayCount    Attr = "playCount"    // 回合出牌張數
	AttrPlayLast     Attr = "playLast"     // 最後出牌卡牌（卡牌引用）
	AttrPlayTotal    Attr = "playTotal"    // 累積出牌張數（查詢函式）
	AttrRoamSize     Attr = "roamSize"     // 遊蕩列表當下大小
	AttrRound        Attr = "round"        // 回合（寫）
	AttrRoundLeft    Attr = "roundLeft"    // 剩餘回合（衍生；寫入轉譯為回合上限）
	AttrRoundMax     Attr = "roundMax"     // 回合上限（寫）
	AttrScore        Attr = "score"        // 餐廳滿意值（寫鎖）
	AttrSeatCount    Attr = "seatCount"    // 回合入座人數
	AttrSeatLast     Attr = "seatLast"     // 最後入座顧客（顧客引用）
	AttrSeatLeft     Attr = "seatLeft"     // 剩餘座位
	AttrSeatSize     Attr = "seatSize"     // 座位列表當下大小
	AttrSelf         Attr = "self"         // 自身（卡牌兼顧客引用）
	AttrTableCount   Attr = "tableCount"   // 桌數查詢（查詢函式）
	AttrTableGuest   Attr = "tableGuest"   // 桌次顧客數（查詢函式）
	AttrTableSize    Attr = "tableSize"    // 桌數總量（衍生）
	AttrTaskCount    Attr = "taskCount"    // 回合行動次數
	AttrTaskGuest    Attr = "taskGuest"    // 最後行動顧客（顧客引用）
	AttrTaskSize     Attr = "taskSize"     // 行動佇列當下大小
	AttrTaskSkill    Attr = "taskSkill"    // 最後行動技能
	AttrWaitSize     Attr = "waitSize"     // 排隊佇列當下大小
)

// AttrRef dot-syntax 引用屬性名稱；對應【營業規格書 | 二十三、屬性清單 | 卡牌引用屬性 / 顧客引用屬性】。
// 以 <引用>.<屬性> 形式存取（如 self.sate、drawLast.cardID、taskGuest.calm）。
type AttrRef string

const (
	// 卡牌引用屬性

	AttrRefCardEffect  AttrRef = "cardEffect"  // 卡牌帶效果數（查詢函式）
	AttrRefCardGroup   AttrRef = "cardGroup"   // 卡牌群組編號
	AttrRefCardID      AttrRef = "cardID"      // 卡牌編號
	AttrRefCardify     AttrRef = "cardify"     // 卡牌化來源顧客
	AttrRefCardSeal    AttrRef = "cardSeal"    // 封印卡牌（鎖）
	AttrRefCost        AttrRef = "cost"        // 出牌費用（寫鎖）
	AttrRefExtraRunMax AttrRef = "extraRunMax" // 額外發動次數上限（寫鎖）
	AttrRefExtraRunMin AttrRef = "extraRunMin" // 額外發動次數下限（寫鎖）
	AttrRefInDeck      AttrRef = "inDeck"      // 位於抽牌牌堆（布林）
	AttrRefInDrop      AttrRef = "inDrop"      // 位於棄牌牌堆（布林）
	AttrRefInExile     AttrRef = "inExile"     // 位於流放牌堆（布林）
	AttrRefInHand      AttrRef = "inHand"      // 位於手牌（布林）
	AttrRefKeep        AttrRef = "keep"        // 不棄卡牌（鎖）
	AttrRefPlayExile   AttrRef = "playExile"   // 出牌後流放（鎖）
	AttrRefUnplayExile AttrRef = "unplayExile" // 未出牌流放（鎖）

	// 顧客引用屬性

	AttrRefCalm         AttrRef = "calm"         // 耐心值（寫鎖）
	AttrRefCalmHit      AttrRef = "calmHit"      // 已觸發耐心門檻（讀取為列表大小）
	AttrRefCalmSeal     AttrRef = "calmSeal"     // 封印耐心技能（鎖）
	AttrRefEffectImmune AttrRef = "effectImmune" // 效果免疫群組（查詢函式）
	AttrRefFreeze       AttrRef = "freeze"       // 凍結起始回合
	AttrRefGuestID      AttrRef = "guestID"      // 顧客編號
	AttrRefMoraleMax    AttrRef = "moraleMax"    // 士氣值上限（寫鎖）
	AttrRefNearSize     AttrRef = "nearSize"     // 鄰桌當下大小
	AttrRefSameSize     AttrRef = "sameSize"     // 同桌當下大小
	AttrRefSate         AttrRef = "sate"         // 飽食值（寫鎖）
	AttrRefSateHit      AttrRef = "sateHit"      // 已觸發飽食門檻（讀取為列表大小）
	AttrRefSateMax      AttrRef = "sateMax"      // 飽食值離場線（寫鎖）
	AttrRefSateSeal     AttrRef = "sateSeal"     // 封印飽食技能（鎖）
	AttrRefScoreMax     AttrRef = "scoreMax"     // 滿意值上限（寫鎖）
	AttrRefSeatID       AttrRef = "seatID"       // 座位編號
	AttrRefSkillImmune  AttrRef = "skillImmune"  // 技能免疫群組（查詢函式）

	// 卡牌與顧客共用

	AttrRefEffectGroup AttrRef = "effectGroup" // 已帶群組效果數（查詢函式）
	AttrRefEffectStack AttrRef = "effectStack" // 已帶效果層數（查詢函式）
	AttrRefMorale      AttrRef = "morale"      // 士氣值（顧客；寫鎖）
	AttrRefScore       AttrRef = "score"       // 滿意值（顧客；寫鎖）
)

// Command 操作命令的英文名稱；對應【營業規格書 | 二十五、操作命令清單】。
//
// 屬性修改命令（運算符語法 = += -= *= /= %= @ #）不在此清單，
// 其可寫屬性見【營業規格書 | 二十三、屬性清單】與 Attr。
type Command string

const (
	// 卡牌屬性 / 效果

	CommandCardCostAdd      Command = "cardCostAdd"      // 出牌費用增減
	CommandCardCostMul      Command = "cardCostMul"      // 出牌費用倍率
	CommandCardCostSet      Command = "cardCostSet"      // 出牌費用設值
	CommandCardEffectAdd    Command = "cardEffectAdd"    // 卡牌增加效果
	CommandCardEffectDel    Command = "cardEffectDel"    // 卡牌移除效果
	CommandCardEffectDelAll Command = "cardEffectDelAll" // 卡牌移除全部效果
	CommandCardify          Command = "cardify"          // 卡牌化顧客
	CommandCardRun          Command = "cardRun"          // 強制發動卡牌
	CommandRestore          Command = "restore"          // 卡牌化還原

	// 抽牌牌堆

	CommandDeckAdd     Command = "deckAdd"     // 加牌入抽牌牌堆
	CommandDeckClone   Command = "deckClone"   // 深複製卡牌至抽牌牌堆
	CommandDeckCopy    Command = "deckCopy"    // 淺複製卡牌至抽牌牌堆
	CommandDeckMorph   Command = "deckMorph"   // 抽牌牌堆變身
	CommandDeckRoll    Command = "deckRoll"    // 抽獎加牌入抽牌牌堆
	CommandDeckShuffle Command = "deckShuffle" // 抽牌牌堆洗牌
	CommandDeckToDrop  Command = "deckToDrop"  // 抽牌牌堆移至棄牌牌堆
	CommandDeckToExile Command = "deckToExile" // 抽牌牌堆移至流放牌堆
	CommandDeckToHand  Command = "deckToHand"  // 抽牌牌堆移至手牌

	// 棄牌牌堆

	CommandDropAdd     Command = "dropAdd"     // 加牌入棄牌牌堆
	CommandDropClone   Command = "dropClone"   // 深複製卡牌至棄牌牌堆
	CommandDropCopy    Command = "dropCopy"    // 淺複製卡牌至棄牌牌堆
	CommandDropMorph   Command = "dropMorph"   // 棄牌牌堆變身
	CommandDropRoll    Command = "dropRoll"    // 抽獎加牌入棄牌牌堆
	CommandDropToDeck  Command = "dropToDeck"  // 棄牌牌堆移至抽牌牌堆
	CommandDropToExile Command = "dropToExile" // 棄牌牌堆移至流放牌堆
	CommandDropToHand  Command = "dropToHand"  // 棄牌牌堆移至手牌

	// 流放牌堆

	CommandExileAdd    Command = "exileAdd"    // 加牌入流放牌堆
	CommandExileClone  Command = "exileClone"  // 深複製卡牌至流放牌堆
	CommandExileCopy   Command = "exileCopy"   // 淺複製卡牌至流放牌堆
	CommandExileMorph  Command = "exileMorph"  // 流放牌堆變身
	CommandExileRoll   Command = "exileRoll"   // 抽獎加牌入流放牌堆
	CommandExileToDeck Command = "exileToDeck" // 流放牌堆移至抽牌牌堆
	CommandExileToDrop Command = "exileToDrop" // 流放牌堆移至棄牌牌堆
	CommandExileToHand Command = "exileToHand" // 流放牌堆移至手牌

	// 手牌

	CommandHandAdd     Command = "handAdd"     // 加牌入手牌
	CommandHandClone   Command = "handClone"   // 深複製卡牌至手牌
	CommandHandCopy    Command = "handCopy"    // 淺複製卡牌至手牌
	CommandHandMorph   Command = "handMorph"   // 手牌變身
	CommandHandRoll    Command = "handRoll"    // 抽獎加牌入手牌
	CommandHandToDeck  Command = "handToDeck"  // 手牌移至抽牌牌堆
	CommandHandToDrop  Command = "handToDrop"  // 手牌移至棄牌牌堆
	CommandHandToExile Command = "handToExile" // 手牌移至流放牌堆

	// 效果佇列

	CommandEffectClear     Command = "effectClear"     // 清除群組效果
	CommandEffectDel       Command = "effectDel"       // 移除效果
	CommandEffectImmuneAdd Command = "effectImmuneAdd" // 賦予效果免疫
	CommandEffectImmuneDel Command = "effectImmuneDel" // 解除效果免疫
	CommandEffectRun       Command = "effectRun"       // 啟動效果

	// 顧客

	CommandGuestExit   Command = "guestExit"   // 顧客離場
	CommandGuestPart   Command = "guestPart"   // 改變顧客部位（規格待後續定案）
	CommandGuestReturn Command = "guestReturn" // 顧客回座
	CommandGuestRoam   Command = "guestRoam"   // 顧客遊蕩
	CommandGuestSeat   Command = "guestSeat"   // 顧客入座
	CommandGuestSkin   Command = "guestSkin"   // 改變顧客外觀（規格待後續定案）
	CommandGuestSpawn  Command = "guestSpawn"  // 產生顧客
	CommandWaitAdd     Command = "waitAdd"     // 加顧客入排隊佇列

	// 技能免疫 / 行動 / 階段

	CommandSkillImmuneAdd Command = "skillImmuneAdd" // 賦予技能免疫
	CommandSkillImmuneDel Command = "skillImmuneDel" // 解除技能免疫
	CommandTaskAdd        Command = "taskAdd"        // 加入行動
	CommandPhaseJump      Command = "phaseJump"      // 階段跳轉
)
