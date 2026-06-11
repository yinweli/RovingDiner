package rules

import (
	"strconv"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// phaseGameStart 營業開始階段(【營業規格書 | 十九、核心流程 | 1. 營業開始階段】): 載入全域設定、回合 / 結算旗標歸零、
// 點數補滿、清空四累積表、依關卡編號建置開局盤面 → 逐一啟動前置技能(此時座位空, 撰寫約束詳見
// 【營業規格書 | 二十一、流程補充 | 前置技能的撰寫約束】)→ 觸發 gameStart。
func phaseGameStart(game *cores.Game) cores.PhaseKind {
	loadSetting(game)
	game.GetRound().Set(0)
	game.Settling = false
	energyFill(game)
	game.GetDrawTotal().Reset()
	game.GetDropTotal().Reset()
	game.GetPlayTotal().Reset()
	game.GetExileTotal().Reset()
	buildStage(game)

	for _, itor := range game.PrefixSkill {
		game.Emit(cores.EventData{Kind: cores.EventScope, Scope: cores.ScopePrefix, SkillID: itor}) // 範圍標題: 前置技能(操作元 = 技能)
		runEffectList(game, game.SkillEffect(itor), skillGroup(game, itor))                         // 啟動技能: 該技能
	} // for

	fireTrigger(game, cores.TriggerGameStart) // 營業開始觸發
	return cores.PhaseRoundStart
}

// buildStage 依本場關卡編號自關卡表格建置開局盤面(【營業規格書 | 四、表格結構 | 關卡（Stage）表格】):
// 實例化手牌 / 抽牌 / 棄牌 / 流放 / 排隊五容器(列表第 1 個 = 頂端 / 隊首, 即 slice 前端; M8 約定)並填前置技能列表。
// 建置屬「設置」非命令: 不觸發 cardDraw / cardDrop / cardExile 時機、不動事件屬性。
// 查無關卡 → 空盤面照走; 單筆查無資料 → 跳過該筆(寬鬆, 嚴格把關交企劃驗證器)。
func buildStage(game *cores.Game) {
	stage := game.GetSheet().Stage.Get(game.StageID)

	if stage == nil {
		return // 查無關卡 → 空盤面照走
	} // if

	game.Hand = stageCard(game, stage.HandID)
	game.Deck = stageCard(game, stage.DeckID)
	game.Drop = stageCard(game, stage.DropID)
	game.Exile = stageCard(game, stage.ExileID)

	for _, itor := range stage.WaitID {
		if guest := cores.NewGuest(game, itor); guest != nil {
			game.Wait = append(game.Wait, guest) // 列表序 = 佇列序(第 1 個 = 隊首)
		} // if
	} // for

	game.PrefixSkill = append([]int32(nil), stage.PrefixSkillID...) // 複本; 不共享靜態表底層
}

// stageCard 依卡牌編號列表實例化卡牌容器(列表序 = slice 序, 第 1 個 = 頂端); 查無卡牌資料跳過該筆。
func stageCard(game *cores.Game, cardID []int32) (result cores.CardList) {
	for _, itor := range cardID {
		if card := cores.NewCard(game, itor); card != nil {
			result = append(result, card)
		} // if
	} // for

	return result
}

// loadSetting 自設定表格載入全域設定初值(【營業規格書 | 四、表格結構 | 設定表格】六鍵)。
func loadSetting(game *cores.Game) {
	game.GetRoundMax().Set(settingNum(game, "RoundMax"))
	game.GetMorale().Set(settingNum(game, "Morale"))
	game.GetMoraleMax().Set(settingNum(game, "MoraleMax"))
	game.GetEnergyMax().Set(settingNum(game, "EnergyMax"))
	game.GetHandMax().Set(settingNum(game, "HandMax"))
	game.GetDrawMax().Set(settingNum(game, "DrawMax"))
}

// settingNum 讀單一數字設定值; 缺鍵 / 空值 / 非數字回 0(寬鬆, 嚴格把關交企劃驗證器)。
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
