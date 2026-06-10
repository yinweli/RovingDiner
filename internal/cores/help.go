package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// Immune 免疫群組計數（群組編號 → 鎖定計數）;效果免疫 / 技能免疫共用
// （【營業規格書 | 五、實例結構 | 顧客（Guest）實例】）。零值可用（Add 自建表）。
type Immune struct {
	count map[int32]int32 // 群組編號 -> 鎖定計數
}

// NewImmune 建構空免疫計數。
func NewImmune() Immune {
	return Immune{count: map[int32]int32{}}
}

// Add 對群組鎖定計數 +1;表未建時自建。
func (this *Immune) Add(group int32) {
	if this.count == nil {
		this.count = map[int32]int32{}
	} // if

	this.count[group]++
}

// Del 對群組鎖定計數 -1,夾 ≥ 0（無鍵 / 空表自然視為 0）。
func (this *Immune) Del(group int32) {
	if this.count[group] > 0 {
		this.count[group]--
	} // if
}

// Get 讀群組鎖定計數;無鍵回 0（>0 / ==0 判斷由呼叫端比較）。
func (this *Immune) Get(group int32) int32 {
	return this.count[group]
}

// Hit 已觸發門檻集合（門檻值 → 已觸發）;飽食 / 耐心門檻共用,擋同門檻重複觸發
// （【營業規格書 | 二十、獨立流程 | 執行結算】）。零值可用（Add 自建表）。
type Hit struct {
	hit map[int32]bool // 門檻值 -> 已觸發
}

// NewHit 建構空門檻集合。
func NewHit() Hit {
	return Hit{hit: map[int32]bool{}}
}

// Add 標記門檻已觸發（M16 執行結算用）;表未建時自建。
func (this *Hit) Add(threshold int32) {
	if this.hit == nil {
		this.hit = map[int32]bool{}
	} // if

	this.hit[threshold] = true
}

// IsHit 回報門檻是否已觸發。
func (this *Hit) IsHit(threshold int32) bool {
	return this.hit[threshold]
}

// Count 讀已觸發門檻數量（sateHit / calmHit 屬性）。
func (this *Hit) Count() int32 {
	return int32(len(this.hit))
}

// IDList 編號列表（有序多重集合,允許重複）;卡牌實例效果列表用,零值可用。
type IDList struct {
	id []int32 // 編號列表（保持加入順序）
}

// NewIDList 以編號集建構列表（複製輸入,不共享底層）。
func NewIDList(id ...int32) IDList {
	return IDList{id: append([]int32(nil), id...)}
}

// Add 尾端加入（可變參數,批次附加共用）。
func (this *IDList) Add(id ...int32) {
	this.id = append(this.id, id...)
}

// DelOne 移除第一個 == id 者;無命中原樣不動。
func (this *IDList) DelOne(id int32) {
	for itor, value := range this.id {
		if value == id {
			this.id = append(this.id[:itor], this.id[itor+1:]...)
			return
		} // if
	} // for
}

// DelAll 移除全部 == id 者。
func (this *IDList) DelAll(id int32) {
	result := []int32{}

	for _, itor := range this.id {
		if itor != id {
			result = append(result, itor)
		} // if
	} // for

	this.id = result
}

// Count 計數 == id 的個數。
func (this *IDList) Count(id int32) int32 {
	result := int32(0)

	for _, itor := range this.id {
		if itor == id {
			result++
		} // if
	} // for

	return result
}

// List 取底層編號列表供迭代（呼叫端唯讀約定）。
func (this *IDList) List() []int32 {
	return this.id
}

// Tally 分組累積計數（群組編號 → 數量）;抽牌 / 棄牌 / 出牌 / 流放的整場累積多重集合共用
// （【營業規格書 | 五、實例結構 | 營業（Game）實例】）。零值可用（Add 自建表）。
type Tally struct {
	count map[int32]int32 // 群組編號 -> 數量
}

// NewTally 建構空累積計數。
func NewTally() Tally {
	return Tally{count: map[int32]int32{}}
}

// Add 對群組數量 +1;表未建時自建。
func (this *Tally) Add(group int32) {
	if this.count == nil {
		this.count = map[int32]int32{}
	} // if

	this.count[group]++
}

// Get 讀群組數量;無鍵回 0。
func (this *Tally) Get(group int32) int32 {
	return this.count[group]
}

// Sum 讀全群組數量加總（查詢函式 N == 0 全加總用）。
func (this *Tally) Sum() (sum int32) {
	for _, v := range this.count {
		sum += v
	} // for

	return sum
}

