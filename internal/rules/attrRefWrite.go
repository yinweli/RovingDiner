package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// attrRefWrite 引用屬性寫入詞彙表(名稱 → 寫入行為); 服務屬性修改命令的引用左值(<引用>.<屬性>)。
// 鍵集僅【營業規格書 | 二十三、屬性清單】卡牌 / 顧客引用屬性子表存取欄為 寫 / 寫鎖 / 鎖 的可寫屬性; 唯讀屬性不在此(Validate 據此擋)。
// 型別不符的引用(以顧客引用寫卡牌屬性等)回 changed=false。寫側不需 Lock 後綴路由(鎖定變更由 @ # 表達)。
// 每一詞條對應一個獨立的 writeRef* 函式(便於逐條單元測試); 本表僅作名稱 → 行為的索引。
var attrRefWrite = map[string]cores.AttrRefWriteFunc{
	// 卡牌引用屬性(cost / extraRun* 寫鎖; cardSeal / keep / playExile / unplayExile 純鎖;
	// 範圍依【營業規格書 | 二十三、屬性清單】於命令寫入路徑夾值, 上限縮減的連動壓回在執行結算)
	"cost":        writeRefCost,
	"extraRunMin": writeRefExtraRunMin,
	"extraRunMax": writeRefExtraRunMax,
	"cardSeal":    writeRefCardSeal,
	"keep":        writeRefKeep,
	"playExile":   writeRefPlayExile,
	"unplayExile": writeRefUnplayExile,

	// 顧客引用屬性(calm / sate / sateMax / morale / moraleMax / score / scoreMax 寫鎖; sateSeal / calmSeal 純鎖)
	"calm":      writeRefCalm,
	"sate":      writeRefSate,
	"sateMax":   writeRefSateMax,
	"morale":    writeRefMorale,
	"moraleMax": writeRefMoraleMax,
	"score":     writeRefScore,
	"scoreMax":  writeRefScoreMax,
	"sateSeal":  writeRefSateSeal,
	"calmSeal":  writeRefCalmSeal,
}

// HasAttrRefWrite 回報引用屬性寫入詞彙表是否登錄 name; 供 games.Validate 檢查屬性修改命令的引用左值屬性可寫性。
func HasAttrRefWrite(name string) bool {
	_, ok := attrRefWrite[name]
	return ok
}

// === 卡牌引用屬性 ===

// writeRefCost 寫卡牌出牌費用(寫鎖; 下限夾 0, 與 cardCost* 命令同則)。
func writeRefCost(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	card, ok := cores.AsCard(ref)

	if ok == false {
		return false
	} // if

	changed := card.GetCost().Apply(op, n)
	card.GetCost().Clamp(0)
	return changed
}

// writeRefExtraRunMin 寫卡牌額外發動次數下限(寫鎖; 下限夾 0)。
func writeRefExtraRunMin(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	card, ok := cores.AsCard(ref)

	if ok == false {
		return false
	} // if

	changed := card.GetExtraRunMin().Apply(op, n)
	card.GetExtraRunMin().Clamp(0)
	return changed
}

// writeRefExtraRunMax 寫卡牌額外發動次數上限(寫鎖; 下限夾 0)。
func writeRefExtraRunMax(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	card, ok := cores.AsCard(ref)

	if ok == false {
		return false
	} // if

	changed := card.GetExtraRunMax().Apply(op, n)
	card.GetExtraRunMax().Clamp(0)
	return changed
}

// writeRefCardSeal 寫卡牌封印(純鎖, 僅 @ #)。
func writeRefCardSeal(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	card, ok := cores.AsCard(ref)

	if ok == false {
		return false
	} // if

	return card.GetSeal().ApplyLockOnly(op)
}

// writeRefKeep 寫卡牌不棄(純鎖, 僅 @ #)。
func writeRefKeep(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	card, ok := cores.AsCard(ref)

	if ok == false {
		return false
	} // if

	return card.GetKeep().ApplyLockOnly(op)
}

