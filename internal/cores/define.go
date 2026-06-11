package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 這三個介面是核心對外的唯一行為邊界; TUI 與測試各自提供實作。
// 對應【營業實作規格書 | 四、解耦的關鍵：邊界介面】。
//
// 靜態表格不走介面: 核心直接吃 Sheeter 生成的 *sheeter.Sheeter(資料 port 本身),
// 以 reader.Get 查詢; 衍生索引由核心(cores 的 NewData)預建。
//
// 核心完全同步、單執行緒、無 channel: 需要玩家輸入時阻塞呼叫 Operator,
// 每跑一個單位呼叫 Presenter.Emit。goroutine + channel 只活在 TUI adapter。
// 三 port 定義於 cores(純資料模型); games 驅動引擎與 infra / TUI 各自實作。

// 詞彙行為的函式型別: rules 各概念檔的行為簽名, 經 Game.Register* 逐詞條裝備。
// 詞彙表為 Game 私有成員(建後唯讀、等同常數); 名稱即詞條鍵, Lock 讀詞條以全名(如 moraleLock)註冊、無後綴路由。

// AttrReadFunc 全域屬性詞條的讀取行為: 以 Game 為 context 求值; arg 供查詢函式型屬性(deckSize…), 純屬性忽略。
type AttrReadFunc func(game *Game, arg []exprs.Value) (result exprs.Value, ok bool)

// AttrWriteFunc 全域屬性詞條的寫入行為: op 為賦值符、n 為已求值的右值(@ # 時忽略)。
type AttrWriteFunc func(game *Game, op AssignKind, n float64) bool

// AttrRefReadFunc 引用屬性詞條的讀取行為: 對 ref(卡牌 / 顧客)以 Game 為 context 取子屬性; arg 供引用查詢函式。
type AttrRefReadFunc func(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool)

// AttrRefWriteFunc 引用屬性詞條的寫入行為: 對 ref 依賦值符 op 與右值 n 變更狀態。
type AttrRefWriteFunc func(game *Game, ref exprs.Ref, op AssignKind, n float64) bool

// CommandFunc 操作命令詞條的執行行為: target 為已解析命令對象(身分集)、arg 為其餘已求值參數。
// target == nil 代表命令對象為 none(ExecOperate 正規化; 見 SelectorNone), 與「篩空的非 nil 空切片」有別——
// 多數詞條不分辨兩者(固定 none 詞條忽略 target、其餘空集合即整體 no-op), 目前僅 effectClear 以 nil 判全域掃描。
type CommandFunc func(game *Game, target []InstanceID, arg []exprs.Value)

// SelectorFunc 命令對象詞條的解析行為: arg 為 [...] 內已求值參數, 產出作用對象集合(身分集)。
type SelectorFunc func(game *Game, arg []exprs.Value) (result []InstanceID)

// SelectorNone 無命令對象的保留詞條名; ExecOperate 對它把 target 正規化為 nil(全域掃描記號),
// 其他命令對象篩空時正規化為非 nil 空切片, 供 verb 以 target == nil 判別「對象為 none」
// (【營業規格書 | 二十五、操作命令清單 | effectClear】對象 = none 為全域掃描)。
const SelectorNone = "none"

// Compiler 把單一命令字串編譯為效果命令執行器; 由 games 注入(cores 無命令解析能力)。語法錯回 error。
// prepareEffect 僅於命令欄非空時呼叫此原語, 故空字串處理不在本型別契約內。
type Compiler func(source string) (command EffectExec, err error)

// EffectExec 預編譯效果命令的執行器; games 注入的 Compiler 把命令字串編成閉包(捕捉 Parse 後 AST + execute 走法 X 分派)、rules 於效果流程(runEffectExec)呼叫。空命令欄為 nil。
type EffectExec func(game *Game)

// Rander 唯一亂數來源(單一 seeded PRNG); 決定性的基礎。
type Rander interface {
	Intn(n int) int                     // 回傳 [0, n) 的隨機整數
	Shuffle(n int, swap func(i, j int)) // 對 n 個元素洗牌
	Weighted(weight []int32) int        // 抽獎 weighted random; 回傳命中索引
}

