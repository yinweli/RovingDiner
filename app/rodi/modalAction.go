package rodi

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// modalAction 行動 modal(【營業顯示規格書 | 7、互動規格 | 7.6 | 7.6.3】): 行動實例僅 3 欄(單欄逐列,
// 偏離 2 欄通則)+ 技能引用的效果以 §5 識別碼橫排(靜態引用無實例段, 不展開完整欄位——歸效果 modal);
// 顧客引用失效(已離場)附 (已離場) 標記。
type modalAction struct {
	action *cores.Action // 檢視對象(實例指標; 開著時引擎暫停, 指標穩定)
}

// Body 組內容行: 行動三欄 + 效果橫排(超寬 wrap 成多列、空列表留空列——區塊恆顯)。
func (this modalAction) Body(game *cores.Game) (title string, row []string) {
	guest := refGuest(game, this.action.GetGuest())

	if guestGone(game, this.action.GetGuest()) {
		guest += " (已離場)"
	} // if

	row = []string{
		"顧客 " + guest,
		"行動類型 " + taskText(this.action.GetKind()),
		"技能 " + cores.IdentSkill(game.GetSheet(), this.action.GetSkillID()),
		modalGroup + "效果",
	}
	token := []string{}

	if meta := game.GetSheet().Skill.Get(this.action.GetSkillID()); meta != nil {
		for _, itor := range meta.EffectID { // 多重集合: 同編號重複引用就重複列出, 不合併
			token = append(token, cores.IdentEffect(game.GetSheet(), itor, cores.NoneID))
		} // for
	} // if

	return "行動檢視", append(row, wrapToken(token)...)
}

// taskText 行動類型全名(飽食 / 耐心; 主畫面單字標記歸 taskName, 越界顯 ?)。
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

// guestGone 回報顧客引用是否已失效(已離場 = 不在座位 / 排隊 / 遊蕩 / 卡牌化任一容器)。
func guestGone(game *cores.Game, guest *cores.Guest) bool {
	if guest == nil {
		return true
	} // if

	for _, itor := range game.Seat {
		if itor == guest {
			return false
		} // if
	} // for

	for _, member := range [][]*cores.Guest{game.Wait, game.Roam, game.Cardify} {
		for _, itor := range member {
			if itor == guest {
				return false
			} // if
		} // for
	} // for

	return true
}
