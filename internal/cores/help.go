package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

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

// groupTotal 求分組累積多重集合中 N 的數量(N == 0 回全加總);包裝 oneInt + totalByGroup。
func groupTotal(total map[int32]int32, arg []exprs.Value) (result exprs.Value, ok bool) {
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	return exprs.NewNum(float64(totalByGroup(total, n))), true
}

// countByGroup 計卡牌容器中卡牌群組編號 == group 的張數;group == 0 回容器全量(不過濾)。
func countByGroup(card []*Card, group int32, data *sheeter.Sheeter) (count int32) {
	if group == 0 {
		return int32(len(card))
	} // if

	for _, itor := range card {
		meta := data.Card.Get(itor.CardID)

		if meta != nil && meta.Group == group {
			count++
		} // if
	} // for

	return count
}

// totalByGroup 取分組累積多重集合的數量;n == 0 回全加總(不過濾)。
func totalByGroup(total map[int32]int32, n int32) (sum int32) {
	if n == 0 {
		for _, v := range total {
			sum += v
		} // for

		return sum
	} // if

	return total[n]
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
	card, ok := asCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if

	return exprs.NewNum(float64(pick(card))), true
}

// guestLock 取顧客引用某屬性的鎖定計數;非顧客引用回 ok=false。
func guestLock(ref exprs.Ref, pick func(guest *Guest) int32) (result exprs.Value, ok bool) {
	guest, ok := asGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if

	return exprs.NewNum(float64(pick(guest))), true
}

// 容器掃描與歸屬判定:卡牌定位、座位占用計數、效果所屬比對。供定位 / 座位 / 效果查詢型屬性共用。

// inContainer 回報卡牌實例是否位於指定容器(以實例編號比對)。供 inHand / inDeck / inDrop / inExile 容器掃描。
func inContainer(card *Card, container []*Card) bool {
	for _, itor := range container {
		if itor.InstanceID == card.InstanceID {
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

// effectSelfIs 回報效果項目的 self 是否就是引用 ref;供 effectStack / effectGroup 比對所屬對象(卡牌或顧客)。
func effectSelfIs(effect *Effect, ref exprs.Ref) bool {
	switch value := ref.(type) {
	case cardRef:
		return effect.Self.Card != nil && effect.Self.Card.InstanceID == value.card.InstanceID
	case guestRef:
		return effect.Self.Guest != nil && effect.Self.Guest.InstanceID == value.guest.InstanceID
	} // switch

	return false
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

// lockDec 對鎖屬性的鎖定計數 - 1(夾 ≥ 0);供解鎖(guestReturn 入列自動鎖回退、restore 不棄回退)共用。
func lockDec(value *Value) {
	value.Unlock()
}

// 切片增刪查:實例切片的定位 / 前端加入 / 移除(以實例編號比對),回新切片不影響原序。供操作命令的容器搬移與實例化共用。

// findCard 自卡牌切片以實例編號找出卡牌;未命中回 nil。供 locateCard 逐牌堆掃描。
func findCard(card []*Card, id InstanceID) *Card {
	for _, itor := range card {
		if itor.InstanceID == id {
			return itor
		} // if
	} // for

	return nil
}

// findGuest 自顧客切片以實例編號找出顧客;未命中回 nil。供 locateGuest 掃描排隊 / 遊蕩 / 卡牌化列表。
func findGuest(guest []*Guest, id InstanceID) *Guest {
	for _, itor := range guest {
		if itor.InstanceID == id {
			return itor
		} // if
	} // for

	return nil
}

// prepend 把卡牌加入切片前端(頂端);對齊牌堆「新進入者置頂」(M8 約定:前端為頂)。
func prepend(card []*Card, add *Card) []*Card {
	return append([]*Card{add}, card...)
}

// removeFrom 自卡牌切片移除指定卡牌(以實例編號比對),回新切片;其餘元素保持原序。
func removeFrom(card []*Card, remove *Card) (result []*Card) {
	for _, itor := range card {
		if itor.InstanceID != remove.InstanceID {
			result = append(result, itor)
		} // if
	} // for

	return result
}

// removeGuest 自顧客切片移除指定顧客(以實例編號比對),回新切片;其餘元素保持原序。供遊蕩 / 排隊 / 卡牌化列表移除共用。
func removeGuest(guest []*Guest, remove *Guest) (result []*Guest) {
	for _, itor := range guest {
		if itor.InstanceID != remove.InstanceID {
			result = append(result, itor)
		} // if
	} // for

	return result
}

// removeEffect 自效果切片移除指定效果（以實例編號比對），回新切片;其餘元素保持原序。供效果佇列的觸發後移除 / 推進 / 清理共用。
func removeEffect(effect []*Effect, remove *Effect) (result []*Effect) {
	for _, itor := range effect {
		if itor.InstanceID != remove.InstanceID {
			result = append(result, itor)
		} // if
	} // for

	return result
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
