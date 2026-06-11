package rodi

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// 詞彙對照(日誌中文化; 【營業顯示規格書 | 6、畫面規格 | 6.10】): 事件只帶英文鍵 / 編號 / 列舉,
// 中文名由顯示端憑本表轉換(M17 拍板: 名稱由前端自查的屬性鍵版本)。
// 名稱依【營業規格書 | 二、英文詞彙對照】與【營業規格書 | 二十三、屬性清單】;
// 查無對照退回原文(寬鬆, 比照識別碼查無顯 ?)。

// 免疫詞條鍵(群組維度的 property 事件: Operand 載群組、前後值載該群組計數; M22 拍板):
// 詞彙對照 / 日誌行文特判共用。
const (
	attrEffectImmune = "effectImmune"
	attrSkillImmune  = "skillImmune"
)

// attrGlobalName 全域屬性詞條鍵 → 中文全名(§6.10: 全域屬性用全名以別於 self 屬性);
// 鍵集 = 全域寫詞條 + 流程白名單(設定載入六鍵已含於寫詞條)。
var attrGlobalName = map[string]string{
	"morale":       "餐廳士氣值",
	"moraleMax":    "餐廳士氣值上限",
	"moraleShield": "餐廳士氣值護盾",
	"moraleBlock":  "餐廳士氣值格擋",
	"score":        "餐廳滿意值",
	"energy":       "出牌點數",
	"energyMax":    "出牌點數上限",
	"energyKeep":   "出牌點數保留",
	"handMax":      "手牌張數上限",
	"drawMax":      "補牌張數上限",
	"round":        "回合",
	"roundMax":     "回合上限",
	"roundLeft":    "剩餘回合",
}

// attrRefName 實例屬性詞條鍵 → 中文名(卡牌 / 顧客引用屬性合表, 鍵集不相撞;
// 同名鍵 morale / moraleMax / score / scoreMax 於此表取顧客義, 與全域表以對象欄分流)。
var attrRefName = map[string]string{
	"cost":           "出牌費用",
	"extraRunMin":    "額外發動次數下限",
	"extraRunMax":    "額外發動次數上限",
	"cardSeal":       "封印卡牌",
	"keep":           "不棄卡牌",
	"playExile":      "出牌後流放",
	"unplayExile":    "未出牌流放",
	"sate":           "飽食值",
	"sateMax":        "飽食值離場線",
	"calm":           "耐心值",
	"morale":         "士氣值",
	"moraleMax":      "士氣值上限",
	"score":          "滿意值",
	"scoreMax":       "滿意值上限",
	"sateSeal":       "封印飽食技能",
	"calmSeal":       "封印耐心技能",
	attrEffectImmune: "效果免疫群組",
	attrSkillImmune:  "技能免疫群組",
}

// triggerName 觸發時機英文鍵 → 中文名(【營業規格書 | 二十二、觸發時機清單】)。
var triggerName = map[cores.TriggerKind]string{
	cores.TriggerGameStart:  "營業開始",
	cores.TriggerRoundReady: "回合準備",
	cores.TriggerRoundStart: "回合開始",
	cores.TriggerGuestSeat:  "顧客入座",
	cores.TriggerCardDraw:   "手牌補充",
	cores.TriggerUserStart:  "玩家開始",
	cores.TriggerUserEnd:    "玩家結束",
	cores.TriggerCardDrop:   "卡牌棄置",
	cores.TriggerCardExile:  "卡牌流放",
	cores.TriggerCardMorph:  "卡牌變身",
	cores.TriggerCardPlay:   "玩家出牌",
	cores.TriggerGuestStart: "顧客開始",
	cores.TriggerGuestTask:  "顧客行動",
	cores.TriggerGuestEnd:   "顧客結束",
	cores.TriggerRoundEnd:   "回合結束",
	cores.TriggerExitAny:    "顧客離場",
	cores.TriggerExitSate:   "飽食離場",
	cores.TriggerExitCalm:   "生氣離場",
	cores.TriggerExitDone:   "顧客離場後",
	cores.TriggerDamage:     "士氣受損",
	cores.TriggerGameSucc:   "營業成功",
	cores.TriggerGameFail:   "營業失敗",
}

// sourceName 選取來源英文鍵 → 中文名(Pick 詞條鍵依【營業規格書 | 二、英文詞彙對照】主體 + 動詞組合,
// discardOver 為手牌上限棄牌流程名); 效果目標選取(EffectID 非零)改顯效果識別碼、不查本表。
var sourceName = map[string]string{
	"guestPick":   "顧客指定",
	"nearPick":    "鄰桌指定",
	"samePick":    "同桌指定",
	"handPick":    "手牌指定",
	"deckPick":    "抽牌堆指定",
	"dropPick":    "棄牌堆指定",
	"exilePick":   "流放堆指定",
	"cardPick":    "卡牌指定",
	"discardOver": "手牌上限",
}

