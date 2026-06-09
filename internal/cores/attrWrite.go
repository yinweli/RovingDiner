package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// attrWriteFunc 全域屬性詞條的寫入行為:以 engine 為 context、op 為賦值符、n 為已求值的右值(@ # 時忽略)。
type attrWriteFunc func(eng *Engine, op AssignKind, n float64) (changed bool)

// attrWrite 全域屬性寫入詞彙表(名稱 → 寫入行為);服務屬性修改命令的全域左值。
// 鍵集僅【營業規格書 | 二十三、屬性清單】主表存取欄為 寫 / 寫鎖 / 鎖 的可寫屬性;唯讀屬性不在此(Validate 據此擋)。
// 寫側不需讀側的 Lock 後綴路由——鎖定變更由賦值符 @ # 表達,左值名即屬性名(無 moraleLock 這類左值)。
// 每一詞條對應一個獨立的 write* 函式(便於逐條單元測試);本表僅作名稱 → 行為的索引。
var attrWrite = map[string]attrWriteFunc{
	// 餐廳 / 出牌全域數值屬性(寫鎖;morale 帶 -= 特例、護盾 / 格擋夾下限 0、energyKeep 為純鎖)
	"morale":       writeMorale,
	"moraleMax":    writeMoraleMax,
	"moraleShield": writeMoraleShield,
	"moraleBlock":  writeMoraleBlock,
	"score":        writeScore,
	"energy":       writeEnergy,
	"energyMax":    writeEnergyMax,
	"energyKeep":   writeEnergyKeep,
	"handMax":      writeHandMax,
	"drawMax":      writeDrawMax,

	// 回合(寫;無鎖定計數;roundLeft 為衍生 → 回合上限)
	"round":     writeRound,
	"roundMax":  writeRoundMax,
	"roundLeft": writeRoundLeft,
}

// HasAttrWrite 回報全域屬性寫入詞彙表是否登錄 name;供 games.Validate 檢查屬性修改命令的全域左值可寫性。
func HasAttrWrite(name string) bool {
	_, ok := attrWrite[name]
	return ok
}

// === 餐廳 / 出牌全域數值屬性 ===

// writeMorale 寫餐廳士氣值;-= 走士氣受損特例(格擋 → 護盾 → morale),其餘為一般寫鎖運算。
// 屬性修改命令路徑的受損來源取 self 顧客(damageSource);guestExit 等流程改傳離場顧客為來源,直接呼叫 moraleDamage。
func writeMorale(eng *Engine, op AssignKind, n float64) (changed bool) {
	if op == AssignSub {
		return moraleDamage(eng, n, damageSource(eng.self))
	} // if

	return writeValue(&eng.runtime.Game.Morale, op, n, nil)
}

// writeMoraleMax 寫餐廳士氣上限。
func writeMoraleMax(eng *Engine, op AssignKind, n float64) (changed bool) {
	return writeValue(&eng.runtime.Game.MoraleMax, op, n, nil)
}

// writeMoraleShield 寫餐廳士氣護盾(下限夾 0)。
func writeMoraleShield(eng *Engine, op AssignKind, n float64) (changed bool) {
	return writeValue(&eng.runtime.Game.MoraleShield, op, n, clampLow0)
}

// writeMoraleBlock 寫餐廳士氣格擋(下限夾 0)。
func writeMoraleBlock(eng *Engine, op AssignKind, n float64) (changed bool) {
	return writeValue(&eng.runtime.Game.MoraleBlock, op, n, clampLow0)
}

// writeScore 寫餐廳滿意值。
func writeScore(eng *Engine, op AssignKind, n float64) (changed bool) {
	return writeValue(&eng.runtime.Game.Score, op, n, nil)
}

// writeEnergy 寫出牌點數。
func writeEnergy(eng *Engine, op AssignKind, n float64) (changed bool) {
	return writeValue(&eng.runtime.Game.Energy, op, n, nil)
}

// writeEnergyMax 寫出牌點數上限。
func writeEnergyMax(eng *Engine, op AssignKind, n float64) (changed bool) {
	return writeValue(&eng.runtime.Game.EnergyMax, op, n, nil)
}

