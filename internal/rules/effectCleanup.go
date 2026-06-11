package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// cleanupEffect 清理效果(【營業規格書 | 二十、獨立流程 | 清理效果】): 以失效對象 dead 篩選「目標類型 != 無目標 && self 同一實例」者、
// 依作用順序排序、逐一退場(結束命令 × 堆疊層數 → 出佇列)。呼叫點: 變身(morph, 變身後以同卡引用清理, 變身前綁定者全中)/
// 顧客離場(guestExitOne)。凍結中顧客不視為失效對象 → 整體 no-op(【營業規格書 | 二十一、流程補充 | 凍結語意】效果凍結)。
func cleanupEffect(game *cores.Game, dead cores.Ref) {
	if guest := dead.GetGuest(); guest != nil && game.IsFrozen(guest) {
		return // 凍結中顧客不視為失效對象
	} // if

	clean := cores.EffectList{}

	for _, itor := range game.Effect {
		meta, ok := game.EffectData(itor.GetEffectID())

		if ok == false || meta.TargetKind == cores.TargetNone {
			continue // 查無編譯資料(防禦)/ 無目標效果不在清理範圍 → 略過
		} // if

		if itor.GetSelf().IsSame(dead) {
			clean = append(clean, itor)
		} // if
	} // for

	clean.Sort(game) // 快照排序; 不擾動 game.Effect(順序無關)

	for _, itor := range clean {
		retireEffect(game, itor)
	} // for
}
