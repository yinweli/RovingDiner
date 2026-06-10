package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// attrRefReadFunc 引用屬性詞條的讀取行為:自 ref(卡牌 / 顧客)以 Game 為 context 取子屬性;arg 供引用查詢函式。
type attrRefReadFunc func(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool)

// attrRefRead 引用屬性值讀取詞彙表(名稱 → 讀取行為);服務 exprs.Resolver 的 AttrRef。
// 涵蓋【營業規格書 | 二十三、屬性清單】卡牌引用屬性 / 顧客引用屬性兩子表;型別不符的引用回 ok=false。
// effectStack / effectGroup 卡牌、顧客共用同一詞條(以 Ref.IsSame 比對所屬對象)。
// 鎖定計數(Lock 後綴)另置 attrRefLockRead,由 Game.AttrRef 剝後綴路由。
// 每一詞條對應一個獨立的 readRef* 函式(便於逐條單元測試);本表僅作名稱 → 行為的索引。
var attrRefRead = map[string]attrRefReadFunc{
	// 卡牌引用屬性
	"cardID":      readRefCardID,
	"cost":        readRefCost,
	"extraRunMin": readRefExtraRunMin,
	"extraRunMax": readRefExtraRunMax,
	"cardSeal":    readRefCardSeal,
	"keep":        readRefKeep,
	"playExile":   readRefPlayExile,
	"unplayExile": readRefUnplayExile,
	"cardify":     readRefCardify,
	"cardGroup":   readRefCardGroup,
	"cardEffect":  readRefCardEffect,
	"inHand":      readRefInHand,
	"inDeck":      readRefInDeck,
	"inDrop":      readRefInDrop,
	"inExile":     readRefInExile,

	// 顧客引用屬性
	"calm":         readRefCalm,
	"sate":         readRefSate,
	"sateMax":      readRefSateMax,
	"morale":       readRefMorale,
	"moraleMax":    readRefMoraleMax,
	"score":        readRefScore,
	"scoreMax":     readRefScoreMax,
	"sateSeal":     readRefSateSeal,
	"calmSeal":     readRefCalmSeal,
	"seatID":       readRefSeatID,
	"guestID":      readRefGuestID,
	"freeze":       readRefFreeze,
	"calmHit":      readRefCalmHit,
	"sateHit":      readRefSateHit,
	"effectImmune": readRefEffectImmune,
	"skillImmune":  readRefSkillImmune,
	"sameSize":     readRefSameSize,
	"nearSize":     readRefNearSize,

	// 卡牌 / 顧客共用查詢函式(以 self 比對所屬對象)
	"effectStack": readRefEffectStack,
	"effectGroup": readRefEffectGroup,
}

// attrRefLockRead 引用屬性鎖定計數讀取詞彙表(基底名 → 讀取行為);僅含【二十三】子表存取欄為「寫鎖 / 鎖」的屬性。
// Game.AttrRef 於值表未命中且名稱以 Lock 結尾時,剝後綴查本表。
var attrRefLockRead = map[string]attrRefReadFunc{
	// 卡牌引用:寫鎖 / 鎖
	"cost":        readRefCostLock,
	"extraRunMin": readRefExtraRunMinLock,
	"extraRunMax": readRefExtraRunMaxLock,
	"cardSeal":    readRefCardSealLock,
	"keep":        readRefKeepLock,
	"playExile":   readRefPlayExileLock,
	"unplayExile": readRefUnplayExileLock,

	// 顧客引用:寫鎖 / 鎖
	"calm":      readRefCalmLock,
	"sate":      readRefSateLock,
	"sateMax":   readRefSateMaxLock,
	"morale":    readRefMoraleLock,
	"moraleMax": readRefMoraleMaxLock,
	"score":     readRefScoreLock,
	"scoreMax":  readRefScoreMaxLock,
	"sateSeal":  readRefSateSealLock,
	"calmSeal":  readRefCalmSealLock,
}

// === 卡牌引用屬性 ===

// readRefCardID 讀卡牌的卡牌編號。
func readRefCardID(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(card.GetCardID())), true
}

// readRefCost 讀卡牌的費用值。
func readRefCost(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(card.GetCost().GetValue())), true
}

// readRefExtraRunMin 讀卡牌的額外執行次數下限值。
func readRefExtraRunMin(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(card.GetExtraRunMin().GetValue())), true
}

// readRefExtraRunMax 讀卡牌的額外執行次數上限值。
func readRefExtraRunMax(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(card.GetExtraRunMax().GetValue())), true
}

// readRefCardSeal 讀卡牌的封印值(恆 0,僅供鎖定承載)。
func readRefCardSeal(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(card.GetSeal().GetValue())), true
}

