package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// attrWrite 全域屬性寫入詞彙表(名稱 → 寫入行為); 服務屬性修改命令的全域左值。
// 鍵集僅【營業規格書 | 二十三、屬性清單】主表存取欄為 寫 / 寫鎖 / 鎖 的可寫屬性; 唯讀屬性不在此(Validate 據此擋)。
// 寫側不需讀側的 Lock 後綴路由——鎖定變更由賦值符 @ # 表達, 左值名即屬性名(無 moraleLock 這類左值)。
// 每一詞條對應一個獨立的 write* 函式(便於逐條單元測試); 本表僅作名稱 → 行為的索引。
var attrWrite = map[string]cores.AttrWriteFunc{
	// 餐廳 / 出牌全域數值屬性(寫鎖; morale 帶 -= 特例、energyKeep 為純鎖; 範圍依【營業規格書 | 二十三、屬性清單】
	// 於命令寫入路徑夾值, 上限縮減的連動壓回在執行結算)
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

	// 回合(寫; 無鎖定計數; roundLeft 為衍生 → 回合上限)
	"round":     writeRound,
	"roundMax":  writeRoundMax,
	"roundLeft": writeRoundLeft,
}

// HasAttrWrite 回報全域屬性寫入詞彙表是否登錄 name; 供 games.Validate 檢查屬性修改命令的全域左值可寫性。
func HasAttrWrite(name string) bool {
	_, ok := attrWrite[name]
	return ok
}

// === 餐廳 / 出牌全域數值屬性 ===

// writeMorale 寫餐廳士氣值(範圍 0 ~ moraleMax); -= 走士氣受損特例(格擋 → 護盾 → morale, 該路徑自夾下限 0), 其餘為一般寫鎖運算。
// 屬性修改命令路徑的受損來源取 self 顧客(damageSource); guestExit 等流程改傳離場顧客為來源, 直接呼叫 moraleDamage。
func writeMorale(game *cores.Game, op cores.AssignKind, n float64) bool {
	if op == cores.AssignSub {
		return moraleDamage(game, n, damageSource(game.GetSelf()))
	} // if

	changed := game.GetMorale().Apply(op, n)
	game.GetMorale().Clamp(0)
	game.GetMorale().ClampMax(game.GetMoraleMax().GetValue())
	return changed
}

// writeMoraleMax 寫餐廳士氣上限(下限夾 0; 縮減使 morale 超標時由執行結算壓回)。
func writeMoraleMax(game *cores.Game, op cores.AssignKind, n float64) bool {
	changed := game.GetMoraleMax().Apply(op, n)
	game.GetMoraleMax().Clamp(0)
	return changed
}

// writeMoraleShield 寫餐廳士氣護盾(下限夾 0)。
func writeMoraleShield(game *cores.Game, op cores.AssignKind, n float64) bool {
	changed := game.GetMoraleShield().Apply(op, n)
	game.GetMoraleShield().Clamp(0)
	return changed
}

// writeMoraleBlock 寫餐廳士氣格擋(下限夾 0)。
func writeMoraleBlock(game *cores.Game, op cores.AssignKind, n float64) bool {
	changed := game.GetMoraleBlock().Apply(op, n)
	game.GetMoraleBlock().Clamp(0)
	return changed
}

// writeScore 寫餐廳滿意值(下限夾 0)。
func writeScore(game *cores.Game, op cores.AssignKind, n float64) bool {
	changed := game.GetScore().Apply(op, n)
	game.GetScore().Clamp(0)
	return changed
}

// writeEnergy 寫出牌點數(範圍 0 ~ energyMax)。
func writeEnergy(game *cores.Game, op cores.AssignKind, n float64) bool {
	changed := game.GetEnergy().Apply(op, n)
	game.GetEnergy().Clamp(0)
	game.GetEnergy().ClampMax(game.GetEnergyMax().GetValue())
	return changed
}

// writeEnergyMax 寫出牌點數上限(下限夾 0; 縮減使 energy 超標時由執行結算壓回)。
func writeEnergyMax(game *cores.Game, op cores.AssignKind, n float64) bool {
	changed := game.GetEnergyMax().Apply(op, n)
	game.GetEnergyMax().Clamp(0)
	return changed
}

// writeEnergyKeep 寫出牌點數保留(純鎖屬性, 數值固定 0, 僅 @ #)。
func writeEnergyKeep(game *cores.Game, op cores.AssignKind, n float64) bool {
	return game.GetEnergyKeep().ApplyLockOnly(op)
}

// writeHandMax 寫手牌張數上限(下限夾 0)。
func writeHandMax(game *cores.Game, op cores.AssignKind, n float64) bool {
	changed := game.GetHandMax().Apply(op, n)
	game.GetHandMax().Clamp(0)
	return changed
}

// writeDrawMax 寫補牌張數上限(下限夾 0)。
func writeDrawMax(game *cores.Game, op cores.AssignKind, n float64) bool {
	changed := game.GetDrawMax().Apply(op, n)
	game.GetDrawMax().Clamp(0)
	return changed
}

// === 回合 ===

// writeRound 寫當前回合數(寫屬性、無鎖定語意; 範圍 1 ~ roundMax——命令路徑限定, 開局系統歸零不經此)。
func writeRound(game *cores.Game, op cores.AssignKind, n float64) bool {
	changed := game.GetRound().ApplyValueOnly(op, n)
	game.GetRound().Clamp(1)
	game.GetRound().ClampMax(game.GetRoundMax().GetValue())
	return changed
}

// writeRoundMax 寫回合上限(寫屬性、無鎖定語意; 下限夾 0——縮減至低於 round 時不壓回 round, 由終止判定收場)。
func writeRoundMax(game *cores.Game, op cores.AssignKind, n float64) bool {
	changed := game.GetRoundMax().ApplyValueOnly(op, n)
	game.GetRoundMax().Clamp(0)
	return changed
}

// writeRoundLeft 寫剩餘回合(衍生): 轉譯為對回合上限的調整(剩餘回合 = N → 回合上限 = 回合 + N),
// 操作後夾使回合上限 >= 回合(【二十三】roundLeft)。基準取原始 回合上限 - 回合(含可負, 使 += N 等同 回合上限 += N),
// 以暫存 Value 套運算後夾 剩餘 >= 0(等價於 回合上限 >= 回合)再寫回。
func writeRoundLeft(game *cores.Game, op cores.AssignKind, n float64) bool {
	round := game.GetRound().GetValue()
	left := cores.NewValue(game.GetRoundMax().GetValue()-round, 0)

	if left.ApplyValueOnly(op, n) == false {
		return false // @ # 或 /= %= 除 0 → no-op
	} // if

	left.Clamp(0)
	return game.GetRoundMax().Set(float64(round + left.GetValue()))
}
