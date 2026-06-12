package rodi

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// textTrigger 「觸發」字面(效果類型名 / 命令標籤 / 群組名 / 門檻表頭共用同一 SSOT 中文)。
const textTrigger = "觸發"

// modalEffect 效果 modal(【營業顯示規格書 | 7、互動規格 | 7.6 | 7.6.4】): 效果實例動態欄 + 該編號的
// 效果資料靜態 def(實例欄不足以 debug, 補入——通則); 恆顯留空: 依【營業規格書 | 四、表格結構】
// 適用類型矩陣, 當前類型不適用的欄保留位置、值空著(版面不隨類型變); 命令字串顯 SSOT 原文。
type modalEffect struct {
	effect *cores.Effect // 檢視對象(實例指標; 開著時引擎暫停, 指標穩定)
}

// Body 組內容行: 實例 / 規格 / 觸發配對先量跨群組 col1 最大寬再排; self / 運算式 / 命令各吃整列、
// 超長 wrap 成多列。
func (this modalEffect) Body(game *cores.Game) (title string, row []string) {
	meta := game.GetSheet().Effect.Get(this.effect.GetEffectID())

	if meta == nil {
		meta = &sheeter.Effect{} // 查無(防禦): 靜態欄全零值照排, 實例欄照顯
	} // if

	kind := cores.EffectKind(meta.Kind)
	queue := kind == cores.EffectTrigger || kind == cores.EffectPersist // 佇列族(觸發 / 常駐)共用的欄
	cond := kind == cores.EffectImmed || kind == cores.EffectTrigger    // 條件族(立即 / 觸發)共用的欄
	pair := [][2]string{
		{labelRow("結束回合", expireText(this.effect)), labelRow("當前層數", num(this.effect.GetStack()))},
		{labelRow("效果類型", kindText(kind)), labelRow("效果群組", onText(queue, num(meta.Group)))},
		{labelRow("目標類型", targetText(cores.TargetKind(meta.TargetKind))), labelRow("目標數量", num(meta.TargetCount))},
		{labelRow("作用回合", onText(queue, num(meta.RunRound))), labelRow("作用順序", num(meta.RunOrder))},
		{labelRow("堆疊層數", onText(queue, num(meta.Stack))), labelRow("堆疊上限", onText(queue, num(meta.StackMax)))},
		{labelRow("觸發時機", onText(kind == cores.EffectTrigger, cores.TriggerText(cores.TriggerKind(meta.TriggerKind)))), labelRow("觸發後行為", onText(kind == cores.EffectTrigger, afterText(cores.TriggerAfter(meta.TriggerAfter))))},
	}
	col1 := 0

	for _, itor := range pair {
		if w := lipgloss.Width(itor[0]); w > col1 {
			col1 = w
		} // if
	} // for

	cell := func(index int) string {
		return padTo(pair[index][0], col1) + "  " + pair[index][1]
	}
	row = []string{
		cell(0),
		"self " + refTarget(game, this.effect),
		modalGroup + "規格",
		cell(1),
		cell(2),
		cell(3),
		cell(4),
		labelRow("堆疊時間", onText(queue, stackText(cores.StackTime(meta.StackTime)))),
		modalGroup + "觸發",
		cell(5),
	}
	row = append(row, wrapText(labelRow("觸發條件", onText(cond, meta.TriggerCond)))...)
	row = append(row, wrapText(labelRow("觸發次數", onText(cond, meta.TriggerCount)))...)
	row = append(row, modalGroup+"命令")
	row = append(row, wrapText(labelRow("立即", onText(kind == cores.EffectImmed, meta.CommandImmed)))...)
	row = append(row, wrapText(labelRow(textTrigger, onText(kind == cores.EffectTrigger, meta.CommandTrigger)))...)
	row = append(row, wrapText(labelRow("啟動", onText(kind == cores.EffectPersist, meta.CommandStart)))...)
	row = append(row, wrapText(labelRow("結束", onText(queue, meta.CommandEnd)))...)
	return "效果檢視 " + cores.IdentEffect(game.GetSheet(), this.effect.GetEffectID(), this.effect.GetInstanceID()), row
}

// onText 恆顯留空原語: 欄位適用當前類型才顯值、不適用回空字串(保留位置不畫記號;
// 依【營業規格書 | 四、表格結構】適用類型矩陣)。
func onText(apply bool, text string) string {
	if apply == false {
		return ""
	} // if

	return text
}

// labelRow 單欄逐列行: 標籤 + 值(值空時只留標籤, 行尾不留空白)。
func labelRow(label, value string) string {
	if value == "" {
		return label
	} // if

	return label + " " + value
}

// refTarget self 物件顯示(依 §5 含實例段): 空物件顯 空; 已失效(顧客離場 / 卡牌移除)附 (已失效)。
func refTarget(game *cores.Game, effect *cores.Effect) string {
	if guest := effect.GetSelf().GetGuest(); guest != nil {
		text := cores.IdentGuest(game.GetSheet(), guest.GetGuestID(), guest.GetInstanceID())

		if guestGone(game, guest) {
			text += " (已失效)"
		} // if

		return text
	} // if

	if card := effect.GetSelf().GetCard(); card != nil {
		text := cores.IdentCard(game.GetSheet(), card.GetCardID(), card.GetInstanceID())

		if cardGone(game, card) {
			text += " (已失效)"
		} // if

		return text
	} // if

	return "空"
}

// cardGone 回報卡牌引用是否已失效(已移除 = 不在手牌 / 抽 / 棄 / 流放任一容器)。
func cardGone(game *cores.Game, card *cores.Card) bool {
	for _, member := range [][]*cores.Card{game.Hand, game.Deck, game.Drop, game.Exile} {
		for _, itor := range member {
			if itor == card {
				return false
			} // if
		} // for
	} // for

	return true
}

// kindText 效果類型中文名(立即 / 觸發 / 常駐; 越界顯 ?)。
func kindText(kind cores.EffectKind) string {
	switch kind {
	case cores.EffectImmed:
		return "立即"

	case cores.EffectTrigger:
		return textTrigger

	case cores.EffectPersist:
		return "常駐"

	default:
		return "?"
	} // switch
}

// targetText 目標類型中文名(【營業規格書 | 八、目標類型】; 越界顯 ?)。
func targetText(target cores.TargetKind) string {
	switch target {
	case cores.TargetNone:
		return "無目標"

	case cores.TargetGuestPick:
		return "新選顧客"

	case cores.TargetGuestSame:
		return "沿用顧客"

	case cores.TargetGuestRand:
		return "隨機顧客"

	case cores.TargetCardPick:
		return "新選手牌"

	case cores.TargetCardSame:
		return "沿用手牌"

	case cores.TargetCardRand:
		return "隨機手牌"

	default:
		return "?"
	} // switch
}

// afterText 觸發後行為中文名(保留 / 移除; 越界顯 ?)。
func afterText(after cores.TriggerAfter) string {
	switch after {
	case cores.TriggerAfterKeep:
		return "保留"

	case cores.TriggerAfterRemove:
		return "移除"

	default:
		return "?"
	} // switch
}

// stackText 堆疊時間中文名(不變 / 刷新; 越界顯 ?)。
func stackText(stack cores.StackTime) string {
	switch stack {
	case cores.StackTimeStay:
		return "不變"

	case cores.StackTimeRefresh:
		return "刷新"

	default:
		return "?"
	} // switch
}
