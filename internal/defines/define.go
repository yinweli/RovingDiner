package defines

// NoneID 空實例編號；表示尚未配發或不存在的實例。
const NoneID InstanceID = 0

const (
	EventInstance  Event = iota // 實例建立 / 銷毀（卡牌 / 顧客 / 效果）
	EventContainer              // 卡牌移動牌堆、顧客入座 / 離場 / 遊蕩 / 卡牌化
	EventProperty               // sate / calm / score / 鎖定計數 變化
	EventTrigger                // 觸發時機到達、某效果觸發
	EventPhase                  // phase 切換
)

// 階段；值採用規格書使用的中文字面值。
const (
	PhaseNone         Phase = ""     // 無階段 / 無跳轉（下一階段清除後的值）
	PhaseGameStart    Phase = "營業開始" // 營業開始階段；啟動前置技能
	PhaseRoundStart   Phase = "回合開始" // 回合開始階段；回合數遞增、入座
	PhasePlayerAction Phase = "玩家行動" // 玩家行動階段；補牌 / 出牌 / 結束
	PhaseGuestAction  Phase = "顧客行動" // 顧客行動階段；彈出行動佇列
	PhaseRoundEnd     Phase = "回合結束" // 回合結束階段；推進效果、執行結算
	PhaseGameSucc     Phase = "營業成功" // 營業成功階段（終止）
	PhaseGameFail     Phase = "營業失敗" // 營業失敗階段（終止）
)

// PhaseJumpLegal phaseJump 命令允許設定的「下一階段」合法值集合（不含空字串）。
// 對應【營業規格書 | 二十五、操作命令清單 | phaseJump】。
var PhaseJumpLegal = map[Phase]bool{
	PhasePlayerAction: true,
	PhaseGuestAction:  true,
	PhaseRoundEnd:     true,
}

// 觸發時機；值採用規格書的英文名稱字面值。
const (
	TriggerGameStart  Trigger = "gameStart"  // 營業開始；前置技能啟動後
	TriggerRoundReady Trigger = "roundReady" // 回合準備；回合數遞增後、回合開始設置前
	TriggerRoundStart Trigger = "roundStart" // 回合開始；回合開始設置後
	TriggerGuestSeat  Trigger = "guestSeat"  // 顧客入座；每彈出 1 位顧客入座後
	TriggerCardDraw   Trigger = "cardDraw"   // 手牌補充；每張卡牌進入手牌後
	TriggerUserStart  Trigger = "userStart"  // 玩家開始；補牌完成後
	TriggerUserEnd    Trigger = "userEnd"    // 玩家結束；進入玩家行動結束流程後、處理剩餘手牌前
	TriggerCardDrop   Trigger = "cardDrop"   // 卡牌棄置；每張卡牌進入棄牌牌堆後
	TriggerCardExile  Trigger = "cardExile"  // 卡牌流放；每張卡牌進入流放牌堆後
	TriggerCardMorph  Trigger = "cardMorph"  // 卡牌變身；卡牌完成變身處理流程後
	TriggerCardPlay   Trigger = "cardPlay"   // 玩家出牌；實例效果列表啟動、卡牌移至棄牌牌堆後
	TriggerGuestStart Trigger = "guestStart" // 顧客開始；進入顧客行動階段後
	TriggerGuestTask  Trigger = "guestTask"  // 顧客行動；每彈出 1 個行動、啟動顧客行動技能後
	TriggerGuestEnd   Trigger = "guestEnd"   // 顧客結束；行動佇列清空後
	TriggerRoundEnd   Trigger = "roundEnd"   // 回合結束；進入回合結束階段後
	TriggerExitAny    Trigger = "exitAny"    // 顧客離場；任何原因離場前
	TriggerExitSate   Trigger = "exitSate"   // 飽食離場；提供滿意值離場前
	TriggerExitCalm   Trigger = "exitCalm"   // 生氣離場；扣士氣離場前
	TriggerExitDone   Trigger = "exitDone"   // 顧客離場後；自所在容器移除後、清理效果前
	TriggerDamage     Trigger = "damage"     // 士氣受損；餐廳士氣值因 -= 實際扣減後
	TriggerGameSucc   Trigger = "gameSucc"   // 營業成功；通關結算判定後
	TriggerGameFail   Trigger = "gameFail"   // 營業失敗；失敗結算判定後
)