// Operator 玩家輸入; 對應所有「暫停流程」點。
type Operator interface {
	PlayerAction(game *Game) *Card                 // 玩家行動: 出牌回該手牌卡、nil 即玩家結束(【營業規格書 | 十九、核心流程 | 3】)
	PickGuest(source []*Guest, count int) []*Guest // 新選顧客 / guestPick / nearPick / samePick
	PickCard(source []*Card, count int) []*Card    // 新選手牌 / *Pick 牌堆類
	PickDiscard(source []*Card, over int) []*Card  // 手牌上限棄牌
}

// Presenter 事件流輸出(yield-per-unit; 步進時可阻塞)。
type Presenter interface {
	Emit(eventData EventData) // 一個命令 / 一次觸發 / 一次 phase 切換 / 一個玩家動作 = 一個 EventData
}

// EventData 核心吐給前端的細粒度投影事件: 單一 struct 以 Kind 分派, 各類只填自己的本體欄位、其餘留零值。
// Round / Phase 為座標欄, 由發射入口 Game.Emit 統一蓋章, 發射點不自帶。
// 識別碼只帶編號、名稱由前端憑同一份靜態表自查; 凡指涉實例同時帶資料編號 + 實例編號(技能無實例編號只帶資料編號)。
// 對應【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 四之二】投影事件分類;
// 欄位形狀對齊事件日誌行角色(【營業顯示規格書 | 6、畫面規格 | 6.10】)。
type EventData struct {
	Kind  EventKind // 投影事件類別
	Round int32     // 座標: 當前回合(Emit 蓋章)
	Phase PhaseKind // 座標: 當前階段(Emit 蓋章)

	// 對象欄(各類的作用對象: instance 本體 / container 移動者 / property 對象(全域屬性留零值)/
	// scope 的卡牌或顧客操作元 / effect 的 self; 無對象留零值)
	DataID     int32      // 對象資料編號
	InstanceID InstanceID // 對象實例編號

	// 成員資格記號(跨類別共用; M21 拍板): EventInstance 實例建立 true / 銷毀 false;
	// EventEffect 結束階段 退層留佇列 true / 退場出佇列 false; EventAction 入列 true / 出列 false。
	Alive bool

	// EventContainer 本體欄位; From == To 為容器重整(洗牌 / 洗回後), Pick 欄載重整後全序(M21 拍板)
	From           ContainerKind // 來源容器(新建直入容器留 ContainerNone)
	To             ContainerKind // 目的容器(銷毀離開容器留 ContainerNone)
	SeatID         int32         // 入座座位編號(To == ContainerSeat 時)
	BindID         int32         // 卡牌化來源顧客資料編號(移動者為已綁定卡牌時; M21 拍板)
	BindInstanceID InstanceID    // 卡牌化來源顧客實例編號(同上)

	// EventProperty 本體欄位(日誌命令行 <屬性> <運算> <值> >> <結果>)
	Attr    string     // 屬性詞條鍵(英文; 中文全名由前端憑詞彙對照轉換)
	Op      AssignKind // 賦值符(@ / # 無算術式, Operand 留零值)
	Operand float64    // 右值(命令求出的運算值)
	Before  float64    // 前值(clamp 後實際變化量 = After - Before)
	After   float64    // 後值(日誌的 >> 結果)

	// EventScope 本體欄位(範圍標題 + 操作元; 卡牌 / 顧客操作元於上方對象欄)
	Scope   ScopeKind   // 範圍類別
	Trigger TriggerKind // 時機名(Scope == ScopeTrigger 時)
	SkillID int32       // 技能操作元資料編號(動作類; 技能無實例編號)

	// EventEffect 本體欄位(效果行; self 於上方對象欄; 表演資訊規格待定, 日後憑 EffectID 查表)
	EffectID         int32       // 效果資料編號
	EffectInstanceID InstanceID  // 效果實例編號
	Stage            EffectStage // 效果階段
	Stack            int32       // 佇列項層數快照(事件後絕對值; 加入 / 結束帶, 退場時為退場層數; M21 拍板)
	Expire           int32       // 佇列項結束回合快照(0 = 整場保留; 加入 / 結束帶; M21 拍板)

	// EventAction 本體欄位(行動佇列入列 / 出列; 對象欄載顧客、SkillID 重用 scope 段、在列記號重用 Alive; M21 拍板)
	Task TaskKind // 行動類型(飽食 / 耐心)

	// EventSelect 本體欄位(玩家輸入紀錄, 日誌 $ 選取 <來源> -> <選中…>; 搭 seed 重現 bug);
	// Pick 另供容器重整事件載重整後全序(M21 拍板)
	Source string     // 選取來源(詞條鍵 / 流程名; 前端轉中文)
	Pick   []PickData // 選中清單 / 重整後全序
}

