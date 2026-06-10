package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// phaseRoundEnd 回合結束階段（【營業規格書 | 十九、核心流程 | 5. 回合結束階段】）:觸發 roundEnd →
// 座位與遊蕩每位顧客耐心 -1（鎖定 → 不扣）、出牌點數保留未鎖定時點數歸零 → 推進效果（advanceEffect）→
// 執行結算（Settle;終止判定命中以哨兵跳出）→ 回回合開始。
func phaseRoundEnd(game *cores.Game) cores.PhaseKind {
	fireTrigger(game, cores.TriggerRoundEnd) // 回合結束觸發

	for _, itor := range game.Seat.Sorted() { // 座位（座位編號序）→ 遊蕩（列表序）,確保決定性
		itor.GetCalm().Sub(1)
	} // for

	for _, itor := range game.Roam {
		itor.GetCalm().Sub(1)
	} // for

	if game.GetEnergyKeep().IsLock() == false {
		game.GetEnergy().Set(0) // 出牌點數歸零（出牌點數保留鎖定 → 保留）
	} // if

	advanceEffect(game)
	Settle(game)
	return cores.PhaseRoundStart
}
