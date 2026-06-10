package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// phaseGuestAction 顧客行動階段（【營業規格書 | 十九、核心流程 | 4. 顧客行動階段】）:觸發 guestStart →
// 行動佇列消費迴圈——佇列空 / 跳轉回合結束 → 觸發 guestEnd 收尾;逐筆彈出設行動事件、封印閘門（sealTask）、
// 啟動顧客行動技能、觸發 guestTask。
func phaseGuestAction(game *cores.Game) cores.PhaseKind {
	fireTrigger(game, cores.TriggerGuestStart) // 顧客開始觸發

	for {
		if len(game.Action) == 0 || game.GetNextPhase() == cores.PhaseRoundEnd {
			fireTrigger(game, cores.TriggerGuestEnd) // 顧客結束觸發
			game.SetNextPhase(cores.PhaseNone)       // 清除跳轉
			return cores.PhaseRoundEnd
		} // if

		action := game.Action.Pop()
		game.EventTask(action.GetGuest(), action.GetSkillID()) // 最後行動顧客 / 技能 / 回合行動次數

		if sealTask(action) == false {
			runEffectList(game, game.SkillEffect(action.GetSkillID()), skillGroup(game, action.GetSkillID())) // 啟動技能:顧客行動技能
		} // if

		fireTrigger(game, cores.TriggerGuestTask) // 顧客行動觸發
	} // for
}

// sealTask 行動封印閘門:依行動類型查顧客的封印飽食 / 耐心技能鎖定計數;封印 → 跳過啟動（行動仍消費完畢）。
// 越界行動類型一律視為封印（防禦;taskAdd 對行動類型未驗範圍）。
func sealTask(action *cores.Action) bool {
	switch action.GetKind() {
	case cores.TaskSate:
		return action.GetGuest().GetSateSeal().IsLock()

	case cores.TaskCalm:
		return action.GetGuest().GetCalmSeal().IsLock()

	default:
		return true // 越界行動類型 → 跳過啟動（防禦）
	} // switch
}
