package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// Game 營業實例; 一場營業的全部: 聚合狀態(全域屬性、事件型狀態、容器、流程旗標)+ 驅動引擎。
// 對應【營業規格書 | 五、實例結構 | 營業（Game）實例】【營業規格書 | 六、容器結構】。
//
// Game 是唯一真相、持有全部實例與容器(零顯示依賴);
// 前端靠事件流 + InstanceID 對照投影畫面, 不持有第二份規則狀態。
// 作為驅動引擎, Game 持遊戲資料與兩 port、委派實作 exprs.Resolver(Attr / AttrRef),
// 對外執行命令(ExecAssign / ExecOperate); 詞彙行為由 rules 套件經 Register* 裝備。
//
// 欄位紀律: 數值屬性私有, 經 Get* 取 *Value 組件讀寫(寫入紀律由 Value 把關);
// 事件欄位只能經 Event* 方法成組寫入(Last + Count + Total 等從不單獨寫)、Get* 讀取;
// 容器欄位公開, 佇列紀律由各列表型別的方法把關; 注入欄位建構時定、詞彙裝備後唯讀、self 為 run-state。
type Game struct {
	// 餐廳 / 出牌全域數值屬性
	morale       Value // 餐廳士氣值
	moraleMax    Value // 餐廳士氣值上限
	moraleShield Value // 餐廳士氣值護盾
	moraleBlock  Value // 餐廳士氣值格擋
	score        Value // 餐廳滿意值
	energy       Value // 出牌點數
	energyMax    Value // 出牌點數上限
	energyKeep   Value // 出牌點數保留(Value 固定 0, 狀態於 Lock)
	handMax      Value // 手牌張數上限
	drawMax      Value // 補牌張數上限

	// 階段 / 回合
	phaseCurr PhaseKind // 當前階段(RunPhase 踏站時經 SetPhase 設定; Emit 座標蓋章來源)
	phaseNext PhaseKind // 下一階段(跳轉目標; 系統於階段轉移時清為 PhaseNone)
	round     Value     // 當前回合數(寫屬性; 無鎖定語意、鎖定計數恆 0)
	roundMax  Value     // 回合上限(寫屬性; 無鎖定語意、鎖定計數恆 0)

	// 士氣受損事件
	damageValue int32  // 士氣受損值(最近一次實際扣減值)
	damageGuest *Guest // 士氣受損顧客(無顧客來源時為 nil)

	// 入座 / 離場 / 行動事件(回合計數於回合開始歸零)
	seatLast     *Guest // 最後入座顧客
	seatCount    int32  // 回合入座人數
	exitLast     *Guest // 最後離場顧客
	exitLastSeat int32  // 最後離場座位(自遊蕩列表離場者為 0)
	exitCount    int32  // 回合離場人數
	taskGuest    *Guest // 最後行動顧客
	taskSkill    int32  // 最後行動技能(技能編號)
	taskCount    int32  // 回合行動次數

	// 卡牌事件(Count 回合計數; Total 依 cardGroup 分組的整場累積多重集合)
	drawLast   *Card // 最後抽出卡牌
	drawCount  int32 // 回合抽牌張數
	drawTotal  Tally // 累積抽牌張數(cardGroup -> 張數)
	dropLast   *Card // 最後棄置卡牌
	dropCount  int32 // 回合棄牌張數
	dropTotal  Tally // 累積棄牌張數
	playLast   *Card // 最後出牌卡牌
	playCount  int32 // 回合出牌張數
	playTotal  Tally // 累積出牌張數
	exileLast  *Card // 最後流放卡牌
	exileCount int32 // 回合流放張數
	exileTotal Tally // 累積流放張數
	morphLast  *Card // 最後變身卡牌
	morphCount int32 // 回合變身次數
	morphOldID int32 // 變身前卡牌編號
	morphNewID int32 // 變身後卡牌編號

	// 卡牌容器
	Hand  CardList // 手牌(玩家檢視序)
	Deck  CardList // 抽牌牌堆(先進後出; 新進入者置頂)
	Drop  CardList // 棄牌牌堆(先進後出; 新進入者置頂)
	Exile CardList // 流放牌堆(先進後出; 新進入者置頂)

	// 顧客容器
	Wait    WaitList  // 排隊佇列(先進先出)
	Seat    SeatList  // 座位列表(座位編號 -> 顧客)
	Roam    GuestList // 遊蕩列表(順序無關)
	Cardify GuestList // 卡牌化列表(順序無關)

	// 效果 / 行動容器
	Effect EffectList // 效果佇列(處理時依【營業規格書 | 十五、作用順序】排序)
	Action ActionList // 行動佇列(先進先出)

	// 流程旗標與設置
	Settling    bool       // 結算旗標; 執行結算的重入防護
	Seed        int64      // 本場 PRNG 種子(執行期狀態, 供顯示 / 重現)
	StageID     int32      // 本場關卡編號(執行期狀態, 供顯示 / 重現); phaseGameStart 據此查關卡表格建置開局
	PrefixSkill []int32    // 前置技能列表(營業開始時逐一啟動; 開局建置自關卡表格填入)
	lastID      InstanceID // 實例編號產生器游標

	// 引擎注入與求值脈絡
	self      *Ref      // 當前求值脈絡的 self 綁定(run-state, 效果流程 save / restore 切換); nil 代表 self 未固定
	data      *Data     // 遊戲資料(原始表 + 衍生索引; 建構注入、營業期唯讀)
	operator  Operator  // 玩家輸入 port; 命令對象 *Pick 暫停流程由玩家選取
	rander    Rander    // 亂數 port; 命令對象 *Rand 隨機選取、deckTop auto-shuffle 洗牌
	presenter Presenter // 事件流輸出 port; 建構時正規化(nil → emptyPresenter), 發射一律經 Emit 蓋章座標

	// 詞彙裝備(rules 經 Register* 逐詞條注入; 裝備後唯讀、等同常數, 不破壞決定性)
	attrRead     map[string]AttrReadFunc     // 全域屬性讀詞彙表(含 Lock 全名詞條)
	attrWrite    map[string]AttrWriteFunc    // 全域屬性寫詞彙表
	attrRefRead  map[string]AttrRefReadFunc  // 引用屬性讀詞彙表(含 Lock 全名詞條)
	attrRefWrite map[string]AttrRefWriteFunc // 引用屬性寫詞彙表
	command      map[string]CommandFunc      // 操作命令詞彙表
	selector     map[string]SelectorFunc     // 命令對象詞彙表
	builtin      map[string]exprs.Builtin    // 內建函式註冊表(Env 帶入 exprs)
}