// 效果類型。
const (
	EffectImmed   Effect = iota // 立即；不進佇列，建立當下檢查觸發條件
	EffectTrigger               // 觸發；進佇列，觸發時機 + 條件符合時執行觸發命令
	EffectPersist               // 常駐；進佇列，以生命週期語意運作（啟動命令 / 結束命令）
)

// 目標類型。
const (
	TargetTypeNone      TargetType = iota // 無目標；self 為空物件
	TargetTypeGuestPick                   // 新選顧客；暫停流程由玩家選取
	TargetTypeGuestSame                   // 沿用顧客；繼承前一效果的 self
	TargetTypeGuestRand                   // 隨機顧客；系統隨機選取
	TargetTypeCardPick                    // 新選手牌；暫停流程由玩家選取
	TargetTypeCardSame                    // 沿用手牌；繼承前一效果的 self
	TargetTypeCardRand                    // 隨機手牌；系統隨機選取
)

// 觸發後行為。
const (
	TriggerAfterKeep   TriggerAfter = iota // 保留；觸發後保留於效果佇列
	TriggerAfterRemove                     // 移除；觸發後自效果佇列移除
)

// 堆疊時間。
const (
	StackTimeStay    StackTime = iota // 不變；堆疊時不刷新作用回合
	StackTimeRefresh                  // 刷新；堆疊時刷新作用回合
)

// 行動類型。
const (
	TaskSate Task = iota // 飽食；查 封印飽食技能 閘門
	TaskCalm             // 耐心；查 封印耐心技能 閘門
)

// 命令對象；對應【營業規格書 | 二十四、命令對象清單】。
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

