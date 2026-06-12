package rodi

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// modalCard 卡牌 modal(【營業顯示規格書 | 7、互動規格 | 7.6 | 7.6.5】): 卡牌實例完整檢視——
// 規格(含靜態補的卡牌群組 / 技能編號)+ 旗標(4 純鎖 2 欄)+ 實例效果列表(每效果一組摘要列,
// 多重集合重複列出); 牌堆項目同用本 modal(單張卡)。標題列嵌 §5 識別碼(含實例段)。
type modalCard struct {
	card *cores.Card // 檢視對象(實例指標; 開著時引擎暫停, 指標穩定)
}

// Body 組內容行: 規格 / 旗標配對先量跨群組 col1 最大寬再排; 卡牌技能 / 卡牌化來源各吃整列
// (非 cardify 卡空引用留空); 效果每項一組摘要列(標頭 + 縮排命令)。
func (this modalCard) Body(game *cores.Game) (title string, row []string) {
	meta := game.GetSheet().Card.Get(this.card.GetCardID())

	if meta == nil {
		meta = &sheeter.Card{} // 查無(防禦): 靜態欄全零值照排, 實例欄照顯
	} // if

	pair := [][2]string{
		{"卡牌群組 " + num(meta.Group), "出牌費用 " + valText(this.card.GetCost())},
		{"額外發動次數下限 " + valText(this.card.GetExtraRunMin()), "額外發動次數上限 " + valText(this.card.GetExtraRunMax())},
		{"不棄卡牌" + lockText(this.card.GetKeep()), "封印卡牌" + lockText(this.card.GetSeal())},
		{"出牌後流放" + lockText(this.card.GetPlayExile()), "未出牌流放" + lockText(this.card.GetUnplayExile())},
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
	cardify := "" // 非 cardify 卡空引用留空(通則; 與計數 modal 的顯 空 規則不同)

	if bind := this.card.GetCardify(); bind != nil {
		cardify = cores.IdentGuest(game.GetSheet(), bind.GetGuestID(), bind.GetInstanceID())
	} // if

	row = []string{
		cell(0),
		cell(1),
		"卡牌技能 " + cores.IdentSkill(game.GetSheet(), meta.SkillID),
		labelRow("卡牌化來源", cardify),
		modalGroup + "旗標",
		cell(2),
		cell(3),
		modalGroup + "效果",
	}
	title = "卡牌檢視 " + cores.IdentCard(game.GetSheet(), this.card.GetCardID(), this.card.GetInstanceID())
	effectID := this.card.GetEffectID().List()

	if len(effectID) == 0 {
		return title, append(row, "") // 空列表: 區塊恆顯、列留空
	} // if

	for _, itor := range effectID { // 多重集合: 同編號重複出現就重複列出, 不合併
		row = append(row, effectRow(game, itor)...)
	} // for

	return title, row
}

// effectRow 單一效果的摘要列(【營業顯示規格書 | 7、互動規格 | 7.6 | 7.6.5】): 卡牌列的效果是表編號
// 靜態引用、場上無實例可開檢視, 直附行為摘要——標頭列 = §5 識別碼 + [類型](觸發類型附時機原文);
// 其下縮排命令列「標籤: 命令原文」, 依適用類型矩陣取欄、空命令略列、超寬 wrap 成多列。
func effectRow(game *cores.Game, effectID int32) (result []string) {
	meta := game.GetSheet().Effect.Get(effectID)

	if meta == nil {
		meta = &sheeter.Effect{} // 查無(防禦): 靜態欄全零值照排
	} // if

	kind := cores.EffectKind(meta.Kind)
	label := kindText(kind)

	if kind == cores.EffectTrigger && meta.TriggerKind != "" {
		label += " " + meta.TriggerKind
	} // if

	result = append(result, cores.IdentEffect(game.GetSheet(), effectID, cores.NoneID)+" ["+label+"]")

	for _, itor := range [][2]string{
		{"命令", onText(kind == cores.EffectImmed, meta.CommandImmed)},
		{"命令", onText(kind == cores.EffectTrigger, meta.CommandTrigger)},
		{"啟動", onText(kind == cores.EffectPersist, meta.CommandStart)},
		{"結束", onText(kind == cores.EffectTrigger || kind == cores.EffectPersist, meta.CommandEnd)},
	} {
		if itor[1] == "" {
			continue // 不適用類型 / 空命令 → 略列
		} // if

		result = append(result, wrapText("  "+itor[0]+": "+itor[1])...)
	} // for

	return result
}
