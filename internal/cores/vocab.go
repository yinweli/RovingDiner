package cores

import (
	"strconv"
)

// 詞彙對照(日誌中文化; 【營業顯示規格書 | 6、畫面規格 | 6.10】): 英文鍵 / 編號 / 列舉轉中文名。
// 名稱依【營業規格書 | 二、英文詞彙對照】與【營業規格書 | 二十三、屬性清單】;
// 查無對照退回原文(寬鬆, 比照識別碼查無顯 ?)。
// 日誌行合成(發射端)與盤面組件(顯示端)共用同一份, 故居 cores。

// 免疫詞條鍵(群組維度的屬性寫入: 運算值載群組、結果載該群組計數; M22 拍板): 詞彙對照 / 日誌行文特判共用。
const (
	AttrEffectImmune = "effectImmune"
	AttrSkillImmune  = "skillImmune"
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
// 同名鍵 morale / moraleMax / score / scoreMax 於此表取顧客義, 與全域表以對象有無分流)。
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
	AttrEffectImmune: "效果免疫群組",
	AttrSkillImmune:  "技能免疫群組",
}

// triggerName 觸發時機英文鍵 → 中文名(【營業規格書 | 二十二、觸發時機清單】)。
var triggerName = map[TriggerKind]string{
	TriggerGameStart:  "營業開始",
	TriggerRoundReady: "回合準備",
	TriggerRoundStart: "回合開始",
	TriggerGuestSeat:  "顧客入座",
	TriggerCardDraw:   "手牌補充",
	TriggerUserStart:  "玩家開始",
	TriggerUserEnd:    "玩家結束",
	TriggerCardDrop:   "卡牌棄置",
	TriggerCardExile:  "卡牌流放",
	TriggerCardMorph:  "卡牌變身",
	TriggerCardPlay:   "玩家出牌",
	TriggerGuestStart: "顧客開始",
	TriggerGuestTask:  "顧客行動",
	TriggerGuestEnd:   "顧客結束",
	TriggerRoundEnd:   "回合結束",
	TriggerExitAny:    "顧客離場",
	TriggerExitSate:   "飽食離場",
	TriggerExitCalm:   "生氣離場",
	TriggerExitDone:   "顧客離場後",
	TriggerDamage:     "士氣受損",
	TriggerGameSucc:   "營業成功",
	TriggerGameFail:   "營業失敗",
}

// sourceName 選取來源英文鍵 → 中文名(Pick 詞條鍵依【營業規格書 | 二、英文詞彙對照】主體 + 動詞組合,
// discardOver 為手牌上限棄牌流程名); 效果目標選取改顯效果識別碼、不查本表。
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

// AttrText 屬性詞條鍵轉中文名: 全域屬性走全域表、實例對象走引用表; 查無退回原文。
func AttrText(attr string, global bool) string {
	table := attrRefName

	if global {
		table = attrGlobalName
	} // if

	if name, ok := table[attr]; ok {
		return name
	} // if

	return attr
}

// TriggerText 觸發時機中文名; 查無退回原文。
func TriggerText(trigger TriggerKind) string {
	if name, ok := triggerName[trigger]; ok {
		return name
	} // if

	return string(trigger)
}

// SourceText 選取來源中文名; 查無退回原文。
func SourceText(source string) string {
	if name, ok := sourceName[source]; ok {
		return name
	} // if

	return source
}

// StageText 效果階段名(【營業顯示規格書 | 6、畫面規格 | 6.10】效果欄位的階段值)。
func StageText(stage EffectStage) string {
	switch stage {
	case EffectStageImmed:
		return "立即"

	case EffectStageJoin:
		return "加入"

	case EffectStageTrigger:
		return "觸發"

	case EffectStageStart:
		return "啟動"

	case EffectStageEnd:
		return "結束"

	case EffectStageCondFail:
		return "條件不成立"

	default:
		return "?"
	} // switch
}

// AssignText 賦值符(【營業規格書 | 十七、命令 | 1. 屬性修改命令】運算符字面)。
func AssignText(op AssignKind) string {
	switch op {
	case AssignSet:
		return "="

	case AssignAdd:
		return "+="

	case AssignSub:
		return "-="

	case AssignMul:
		return "*="

	case AssignDiv:
		return "/="

	case AssignMod:
		return "%="

	case AssignLock:
		return "@"

	case AssignUnlock:
		return "#"

	default:
		return "?"
	} // switch
}

// ContainerText 容器名(與盤面各區標籤同名); None 只作去向出現(顧客離場), 顯 離場。
func ContainerText(kind ContainerKind) string {
	switch kind {
	case ContainerHand:
		return "手牌"

	case ContainerDeck:
		return "抽牌堆"

	case ContainerDrop:
		return "棄牌堆"

	case ContainerExile:
		return "流放堆"

	case ContainerWait:
		return "排隊"

	case ContainerSeat:
		return "座位"

	case ContainerRoam:
		return "遊蕩"

	case ContainerCardify:
		return "卡牌化"

	case ContainerNone:
		return "離場"

	default:
		return "?"
	} // switch
}

// TaskText 行動類型全名(日誌用; 行動區的單字短形另見 panelAction 的 taskName)。
func TaskText(task TaskKind) string {
	switch task {
	case TaskSate:
		return "飽食"

	case TaskCalm:
		return "耐心"

	default:
		return "?"
	} // switch
}

// PhaseName 階段中文全名(【營業顯示規格書 | 6、畫面規格 | 6.10】座標用全名); 無階段 / 未知顯 -。
func PhaseName(phase PhaseKind) string {
	switch phase {
	case PhaseGameStart:
		return "營業開始"

	case PhaseRoundStart:
		return "回合開始"

	case PhasePlayerAction:
		return "玩家行動"

	case PhaseGuestAction:
		return "顧客行動"

	case PhaseRoundEnd:
		return "回合結束"

	case PhaseGameSucc:
		return "營業成功"

	case PhaseGameFail:
		return "營業失敗"

	default:
		return "-"
	} // switch
}

// NumText 數值轉日誌字串(整數去小數位、運算值可為小數)。
func NumText(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