// NewGame 建構營業實例: 盤面空白(屬性零值、容器空、四張累積表零值可用),
// 注入本場身分(seed / 關卡編號; 同 seed + 同關卡 + 同玩家輸入 = 同一局)、遊戲資料(原始表 + 衍生索引, nil 補空殼)
// 與玩家輸入 / 亂數 / 事件流三 port(presenter nil 正規化為無輸出替身), 並預建七張空詞彙表(Register* 填入)。
// 開局盤面(手牌 / 三堆 / 排隊 / 前置技能)由 phaseGameStart 依關卡編號自關卡表格建置; self 為求值脈絡 run-state、
// 詞彙為裝備(rules.Register), 皆不入建構。
func NewGame(seed int64, stageID int32, data *Data, operator Operator, rander Rander, presenter Presenter) *Game {
	if data == nil {
		data = NewData(nil, nil)
	} // if

	if presenter == nil {
		presenter = emptyPresenter{}
	} // if

	return &Game{
		Seat:         SeatList{},
		Seed:         seed,
		StageID:      stageID,
		data:         data,
		operator:     operator,
		rander:       rander,
		presenter:    presenter,
		attrRead:     map[string]AttrReadFunc{},
		attrWrite:    map[string]AttrWriteFunc{},
		attrRefRead:  map[string]AttrRefReadFunc{},
		attrRefWrite: map[string]AttrRefWriteFunc{},
		command:      map[string]CommandFunc{},
		selector:     map[string]SelectorFunc{},
		builtin:      map[string]exprs.Builtin{},
	}
}