// 全域屬性名稱；對應【營業規格書 | 二十三、屬性清單】主表。
const (
	PropertyCardifySize  Property = "cardifySize"  // 卡牌化列表當下大小
	PropertyDamageGuest  Property = "damageGuest"  // 士氣受損顧客（顧客引用）
	PropertyDamageValue  Property = "damageValue"  // 士氣受損值
	PropertyDeckSize     Property = "deckSize"     // 抽牌牌堆卡牌張數（查詢函式）
	PropertyDrawCount    Property = "drawCount"    // 回合抽牌張數
	PropertyDrawLast     Property = "drawLast"     // 最後抽出卡牌（卡牌引用）
	PropertyDrawMax      Property = "drawMax"      // 補牌張數上限（寫鎖）
	PropertyDrawTotal    Property = "drawTotal"    // 累積抽牌張數（查詢函式）
	PropertyDropCount    Property = "dropCount"    // 回合棄牌張數
	PropertyDropLast     Property = "dropLast"     // 最後棄置卡牌（卡牌引用）
	PropertyDropSize     Property = "dropSize"     // 棄牌牌堆卡牌張數（查詢函式）
	PropertyDropTotal    Property = "dropTotal"    // 累積棄牌張數（查詢函式）
	PropertyEnergy       Property = "energy"       // 出牌點數（寫鎖）
	PropertyEnergyMax    Property = "energyMax"    // 出牌點數上限（寫鎖）
	PropertyEnergyKeep   Property = "energyKeep"   // 出牌點數保留（鎖）
	PropertyExileCount   Property = "exileCount"   // 回合流放張數
	PropertyExileLast    Property = "exileLast"    // 最後流放卡牌（卡牌引用）
	PropertyExileSize    Property = "exileSize"    // 流放牌堆卡牌張數（查詢函式）
	PropertyExileTotal   Property = "exileTotal"   // 累積流放張數（查詢函式）
	PropertyExitCount    Property = "exitCount"    // 回合離場人數
	PropertyExitLast     Property = "exitLast"     // 最後離場顧客（顧客引用）
	PropertyExitLastSeat Property = "exitLastSeat" // 最後離場座位
	PropertyGuestSize    Property = "guestSize"    // 店內顧客當下總數（衍生）
	PropertyHandMax      Property = "handMax"      // 手牌張數上限（寫鎖）
	PropertyHandSize     Property = "handSize"     // 手牌卡牌張數（查詢函式）
	PropertyMorale       Property = "morale"       // 餐廳士氣值（寫鎖）
	PropertyMoraleBlock  Property = "moraleBlock"  // 餐廳士氣值格擋（寫鎖）
	PropertyMoraleMax    Property = "moraleMax"    // 餐廳士氣值上限（寫鎖）
	PropertyMoraleShield Property = "moraleShield" // 餐廳士氣值護盾（寫鎖）
	PropertyMorphCount   Property = "morphCount"   // 回合變身次數
	PropertyMorphLast    Property = "morphLast"    // 最後變身卡牌（卡牌引用）
	PropertyMorphNewID   Property = "morphNewID"   // 變身後卡牌編號
	PropertyMorphOldID   Property = "morphOldID"   // 變身前卡牌編號
	PropertyNextPhase    Property = "nextPhase"    // 下一階段（唯讀；僅 phaseJump 可寫）
	PropertyPlayCount    Property = "playCount"    // 回合出牌張數
	PropertyPlayLast     Property = "playLast"     // 最後出牌卡牌（卡牌引用）
	PropertyPlayTotal    Property = "playTotal"    // 累積出牌張數（查詢函式）
	PropertyRoamSize     Property = "roamSize"     // 遊蕩列表當下大小
	PropertyRound        Property = "round"        // 回合（寫）
	PropertyRoundLeft    Property = "roundLeft"    // 剩餘回合（衍生；寫入轉譯為回合上限）
	PropertyRoundMax     Property = "roundMax"     // 回合上限（寫）
	PropertyScore        Property = "score"        // 餐廳滿意值（寫鎖）
	PropertySeatCount    Property = "seatCount"    // 回合入座人數
	PropertySeatLast     Property = "seatLast"     // 最後入座顧客（顧客引用）
	PropertySeatLeft     Property = "seatLeft"     // 剩餘座位
	PropertySeatSize     Property = "seatSize"     // 座位列表當下大小
	PropertySelf         Property = "self"         // 自身（卡牌兼顧客引用）
	PropertyTableCount   Property = "tableCount"   // 桌數查詢（查詢函式）
	PropertyTableGuest   Property = "tableGuest"   // 桌次顧客數（查詢函式）
	PropertyTableSize    Property = "tableSize"    // 桌數總量（衍生）
	PropertyTaskCount    Property = "taskCount"    // 回合行動次數
	PropertyTaskGuest    Property = "taskGuest"    // 最後行動顧客（顧客引用）
	PropertyTaskSize     Property = "taskSize"     // 行動佇列當下大小
	PropertyTaskSkill    Property = "taskSkill"    // 最後行動技能
	PropertyWaitSize     Property = "waitSize"     // 排隊佇列當下大小
)