// 參數取值:自 []exprs.Value 取出型別化參數(嚴格數量 / 首參寬鬆 / 集合)。供查詢函式、命令對象、操作命令位置參數共用。

// oneInt 取單一整數參數;數量 / 型別不符回 ok=false。供單參數查詢函式(handSize / tableGuest / effectImmune…)共用。
func oneInt(arg []exprs.Value) (num int32, ok bool) {
	if len(arg) != 1 {
		return 0, false
	} // if

	if arg[0].IsNum() == false {
		return 0, false
	} // if

	return int32(arg[0].Num()), true
}

// twoInt 取兩個整數參數;數量 / 型別不符回 ok=false。供雙參數命令對象(handPick / deckRand…的 N、卡牌編號)共用參數校驗。
func twoInt(arg []exprs.Value) (a, b int32, ok bool) {
	if len(arg) != 2 {
		return 0, 0, false
	} // if

	if arg[0].IsNum() == false || arg[1].IsNum() == false {
		return 0, 0, false
	} // if

	return int32(arg[0].Num()), int32(arg[1].Num()), true
}

// argInt 取首個參數為整數;缺漏 / 非數值回 ok=false。需讀第 k 個參數時呼叫端傳 arg[k:]。供操作命令位置參數取值共用。
func argInt(arg []exprs.Value) (n int32, ok bool) {
	if len(arg) == 0 || arg[0].IsNum() == false {
		return 0, false
	} // if

	return int32(arg[0].Num()), true
}

// argNum 取首個參數為浮點(供倍率等可帶小數者);缺漏 / 非數值回 ok=false。需讀第 k 個參數時呼叫端傳 arg[k:]。
func argNum(arg []exprs.Value) (n float64, ok bool) {
	if len(arg) == 0 || arg[0].IsNum() == false {
		return 0, false
	} // if

	return arg[0].Num(), true
}

// argBool 取首個參數為布林;缺漏 / 非布林回 false。需讀第 k 個參數時呼叫端傳 arg[k:]。供帶「洗牌」等布林旗標的操作命令共用。
func argBool(arg []exprs.Value) bool {
	return len(arg) > 0 && arg[0].IsBool() && arg[0].Bool()
}

// intList 把運算式值列表轉為整數列表(略過非數值);供 varargs 整數參數(附加效果編號)取值。
func intList(arg []exprs.Value) (result []int32) {
	for _, itor := range arg {
		if itor.IsNum() {
			result = append(result, int32(itor.Num()))
		} // if
	} // for

	return result
}

// argTail 安全取參數列表自 from 起的尾切片;from 越界回 nil(避免 arg[from:] 越界 panic)。供讀「第 k 個之後」的位置 / varargs 參數。
func argTail(arg []exprs.Value, from int) []exprs.Value {
	if from >= len(arg) {
		return nil
	} // if

	return arg[from:]
}

// 分組計數與比較:卡牌分組計數、累積多重集合查詢、比較運算符。供 attrRead.go 查詢函式型屬性共用。

// groupSize 求容器中卡牌群組編號 == N 的張數(N == 0 回容器全量);包裝 oneInt + countByGroup。
func groupSize(card []*Card, data *sheeter.Sheeter, arg []exprs.Value) (result exprs.Value, ok bool) {
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	return exprs.NewNum(float64(countByGroup(card, n, data))), true
}

// groupTotal 求分組累積多重集合中 N 的數量(N == 0 回全加總);包裝 oneInt + Tally 查詢。
func groupTotal(total *Tally, arg []exprs.Value) (result exprs.Value, ok bool) {
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	if n == 0 {
		return exprs.NewNum(float64(total.Sum())), true
	} // if

	return exprs.NewNum(float64(total.Get(n))), true
}

// countByGroup 計卡牌容器中卡牌群組編號 == group 的張數;group == 0 回容器全量(不過濾)。
func countByGroup(card []*Card, group int32, data *sheeter.Sheeter) (count int32) {
	if group == 0 {
		return int32(len(card))
	} // if

	for _, itor := range card {
		meta := data.Card.Get(itor.GetCardID())

		if meta != nil && meta.Group == group {
			count++
		} // if
	} // for

	return count
}

// compareOp 以字串運算符比較 a 與 b(對齊【二十七、運算式 | 2】比較運算符);未知運算符回 ok=false。
func compareOp(op string, a, b int32) (result, ok bool) {
	switch op {
	case "<":
		return a < b, true

	case ">":
		return a > b, true

	case "<=":
		return a <= b, true

	case ">=":
		return a >= b, true

	case "==":
		return a == b, true

	case "!=":
		return a != b, true
	} // switch

	return false, false
}