// NextID 配發下一個唯一實例編號(卡牌 / 顧客 / 效果共用同一序列)。
func (this *Game) NextID() InstanceID {
	this.lastID++
	return this.lastID
}

// RegisterAttrRead 註冊全域屬性讀詞條(Lock 詞條以全名為鍵, 如 moraleLock); 重複註冊後者覆蓋。表由 NewGame 預建。
func (this *Game) RegisterAttrRead(name string, read AttrReadFunc) {
	this.attrRead[name] = read
}

// RegisterAttrWrite 註冊全域屬性寫詞條; 重複註冊後者覆蓋。
func (this *Game) RegisterAttrWrite(name string, write AttrWriteFunc) {
	this.attrWrite[name] = write
}

// RegisterAttrRefRead 註冊引用屬性讀詞條(Lock 詞條以全名為鍵); 重複註冊後者覆蓋。
func (this *Game) RegisterAttrRefRead(name string, read AttrRefReadFunc) {
	this.attrRefRead[name] = read
}

// RegisterAttrRefWrite 註冊引用屬性寫詞條; 重複註冊後者覆蓋。
func (this *Game) RegisterAttrRefWrite(name string, write AttrRefWriteFunc) {
	this.attrRefWrite[name] = write
}

// RegisterCommand 註冊操作命令詞條; 重複註冊後者覆蓋。
func (this *Game) RegisterCommand(name string, run CommandFunc) {
	this.command[name] = run
}

// RegisterSelector 註冊命令對象詞條; 重複註冊後者覆蓋。
func (this *Game) RegisterSelector(name string, resolve SelectorFunc) {
	this.selector[name] = resolve
}

// RegisterBuiltin 註冊內建函式詞條(Env 帶入 exprs 求值); 重複註冊後者覆蓋。
func (this *Game) RegisterBuiltin(name string, run exprs.Builtin) {
	this.builtin[name] = run
}

// GetSheet 取原始靜態表(委派遊戲資料; 卡牌 / 顧客 / 座位 / 技能 / 設定查詢)。
func (this *Game) GetSheet() *sheeter.Sheeter {
	return this.data.GetSheet()
}

// GetSelf 取當前求值脈絡的 self 綁定; nil 代表 self 未固定。
func (this *Game) GetSelf() *Ref {
	return this.self
}

// SetSelf 設定求值脈絡 self 並回傳還原函式; 呼叫端 defer restore() 逐層還原(結算重入安全)。
// self 的唯一寫入門: 存舊值與還原目標都鎖在方法內, 呼叫端只需記得 defer。
func (this *Game) SetSelf(self *Ref) (restore func()) {
	prev := this.self
	this.self = self
	return func() { this.self = prev }
}

// GetOperator 取玩家輸入 port(命令對象 *Pick / 目標選取暫停流程用)。
func (this *Game) GetOperator() Operator {
	return this.operator
}

// GetRander 取亂數 port(*Rand 隨機選取 / 洗牌 / 隨機空位用)。
func (this *Game) GetRander() Rander {
	return this.rander
}

// Emit 發射投影事件: 蓋章座標(當前回合 / 階段)後轉交事件流輸出 port。
// 發射點只填事件本體欄位, Round / Phase 由此統一蓋章、不會漏; presenter 建構時已正規化, 免判 nil。
func (this *Game) Emit(eventData EventData) {
	eventData.Round = this.round.GetValue()
	eventData.Phase = this.phaseCurr
	this.presenter.Emit(eventData)
}

// GetMorale 取餐廳士氣值。
func (this *Game) GetMorale() *Value {
	return &this.morale
}

// GetMoraleMax 取餐廳士氣值上限。
func (this *Game) GetMoraleMax() *Value {
	return &this.moraleMax
}

// GetMoraleShield 取餐廳士氣值護盾。
func (this *Game) GetMoraleShield() *Value {
	return &this.moraleShield
}

// GetMoraleBlock 取餐廳士氣值格擋。
func (this *Game) GetMoraleBlock() *Value {
	return &this.moraleBlock
}

// GetScore 取餐廳滿意值。
func (this *Game) GetScore() *Value {
	return &this.score
}

