package rodi

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// modalGuest 顧客 modal(【營業顯示規格書 | 7、互動規格 | 7.6 | 7.6.2】): 顧客實例的完整檢視——
// 顧客身分屬性區(2 欄)+ 飽食 / 耐心技能(門檻配對 x 是否已觸發)+ 免疫(類型 x 群組 x 計數)+
// self active 效果列表; 除門檻技能配置外其餘靜態定義不進 modal。標題列嵌 §5 識別碼(含實例段)。
type modalGuest struct {
	guest *cores.Guest // 檢視對象(實例指標; 開著時引擎暫停, 指標穩定)
}

// Body 組內容行: 身分配對先量 col1 最大寬再排, 三張表各自 content-fit; 技能列序 = 飽食門檻低到高、
// 耐心門檻高到低(對齊執行結算的觸發序; GuestData 已照此排序)。
func (this modalGuest) Body(game *cores.Game) (title string, row []string) {
	pair := [][2]string{
		{"滿意值 " + valText(this.guest.GetScore()) + "/" + valText(this.guest.GetScoreMax()), "士氣值 " + valText(this.guest.GetMorale()) + "/" + valText(this.guest.GetMoraleMax())},
		{"飽食值 " + valText(this.guest.GetSate()) + "/" + valText(this.guest.GetSateMax()), "封印飽食技能" + lockText(this.guest.GetSateSeal())},
		{"耐心值 " + valText(this.guest.GetCalm()), "封印耐心技能" + lockText(this.guest.GetCalmSeal())},
		{seatText(game, this.guest), "凍結起始回合 " + num(this.guest.GetFreeze())},
	}
	col1 := 0

	for _, itor := range pair {
		if w := lipgloss.Width(itor[0]); w > col1 {
			col1 = w
		} // if
	} // for

	for _, itor := range pair {
		row = append(row, padTo(itor[0], col1)+"  "+itor[1])
	} // for

	meta, _ := game.GuestData(this.guest.GetGuestID()) // 查無即無門檻(表恆顯、列留空)
	row = append(row, modalGroup+"飽食技能")
	row = append(row, alignTable(hitTable(game, meta.Sate, this.guest.GetSateHit()))...)
	row = append(row, modalGroup+"耐心技能")
	row = append(row, alignTable(hitTable(game, meta.Calm, this.guest.GetCalmHit()))...)
	row = append(row, modalGroup+"免疫")
	row = append(row, alignTable(immuneTable(this.guest))...)
	row = append(row, modalGroup+"效果")
	row = append(row, alignTable(effectTable(game, this.guest))...)
	return "顧客檢視 " + cores.IdentGuest(game.GetSheet(), this.guest.GetGuestID(), this.guest.GetInstanceID()), row
}

// seatText 位置顯示: 在座顯 座位 N (桌M); SeatID = 0 依所在容器標 排隊 / 遊蕩 / 卡牌化,
// 都不在(防禦)即已離場。
func seatText(game *cores.Game, guest *cores.Guest) string {
	if seatID := guest.GetSeatID(); seatID > 0 {
		text := fmt.Sprintf("座位 %v", seatID)

		if meta := game.GetSheet().Seat.Get(seatID); meta != nil {
			text += fmt.Sprintf(" (桌%v)", meta.TableID)
		} // if

		return text
	} // if

	for _, itor := range game.Wait {
		if itor == guest {
			return "座位 排隊"
		} // if
	} // for

	for _, itor := range game.Roam {
		if itor == guest {
			return "座位 遊蕩"
		} // if
	} // for

	for _, itor := range game.Cardify {
		if itor == guest {
			return "座位 卡牌化"
		} // if
	} // for

	return "座位 離場"
}

// hitTable 門檻技能表(表頭 + 配對列: 門檻值 / 技能 / 觸發——該門檻已觸發標 v、否則留空)。
func hitTable(game *cores.Game, threshold []cores.Threshold, hit *cores.Hit) [][]string {
	result := [][]string{{"門檻值", "技能", textTrigger}}

	for _, itor := range threshold {
		mark := ""

		if hit.IsHit(itor.Value) {
			mark = "v"
		} // if

		result = append(result, []string{num(itor.Value), cores.IdentSkill(game.GetSheet(), itor.SkillID), mark})
	} // for

	return result
}

// immuneTable 免疫表(表頭 + 效果 / 技能兩類的鎖定中群組列; 兩類皆空仍顯表頭——恆顯, 通則)。
func immuneTable(guest *cores.Guest) [][]string {
	result := [][]string{{"類型", "群組", "計數"}}

	for _, itor := range guest.GetEffectImmune().Group() {
		result = append(result, []string{"效果", num(itor), num(guest.GetEffectImmune().Get(itor))})
	} // for

	for _, itor := range guest.GetSkillImmune().Group() {
		result = append(result, []string{"技能", num(itor), num(guest.GetSkillImmune().Get(itor))})
	} // for

	return result
}

// effectTable self active 效果表(表頭 + self == 本顧客的佇列項; 結束回合 0 = 整場保留顯 永)。
func effectTable(game *cores.Game, guest *cores.Guest) [][]string {
	result := [][]string{{"效果", "結束回合", "當前層數"}}

	for _, itor := range game.Effect {
		if itor.GetSelf().GetGuest() != guest {
			continue
		} // if

		result = append(result, []string{
			cores.IdentEffect(game.GetSheet(), itor.GetEffectID(), itor.GetInstanceID()),
			expireText(itor),
			num(itor.GetStack()),
		})
	} // for

	return result
}

// expireText 結束回合顯示: Expire = 0(整場保留)顯 永、其餘顯原值。
func expireText(effect *cores.Effect) string {
	if effect.GetExpire() == 0 {
		return "永"
	} // if

	return num(effect.GetExpire())
}
