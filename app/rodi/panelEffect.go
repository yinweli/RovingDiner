package rodi

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// panelEffect 效果佇列組件(區 4; 【營業顯示規格書 | 6、畫面規格 | 6.6】): 依作用順序排序(大者優先、
// 同序效果編號小者優先, 同【營業規格書 | 十五、作用順序】; 排序用本地複本, 不動引擎佇列);
// 每項 2 行垂直區塊橫向並排——第 1 行 效果識別碼xStack (剩餘回合)(層數 1 省 xN、整場保留顯 永),
// 第 2 行 self 識別碼 + 類型(觸發 / 常駐); 欄寬 content-fit、欄距 2; 超寬固定窗截斷補右緣 >(M21 拍板⑦)。
type panelEffect struct{}

// View 渲染標題列(含佇列數) + 2 行; 空佇列兩行留白(高度穩定)。
func (this panelEffect) View(game *cores.Game, width int) string {
	sorted := append([]*cores.Effect{}, game.Effect...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left, right := effectOrder(game.GetSheet(), sorted[i].GetEffectID()), effectOrder(game.GetSheet(), sorted[j].GetEffectID())

		if left != right {
			return left > right
		} // if

		return sorted[i].GetEffectID() < sorted[j].GetEffectID()
	})

	row1 := []string{}
	row2 := []string{}

	for _, itor := range sorted {
		text1 := identEffect(game.GetSheet(), itor.GetEffectID(), cores.NoneID)

		if itor.GetStack() > 1 {
			text1 += "x" + num(itor.GetStack())
		} // if

		if itor.GetExpire() > 0 {
			text1 += " (" + num(itor.GetExpire()-game.GetRound().GetValue()) + ")"
		} else {
			text1 += " (永)"
		} // if

		text2 := effectSelf(game.GetSheet(), itor) + " " + effectKindName(game.GetSheet(), itor.GetEffectID())
		size := max(lipgloss.Width(text1), lipgloss.Width(text2))
		row1 = append(row1, padTo(text1, size))
		row2 = append(row2, padTo(text2, size))
	} // for

	return strings.Join([]string{
		panelTitle(fmt.Sprintf("效果佇列(%v)", len(game.Effect)), width),
		truncMark(strings.TrimRight(strings.Join(row1, "  "), " "), width),
		truncTo(strings.TrimRight(strings.Join(row2, "  "), " "), width),
	}, "\n")
}

// effectOrder 查效果作用順序(查無回 0, 防禦; 與 EffectList.Sort 同源規則)。
func effectOrder(sheet *sheeter.Sheeter, effectID int32) int32 {
	meta := sheet.Effect.Get(effectID)

	if meta == nil {
		return 0
	} // if

	return meta.RunOrder
}

// effectSelf 佇列項 self 的主畫面摘要: Ref 自帶辨型(卡牌 / 顧客); 空物件顯「空」。
func effectSelf(sheet *sheeter.Sheeter, effect *cores.Effect) string {
	self := effect.GetSelf()

	if guest := self.GetGuest(); guest != nil {
		return identGuest(sheet, guest.GetGuestID(), cores.NoneID)
	} // if

	if card := self.GetCard(); card != nil {
		return identCard(sheet, card.GetCardID(), cores.NoneID)
	} // if

	return "空"
}

// effectKindName 效果類型標記(觸發 / 常駐; 立即不入列, 查無 / 越界顯 ?)。
func effectKindName(sheet *sheeter.Sheeter, effectID int32) string {
	meta := sheet.Effect.Get(effectID)

	if meta == nil {
		return "?"
	} // if

	switch cores.EffectKind(meta.Kind) {
	case cores.EffectTrigger:
		return "觸發"

	case cores.EffectPersist:
		return "常駐"

	default:
		return "?"
	} // switch
}
