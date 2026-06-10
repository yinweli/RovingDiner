package cores

// Game 營業實例；持有全域屬性與事件型狀態。
// 對應【營業規格書 | 五、實例結構 | 營業（Game）實例】。
//
// 欄位私有：數值屬性經 Get* 取 *Value 組件讀寫（寫入紀律由 Value 把關）；
// 事件欄位只能經 Event* 方法成組寫入（Last + Count + Total 等從不單獨寫）、Get* 讀取。
type Game struct {
	// 餐廳 / 出牌全域數值屬性
	morale       Value // 餐廳士氣值
	moraleMax    Value // 餐廳士氣值上限
	moraleShield Value // 餐廳士氣值護盾
	moraleBlock  Value // 餐廳士氣值格擋
	score        Value // 餐廳滿意值
	energy       Value // 出牌點數
	energyMax    Value // 出牌點數上限
	energyKeep   Value // 出牌點數保留（Value 固定 0，狀態於 Lock）
	handMax      Value // 手牌張數上限
	drawMax      Value // 補牌張數上限

	// 階段 / 回合
	nextPhase PhaseKind // 下一階段（跳轉目標；系統於階段轉移時清為 PhaseNone）
	round     Value     // 當前回合數（寫屬性；無鎖定語意、鎖定計數恆 0）
	roundMax  Value     // 回合上限（寫屬性；無鎖定語意、鎖定計數恆 0）

	// 士氣受損事件
	damageValue int32  // 士氣受損值（最近一次實際扣減值）
	damageGuest *Guest // 士氣受損顧客（無顧客來源時為 nil）

	// 入座 / 離場 / 行動事件（回合計數於回合開始歸零）
	seatLast     *Guest // 最後入座顧客
	seatCount    int32  // 回合入座人數
	exitLast     *Guest // 最後離場顧客
	exitLastSeat int32  // 最後離場座位（自遊蕩列表離場者為 0）
	exitCount    int32  // 回合離場人數
	taskGuest    *Guest // 最後行動顧客
	taskSkill    int32  // 最後行動技能（技能編號）
	taskCount    int32  // 回合行動次數

	// 卡牌事件（Count 回合計數；Total 依 cardGroup 分組的整場累積多重集合）
	drawLast   *Card // 最後抽出卡牌
	drawCount  int32 // 回合抽牌張數
	drawTotal  Tally // 累積抽牌張數（cardGroup -> 張數）
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
}

// NewGame 建構空白營業實例（屬性零值；四張累積表零值可用、Add 自建）。
func NewGame() *Game {
	return &Game{}
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

// GetEnergyKeep 取出牌點數保留（純鎖屬性）。
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

// GetNextPhase 取下一階段（無跳轉目標時為 PhaseNone）。
func (this *Game) GetNextPhase() PhaseKind {
	return this.nextPhase
}

// SetNextPhase 設下一階段；合法性由呼叫端把關（phaseJump 驗 PhaseJumpLegal、系統轉移時清 PhaseNone）。
func (this *Game) SetNextPhase(phase PhaseKind) {
	this.nextPhase = phase
}

// GetRound 取當前回合數。
func (this *Game) GetRound() *Value {
	return &this.round
}

// GetRoundMax 取回合上限。
func (this *Game) GetRoundMax() *Value {
	return &this.roundMax
}

// GetDamageValue 取士氣受損值（最近一次實際扣減值）。
func (this *Game) GetDamageValue() int32 {
	return this.damageValue
}

// GetDamageGuest 取士氣受損顧客（無顧客來源回 nil）。
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

// GetExitLastSeat 取最後離場座位（自遊蕩列表離場者為 0）。
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

// EventDamage 設置士氣受損事件（成組：受損值 + 來源顧客）；moraleDamage 於實際扣減 > 0 時呼叫。
func (this *Game) EventDamage(value int32, source *Guest) {
	this.damageValue = value
	this.damageGuest = source
}

// EventSeat 設置入座事件（成組：最後入座 + 回合入座人數）。
func (this *Game) EventSeat(guest *Guest) {
	this.seatLast = guest
	this.seatCount++
}

// EventExit 設置離場事件（成組：最後離場 + 離場座位 + 回合離場人數）；
// 離場座位取自顧客當下 seatID，故須在自座位列表移除前呼叫（自遊蕩 / 排隊離場者為 0）。
func (this *Game) EventExit(guest *Guest) {
	this.exitLast = guest
	this.exitLastSeat = guest.GetSeatID()
	this.exitCount++
}

// EventTask 設置行動事件（成組：最後行動顧客 + 行動技能 + 回合行動次數）；寫入點於玩家行動主迴圈（M14）。
func (this *Game) EventTask(guest *Guest, skillID int32) {
	this.taskGuest = guest
	this.taskSkill = skillID
	this.taskCount++
}

// EventDraw 設置抽牌事件（成組：最後抽出 + 回合張數 + 分組累積）；group 由呼叫端以 cardGroup 取得（查靜態表不進 Game）。
func (this *Game) EventDraw(card *Card, group int32) {
	this.drawLast = card
	this.drawCount++
	this.drawTotal.Add(group)
}

// EventDrop 設置棄牌事件（成組：最後棄置 + 回合張數 + 分組累積）。
func (this *Game) EventDrop(card *Card, group int32) {
	this.dropLast = card
	this.dropCount++
	this.dropTotal.Add(group)
}

// EventPlay 設置出牌事件（成組：最後出牌 + 回合張數 + 分組累積）。
func (this *Game) EventPlay(card *Card, group int32) {
	this.playLast = card
	this.playCount++
	this.playTotal.Add(group)
}

// EventExile 設置流放事件（成組：最後流放 + 回合張數 + 分組累積）。
func (this *Game) EventExile(card *Card, group int32) {
	this.exileLast = card
	this.exileCount++
	this.exileTotal.Add(group)
}

// EventMorph 設置變身事件（成組：最後變身 + 前後卡牌編號 + 回合次數）。
func (this *Game) EventMorph(card *Card, oldID, newID int32) {
	this.morphLast = card
	this.morphOldID = oldID
	this.morphNewID = newID
	this.morphCount++
}

// RoundReset 回合開始歸零全部回合計數（入座 / 離場 / 行動 / 抽牌 / 棄牌 / 出牌 / 流放 / 變身）；
// Last 引用與整場累積（Total）不在此列。呼叫點於回合開始（M14）。
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
