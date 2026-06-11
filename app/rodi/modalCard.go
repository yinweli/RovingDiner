package rodi

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// modalCard 卡牌 modal(【營業顯示規格書 | 7、互動規格 | 7.6 | 7.6.5】): 卡牌實例完整檢視——
// 規格(含靜態補的卡牌群組 / 技能編號)+ 旗標(4 純鎖 2 欄)+ 實例效果列表(§5 識別碼橫排,
// 多重集合重複列出); 牌堆項目同用本 modal(單張卡)。標題列嵌 §5 識別碼(含實例段)。
type modalCard struct {
	card *cores.Card // 檢視對象(實例指標; 開著時引擎暫停, 指標穩定)
}

// Body 組內容行: 規格 / 旗標配對先量跨群組 col1 最大寬再排; 卡牌技能 / 卡牌化來源各吃整列
// (非 cardify 卡空引用留空); 效果橫排超寬 wrap 成多列。
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
	token := []string{}

	for _, itor := range this.card.GetEffectID().List() { // 多重集合: 同編號重複出現就重複列出, 不合併
		token = append(token, cores.IdentEffect(game.GetSheet(), itor, cores.NoneID))
	} // for

	return "卡牌檢視 " + cores.IdentCard(game.GetSheet(), this.card.GetCardID(), this.card.GetInstanceID()), append(row, wrapToken(token)...)
}