// PickData 玩家選取結果的一筆: 資料編號 + 實例編號(EventSelect 的選中清單元素)。
type PickData struct {
	DataID     int32      // 資料編號
	InstanceID InstanceID // 實例編號
}

// InstanceID 實例的唯一識別碼; 卡牌實例 / 顧客實例 / 效果實例共用。
// 對應【營業規格書 | 五、實例結構】各實例的「實例編號」欄位。
type InstanceID int64

// NoneID 空實例編號; 表示尚未配發或不存在的實例。
const NoneID InstanceID = 0

// ContainerKind 容器種類; 供 locate 回報實例所在容器、操作命令做「位置不符該項 no-op」判定。
// 對應【營業規格書 | 六、容器結構】: 卡牌四牌堆 + 顧客四容器。顯示層 EventContainer 投影亦復用本型別(EventData 的 From / To 欄)。
type ContainerKind int

const (
	ContainerNone    ContainerKind = iota // 不在任何容器(未找到)
	ContainerHand                         // 手牌
	ContainerDeck                         // 抽牌牌堆
	ContainerDrop                         // 棄牌牌堆
	ContainerExile                        // 流放牌堆
	ContainerWait                         // 排隊佇列
	ContainerSeat                         // 座位列表
	ContainerRoam                         // 遊蕩列表
	ContainerCardify                      // 卡牌化列表
)

// EventKind 投影事件類別; 對應【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 四之二】投影事件分類,
// 與事件日誌行角色對齊(【營業顯示規格書 | 6、畫面規格 | 6.10】); 類內變體由 ScopeKind / EffectStage 欄位表達。
type EventKind int

const (
	EventInstance  EventKind = iota // 實例建立 / 銷毀; 僅載位置不變的身分變更(morph 銷毀舊 + 建立新), 一般生滅由 EventContainer 的 From / To == ContainerNone 表達、效果生滅由 EventEffect 的 加入 / 結束 表達, 皆不雙發(M18 拍板)
	EventContainer                  // 卡牌移動牌堆、顧客入座 / 離場 / 遊蕩 / 卡牌化; 日誌容器類命令行
	EventProperty                   // 屬性變化(前後值); 日誌屬性類命令行
	EventScope                      // 範圍起點(頂層單位); 日誌範圍標題 + 操作元
	EventEffect                     // 效果 + 階段; 日誌效果行
	EventSelect                     // 玩家選取紀錄; 日誌 $ 選取行
	EventPhase                      // phase 切換(無本體欄位, 座標欄即新階段); 前端更新狀態列 / 標題前綴
	EventAction                     // 行動佇列入列 / 出列(M21 拍板, 七類擴八類); 前端更新行動區
)

// ScopeKind 範圍事件類別(EventScope 的類內變體); 對應【營業顯示規格書 | 6、畫面規格 | 6.10】範圍事件表。
type ScopeKind int

const (
	ScopeNone    ScopeKind = iota // 非範圍事件(零值)
	ScopePlay                     // 玩家出牌(動作; 操作元 = 卡牌 + 技能)
	ScopeGuest                    // 顧客行動(動作; 操作元 = 顧客 + 技能)
	ScopePrefix                   // 前置技能(動作; 操作元 = 技能)
	ScopeTrigger                  // 時機(Trigger 欄帶時機名)
	ScopeSettle                   // 執行結算(流程)
	ScopeManual                   // 手動結束(流程)
)

