package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// phaseRoundEnd 回合結束階段(【營業規格書 | 十九、核心流程 | 5. 回合結束階段】): 觸發 roundEnd →
// 座位與遊蕩每位顧客耐心 -1(鎖定 → 不扣)、出牌點數保留未鎖定時點數歸零 → 推進效果(advanceEffect)→
// 執行結算(Settle; 終止判定命中以哨兵跳出)→ 回回合開始。耐心 -1 與點數歸零屬流程寫入白名單(M18 拍板)。
func phaseRoundEnd(game *cores.Game) cores.PhaseKind {
	fireTrigger(game, cores.TriggerRoundEnd) // 回合結束觸發

	for _, itor := range game.Seat.Sorted() { // 座位(座位編號序)→ 遊蕩(列表序), 確保決定性
		calmDrop(game, itor)
	} // for

	for _, itor := range game.Roam {
		calmDrop(game, itor)
	} // for

	if game.GetEnergyKeep().IsLock() == false {
		game.GetEnergy().Set(0) // 出牌點數歸零(出牌點數保留鎖定 → 保留)
		cores.EmitProperty(game, 0, cores.NoneID, "energy", cores.AssignSet, 0, float64(game.GetEnergy().GetValue()))
	} // if

	advanceEffect(game)
	Settle(game)
	return cores.PhaseRoundStart
}

// calmDrop 回合結束的單位顧客耐心 -1(鎖定 → 不扣, 行結果值不變); 流程寫入白名單, 逐位發屬性行。
func calmDrop(game *cores.Game, guest *cores.Guest) {
	guest.GetCalm().Sub(1)
	cores.EmitProperty(game, guest.GetGuestID(), guest.GetInstanceID(), "calm", cores.AssignSub, 1, float64(guest.GetCalm().GetValue()))
}
