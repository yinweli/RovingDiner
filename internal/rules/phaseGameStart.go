package rules

import (
	"strconv"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// phaseGameStart 營業開始階段（【營業規格書 | 十九、核心流程 | 1. 營業開始階段】）:載入全域設定、回合 / 結算旗標歸零、
// 點數補滿、清空四累積表 → 逐一啟動前置技能（此時座位與手牌皆空,撰寫約束詳見【營業規格書 | 二十一、流程補充 | 前置技能的撰寫約束】）→
// 觸發 gameStart。初始牌堆 / 排隊佇列 / 前置技能列表由組裝層（infra / tester）注入,本站只消費。
func phaseGameStart(game *cores.Game) cores.PhaseKind {
	loadSetting(game)
	game.GetRound().Set(0)
	game.Settling = false
	energyFill(game)
	game.GetDrawTotal().Reset()
	game.GetDropTotal().Reset()
	game.GetPlayTotal().Reset()
	game.GetExileTotal().Reset()

	for _, itor := range game.PrefixSkill {
		runEffectList(game, game.SkillEffect(itor), skillGroup(game, itor)) // 啟動技能:該技能
	} // for

	fireTrigger(game, cores.TriggerGameStart) // 營業開始觸發
	return cores.PhaseRoundStart
}

// loadSetting 自設定表格載入全域設定初值（【營業規格書 | 四、表格結構 | 設定表格】六鍵）。
func loadSetting(game *cores.Game) {
	game.GetRoundMax().Set(settingNum(game, "RoundMax"))
	game.GetMorale().Set(settingNum(game, "Morale"))
	game.GetMoraleMax().Set(settingNum(game, "MoraleMax"))
	game.GetEnergyMax().Set(settingNum(game, "EnergyMax"))
	game.GetHandMax().Set(settingNum(game, "HandMax"))
	game.GetDrawMax().Set(settingNum(game, "DrawMax"))
}

// settingNum 讀單一數字設定值;缺鍵 / 空值 / 非數字回 0（寬鬆,嚴格把關交企劃驗證器）。
func settingNum(game *cores.Game, name string) float64 {
	meta := game.GetSheet().Setting.Get(name)

	if meta == nil || len(meta.Value) == 0 {
		return 0
	} // if

	result, err := strconv.ParseFloat(meta.Value[0], 64)

	if err != nil {
		return 0
	} // if

	return result
}