// GetEnergy 取出牌點數。
func (this *Game) GetEnergy() *Value {
	return &this.energy
}

// GetEnergyMax 取出牌點數上限。
func (this *Game) GetEnergyMax() *Value {
	return &this.energyMax
}

// GetEnergyKeep 取出牌點數保留(純鎖屬性)。
func (this *Game) GetEnergyKeep() *Value {
	return &this.energyKeep
}

// GetHandMax 取手牌張數上限。
func (this *Game) GetHandMax() *Value {
	return &this.handMax
}

// GetDrawMax 取補牌張數上限。
func (this *Game) GetDrawMax() *Value {
	return &this.drawMax
}

// GetPhase 取當前階段(未踏站前為 PhaseNone)。
func (this *Game) GetPhase() PhaseKind {
	return this.phaseCurr
}

// SetPhase 設當前階段; RunPhase 踏站時設定, Emit 據此蓋章座標。
func (this *Game) SetPhase(phase PhaseKind) {
	this.phaseCurr = phase
}

// GetNextPhase 取下一階段(無跳轉目標時為 PhaseNone)。
func (this *Game) GetNextPhase() PhaseKind {
	return this.phaseNext
}

// SetNextPhase 設下一階段; 合法性由呼叫端把關(phaseJump 驗 PhaseJumpLegal、系統轉移時清 PhaseNone)。
func (this *Game) SetNextPhase(phase PhaseKind) {
	this.phaseNext = phase
}

// GetRound 取當前回合數。
func (this *Game) GetRound() *Value {
	return &this.round
}

// GetRoundMax 取回合上限。
func (this *Game) GetRoundMax() *Value {
	return &this.roundMax
}

// GetDamageValue 取士氣受損值(最近一次實際扣減值)。
func (this *Game) GetDamageValue() int32 {
	return this.damageValue
}

// GetDamageGuest 取士氣受損顧客(無顧客來源回 nil)。
func (this *Game) GetDamageGuest() *Guest {
	return this.damageGuest
}

// GetSeatLast 取最後入座顧客。
func (this *Game) GetSeatLast() *Guest {
	return this.seatLast
}

// GetSeatCount 取回合入座人數。
func (this *Game) GetSeatCount() int32 {
	return this.seatCount
}

// GetExitLast 取最後離場顧客。
func (this *Game) GetExitLast() *Guest {
	return this.exitLast
}

// GetExitLastSeat 取最後離場座位(自遊蕩列表離場者為 0)。
func (this *Game) GetExitLastSeat() int32 {
	return this.exitLastSeat
}

// GetExitCount 取回合離場人數。
func (this *Game) GetExitCount() int32 {
	return this.exitCount
}

// GetTaskGuest 取最後行動顧客。
func (this *Game) GetTaskGuest() *Guest {
	return this.taskGuest
}

// GetTaskSkill 取最後行動技能編號。
func (this *Game) GetTaskSkill() int32 {
	return this.taskSkill
}

// GetTaskCount 取回合行動次數。
func (this *Game) GetTaskCount() int32 {
	return this.taskCount
}

// GetDrawLast 取最後抽出卡牌。
func (this *Game) GetDrawLast() *Card {
	return this.drawLast
}

// GetDrawCount 取回合抽牌張數。
func (this *Game) GetDrawCount() int32 {
	return this.drawCount
}

// GetDrawTotal 取累積抽牌張數計數。
func (this *Game) GetDrawTotal() *Tally {
	return &this.drawTotal
}

// GetDropLast 取最後棄置卡牌。
func (this *Game) GetDropLast() *Card {
	return this.dropLast
}

// GetDropCount 取回合棄牌張數。
func (this *Game) GetDropCount() int32 {
	return this.dropCount
}

// GetDropTotal 取累積棄牌張數計數。
func (this *Game) GetDropTotal() *Tally {
	return &this.dropTotal
}

// GetPlayLast 取最後出牌卡牌。
func (this *Game) GetPlayLast() *Card {
	return this.playLast
}

// GetPlayCount 取回合出牌張數。
func (this *Game) GetPlayCount() int32 {
	return this.playCount
}