// writeEnergyKeep 寫出牌點數保留(純鎖屬性,Value 固定 0,僅 @ #)。
func writeEnergyKeep(eng *Engine, op AssignKind, n float64) (changed bool) {
	return writeLockOnly(&eng.runtime.Game.EnergyKeep, op)
}

// writeHandMax 寫手牌張數上限。
func writeHandMax(eng *Engine, op AssignKind, n float64) (changed bool) {
	return writeValue(&eng.runtime.Game.HandMax, op, n, nil)
}

// writeDrawMax 寫補牌張數上限。
func writeDrawMax(eng *Engine, op AssignKind, n float64) (changed bool) {
	return writeValue(&eng.runtime.Game.DrawMax, op, n, nil)
}

// === 回合 ===

// writeRound 寫當前回合數(整數欄位、無鎖定計數)。
func writeRound(eng *Engine, op AssignKind, n float64) (changed bool) {
	return writeInt(&eng.runtime.Game.Round, op, n)
}

// writeRoundMax 寫回合上限(整數欄位、無鎖定計數)。
func writeRoundMax(eng *Engine, op AssignKind, n float64) (changed bool) {
	return writeInt(&eng.runtime.Game.RoundMax, op, n)
}

// writeRoundLeft 寫剩餘回合(衍生):轉譯為對回合上限的調整(剩餘回合 = N → 回合上限 = 回合 + N),
// 操作後夾使回合上限 >= 回合(【二十三】roundLeft)。基準取原始 回合上限 - 回合(含可負,使 += N 等同 回合上限 += N)。
func writeRoundLeft(eng *Engine, op AssignKind, n float64) (changed bool) {
	game := eng.runtime.Game
	result, ok := applyOp(op, float64(game.RoundMax-game.Round), n)

	if ok == false {
		return false // @ # 或 /= %= 除 0 → no-op
	} // if

	newMax := game.Round + exprs.Round(result)

	if newMax < game.Round { // clamp 回合上限 >= 回合
		newMax = game.Round
	} // if

	game.RoundMax = newMax
	return true
}

// === 餐廳士氣值 -= 特例 ===

// moraleDamage 餐廳士氣值 -= 特例(【十七、命令 | 1】特例):依 格擋 → 護盾 → morale 順序消耗扣減值 N;
// 實際扣減 > 0 時設置 damageValue / damageGuest(來源 source 由呼叫端決定:命令路徑取 self 顧客、guestExit 取離場顧客),並標記士氣受損時機。
// morale 鎖定時格擋 / 護盾仍消耗、morale 不動、無實際扣減(對齊目前解讀)。
// N 先四捨五入為整數扣減值,使格擋 / 護盾 / morale 的整數消耗自洽(小數扣減值的捨入時點待規格確認)。
func moraleDamage(eng *Engine, n float64, source *Guest) (changed bool) {
	game := eng.runtime.Game
	damage := exprs.Round(n)

	if game.MoraleBlock.Value > 0 { // 格擋優先:格擋 -= 1,本次無視 N
		game.MoraleBlock.Value--
		return true
	} // if

	if damage > 0 { // 護盾消耗:d = min(N, 護盾) → 護盾 -= d → N -= d
		d := min(damage, game.MoraleShield.Value)

		if d > 0 {
			game.MoraleShield.Value -= d
			damage -= d
			changed = true
		} // if
	} // if

	if damage > 0 && game.Morale.Locked() == false { // 殘餘對 morale 一般 -= 運算(鎖定 → 不扣)
		before := game.Morale.Value
		game.Morale.Value -= damage
		actual := before - game.Morale.Value

		if actual > 0 {
			game.DamageValue = actual
			game.DamageGuest = source
			changed = true
			fireTrigger(eng, TriggerDamage) // 士氣受損時機
		} // if
	} // if

	return changed
}

// damageSource 取士氣受損來源顧客:self 為顧客時回該顧客,否則空物件(nil);供 Phase 4 命令路徑的 morale -= 使用。
func damageSource(self *Self) (guest *Guest) {
	if self != nil && self.Guest != nil {
		return self.Guest
	} // if

	return nil
}
