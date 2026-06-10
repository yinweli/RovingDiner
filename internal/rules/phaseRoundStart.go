package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// phaseRoundStart 回合開始階段（【營業規格書 | 十九、核心流程 | 2. 回合開始階段】）:回合 +1、回合計數歸零 → 觸發 roundReady →
// 點數補滿 → 觸發 roundStart → 排隊入座迴圈（有空位且有人排隊,逐位入座）→ 依下一階段分派（顧客行動 / 回合結束 / 預設玩家行動,皆清除跳轉）。
func phaseRoundStart(game *cores.Game) cores.PhaseKind {
	game.GetRound().Add(1)
	game.RoundReset()
	fireTrigger(game, cores.TriggerRoundReady) // 回合準備觸發

	energyFill(game)
	fireTrigger(game, cores.TriggerRoundStart) // 回合開始觸發

	for guestSeatOne(game) { // 剩餘座位 > 0 && 排隊佇列 > 0 → 逐位入座
	} // for

	next := game.GetNextPhase()
	game.SetNextPhase(cores.PhaseNone) // 各分支皆清除跳轉

	switch next {
	case cores.PhaseGuestAction, cores.PhaseRoundEnd:
		return next

	default:
		return cores.PhasePlayerAction
	} // switch
}