// GetPlayTotal 取累積出牌張數計數。
func (this *Game) GetPlayTotal() *Tally {
	return &this.playTotal
}

// GetExileLast 取最後流放卡牌。
func (this *Game) GetExileLast() *Card {
	return this.exileLast
}

// GetExileCount 取回合流放張數。
func (this *Game) GetExileCount() int32 {
	return this.exileCount
}

// GetExileTotal 取累積流放張數計數。
func (this *Game) GetExileTotal() *Tally {
	return &this.exileTotal
}

// GetMorphLast 取最後變身卡牌。
func (this *Game) GetMorphLast() *Card {
	return this.morphLast
}

// GetMorphCount 取回合變身次數。
func (this *Game) GetMorphCount() int32 {
	return this.morphCount
}

// GetMorphOldID 取變身前卡牌編號。
func (this *Game) GetMorphOldID() int32 {
	return this.morphOldID
}

// GetMorphNewID 取變身後卡牌編號。
func (this *Game) GetMorphNewID() int32 {
	return this.morphNewID
}

// EventDamage 設置士氣受損事件(成組: 受損值 + 來源顧客); moraleDamage 於實際扣減 > 0 時呼叫。
func (this *Game) EventDamage(value int32, source *Guest) {
	this.damageValue = value
	this.damageGuest = source
}

// EventSeat 設置入座事件(成組: 最後入座 + 回合入座人數)。
func (this *Game) EventSeat(guest *Guest) {
	this.seatLast = guest
	this.seatCount++
}

// EventExit 設置離場事件(成組: 最後離場 + 離場座位 + 回合離場人數);
// 離場座位取自顧客當下 seatID, 故須在自座位列表移除前呼叫(自遊蕩 / 排隊離場者為 0)。
func (this *Game) EventExit(guest *Guest) {
	this.exitLast = guest
	this.exitLastSeat = guest.GetSeatID()
	this.exitCount++
}

// EventTask 設置行動事件(成組: 最後行動顧客 + 行動技能 + 回合行動次數); 寫入點於玩家行動主迴圈(M14)。
func (this *Game) EventTask(guest *Guest, skillID int32) {
	this.taskGuest = guest
	this.taskSkill = skillID
	this.taskCount++
}

// EventDraw 設置抽牌事件(成組: 最後抽出 + 回合張數 + 分組累積); group 由呼叫端以 cardGroup 取得(查靜態表不進 Game)。
func (this *Game) EventDraw(card *Card, group int32) {
	this.drawLast = card
	this.drawCount++
	this.drawTotal.Add(group)
}

// EventDrop 設置棄牌事件(成組: 最後棄置 + 回合張數 + 分組累積)。
func (this *Game) EventDrop(card *Card, group int32) {
	this.dropLast = card
	this.dropCount++
	this.dropTotal.Add(group)
}

// EventPlay 設置出牌事件(成組: 最後出牌 + 回合張數 + 分組累積)。
func (this *Game) EventPlay(card *Card, group int32) {
	this.playLast = card
	this.playCount++
	this.playTotal.Add(group)
}

// EventExile 設置流放事件(成組: 最後流放 + 回合張數 + 分組累積)。
func (this *Game) EventExile(card *Card, group int32) {
	this.exileLast = card
	this.exileCount++
	this.exileTotal.Add(group)
}

// EventMorph 設置變身事件(成組: 最後變身 + 前後卡牌編號 + 回合次數)。
func (this *Game) EventMorph(card *Card, oldID, newID int32) {
	this.morphLast = card
	this.morphOldID = oldID
	this.morphNewID = newID
	this.morphCount++
}

// RoundReset 回合開始歸零全部回合計數(入座 / 離場 / 行動 / 抽牌 / 棄牌 / 出牌 / 流放 / 變身);
// Last 引用與整場累積(Total)不在此列。呼叫點於回合開始(M14)。
func (this *Game) RoundReset() {
	this.seatCount = 0
	this.exitCount = 0
	this.taskCount = 0
	this.drawCount = 0
	this.dropCount = 0
	this.playCount = 0
	this.exileCount = 0
	this.morphCount = 0
}

