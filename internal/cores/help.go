package cores

import (
	"math"

	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// 全域屬性查詢輔助:解析查詢函式參數、分組計數、比較運算。供 attrRead.go 的查詢函式型屬性共用。

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

// 屬性寫入輔助:依賦值符與存取等級(寫 / 寫鎖 / 鎖)變更屬性。供 attrWrite.go / attrRefWrite.go 的寫側詞條共用。
// 對應【營業規格書 | 十七、命令 | 1】賦值符語意與【二十三、屬性清單】存取欄。

// applyOp 依帶值賦值符對舊值套用算術,回未捨入 / 未夾的新值;除 0 / 取餘 0 → ok=false(對齊 exprs 算術失敗);
// @ # 非帶值賦值落入 default → ok=false。中間值為 float64、過程不捨入(捨入由呼叫方以 exprs.Round 處理)。
func applyOp(op AssignKind, old, n float64) (result float64, ok bool) {
	switch op {
	case AssignSet:
		return n, true

	case AssignAdd:
		return old + n, true

	case AssignSub:
		return old - n, true

	case AssignMul:
		return old * n, true

	case AssignDiv:
		if n == 0 {
			return 0, false
		} // if

		return old / n, true

	case AssignMod:
		if n == 0 {
			return 0, false
		} // if

		return math.Mod(old, n), true

	default:
		return 0, false
	} // switch
}

// applyLock 套用鎖定 / 解鎖(@ 計數 +1、# 計數 -1 且夾 >= 0,對 0 計數的 # 為 no-op);
// 非鎖定符回 handled=false,交呼叫方處理帶值賦值。
func applyLock(target *Value, op AssignKind) (changed, handled bool) {
	switch op {
	case AssignLock:
		target.Lock++
		return true, true

	case AssignUnlock:
		if target.Lock > 0 {
			target.Lock--
			return true, true
		} // if

		return false, true

	default:
		return false, false
	} // switch
}

// writeValue 寫鎖屬性的寫入:先試鎖定 / 解鎖;帶值賦值於鎖定計數 > 0 時 no-op,否則套算術 → 捨入 → 夾(clamp 為 nil 不夾)→ 寫回。
func writeValue(target *Value, op AssignKind, n float64, clamp func(value int32) int32) (changed bool) {
	if locked, handled := applyLock(target, op); handled {
		return locked
	} // if

	if target.Locked() {
		return false // 鎖定計數 > 0 → 帶值賦值 no-op
	} // if

	result, ok := applyOp(op, float64(target.Value), n)

	if ok == false {
		return false
	} // if

	value := exprs.Round(result)

	if clamp != nil {
		value = clamp(value)
	} // if

	target.Value = value
	return true
}

// writeLockOnly 鎖屬性(純計數封印,Value 固定 0)的寫入:僅鎖定 / 解鎖;帶值賦值不適用故 no-op(可寫性由 Validate 先擋)。
func writeLockOnly(target *Value, op AssignKind) (changed bool) {
	locked, _ := applyLock(target, op)
	return locked
}

// writeInt 寫屬性(無鎖定計數的整數欄位,如 round / roundMax)的寫入:僅帶值賦值;@ # 經 applyOp default → no-op。
func writeInt(target *int32, op AssignKind, n float64) (changed bool) {
	result, ok := applyOp(op, float64(*target), n)

	if ok == false {
		return false
	} // if

	*target = exprs.Round(result)
	return true
}

// clampLow0 夾下限 0;供 護盾 / 格擋 等規格明寫「下限夾 0」的屬性使用(其餘屬性不臆測 clamp)。
func clampLow0(value int32) int32 {
	if value < 0 {
		return 0
	} // if

	return value
}
