package rules

import (
	"slices"

	"github.com/yinweli/RovingDiner/internal/cores"
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

// numArg 取 >= 2 個數值參數為 float64 切片;少於 2 個或任一非數值即失敗。供多參數內建函式(min / max)共用參數校驗。
func numArg(arg []exprs.Value) (num []float64, ok bool) {
	if len(arg) < 2 {
		return nil, false
	} // if

	num = make([]float64, 0, len(arg))

	for _, itor := range arg {
		if itor.IsNum() == false {
			return nil, false
		} // if

		num = append(num, itor.Num())
	} // for

	return num, true
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

// groupSize 求牌堆中卡牌群組編號 == N 的張數(N == 0 回牌堆全量);包裝 oneInt + cores.CardList.CountGroup。
func groupSize(card cores.CardList, data *sheeter.Sheeter, arg []exprs.Value) (result exprs.Value, ok bool) {
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	return exprs.NewNum(float64(card.CountGroup(n, data))), true
}

// groupTotal 求分組累積多重集合中 N 的數量(N == 0 回全加總);包裝 oneInt + cores.Tally 查詢。
func groupTotal(total *cores.Tally, arg []exprs.Value) (result exprs.Value, ok bool) {
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	if n == 0 {
		return exprs.NewNum(float64(total.Sum())), true
	} // if

	return exprs.NewNum(float64(total.Get(n))), true
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

// 引用鎖定計數取值:自卡牌 / 顧客引用取出屬性的鎖定計數(透過 ref.go 的 asCard / asGuest 拆解)。供 attrRefRead 的鎖定讀取詞條共用。

// cardLock 取卡牌引用某屬性的鎖定計數;非卡牌引用回 ok=false。pick 自卡牌實例取出對應 Value.Lock。
func cardLock(ref exprs.Ref, pick func(card *cores.Card) int32) (result exprs.Value, ok bool) {
	card, ok := cores.AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if

	return exprs.NewNum(float64(pick(card))), true
}

// guestLock 取顧客引用某屬性的鎖定計數;非顧客引用回 ok=false。
func guestLock(ref exprs.Ref, pick func(guest *cores.Guest) int32) (result exprs.Value, ok bool) {
	guest, ok := cores.AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if

	return exprs.NewNum(float64(pick(guest))), true
}

// 卡牌容器原語:四牌堆的群組查詢 / 移除 / 加入 / 洗牌底層操作。供搬移(commandMove)、處理流程(commandFlow)、實例化(commandInstance)、deckShuffle 與 deckTop auto-shuffle(command / selector)共用。

// cardGroup 取卡牌的群組編號（供 *Total 多重集合鍵）；靜態資料缺失回 0。
func cardGroup(game *cores.Game, card *cores.Card) int32 {
	meta := game.GetSheet().Card.Get(card.GetCardID())

	if meta == nil {
		return 0
	} // if

	return meta.Group
}

// removeCard 自 source 牌堆移除指定卡牌（以實例編號比對）。
func removeCard(game *cores.Game, source cores.ContainerKind, card *cores.Card) {
	switch source {
	case cores.ContainerHand:
		game.Hand.Remove(card.GetInstanceID())

	case cores.ContainerDeck:
		game.Deck.Remove(card.GetInstanceID())

	case cores.ContainerDrop:
		game.Drop.Remove(card.GetInstanceID())

	case cores.ContainerExile:
		game.Exile.Remove(card.GetInstanceID())

	default:
		// 不可達：removeCard 僅以四牌堆 source 呼叫
	} // switch
}

// placeCard 把卡牌加入 dest 牌堆頂端（前端），並依目的設事件與 system 觸發（進抽牌牌堆無事件屬性與觸發）。
// 卡牌容器事件的單點收口（M18 拍板）:from 由呼叫端供給（剛移出的來源牌堆;新建直入傳 ContainerNone）,
// 入容器後、事件屬性與觸發前發射（消費端見事件時盤面已就位）。
func placeCard(game *cores.Game, from, dest cores.ContainerKind, card *cores.Card) {
	switch dest {
	case cores.ContainerHand:
		game.Hand.Push(card)

	case cores.ContainerDeck:
		game.Deck.Push(card)

	case cores.ContainerDrop:
		game.Drop.Push(card)

	case cores.ContainerExile:
		game.Exile.Push(card)

	default:
		return // 不可達：placeCard 僅以四牌堆 dest 呼叫
	} // switch

	game.Emit(cores.EventData{Kind: cores.EventContainer, DataID: card.GetCardID(), InstanceID: card.GetInstanceID(), From: from, To: dest})

	switch dest {
	case cores.ContainerHand:
		game.EventDraw(card, cardGroup(game, card))
		fireTrigger(game, cores.TriggerCardDraw) // 卡牌進手牌觸發

	case cores.ContainerDrop:
		game.EventDrop(card, cardGroup(game, card))
		fireTrigger(game, cores.TriggerCardDrop) // 卡牌進棄牌牌堆觸發

	case cores.ContainerExile:
		game.EventExile(card, cardGroup(game, card))
		fireTrigger(game, cores.TriggerCardExile) // 卡牌進流放牌堆觸發

	default:
		// 進抽牌牌堆:無事件屬性與觸發
	} // switch
}

// shuffleCard 經 Rander 就地洗牌卡牌切片。
func shuffleCard(game *cores.Game, card []*cores.Card) {
	game.GetRander().Shuffle(len(card), func(i, j int) { card[i], card[j] = card[j], card[i] })
}

// 餐廳士氣值 -= 特例:格擋 → 護盾 → morale 的扣減消耗鏈。供屬性修改命令(attrWrite 的 morale -=)與生氣離場(commandFlow 的 guestExit)共用。

// moraleDamage 餐廳士氣值 -= 特例(【十七、命令 | 1】特例):依 格擋 → 護盾 → morale 順序消耗扣減值 N;
// 實際扣減 > 0 時設置 damageValue / damageGuest(來源 source 由呼叫端決定:命令路徑取 self 顧客、guestExit 取離場顧客),並標記士氣受損時機。
// morale 鎖定時格擋 / 護盾仍消耗、morale 不動、無實際扣減(對齊目前解讀)。
// N 先四捨五入為整數扣減值,使格擋 / 護盾 / morale 的整數消耗自洽(小數扣減值的捨入時點待規格確認)。
// 三段消耗皆經守衛寫入、各自鎖定時不消耗:格擋鎖定仍無視本次 N 但不減層、護盾鎖定則殘餘全進 morale、morale 鎖定不扣。
func moraleDamage(game *cores.Game, n float64, source *cores.Guest) bool {
	damage := exprs.Round(n)
	changed := false

	if game.GetMoraleBlock().GetValue() > 0 { // 格擋優先:格擋 -= 1(鎖定 → 不減層),本次無視 N
		return game.GetMoraleBlock().Sub(1)
	} // if

	if damage > 0 { // 護盾消耗:d = min(N, 護盾) → 護盾 -= d → N -= d(鎖定 → 不消耗、殘餘進 morale)
		d := min(damage, game.GetMoraleShield().GetValue())

		if d > 0 && game.GetMoraleShield().Sub(float64(d)) {
			damage -= d
			changed = true
		} // if
	} // if

	if damage > 0 && game.GetMorale().IsLock() == false { // 殘餘對 morale 一般 -= 運算(鎖定 → 不扣)
		before := game.GetMorale().GetValue()
		game.GetMorale().Sub(float64(damage))
		actual := before - game.GetMorale().GetValue()

		if actual > 0 {
			game.EventDamage(actual, source)
			changed = true
			fireTrigger(game, cores.TriggerDamage) // 士氣受損時機
		} // if
	} // if

	return changed
}

// damageSource 取士氣受損來源顧客:self 為顧客時回該顧客,否則空物件(nil);供 Phase 4 命令路徑的 morale -= 使用。
func damageSource(self *cores.Ref) *cores.Guest {
	if self != nil {
		return self.GetGuest()
	} // if

	return nil
}

// 效果佇列共用:凍結判定、退場核心、身分取值。供觸發(effectTrigger)、推進(effectAdvance)、清理(effectCleanup)與效果佇列命令(commandEffect)共用。

// frozenSelf 回報效果的 self 是否為凍結中顧客(效果凍結判定;【營業規格書 | 二十一、流程補充 | 凍結語意】)。
// 觸發(fireTrigger)/ 推進(advanceEffect)/ 清除(effectClear)三處篩選共用;清理以失效對象判凍結,另於 cleanupEffect 把關。
func frozenSelf(game *cores.Game, effect *cores.Effect) bool {
	guest := effect.GetSelf().GetGuest()
	return guest != nil && game.IsFrozen(guest)
}

// retireEffect 效果退場核心:綁該效果 self → 結束命令 × 當前堆疊層數 → 出佇列(【營業規格書 | 十六、堆疊規則】離開佇列語意);
// 供推進 / 清理 / effectClear 的逐效果退場共用,呼叫端須先確認編譯資料存在。
func retireEffect(game *cores.Game, effect *cores.Effect) {
	meta, _ := game.EffectData(effect.GetEffectID()) // 呼叫端已確認存在

	runEffectEnd(game, effect, meta.End, effect.GetStack())
	game.Effect.Remove(effect.GetInstanceID())
}

// runEffectEnd 以效果自身 self 綁定執行結束命令 times 次(綁定逐層 save / restore;退場與 effectDel 退層共用)。
// 效果事件:結束 階段於此發(推進 / 清理 / effectClear 經 retireEffect、effectDel 退層皆收口於此;
// 觸發後移除路徑因 self 已綁定、由 fireOne 自發)。
func runEffectEnd(game *cores.Game, effect *cores.Effect, end cores.EffectExec, times int32) {
	self := effect.GetSelf()
	restore := game.SetSelf(&self)

	defer restore()

	emitEffect(game, effect.GetEffectID(), effect.GetInstanceID(), self, cores.EffectStageEnd)
	runEffectExec(game, end, times)
}

// instanceRef 以實例編號自卡牌四牌堆 / 顧客四容器找出實例引用;未命中回 ok=false。
// 供以「命令對象元素為 self」的效果佇列命令(effectDel / effectRun)取 self 引用。
func instanceRef(game *cores.Game, id cores.InstanceID) (ref cores.Ref, ok bool) {
	if card, _, found := game.LocateCard(id); found {
		return cores.NewRefCard(card), true
	} // if

	if guest, _, found := game.LocateGuest(id); found {
		return cores.NewRefGuest(guest), true
	} // if

	return cores.Ref{}, false
}

// selfIn 回報效果的 self 是否屬於命令對象身分集(空物件 self 不屬於任何身分集);供 effectClear 非全域掃描篩選。
func selfIn(effect *cores.Effect, target []cores.InstanceID) bool {
	id := cores.NoneID

	if card := effect.GetSelf().GetCard(); card != nil {
		id = card.GetInstanceID()
	} // if

	if guest := effect.GetSelf().GetGuest(); guest != nil {
		id = guest.GetInstanceID()
	} // if

	if id == cores.NoneID {
		return false
	} // if

	return slices.Contains(target, id)
}