// 引用鎖定計數取值:自卡牌 / 顧客引用取出屬性的鎖定計數(透過 ref.go 的 asCard / asGuest 拆解)。供 attrRefLockRead 的鎖定讀取詞條共用。

// cardLock 取卡牌引用某屬性的鎖定計數;非卡牌引用回 ok=false。pick 自卡牌實例取出對應 Value.Lock。
func cardLock(ref exprs.Ref, pick func(card *Card) int32) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if

	return exprs.NewNum(float64(pick(card))), true
}

// guestLock 取顧客引用某屬性的鎖定計數;非顧客引用回 ok=false。
func guestLock(ref exprs.Ref, pick func(guest *Guest) int32) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if

	return exprs.NewNum(float64(pick(guest))), true
}

// 容器掃描與歸屬判定:卡牌定位、座位占用計數。供定位 / 座位查詢型屬性共用。

// inContainer 回報卡牌實例是否位於指定容器(以實例編號比對)。供 inHand / inDeck / inDrop / inExile 容器掃描。
func inContainer(card *Card, container []*Card) bool {
	for _, itor := range container {
		if itor.GetInstanceID() == card.GetInstanceID() {
			return true
		} // if
	} // for

	return false
}

// occupiedAmong 計座位編號列表中當下有顧客占用的座位數。供 sameSize / nearSize 使用。
func occupiedAmong(eng *Engine, seatID []int32) (count int32) {
	for _, itor := range seatID {
		if eng.runtime.Seat[itor] != nil {
			count++
		} // if
	} // for

	return count
}

// 屬性寫入:依賦值符與存取等級(寫 / 寫鎖 / 鎖)變更屬性。供 attrWrite.go / attrRefWrite.go 的寫側詞條共用。
// 對應【營業規格書 | 十七、命令 | 1】賦值符語意與【二十三、屬性清單】存取欄。

// writeLockOnly 鎖屬性(純計數封印,數值固定 0)的寫入:僅鎖定 / 解鎖;帶值賦值不適用故 no-op(可寫性由 Validate 先擋)。
func writeLockOnly(target *Value, op AssignKind) bool {
	switch op {
	case AssignLock:
		target.Lock()
		return true

	case AssignUnlock:
		return target.Unlock()

	default:
		return false // 帶值賦值不適用純鎖屬性 → no-op
	} // switch
}

// writeValueOnly 寫屬性(無鎖定語意,如 round / roundMax,鎖定計數恆 0)的寫入:僅帶值賦值;@ # 不適用故 no-op。
func writeValueOnly(target *Value, op AssignKind, n float64) bool {
	if op == AssignLock || op == AssignUnlock {
		return false // 鎖定 / 解鎖不適用寫屬性 → no-op
	} // if

	return target.Apply(op, n)
}

// 切片增刪查:實例切片的定位(以實例編號比對)。供引擎跨容器分派共用。

// findGuest 自顧客切片以實例編號找出顧客;未命中回 nil。供 locateGuest 掃描排隊 / 遊蕩 / 卡牌化列表。
func findGuest(guest []*Guest, id InstanceID) *Guest {
	for _, itor := range guest {
		if itor.GetInstanceID() == id {
			return itor
		} // if
	} // for

	return nil
}

// 技能靜態查詢:自卡牌經 SkillID 取技能靜態欄位(效果編號列表 / 技能群組編號)。供 newCard 載入實例效果列表、cardMorph 重設、與效果列表啟動的 skillImmune 排除共用。

// skillEffect 取技能的效果編號列表複本(新卡實例效果列表來源 = Card.SkillID → Skill.EffectID);技能不存在回 nil。複製以免共享靜態表切片。供 newCard 載入卡牌實例效果列表、cardMorph 變身後重設效果共用。
func skillEffect(eng *Engine, skillID int32) []int32 {
	skill := eng.data.Skill.Get(skillID)

	if skill == nil {
		return nil
	} // if

	return append([]int32(nil), skill.EffectID...)
}

// cardSkillGroup 取卡牌的技能群組編號（卡牌資料.SkillID → Skill.Group）;卡牌 / 技能資料不存在回 0。供 cardRun 啟動效果列表時 skillImmune 排除免疫顧客。
func cardSkillGroup(eng *Engine, cardID int32) int32 {
	card := eng.data.Card.Get(cardID)

	if card == nil {
		return 0
	} // if

	skill := eng.data.Skill.Get(card.SkillID)

	if skill == nil {
		return 0
	} // if

	return skill.Group
}