// EffectStage 效果階段(EventEffect 的類內變體; 當前階段, 非靜態效果類型);
// 對應【營業顯示規格書 | 6、畫面規格 | 6.10】效果欄位的階段值。
type EffectStage int

const (
	EffectStageNone     EffectStage = iota // 非效果事件(零值)
	EffectStageImmed                       // 立即
	EffectStageJoin                        // 加入(進佇列)
	EffectStageTrigger                     // 觸發
	EffectStageStart                       // 啟動
	EffectStageEnd                         // 結束
	EffectStageCondFail                    // 條件不成立
)

// PhaseKind 營業的執行階段; 對應【營業規格書 | 十九、核心流程】。
//
// 值採用規格書使用的中文字面值: 全域屬性「下一階段」(NextPhase) 即以此型別儲存,
// 合法跳轉值為 玩家行動 / 顧客行動 / 回合結束 / 空字串(無跳轉)。
type PhaseKind string

const (
	PhaseNone         PhaseKind = ""     // 無階段 / 無跳轉(下一階段清除後的值)
	PhaseGameStart    PhaseKind = "營業開始" // 營業開始階段; 啟動前置技能
	PhaseRoundStart   PhaseKind = "回合開始" // 回合開始階段; 回合數遞增、入座
	PhasePlayerAction PhaseKind = "玩家行動" // 玩家行動階段; 補牌 / 出牌 / 結束
	PhaseGuestAction  PhaseKind = "顧客行動" // 顧客行動階段; 彈出行動佇列
	PhaseRoundEnd     PhaseKind = "回合結束" // 回合結束階段; 推進效果、執行結算
	PhaseGameSucc     PhaseKind = "營業成功" // 營業成功階段(終止)
	PhaseGameFail     PhaseKind = "營業失敗" // 營業失敗階段(終止)
)

// PhaseJumpLegal phaseJump 命令允許設定的「下一階段」合法值集合(不含空字串)。
// 對應【營業規格書 | 二十五、操作命令清單 | phaseJump】。
var PhaseJumpLegal = map[PhaseKind]bool{
	PhasePlayerAction: true,
	PhaseGuestAction:  true,
	PhaseRoundEnd:     true,
}

// TriggerKind 觸發時機; 對應【營業規格書 | 二十二、觸發時機清單】。
//
// 觸發時機僅作為時間訊號、不攜帶資料。值採用規格書的英文名稱字面值。
// 效果靜態表格 Effect.TriggerKind 為字串欄(同本型別底層), 載入時直接轉型、無對照表。
type TriggerKind string

const (
	TriggerGameStart  TriggerKind = "gameStart"  // 營業開始; 前置技能啟動後
	TriggerRoundReady TriggerKind = "roundReady" // 回合準備; 回合數遞增後、回合開始設置前
	TriggerRoundStart TriggerKind = "roundStart" // 回合開始; 回合開始設置後
	TriggerGuestSeat  TriggerKind = "guestSeat"  // 顧客入座; 每彈出 1 位顧客入座後
	TriggerCardDraw   TriggerKind = "cardDraw"   // 手牌補充; 每張卡牌進入手牌後
	TriggerUserStart  TriggerKind = "userStart"  // 玩家開始; 補牌完成後
	TriggerUserEnd    TriggerKind = "userEnd"    // 玩家結束; 進入玩家行動結束流程後、處理剩餘手牌前
	TriggerCardDrop   TriggerKind = "cardDrop"   // 卡牌棄置; 每張卡牌進入棄牌牌堆後
	TriggerCardExile  TriggerKind = "cardExile"  // 卡牌流放; 每張卡牌進入流放牌堆後
	TriggerCardMorph  TriggerKind = "cardMorph"  // 卡牌變身; 卡牌完成變身處理流程後
	TriggerCardPlay   TriggerKind = "cardPlay"   // 玩家出牌; 實例效果列表啟動、卡牌移至棄牌牌堆後
	TriggerGuestStart TriggerKind = "guestStart" // 顧客開始; 進入顧客行動階段後
	TriggerGuestTask  TriggerKind = "guestTask"  // 顧客行動; 每彈出 1 個行動、啟動顧客行動技能後
	TriggerGuestEnd   TriggerKind = "guestEnd"   // 顧客結束; 行動佇列清空後
	TriggerRoundEnd   TriggerKind = "roundEnd"   // 回合結束; 進入回合結束階段後
	TriggerExitAny    TriggerKind = "exitAny"    // 顧客離場; 任何原因離場前
	TriggerExitSate   TriggerKind = "exitSate"   // 飽食離場; 提供滿意值離場前
	TriggerExitCalm   TriggerKind = "exitCalm"   // 生氣離場; 扣士氣離場前
	TriggerExitDone   TriggerKind = "exitDone"   // 顧客離場後; 自所在容器移除後、清理效果前
	TriggerDamage     TriggerKind = "damage"     // 士氣受損; 餐廳士氣值因 -= 實際扣減後
	TriggerGameSucc   TriggerKind = "gameSucc"   // 營業成功; 通關結算判定後
	TriggerGameFail   TriggerKind = "gameFail"   // 營業失敗; 失敗結算判定後
)