// SkillEffect 取技能的效果編號列表複本(新卡實例效果列表來源 = Card.SkillID → Skill.EffectID); 技能不存在回 nil。
// 複製以免共享靜態表切片。供 NewCard 載入卡牌實例效果列表、Morph 變身後重設效果共用。
func (this *Game) SkillEffect(skillID int32) []int32 {
	skill := this.GetSheet().Skill.Get(skillID)

	if skill == nil {
		return nil
	} // if

	return append([]int32(nil), skill.EffectID...)
}

// CardSkillGroup 取卡牌的技能群組編號(卡牌資料.SkillID → Skill.Group); 卡牌 / 技能資料不存在回 0。
// 供 cardRun 啟動效果列表時 skillImmune 排除免疫顧客。
func (this *Game) CardSkillGroup(cardID int32) int32 {
	card := this.GetSheet().Card.Get(cardID)

	if card == nil {
		return 0
	} // if

	skill := this.GetSheet().Skill.Get(card.SkillID)

	if skill == nil {
		return 0
	} // if

	return skill.Group
}

// EffectData 查預編譯效果(委派遊戲資料; 效果流程查 Kind / 命令 / 條件); 查無回 ok=false。
func (this *Game) EffectData(effectID int32) (meta EffectData, ok bool) {
	return this.data.GetEffect(effectID)
}

// GuestData 查顧客門檻配對(委派遊戲資料; 執行結算的飽食 / 耐心門檻迭代用); 查無回 ok=false(即無門檻)。
func (this *Game) GuestData(guestID int32) (meta GuestData, ok bool) {
	return this.data.GetGuest(guestID)
}

// RollCard 對抽獎群組做 weighted random 抽一張卡牌編號(deckRoll / handRoll / *Morph 用);
// 群組不存在 / 總權重 0 回 ok=false。
func (this *Game) RollCard(group int32) (cardID int32, ok bool) {
	award, found := this.data.GetAward(group)

	if found == false || len(award.weight) == 0 {
		return 0, false
	} // if

	return award.cardID[this.rander.Weighted(award.weight)], true
}

// Attr 委派全域屬性詞彙表, 以自身為 context 求值; 名稱即詞條鍵(Lock 詞條以全名註冊, 無後綴路由)。
func (this *Game) Attr(name string, arg []exprs.Value) (result exprs.Value, ok bool) {
	if read, known := this.attrRead[name]; known {
		return read(this, arg)
	} // if

	return exprs.Value{}, false
}

// AttrRef 以引用 ref 為主體委派引用屬性詞彙表; 名稱即詞條鍵。
func (this *Game) AttrRef(ref exprs.Ref, name string, arg []exprs.Value) (result exprs.Value, ok bool) {
	if read, known := this.attrRefRead[name]; known {
		return read(this, ref, arg)
	} // if

	return exprs.Value{}, false
}

// attrNum 讀全域屬性的屬性事件前後值(ExecAssign 收口用): @ / # 改讀鎖定計數(Lock 全名讀鍵);
// 查無詞條 / 非數值回 0(無鎖屬性的鎖定變更前後值即留零值)。
func (this *Game) attrNum(name string, op AssignKind) float64 {
	if op == AssignLock || op == AssignUnlock {
		name += "Lock"
	} // if

	result, ok := this.Attr(name, nil)

	if ok == false || result.IsNum() == false {
		return 0
	} // if

	return result.Num()
}

// attrRefNum 讀引用屬性的屬性事件前後值(ExecAssign 收口用); 規則同 attrNum。
func (this *Game) attrRefNum(ref exprs.Ref, name string, op AssignKind) float64 {
	if op == AssignLock || op == AssignUnlock {
		name += "Lock"
	} // if

	result, ok := this.AttrRef(ref, name, nil)

	if ok == false || result.IsNum() == false {
		return 0
	} // if

	return result.Num()
}

