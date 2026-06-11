package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// 核心流程家族(phase*.go): 【營業規格書 | 十九、核心流程】一階段一檔一函式, 檔名 = 函式名 = cores.PhaseKind 列舉名。
// 每站只回報下一站、不互相呼叫(單站可獨立測試); 營業成功 / 失敗觸發後回 PhaseNone 作停機訊號。
// 本檔持分派入口(RunPhase, games.Run 的驅動接縫)與跨站共用 helper; 驅動迴圈與成敗記錄屬 games.Run。

// RunPhase 跑指定階段一站、回下一站; PhaseNone / 未知階段回 PhaseNone(停機, 不踏站不發事件)。
// 已知階段踏站時先 SetPhase(Emit 座標蓋章來源)再發 phase 切換事件(無本體欄位, 座標即新階段; M17 拍板、M18 接線)。
// 終止判定的哨兵(gameEnd)在此單點 recover → 回對應終止站; 非哨兵 panic 原樣重拋(真 bug 不被吞)。
func RunPhase(game *cores.Game, phase cores.PhaseKind) (next cores.PhaseKind) {
	defer func() {
		if cause := recover(); cause != nil {
			end, ok := cause.(gameEnd)

			if ok == false {
				panic(cause) // 非哨兵 → 原樣重拋
			} // if

			next = end.phase
		} // if
	}()

	var station func(game *cores.Game) cores.PhaseKind

	switch phase {
	case cores.PhaseGameStart:
		station = phaseGameStart

	case cores.PhaseRoundStart:
		station = phaseRoundStart

	case cores.PhasePlayerAction:
		station = phasePlayerAction

	case cores.PhaseGuestAction:
		station = phaseGuestAction

	case cores.PhaseRoundEnd:
		station = phaseRoundEnd

	case cores.PhaseGameSucc:
		station = phaseGameSucc

	case cores.PhaseGameFail:
		station = phaseGameFail

	default:
		return cores.PhaseNone // 無階段 / 未知 → 停機
	} // switch

	game.SetPhase(phase)
	game.Emit(cores.EventData{Kind: cores.EventPhase})
	return station(game)
}

// energyFill 點數補滿: 出牌點數低於上限時補至上限(「出牌點數 = max(出牌點數, 出牌點數上限)」; 鎖定 → 不補)。
// 營業開始 / 回合開始設置共用。流程寫入白名單: 發屬性事件(鎖定不補以 Before == After 表達; M18 拍板)。
func energyFill(game *cores.Game) {
	if game.GetEnergy().GetValue() < game.GetEnergyMax().GetValue() {
		before := float64(game.GetEnergy().GetValue())
		game.GetEnergy().Set(float64(game.GetEnergyMax().GetValue()))
		emitProperty(game, 0, cores.NoneID, "energy", cores.AssignSet, float64(game.GetEnergyMax().GetValue()), before, float64(game.GetEnergy().GetValue()))
	} // if
}

// skillGroup 取技能的技能群組編號(供 skillImmune 排除; 【營業規格書 | 二十、獨立流程 | 啟動技能】); 技能資料缺失回 0。
// 前置技能(營業開始)與顧客行動技能(顧客行動)共用; 卡牌出牌走 Game.CardSkillGroup。
func skillGroup(game *cores.Game, skillID int32) int32 {
	meta := game.GetSheet().Skill.Get(skillID)

	if meta == nil {
		return 0
	} // if

	return meta.Group
}

// cardSkill 取卡牌的技能編號(玩家出牌範圍事件的技能操作元; 【營業顯示規格書 | 6、畫面規格 | 6.10】範圍事件表);
// 卡牌資料缺失回 0。
func cardSkill(game *cores.Game, cardID int32) int32 {
	meta := game.GetSheet().Card.Get(cardID)

	if meta == nil {
		return 0
	} // if

	return meta.SkillID
}