// writeRefPlayExile 寫卡牌出牌後流放(純鎖, 僅 @ #)。
func writeRefPlayExile(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	card, ok := cores.AsCard(ref)

	if ok == false {
		return false
	} // if

	return card.GetPlayExile().ApplyLockOnly(op)
}

// writeRefUnplayExile 寫卡牌未出牌流放(純鎖, 僅 @ #)。
func writeRefUnplayExile(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	card, ok := cores.AsCard(ref)

	if ok == false {
		return false
	} // if

	return card.GetUnplayExile().ApplyLockOnly(op)
}

// === 顧客引用屬性 ===

// writeRefCalm 寫顧客耐心值(寫鎖; 下限夾 0)。
func writeRefCalm(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	guest, ok := cores.AsGuest(ref)

	if ok == false {
		return false
	} // if

	changed := guest.GetCalm().Apply(op, n)
	guest.GetCalm().Clamp(0)
	return changed
}

// writeRefSate 寫顧客飽食值(寫鎖; 範圍 0 ~ sateMax)。
func writeRefSate(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	guest, ok := cores.AsGuest(ref)

	if ok == false {
		return false
	} // if

	changed := guest.GetSate().Apply(op, n)
	guest.GetSate().Clamp(0)
	guest.GetSate().ClampMax(guest.GetSateMax().GetValue())
	return changed
}

// writeRefSateMax 寫顧客飽食值離場線(寫鎖; 下限夾 0——縮減使 sate 達線時由飽食離場收場, 不壓回)。
func writeRefSateMax(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	guest, ok := cores.AsGuest(ref)

	if ok == false {
		return false
	} // if

	changed := guest.GetSateMax().Apply(op, n)
	guest.GetSateMax().Clamp(0)
	return changed
}

// writeRefMorale 寫顧客士氣值(寫鎖; 範圍 0 ~ 該顧客 moraleMax); 引用屬性的 -= 走一般運算, 不啟動餐廳 morale 特例。
func writeRefMorale(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	guest, ok := cores.AsGuest(ref)

	if ok == false {
		return false
	} // if

	changed := guest.GetMorale().Apply(op, n)
	guest.GetMorale().Clamp(0)
	guest.GetMorale().ClampMax(guest.GetMoraleMax().GetValue())
	return changed
}

// writeRefMoraleMax 寫顧客士氣值上限(寫鎖; 下限夾 0; 縮減使該顧客 morale 超標時由執行結算壓回)。
func writeRefMoraleMax(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	guest, ok := cores.AsGuest(ref)

	if ok == false {
		return false
	} // if

	changed := guest.GetMoraleMax().Apply(op, n)
	guest.GetMoraleMax().Clamp(0)
	return changed
}

// writeRefScore 寫顧客滿意值(寫鎖; 下限夾 0)。
func writeRefScore(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	guest, ok := cores.AsGuest(ref)

	if ok == false {
		return false
	} // if

	changed := guest.GetScore().Apply(op, n)
	guest.GetScore().Clamp(0)
	return changed
}

// writeRefScoreMax 寫顧客滿意值上限(寫鎖; 下限夾 0)。
func writeRefScoreMax(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	guest, ok := cores.AsGuest(ref)

	if ok == false {
		return false
	} // if

	changed := guest.GetScoreMax().Apply(op, n)
	guest.GetScoreMax().Clamp(0)
	return changed
}

// writeRefSateSeal 寫顧客封印飽食技能(純鎖, 僅 @ #)。
func writeRefSateSeal(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	guest, ok := cores.AsGuest(ref)

	if ok == false {
		return false
	} // if

	return guest.GetSateSeal().ApplyLockOnly(op)
}

// writeRefCalmSeal 寫顧客封印耐心技能(純鎖, 僅 @ #)。
func writeRefCalmSeal(game *cores.Game, ref exprs.Ref, op cores.AssignKind, n float64) bool {
	guest, ok := cores.AsGuest(ref)

	if ok == false {
		return false
	} // if

	return guest.GetCalmSeal().ApplyLockOnly(op)
}