// ExecAssign 執行屬性修改命令(【營業規格書 | 十七、命令 | 1】); 回報是否實際寫入。
// 帶值賦值先求值右值(評估失敗 / 右值非數值 → no-op); 引用左值先解析引用主體(空物件 / 型別不符 / 不存在 → no-op),
// 主體為凍結中顧客一律 no-op(屬性凍結; 【營業規格書 | 二十一、流程補充 | 凍結語意】);
// 再經寫入詞彙表(全域 attrWrite / 引用 attrRefWrite)依賦值符變更狀態。名稱可寫性由 games.Validate 先行檢查。
// 屬性事件於此收口: 進了寫入詞條就發、沒進就不發(閘門前夭折無事件; 鎖定拒寫以 Before == After 表達; M18 拍板),
// 前後值經讀詞條取得、@ / # 載鎖定計數, 引用左值帶對象編號、全域留零值。
func (this *Game) ExecAssign(base, refAttr string, isRef bool, op AssignKind, value *exprs.Expr) (changed bool) {
	n := float64(0)

	if op != AssignLock && op != AssignUnlock { // @ # 不帶右值
		result, ok := value.Eval(this.Env())

		if ok == false || result.IsNum() == false {
			return false // 算術評估失敗 / 右值非數值 → no-op
		} // if

		n = result.Num()
	} // if

	if isRef {
		owner, ok := this.Attr(base, nil)

		if ok == false || owner.IsRef() == false {
			return false // 引用解析為空物件 / 型別不符 / 不存在 → no-op
		} // if

		if guest, isGuest := AsGuest(owner.Ref()); isGuest && this.IsFrozen(guest) {
			return false // 屬性凍結: 寫入主體為凍結中顧客 → no-op
		} // if

		write, known := this.attrRefWrite[refAttr]

		if known == false {
			return false
		} // if

		dataID, instanceID := RefTarget(owner.Ref())
		before := this.attrRefNum(owner.Ref(), refAttr, op)
		changed = write(this, owner.Ref(), op, n)
		this.Emit(EventData{Kind: EventProperty, DataID: dataID, InstanceID: instanceID, Attr: refAttr, Op: op, Operand: n, Before: before, After: this.attrRefNum(owner.Ref(), refAttr, op)})
		return changed
	} // if

	write, known := this.attrWrite[base]

	if known == false {
		return false
	} // if

	before := this.attrNum(base, op)
	changed = write(this, op, n)
	this.Emit(EventData{Kind: EventProperty, Attr: base, Op: op, Operand: n, Before: before, After: this.attrNum(base, op)})
	return changed
}

// ExecOperate 執行操作命令(【營業規格書 | 十七、命令 | 2】【二十五、操作命令清單】); 走法 X: games 持 AST 型別 switch、引擎做分派。
// 流程: 求值命令對象 [...] 參數 → selectObject 解析作用集合 → 求值其餘參數 → 查 command 表 → 對集合 fan-out。
// **任一參數評估失敗 → 整動作 no-op**(對齊規格); 命令對象 / verb 未登錄(Validate 應先擋)亦 no-op。
// 逐元素的型別 / 位置不符 no-op、空集合整體 no-op 由各 verb 本體處理。
func (this *Game) ExecOperate(verb, selectorName string, selectorParam, arg []*exprs.Expr) {
	selectorValue, ok := this.evalAll(selectorParam)

	if ok == false {
		return // 命令對象參數評估失敗 → 整動作 no-op
	} // if

	target, known := this.selectObject(selectorName, selectorValue)

	if known == false {
		return // 命令對象未登錄 → no-op
	} // if

	// none / 篩空正規化: none → nil(全域掃描記號)、其他命令對象篩空 → 非 nil 空切片(verb 以 target == nil 判 none; 見 SelectorNone)
	if selectorName == SelectorNone {
		target = nil
	} else if target == nil {
		target = []InstanceID{}
	} // if

	argValue, ok := this.evalAll(arg)

	if ok == false {
		return // 其餘參數評估失敗 → 整動作 no-op
	} // if

	run, found := this.command[verb]

	if found == false {
		return // 未知命令 → no-op
	} // if

	run(this, target, argValue)
}