// attrText 屬性詞條鍵轉中文名: 對象欄零值走全域表、實例對象走引用表; 查無退回原文。
func attrText(attr string, global bool) string {
	table := attrRefName

	if global {
		table = attrGlobalName
	} // if

	if name, ok := table[attr]; ok {
		return name
	} // if

	return attr
}

// triggerText 觸發時機中文名; 查無退回原文。
func triggerText(trigger cores.TriggerKind) string {
	if name, ok := triggerName[trigger]; ok {
		return name
	} // if

	return string(trigger)
}

// sourceText 選取來源中文名; 查無退回原文。
func sourceText(source string) string {
	if name, ok := sourceName[source]; ok {
		return name
	} // if

	return source
}

// scopeText 範圍事件名(標題的 {事件}; 【營業顯示規格書 | 6、畫面規格 | 6.10】範圍事件表)。
func scopeText(scope cores.ScopeKind, trigger cores.TriggerKind) string {
	switch scope {
	case cores.ScopePlay:
		return "玩家出牌"

	case cores.ScopeGuest:
		return "顧客行動"

	case cores.ScopePrefix:
		return "前置技能"

	case cores.ScopeTrigger:
		return "時機:" + triggerText(trigger)

	case cores.ScopeSettle:
		return "執行結算"

	case cores.ScopeManual:
		return "手動結束"

	default:
		return "?"
	} // switch
}

// stageText 效果階段名(【營業顯示規格書 | 6、畫面規格 | 6.10】效果欄位的階段值)。
func stageText(stage cores.EffectStage) string {
	switch stage {
	case cores.EffectStageImmed:
		return "立即"

	case cores.EffectStageJoin:
		return "加入"

	case cores.EffectStageTrigger:
		return "觸發"

	case cores.EffectStageStart:
		return "啟動"

	case cores.EffectStageEnd:
		return "結束"

	case cores.EffectStageCondFail:
		return "條件不成立"

	default:
		return "?"
	} // switch
}

// assignText 賦值符(【營業規格書 | 十七、命令 | 1. 屬性修改命令】運算符字面)。
func assignText(op cores.AssignKind) string {
	switch op {
	case cores.AssignSet:
		return "="

	case cores.AssignAdd:
		return "+="

	case cores.AssignSub:
		return "-="

	case cores.AssignMul:
		return "*="

	case cores.AssignDiv:
		return "/="

	case cores.AssignMod:
		return "%="

	case cores.AssignLock:
		return "@"

	case cores.AssignUnlock:
		return "#"

	default:
		return "?"
	} // switch
}

// containerText 容器名(與盤面各區標籤同名); None 只作去向出現(顧客離場), 顯 離場。
func containerText(kind cores.ContainerKind) string {
	switch kind {
	case cores.ContainerHand:
		return "手牌"

	case cores.ContainerDeck:
		return "抽牌堆"

	case cores.ContainerDrop:
		return "棄牌堆"

	case cores.ContainerExile:
		return "流放堆"

	case cores.ContainerWait:
		return "排隊"

	case cores.ContainerSeat:
		return "座位"

	case cores.ContainerRoam:
		return "遊蕩"

	case cores.ContainerCardify:
		return "卡牌化"

	case cores.ContainerNone:
		return "離場"

	default:
		return "?"
	} // switch
}

// taskText 行動類型全名(日誌用; 行動區的單字短形另見 panelAction 的 taskName)。
func taskText(task cores.TaskKind) string {
	switch task {
	case cores.TaskSate:
		return "飽食"

	case cores.TaskCalm:
		return "耐心"

	default:
		return "?"
	} // switch
}

// phaseName 階段中文全名(【營業顯示規格書 | 6、畫面規格 | 6.10】座標用全名); 無階段 / 未知顯 -。
func phaseName(phase cores.PhaseKind) string {
	switch phase {
	case cores.PhaseGameStart:
		return "營業開始"

	case cores.PhaseRoundStart:
		return "回合開始"

	case cores.PhasePlayerAction:
		return "玩家行動"

	case cores.PhaseGuestAction:
		return "顧客行動"

	case cores.PhaseRoundEnd:
		return "回合結束"

	case cores.PhaseGameSucc:
		return "營業成功"

	case cores.PhaseGameFail:
		return "營業失敗"

	default:
		return "-"
	} // switch
}