// readRefKeep 讀卡牌的保留值(恆 0,僅供鎖定承載)。
func readRefKeep(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(card.GetKeep().GetValue())), true
}

// readRefPlayExile 讀卡牌的出牌流放值(恆 0,僅供鎖定承載)。
func readRefPlayExile(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(card.GetPlayExile().GetValue())), true
}

// readRefUnplayExile 讀卡牌的未出牌流放值(恆 0,僅供鎖定承載)。
func readRefUnplayExile(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(card.GetUnplayExile().GetValue())), true
}

// readRefCardify 讀卡牌所卡牌化的顧客引用(無則 none)。
func readRefCardify(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return NewRefGuest(card.GetCardify()).Value(), true
}

// readRefCardGroup 讀卡牌的群組編號(自靜態表;資料不存在回失敗)。
func readRefCardGroup(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	meta := game.data.Card.Get(card.GetCardID())

	if meta == nil {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(meta.Group)), true
}

// readRefCardEffect 讀卡牌效果列表中效果編號 == N 的個數(N == 0 不命中)。
func readRefCardEffect(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	if n == 0 { // N == 0 不命中
		return exprs.NewNum(0), true
	} // if
	return exprs.NewNum(float64(card.GetEffectID().Count(n))), true
}

// readRefInHand 回報卡牌是否在手牌。
func readRefInHand(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewBool(game.Hand.Has(card.GetInstanceID())), true
}

// readRefInDeck 回報卡牌是否在抽牌牌堆。
func readRefInDeck(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewBool(game.Deck.Has(card.GetInstanceID())), true
}

// readRefInDrop 回報卡牌是否在棄牌牌堆。
func readRefInDrop(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewBool(game.Drop.Has(card.GetInstanceID())), true
}

// readRefInExile 回報卡牌是否在流放牌堆。
func readRefInExile(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	card, ok := AsCard(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewBool(game.Exile.Has(card.GetInstanceID())), true
}

// === 顧客引用屬性 ===

// readRefCalm 讀顧客的耐心值。
func readRefCalm(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetCalm().GetValue())), true
}

// readRefSate 讀顧客的飽食值。
func readRefSate(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetSate().GetValue())), true
}

// readRefSateMax 讀顧客的飽食上限。
func readRefSateMax(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetSateMax().GetValue())), true
}

// readRefMorale 讀顧客的士氣值。
func readRefMorale(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetMorale().GetValue())), true
}

// readRefMoraleMax 讀顧客的士氣上限。
func readRefMoraleMax(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetMoraleMax().GetValue())), true
}

// readRefScore 讀顧客的分數值。
func readRefScore(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetScore().GetValue())), true
}

// readRefScoreMax 讀顧客的分數上限。
func readRefScoreMax(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetScoreMax().GetValue())), true
}

// readRefSateSeal 讀顧客的封印飽食值(恆 0,僅供鎖定承載)。
func readRefSateSeal(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetSateSeal().GetValue())), true
}

// readRefCalmSeal 讀顧客的封印耐心值(恆 0,僅供鎖定承載)。
func readRefCalmSeal(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetCalmSeal().GetValue())), true
}

// readRefSeatID 讀顧客的座位編號。
func readRefSeatID(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetSeatID())), true
}

// readRefGuestID 讀顧客的顧客編號。
func readRefGuestID(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetGuestID())), true
}

// readRefFreeze 讀顧客的凍結回合數。
func readRefFreeze(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetFreeze())), true
}

// readRefCalmHit 讀顧客的耐心閾值命中數。
func readRefCalmHit(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetCalmHit().Count())), true
}

// readRefSateHit 讀顧客的飽食閾值命中數。
func readRefSateHit(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetSateHit().Count())), true
}

// readRefEffectImmune 讀顧客對效果 N 的免疫計數。
func readRefEffectImmune(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetEffectImmune().Get(n))), true
}

// readRefSkillImmune 讀顧客對技能 N 的免疫計數。
func readRefSkillImmune(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if
	return exprs.NewNum(float64(guest.GetSkillImmune().Get(n))), true
}

// readRefSameSize 讀顧客同桌占用座位數(非入座回 0;含自身語意依座位表)。
func readRefSameSize(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	meta := game.data.Seat.Get(guest.GetSeatID())

	if meta == nil { // 非入座(SeatID=0 / 座位不存在)→ 0
		return exprs.NewNum(0), true
	} // if
	return exprs.NewNum(float64(game.Seat.Occupied(meta.SameSeatID))), true
}

// readRefNearSize 讀顧客鄰桌占用座位數(非入座回 0)。
func readRefNearSize(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	guest, ok := AsGuest(ref)

	if ok == false {
		return exprs.Value{}, false
	} // if
	meta := game.data.Seat.Get(guest.GetSeatID())

	if meta == nil {
		return exprs.NewNum(0), true
	} // if
	return exprs.NewNum(float64(game.Seat.Occupied(meta.NearSeatID))), true
}