// dot-syntax 引用屬性名稱；對應【營業規格書 | 二十三、屬性清單 | 卡牌引用屬性 / 顧客引用屬性】。
const (
	// 卡牌引用屬性

	AttrCardEffect  Attr = "cardEffect"  // 卡牌帶效果數（查詢函式）
	AttrCardGroup   Attr = "cardGroup"   // 卡牌群組編號
	AttrCardID      Attr = "cardID"      // 卡牌編號
	AttrCardify     Attr = "cardify"     // 卡牌化來源顧客
	AttrCardSeal    Attr = "cardSeal"    // 封印卡牌（鎖）
	AttrCost        Attr = "cost"        // 出牌費用（寫鎖）
	AttrExtraRunMax Attr = "extraRunMax" // 額外發動次數上限（寫鎖）
	AttrExtraRunMin Attr = "extraRunMin" // 額外發動次數下限（寫鎖）
	AttrInDeck      Attr = "inDeck"      // 位於抽牌牌堆（布林）
	AttrInDrop      Attr = "inDrop"      // 位於棄牌牌堆（布林）
	AttrInExile     Attr = "inExile"     // 位於流放牌堆（布林）
	AttrInHand      Attr = "inHand"      // 位於手牌（布林）
	AttrKeep        Attr = "keep"        // 不棄卡牌（鎖）
	AttrPlayExile   Attr = "playExile"   // 出牌後流放（鎖）
	AttrUnplayExile Attr = "unplayExile" // 未出牌流放（鎖）

	// 顧客引用屬性

	AttrCalm         Attr = "calm"         // 耐心值（寫鎖）
	AttrCalmHit      Attr = "calmHit"      // 已觸發耐心門檻（讀取為列表大小）
	AttrCalmSeal     Attr = "calmSeal"     // 封印耐心技能（鎖）
	AttrEffectImmune Attr = "effectImmune" // 效果免疫群組（查詢函式）
	AttrFreeze       Attr = "freeze"       // 凍結起始回合
	AttrGuestID      Attr = "guestID"      // 顧客編號
	AttrMoraleMax    Attr = "moraleMax"    // 士氣值上限（寫鎖）
	AttrNearSize     Attr = "nearSize"     // 鄰桌當下大小
	AttrSameSize     Attr = "sameSize"     // 同桌當下大小
	AttrSate         Attr = "sate"         // 飽食值（寫鎖）
	AttrSateHit      Attr = "sateHit"      // 已觸發飽食門檻（讀取為列表大小）
	AttrSateMax      Attr = "sateMax"      // 飽食值離場線（寫鎖）
	AttrSateSeal     Attr = "sateSeal"     // 封印飽食技能（鎖）
	AttrScoreMax     Attr = "scoreMax"     // 滿意值上限（寫鎖）
	AttrSeatID       Attr = "seatID"       // 座位編號
	AttrSkillImmune  Attr = "skillImmune"  // 技能免疫群組（查詢函式）

	// 卡牌與顧客共用

	AttrEffectGroup Attr = "effectGroup" // 已帶群組效果數（查詢函式）
	AttrEffectStack Attr = "effectStack" // 已帶效果層數（查詢函式）
	AttrMorale      Attr = "morale"      // 士氣值（顧客；寫鎖）
	AttrScore       Attr = "score"       // 滿意值（顧客；寫鎖）
)

// 操作命令；對應【營業規格書 | 二十五、操作命令清單】。
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
	CommandGuestPart   Command = "guestPart"   // 改變顧客部位（規格待定）
	CommandGuestReturn Command = "guestReturn" // 顧客回座
	CommandGuestRoam   Command = "guestRoam"   // 顧客遊蕩
	CommandGuestSeat   Command = "guestSeat"   // 顧客入座
	CommandGuestSkin   Command = "guestSkin"   // 改變顧客外觀（規格待定）
	CommandGuestSpawn  Command = "guestSpawn"  // 產生顧客
	CommandWaitAdd     Command = "waitAdd"     // 加顧客入排隊佇列

	// 技能免疫 / 行動 / 階段

	CommandSkillImmuneAdd Command = "skillImmuneAdd" // 賦予技能免疫
	CommandSkillImmuneDel Command = "skillImmuneDel" // 解除技能免疫
	CommandTaskAdd        Command = "taskAdd"        // 加入行動
	CommandPhaseJump      Command = "phaseJump"      // 階段跳轉
)
