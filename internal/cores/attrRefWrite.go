package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// attrRefWriteFunc 引用屬性詞條的寫入行為:自 ref(卡牌 / 顧客)以 engine 為 context 寫子屬性;op 為賦值符、n 為已求值的右值。
type attrRefWriteFunc func(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool)

// attrRefWrite 引用屬性寫入詞彙表(名稱 → 寫入行為);服務屬性修改命令的引用左值(<引用>.<屬性>)。
// 鍵集僅【營業規格書 | 二十三、屬性清單】卡牌 / 顧客引用屬性子表存取欄為 寫 / 寫鎖 / 鎖 的可寫屬性;唯讀屬性不在此(Validate 據此擋)。
// 型別不符的引用(以顧客引用寫卡牌屬性等)回 changed=false。寫側不需 Lock 後綴路由(鎖定變更由 @ # 表達)。
// 每一詞條對應一個獨立的 writeRef* 函式(便於逐條單元測試);本表僅作名稱 → 行為的索引。
var attrRefWrite = map[string]attrRefWriteFunc{
	// 卡牌引用屬性(cost / extraRun* 寫鎖;cardSeal / keep / playExile / unplayExile 純鎖)
	"cost":        writeRefCost,
	"extraRunMin": writeRefExtraRunMin,
	"extraRunMax": writeRefExtraRunMax,
	"cardSeal":    writeRefCardSeal,
	"keep":        writeRefKeep,
	"playExile":   writeRefPlayExile,
	"unplayExile": writeRefUnplayExile,

	// 顧客引用屬性(calm / sate / sateMax / morale / moraleMax / score / scoreMax 寫鎖;sateSeal / calmSeal 純鎖)
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

// HasAttrRefWrite 回報引用屬性寫入詞彙表是否登錄 name;供 games.Validate 檢查屬性修改命令的引用左值屬性可寫性。
func HasAttrRefWrite(name string) bool {
	_, ok := attrRefWrite[name]
	return ok
}

// === 卡牌引用屬性 ===

// writeRefCost 寫卡牌出牌費用(寫鎖)。
func writeRefCost(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	card, ok := asCard(ref)

	if ok == false {
		return false
	} // if

	return writeValue(&card.Cost, op, n, nil)
}

// writeRefExtraRunMin 寫卡牌額外發動次數下限(寫鎖)。
func writeRefExtraRunMin(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	card, ok := asCard(ref)

	if ok == false {
		return false
	} // if

	return writeValue(&card.ExtraRunMin, op, n, nil)
}

// writeRefExtraRunMax 寫卡牌額外發動次數上限(寫鎖)。
func writeRefExtraRunMax(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	card, ok := asCard(ref)

	if ok == false {
		return false
	} // if

	return writeValue(&card.ExtraRunMax, op, n, nil)
}

// writeRefCardSeal 寫卡牌封印(純鎖,僅 @ #)。
func writeRefCardSeal(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	card, ok := asCard(ref)

	if ok == false {
		return false
	} // if

	return writeLockOnly(&card.Seal, op)
}

// writeRefKeep 寫卡牌不棄(純鎖,僅 @ #)。
func writeRefKeep(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	card, ok := asCard(ref)

	if ok == false {
		return false
	} // if

	return writeLockOnly(&card.Keep, op)
}

// writeRefPlayExile 寫卡牌出牌後流放(純鎖,僅 @ #)。
func writeRefPlayExile(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	card, ok := asCard(ref)

	if ok == false {
		return false
	} // if

	return writeLockOnly(&card.PlayExile, op)
}

// writeRefUnplayExile 寫卡牌未出牌流放(純鎖,僅 @ #)。
func writeRefUnplayExile(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	card, ok := asCard(ref)

	if ok == false {
		return false
	} // if

	return writeLockOnly(&card.UnplayExile, op)
}

// === 顧客引用屬性 ===

// writeRefCalm 寫顧客耐心值(寫鎖)。
func writeRefCalm(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	guest, ok := asGuest(ref)

	if ok == false {
		return false
	} // if

	return writeValue(&guest.Calm, op, n, nil)
}

// writeRefSate 寫顧客飽食值(寫鎖)。
func writeRefSate(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	guest, ok := asGuest(ref)

	if ok == false {
		return false
	} // if

	return writeValue(&guest.Sate, op, n, nil)
}

// writeRefSateMax 寫顧客飽食值離場線(寫鎖)。
func writeRefSateMax(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	guest, ok := asGuest(ref)

	if ok == false {
		return false
	} // if

	return writeValue(&guest.SateMax, op, n, nil)
}

// writeRefMorale 寫顧客士氣值(寫鎖);引用屬性的 -= 走一般運算,不啟動餐廳 morale 特例。
func writeRefMorale(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	guest, ok := asGuest(ref)

	if ok == false {
		return false
	} // if

	return writeValue(&guest.Morale, op, n, nil)
}

// writeRefMoraleMax 寫顧客士氣值上限(寫鎖)。
func writeRefMoraleMax(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	guest, ok := asGuest(ref)

	if ok == false {
		return false
	} // if

	return writeValue(&guest.MoraleMax, op, n, nil)
}

// writeRefScore 寫顧客滿意值(寫鎖)。
func writeRefScore(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	guest, ok := asGuest(ref)

	if ok == false {
		return false
	} // if

	return writeValue(&guest.Score, op, n, nil)
}

// writeRefScoreMax 寫顧客滿意值上限(寫鎖)。
func writeRefScoreMax(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	guest, ok := asGuest(ref)

	if ok == false {
		return false
	} // if

	return writeValue(&guest.ScoreMax, op, n, nil)
}

// writeRefSateSeal 寫顧客封印飽食技能(純鎖,僅 @ #)。
func writeRefSateSeal(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	guest, ok := asGuest(ref)

	if ok == false {
		return false
	} // if

	return writeLockOnly(&guest.SateSeal, op)
}

// writeRefCalmSeal 寫顧客封印耐心技能(純鎖,僅 @ #)。
func writeRefCalmSeal(eng *Engine, ref exprs.Ref, op AssignKind, n float64) (changed bool) {
	guest, ok := asGuest(ref)

	if ok == false {
		return false
	} // if

	return writeLockOnly(&guest.CalmSeal, op)
}
