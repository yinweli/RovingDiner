package cores

// Value 數值實例；執行期可變屬性的統一容器。
// 對應【營業規格書 | 五、實例結構 | 數值（Value）實例】。
//
// 所有可變屬性都以本型別表達：一般運算（= += -= *= /= %=）改 Value，
// 鎖定 / 解鎖（@ #）改 Lock。Lock > 0 時一般運算對該屬性為 no-op。
// 純計數型屬性（如 封印卡牌 / 出牌點數保留）Value 固定為 0，狀態全表達於 Lock。
type Value struct {
	Value int32 // 屬性當前數值（整數）
	Lock  int32 // 屬性鎖定狀態計數；由 @ / # 控制；> 0 時一般運算為 no-op
}

// Locked 回傳屬性是否處於鎖定狀態（鎖定計數 > 0）。
func (this Value) Locked() bool {
	return this.Lock > 0
}

// Self 效果建立時綁定的對象（卡牌兼顧客引用）；對應【營業規格書 | 八、目標類型】。
// 兩欄皆 nil 代表空物件（無目標）；至多一欄非 nil。
type Self struct {
	Card  *Card  // self 為卡牌時非 nil
	Guest *Guest // self 為顧客時非 nil
}

// IsNone 回傳 self 是否為空物件。
func (this Self) IsNone() bool {
	return this.Card == nil && this.Guest == nil
}

// Game 營業實例；持有全域屬性與事件型狀態。
// 對應【營業規格書 | 五、實例結構 | 營業（Game）實例】。
type Game struct {
	// 餐廳 / 出牌全域數值屬性
	Morale       Value // 餐廳士氣值
	MoraleMax    Value // 餐廳士氣值上限
	MoraleShield Value // 餐廳士氣值護盾
	MoraleBlock  Value // 餐廳士氣值格擋
	Score        Value // 餐廳滿意值
	Energy       Value // 出牌點數
	EnergyMax    Value // 出牌點數上限
	EnergyKeep   Value // 出牌點數保留（Value 固定 0，狀態於 Lock）
	HandMax      Value // 手牌張數上限
	DrawMax      Value // 補牌張數上限

	// 階段 / 回合
	NextPhase PhaseType // 下一階段（跳轉目標；系統於階段轉移時清為 PhaseTypeNone）
	Round     int32     // 當前回合數
	RoundMax  int32     // 回合上限

	// 士氣受損事件
	DamageValue int32  // 士氣受損值（最近一次實際扣減值）
	DamageGuest *Guest // 士氣受損顧客（無顧客來源時為 nil）

	// 入座 / 離場 / 行動事件（回合計數於回合開始歸零）
	SeatLast     *Guest // 最後入座顧客
	SeatCount    int32  // 回合入座人數
	ExitLast     *Guest // 最後離場顧客
	ExitLastSeat int32  // 最後離場座位（自遊蕩列表離場者為 0）
	ExitCount    int32  // 回合離場人數
	TaskGuest    *Guest // 最後行動顧客
	TaskSkill    int32  // 最後行動技能（技能編號）
	TaskCount    int32  // 回合行動次數

	// 卡牌事件（Count 回合計數；Total 依 cardGroup 分組的整場累積多重集合）
	DrawLast   *Card           // 最後抽出卡牌
	DrawCount  int32           // 回合抽牌張數
	DrawTotal  map[int32]int32 // 累積抽牌張數（cardGroup -> 張數）
	DropLast   *Card           // 最後棄置卡牌
	DropCount  int32           // 回合棄牌張數
	DropTotal  map[int32]int32 // 累積棄牌張數
	PlayLast   *Card           // 最後出牌卡牌
	PlayCount  int32           // 回合出牌張數
	PlayTotal  map[int32]int32 // 累積出牌張數
	ExileLast  *Card           // 最後流放卡牌
	ExileCount int32           // 回合流放張數
	ExileTotal map[int32]int32 // 累積流放張數
	MorphLast  *Card           // 最後變身卡牌
	MorphCount int32           // 回合變身次數
	MorphOldID int32           // 變身前卡牌編號
	MorphNewID int32           // 變身後卡牌編號
}

// Card 卡牌實例；對應【營業規格書 | 五、實例結構 | 卡牌（Card）實例】。
type Card struct {
	InstanceID  InstanceID // 實例唯一識別碼
	CardID      int32      // 引用對應卡牌資料
	Cost        Value      // 出牌費用
	ExtraRunMin Value      // 額外發動次數下限
	ExtraRunMax Value      // 額外發動次數上限
	Keep        Value      // 不棄卡牌（Value 固定 0，狀態於 Lock）
	Seal        Value      // 封印卡牌（Value 固定 0，狀態於 Lock）
	PlayExile   Value      // 出牌後流放（Value 固定 0，狀態於 Lock）
	UnplayExile Value      // 未出牌流放（Value 固定 0，狀態於 Lock）
	EffectID    []int32    // 實例效果列表（效果編號多重集合；允許重複）
	Cardify     *Guest     // 卡牌化來源顧客（初值 nil）
}

// Guest 顧客實例；對應【營業規格書 | 五、實例結構 | 顧客（Guest）實例】。
//
// SateMax（飽食值離場線）為【營業規格書 | 二十三、屬性清單 | 顧客引用屬性】登記的
// 「寫鎖」屬性 sateMax，故以 Value 儲存；§五 顧客實例表目前漏列此欄（待規格補正）。
type Guest struct {
	InstanceID   InstanceID      // 實例唯一識別碼
	GuestID      int32           // 引用對應顧客資料
	SeatID       int32           // 占用的座位編號（位於遊蕩 / 卡牌化列表時為 0）
	Score        Value           // 滿意值
	ScoreMax     Value           // 滿意值上限
	Morale       Value           // 士氣值
	MoraleMax    Value           // 士氣值上限
	Sate         Value           // 飽食值
	SateMax      Value           // 飽食值離場線
	Calm         Value           // 耐心值
	SateSeal     Value           // 封印飽食技能（Value 固定 0，狀態於 Lock）
	CalmSeal     Value           // 封印耐心技能（Value 固定 0，狀態於 Lock）
	SateHit      map[int32]bool  // 已觸發飽食門檻集合（避免重複觸發）
	CalmHit      map[int32]bool  // 已觸發耐心門檻集合
	EffectImmune map[int32]int32 // 效果免疫群組（效果群組編號 -> 鎖定計數）
	SkillImmune  map[int32]int32 // 技能免疫群組（技能群組編號 -> 鎖定計數）
	Freeze       int32           // 凍結起始回合（卡牌化時記錄；解凍後重置 0）
}

// Effect 效果實例；對應【營業規格書 | 五、實例結構 | 效果（Effect）實例】。
// 效果的靜態類型（立即 / 觸發 / 常駐）見靜態表格 Effect.Type 與 define.go 的 EffectType。
type Effect struct {
	InstanceID InstanceID // 實例唯一識別碼
	EffectID   int32      // 引用對應效果資料
	Expire     int32      // 結束回合（作用回合 = 0 時為 0，代表整場保留）
	Stack      int32      // 當前堆疊層數
	Self       Self       // self 物件（空物件 / 顧客 / 卡牌）
}

// Action 行動實例；對應【營業規格書 | 五、實例結構 | 行動（Action）實例】。
type Action struct {
	Guest    *Guest   // 顧客實例
	TaskType TaskType // 行動類型（飽食 / 耐心）
	SkillID  int32    // 顧客行動階段彈出後啟動的技能編號
}