// EffectKind 效果類型; 對應【營業規格書 | 七、效果類型】。
// (效果實例型別見 effect.go 的 Effect; 此 enum 表達其靜態類型, 故以 Kind 為後綴避免撞名。)
type EffectKind int32

const (
	EffectImmed   EffectKind = iota // 立即; 不進佇列, 建立當下檢查觸發條件
	EffectTrigger                   // 觸發; 進佇列, 觸發時機 + 條件符合時執行觸發命令
	EffectPersist                   // 常駐; 進佇列, 以生命週期語意運作(啟動命令 / 結束命令)
)

// TargetKind 目標類型; 對應【營業規格書 | 八、目標類型】。
type TargetKind int32

const (
	TargetNone      TargetKind = iota // 無目標; self 為空物件
	TargetGuestPick                   // 新選顧客; 暫停流程由玩家選取
	TargetGuestSame                   // 沿用顧客; 繼承前一效果的 self
	TargetGuestRand                   // 隨機顧客; 系統隨機選取
	TargetCardPick                    // 新選手牌; 暫停流程由玩家選取
	TargetCardSame                    // 沿用手牌; 繼承前一效果的 self
	TargetCardRand                    // 隨機手牌; 系統隨機選取
)

// TriggerAfter 觸發後行為; 對應【營業規格書 | 十三、觸發後行為】(僅觸發類型適用)。
type TriggerAfter int32

const (
	TriggerAfterKeep   TriggerAfter = iota // 保留; 觸發後保留於效果佇列
	TriggerAfterRemove                     // 移除; 觸發後自效果佇列移除
)

// StackTime 堆疊時間; 對應【營業規格書 | 十六、堆疊規則】。
type StackTime int32

const (
	StackTimeStay    StackTime = iota // 不變; 堆疊時不刷新作用回合
	StackTimeRefresh                  // 刷新; 堆疊時刷新作用回合
)

// TaskKind 行動類型; 對應【營業規格書 | 五、實例結構 | 行動（Action）實例】。
// 於顧客行動階段決定封印閘門查詢對象(飽食 / 耐心)。
type TaskKind int32

const (
	TaskSate TaskKind = iota // 飽食; 查 封印飽食技能 閘門
	TaskCalm                 // 耐心; 查 封印耐心技能 閘門
)

// AssignKind 屬性修改命令的賦值符種類; 對應【營業規格書 | 十七、命令 | 1】。
//
// 由 games 的命令解析器(parse.go)產出、cores 的數值原語(value.go 的 Apply 系列)據此分派;
// 賦值符是 cores 分派的語意, 故型別下沉 cores 作單一來源。
// AssignLock(@)/ AssignUnlock(#)不帶算術式, 其餘必帶。
type AssignKind int

const (
	AssignSet    AssignKind = iota // =
	AssignAdd                      // +=
	AssignSub                      // -=
	AssignMul                      // *=
	AssignDiv                      // /=
	AssignMod                      // %=
	AssignLock                     // @ 鎖定(不帶算術式)
	AssignUnlock                   // # 解鎖(不帶算術式)
)
