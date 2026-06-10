package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// 核心流程家族（phase*.go）:【營業規格書 | 十九、核心流程】一階段一檔一函式,檔名 = 函式名 = cores.PhaseKind 列舉名。
// 每站只回報下一站、不互相呼叫（單站可獨立測試）;營業成功 / 失敗觸發後回 PhaseNone 作停機訊號。
// 本檔持分派入口（runPhase）與跨站共用 helper;驅動迴圈與匯出接縫屬 games.Run（M16）。

// runPhase 跑指定階段一站、回下一站;PhaseNone / 未知階段回 PhaseNone（停機）。
func runPhase(game *cores.Game, phase cores.PhaseKind) cores.PhaseKind {
	switch phase {
	case cores.PhaseGameStart:
		return phaseGameStart(game)

	case cores.PhaseRoundStart:
		return phaseRoundStart(game)

	case cores.PhasePlayerAction:
		return phasePlayerAction(game)

	case cores.PhaseGuestAction:
		return phaseGuestAction(game)

	case cores.PhaseRoundEnd:
		return phaseRoundEnd(game)

	case cores.PhaseGameSucc:
		return phaseGameSucc(game)

	case cores.PhaseGameFail:
		return phaseGameFail(game)

	default:
		return cores.PhaseNone // 無階段 / 未知 → 停機
	} // switch
}

// energyFill 點數補滿:出牌點數低於上限時補至上限（「出牌點數 = max(出牌點數, 出牌點數上限)」;鎖定 → 不補）。
// 營業開始 / 回合開始設置共用。
func energyFill(game *cores.Game) {
	if game.GetEnergy().GetValue() < game.GetEnergyMax().GetValue() {
		game.GetEnergy().Set(float64(game.GetEnergyMax().GetValue()))
	} // if
}

// skillGroup 取技能的技能群組編號（供 skillImmune 排除;【營業規格書 | 二十、獨立流程 | 啟動技能】）;技能資料缺失回 0。
// 前置技能（營業開始）與顧客行動技能（顧客行動）共用;卡牌出牌走 Game.CardSkillGroup。
func skillGroup(game *cores.Game, skillID int32) int32 {
	meta := game.GetSheet().Skill.Get(skillID)

	if meta == nil {
		return 0
	} // if

	return meta.Group
}