// === 卡牌 / 顧客共用查詢函式(以 self 比對所屬對象) ===

// readRefEffectStack 讀效果佇列中 self == ref 且效果編號 == N 的層數加總。
func readRefEffectStack(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if
	stack := int32(0)

	for _, effect := range game.Effect {
		if effect.GetEffectID() == n && effect.GetSelf().IsSame(ref) {
			stack += effect.GetStack()
		} // if
	} // for
	return exprs.NewNum(float64(stack)), true
}

// readRefEffectGroup 讀效果佇列中 self == ref 且效果群組 == N 的項目數(不加層;N == 0 不命中)。
func readRefEffectGroup(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	n, valid := oneInt(arg)

	if valid == false {
		return exprs.Value{}, false
	} // if

	if n == 0 { // N == 0 不命中
		return exprs.NewNum(0), true
	} // if
	count := int32(0)

	for _, effect := range game.Effect {
		if effect.GetSelf().IsSame(ref) == false {
			continue
		} // if
		meta, found := game.effectData[effect.GetEffectID()]

		if found && meta.Group == n {
			count++
		} // if
	} // for
	return exprs.NewNum(float64(count)), true
}

// === 卡牌引用屬性鎖定計數(基底名 → .GetLock()) ===

// readRefCostLock 讀卡牌費用的鎖定計數。
func readRefCostLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardLock(ref, func(c *Card) int32 { return c.GetCost().GetLock() })
}

// readRefExtraRunMinLock 讀卡牌額外執行次數下限的鎖定計數。
func readRefExtraRunMinLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardLock(ref, func(c *Card) int32 { return c.GetExtraRunMin().GetLock() })
}

// readRefExtraRunMaxLock 讀卡牌額外執行次數上限的鎖定計數。
func readRefExtraRunMaxLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardLock(ref, func(c *Card) int32 { return c.GetExtraRunMax().GetLock() })
}

// readRefCardSealLock 讀卡牌封印的鎖定計數。
func readRefCardSealLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardLock(ref, func(c *Card) int32 { return c.GetSeal().GetLock() })
}

// readRefKeepLock 讀卡牌保留的鎖定計數。
func readRefKeepLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardLock(ref, func(c *Card) int32 { return c.GetKeep().GetLock() })
}

// readRefPlayExileLock 讀卡牌出牌流放的鎖定計數。
func readRefPlayExileLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardLock(ref, func(c *Card) int32 { return c.GetPlayExile().GetLock() })
}

// readRefUnplayExileLock 讀卡牌未出牌流放的鎖定計數。
func readRefUnplayExileLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return cardLock(ref, func(c *Card) int32 { return c.GetUnplayExile().GetLock() })
}

// === 顧客引用屬性鎖定計數(基底名 → .GetLock()) ===

// readRefCalmLock 讀顧客耐心的鎖定計數。
func readRefCalmLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestLock(ref, func(g *Guest) int32 { return g.GetCalm().GetLock() })
}

// readRefSateLock 讀顧客飽食的鎖定計數。
func readRefSateLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestLock(ref, func(g *Guest) int32 { return g.GetSate().GetLock() })
}

// readRefSateMaxLock 讀顧客飽食上限的鎖定計數。
func readRefSateMaxLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestLock(ref, func(g *Guest) int32 { return g.GetSateMax().GetLock() })
}

// readRefMoraleLock 讀顧客士氣的鎖定計數。
func readRefMoraleLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestLock(ref, func(g *Guest) int32 { return g.GetMorale().GetLock() })
}

// readRefMoraleMaxLock 讀顧客士氣上限的鎖定計數。
func readRefMoraleMaxLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestLock(ref, func(g *Guest) int32 { return g.GetMoraleMax().GetLock() })
}

// readRefScoreLock 讀顧客分數的鎖定計數。
func readRefScoreLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestLock(ref, func(g *Guest) int32 { return g.GetScore().GetLock() })
}

// readRefScoreMaxLock 讀顧客分數上限的鎖定計數。
func readRefScoreMaxLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestLock(ref, func(g *Guest) int32 { return g.GetScoreMax().GetLock() })
}

// readRefSateSealLock 讀顧客封印飽食的鎖定計數。
func readRefSateSealLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestLock(ref, func(g *Guest) int32 { return g.GetSateSeal().GetLock() })
}

// readRefCalmSealLock 讀顧客封印耐心的鎖定計數。
func readRefCalmSealLock(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
	return guestLock(ref, func(g *Guest) int32 { return g.GetCalmSeal().GetLock() })
}
