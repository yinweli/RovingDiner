package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// advanceEffect 推進效果(【營業規格書 | 二十、獨立流程 | 推進效果】): 自效果佇列篩選「結束回合 > 0 && 結束回合 <= 當前回合」者、
// 依作用順序排序、逐一退場(結束命令 × 堆疊層數 → 出佇列)。呼叫點為回合結束階段(【營業規格書 | 十九、核心流程 | 5】, M14 接線)。
// 凍結中顧客的效果一律跳過——結束回合不推進, 解凍時由 restore 補回差額(【營業規格書 | 二十一、流程補充 | 凍結語意】效果凍結)。
func advanceEffect(game *cores.Game) {
	round := game.GetRound().GetValue()
	end := cores.EffectList{}

	for _, itor := range game.Effect {
		if _, ok := game.EffectData(itor.GetEffectID()); ok == false {
			continue // 查無編譯資料(防禦; 比照 fireTrigger)→ 略過
		} // if

		if frozenSelf(game, itor) {
			continue // 效果凍結: 凍結中顧客的效果不推進
		} // if

		if itor.GetExpire() > 0 && itor.GetExpire() <= round {
			end = append(end, itor)
		} // if
	} // for

	end.Sort(game) // 快照排序; 不擾動 game.Effect(順序無關)

	for _, itor := range end {
		retireEffect(game, itor)
	} // for
}
