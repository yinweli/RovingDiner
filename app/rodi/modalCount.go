package rodi

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// modalCount 計數 modal(【營業顯示規格書 | 7、互動規格 | 7.6 | 7.6.1】): 營業實例的完整檢視——
// 回合/階段 / 士氣 / 卡牌 屬性區(2 欄、col2 跨群組對齊)+ 回合計數(8 項 x 次數 / 最後對象)+
// 歷史計數(群組 x 抽/棄/出/流放 樞紐表); 僅【營業規格書 | 五、實例結構】stored 欄位,
// 衍生 / 查詢屬性歸各面板標題。標題列嵌 seed(同 seed + 同輸入 = 同一局, 重現對帳用)。
type modalCount struct {
	seed int64 // 本場 seed(標題列嵌入; 直讀 game.Seed 由呼叫端帶入, 顯示與 scrollback 同為十進位)
}

// Body 組內容行: 屬性配對先量跨群組 col1 最大寬再排(col2 對齊同一欄), 單身屬性佔整列;
// 兩張表各自 content-fit。
func (this modalCount) Body(game *cores.Game) (title string, row []string) {
	phase := cores.PhaseName(game.GetPhase())

	if next := game.GetNextPhase(); next != cores.PhaseNone {
		phase += " > " + cores.PhaseName(next)
	} // if

	pair := [][2]string{
		{"回合 " + valText(game.GetRound()), "回合上限 " + valText(game.GetRoundMax())},
		{"滿意值 " + valText(game.GetScore()), "階段 " + phase},
		{"士氣值 " + valText(game.GetMorale()), "士氣值上限 " + valText(game.GetMoraleMax())},
		{"士氣值護盾 " + valText(game.GetMoraleShield()), "士氣值格擋 " + valText(game.GetMoraleBlock())},
		{"手牌上限 " + valText(game.GetHandMax()), "補牌上限 " + valText(game.GetDrawMax())},
		{"出牌點數 " + valText(game.GetEnergy()), "出牌點數上限 " + valText(game.GetEnergyMax())},
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
		cell(1),
		modalGroup + "士氣",
		cell(2),
		cell(3),
		"士氣受損 " + num(game.GetDamageValue()) + " (" + refGuest(game, game.GetDamageGuest()) + ")",
		modalGroup + "卡牌",
		cell(4),
		cell(5),
		"出牌點數保留" + lockText(game.GetEnergyKeep()),
		modalGroup + "回合計數",
	}
	row = append(row, alignTable(countTable(game))...)
	row = append(row, modalGroup+"歷史計數")
	row = append(row, alignTable(historyTable(game))...)
	return fmt.Sprintf("計數檢視 (seed %v)", this.seed), row
}

// valText 數值實例顯示(值 + 鎖; 【營業顯示規格書 | 7、互動規格 | 7.6】通則): 鎖定計數 > 0 時值後緊接 [鎖N]。
func valText(value *cores.Value) string {
	text := num(value.GetValue())

	if value.IsLock() {
		text += "[鎖" + num(value.GetLock()) + "]"
	} // if

	return text
}

// lockText 純鎖屬性顯示(值恆 0; 【營業顯示規格書 | 7、互動規格 | 7.6】通則): 一律顯 [鎖N](含 [鎖0])。
func lockText(value *cores.Value) string {
	return "[鎖" + num(value.GetLock()) + "]"
}

// refGuest 顧客引用顯示(依【營業顯示規格書 | 5、識別碼】含實例段); 空物件顯 空。
func refGuest(game *cores.Game, guest *cores.Guest) string {
	if guest == nil {
		return "空"
	} // if

	return cores.IdentGuest(game.GetSheet(), guest.GetGuestID(), guest.GetInstanceID())
}

// refCard 卡牌引用顯示(依【營業顯示規格書 | 5、識別碼】含實例段); 空引用回空字串(尚無此項目, 欄空著)。
func refCard(game *cores.Game, card *cores.Card) string {
	if card == nil {
		return ""
	} // if

	return cores.IdentCard(game.GetSheet(), card.GetCardID(), card.GetInstanceID())
}

// countTable 回合計數表(表頭 + 8 項; 次數 = 當回計數, 最後對象空引用欄空著、附帶資訊以 (...) 接於識別碼後)。
func countTable(game *cores.Game) [][]string {
	seat := ""

	if last := game.GetSeatLast(); last != nil {
		seat = refGuest(game, last)
	} // if

	exit := ""

	if last := game.GetExitLast(); last != nil {
		exit = refGuest(game, last)

		if meta := game.GetSheet().Seat.Get(game.GetExitLastSeat()); meta != nil {
			exit += fmt.Sprintf(" (桌%v)", meta.TableID)
		} // if
	} // if

	task := ""

	if last := game.GetTaskGuest(); last != nil {
		task = refGuest(game, last) + " (" + cores.IdentSkill(game.GetSheet(), game.GetTaskSkill()) + ")"
	} // if

	morph := ""

	if last := game.GetMorphLast(); last != nil {
		morph = refCard(game, last) + fmt.Sprintf(" (%v->%v)", game.GetMorphOldID(), game.GetMorphNewID())
	} // if

	return [][]string{
		{"項目", "次數", "最後對象"},
		{"入座", num(game.GetSeatCount()), seat},
		{"離場", num(game.GetExitCount()), exit},
		{"行動", num(game.GetTaskCount()), task},
		{"抽牌", num(game.GetDrawCount()), refCard(game, game.GetDrawLast())},
		{"棄牌", num(game.GetDropCount()), refCard(game, game.GetDropLast())},
		{"出牌", num(game.GetPlayCount()), refCard(game, game.GetPlayLast())},
		{"流放", num(game.GetExileCount()), refCard(game, game.GetExileLast())},
		{"變身", num(game.GetMorphCount()), morph},
	}
}

// historyTable 歷史計數樞紐表(表頭 + 群組列; 列 = 四動作累積出現過的群組聯集升序、欄 = 抽/棄/出/流放,
// 無群組仍顯表頭——恆顯, 通則)。
func historyTable(game *cores.Game) [][]string {
	merge := map[int32]bool{}

	for _, tally := range []*cores.Tally{game.GetDrawTotal(), game.GetDropTotal(), game.GetPlayTotal(), game.GetExileTotal()} {
		for _, itor := range tally.Group() {
			merge[itor] = true
		} // for
	} // for

	group := []int32{}

	for k := range merge {
		group = append(group, k)
	} // for

	sort.Slice(group, func(i, j int) bool { return group[i] < group[j] })
	result := [][]string{{"群組", "抽牌張數", "棄牌張數", "出牌張數", "流放張數"}}

	for _, itor := range group {
		result = append(result, []string{
			num(itor),
			num(game.GetDrawTotal().Get(itor)),
			num(game.GetDropTotal().Get(itor)),
			num(game.GetPlayTotal().Get(itor)),
			num(game.GetExileTotal().Get(itor)),
		})
	} // for

	return result
}