// selectObject 解析命令對象為作用對象集合(【營業規格書 | 二十四、命令對象清單】); ok=false 代表命令對象名稱未登錄。
// arg 為 [...] 內參數的已求值結果(求值由呼叫端 M9 ExecOperate 負責); selector 自身不碰 exprs。
// 回身分集 []InstanceID(非具型別實例): M9 verb 自行 locate 取實例 + 查容器位置(位置不符 no-op 本就要查),
// 使本表保持同構、與 Self / effect 的 InstanceID 身分模型一致。
func (this *Game) selectObject(name string, arg []exprs.Value) (result []InstanceID, ok bool) {
	resolve, known := this.selector[name]

	if known == false {
		return nil, false
	} // if

	return resolve(this, arg), true
}

// evalAll 依序求值一串算術式; 任一失敗回 ok=false(供 ExecOperate 套用「任一參數評估失敗 → 整動作 no-op」)。
func (this *Game) evalAll(expr []*exprs.Expr) (result []exprs.Value, ok bool) {
	result = make([]exprs.Value, 0, len(expr))

	for _, itor := range expr {
		value, valid := itor.Eval(this.Env())

		if valid == false {
			return nil, false
		} // if

		result = append(result, value)
	} // for

	return result, true
}

// LocateCard 以實例編號自四牌堆找出卡牌及其所在容器; 未命中回 (nil, ContainerNone, false)。
// 供操作命令取實例 + 查容器位置(位置不符 no-op); 卡牌僅存在於 手牌 / 抽牌 / 棄牌 / 流放。
func (this *Game) LocateCard(id InstanceID) (card *Card, where ContainerKind, ok bool) {
	if found := this.Hand.Find(id); found != nil {
		return found, ContainerHand, true
	} // if

	if found := this.Deck.Find(id); found != nil {
		return found, ContainerDeck, true
	} // if

	if found := this.Drop.Find(id); found != nil {
		return found, ContainerDrop, true
	} // if

	if found := this.Exile.Find(id); found != nil {
		return found, ContainerExile, true
	} // if

	return nil, ContainerNone, false
}

// LocateGuest 以實例編號自顧客四容器找出顧客及其所在容器; 未命中回 (nil, ContainerNone, false)。
// 供操作命令取顧客實例 + 查容器位置(型別不符 / 位置不符 no-op); 顧客存在於 座位 / 排隊 / 遊蕩 / 卡牌化。
func (this *Game) LocateGuest(id InstanceID) (guest *Guest, where ContainerKind, ok bool) {
	if found := this.Seat.Find(id); found != nil {
		return found, ContainerSeat, true
	} // if

	if found := this.Wait.Find(id); found != nil {
		return found, ContainerWait, true
	} // if

	if found := this.Roam.Find(id); found != nil {
		return found, ContainerRoam, true
	} // if

	if found := this.Cardify.Find(id); found != nil {
		return found, ContainerCardify, true
	} // if

	return nil, ContainerNone, false
}

// IsFrozen 回報顧客是否凍結中(位於卡牌化列表; 【營業規格書 | 二十一、流程補充 | 凍結語意】)。
// 容器成員身分即凍結狀態的單一來源; freeze 欄位僅作解凍補回結束回合的起算錨點、不兼任旗標
// (round 0 期間卡牌化時 freeze == 0, 與未凍結零值無法區分)。
func (this *Game) IsFrozen(guest *Guest) bool {
	return this.Cardify.Find(guest.GetInstanceID()) != nil
}

// Env 組裝求值期環境: 以自身為條件對象 Resolver、帶入裝備的內建函式註冊表。
func (this *Game) Env() exprs.Env {
	return exprs.Env{Resolver: this, Builtin: this.builtin}
}

// emptyPresenter 無輸出替身: NewGame 對 presenter == nil 正規化為此(比照 data nil 補空殼)、Emit 靜默丟棄。
// Presenter 為純輸出 port(Emit 無回傳、零決定性影響), no-op 在語意上無損, 發射端因此免判 nil。
type emptyPresenter struct{}

func (this emptyPresenter) Emit(eventData EventData) {
}

// 編譯期確認 Game 滿足 exprs.Resolver(條件對象求值的接縫)。
var _ exprs.Resolver = (*Game)(nil)
