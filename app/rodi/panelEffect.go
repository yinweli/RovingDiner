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
// 第 2 行 self 識別碼 + 類型(觸發 / 常駐); 欄寬 content-fit、欄距 2。游標單列左右移(M26 R2, 索引對排序後
// 順序), 游標態反白整個項目區塊(2 行); 窗格跟游標捲、左緣 < 兩行同縮排、超寬補右緣 >。
type panelEffect struct {
	cursor int // 游標索引(對排序後順序; 自持 UI 狀態、讀取時夾界)
}

// View 渲染標題列(含佇列數) + 2 行; 空佇列兩行留白(高度穩定)。
func (this *panelEffect) View(game *cores.Game, width int, focus bool) string {
	sorted := effectSorted(game)
	cursor := clampIndex(this.cursor, len(sorted))
	row1 := []string{}
	row2 := []string{}
	size := []int{}

	for index, itor := range sorted {
		text1 := cores.IdentEffect(game.GetSheet(), itor.GetEffectID(), cores.NoneID)

		if itor.GetStack() > 1 {
			text1 += "x" + num(itor.GetStack())
		} // if

		if itor.GetExpire() > 0 {
			text1 += " (" + num(itor.GetExpire()-game.GetRound().GetValue()) + ")"
		} else {
			text1 += " (永)"
		} // if

		text2 := effectSelf(game.GetSheet(), itor) + " " + effectKindName(game.GetSheet(), itor.GetEffectID())
		w := max(lipgloss.Width(text1), lipgloss.Width(text2))
		cell1 := padTo(text1, w)
		cell2 := padTo(text2, w)

		if focus && index == cursor {
			cell1 = restyle(&styleCursor, cell1)
			cell2 = restyle(&styleCursor, cell2)
		} // if

		row1 = append(row1, cell1)
		row2 = append(row2, cell2)
		size = append(size, w)
	} // for

	first := stripFirst(size, 2, width-4, cursor)
	head1, head2 := "", ""

	if first > 0 {
		head1, head2 = markHead, markIndent // 左緣記號佔位: 兩行同縮排, 項目區塊上下對齊
	} // if

	return strings.Join([]string{
		panelTitle(fmt.Sprintf("效果佇列(%v)", len(game.Effect)), width),
		boxMark(head1+strings.Join(row1[first:], "  "), width),
		boxTrunc(head2+strings.Join(row2[first:], "  "), width),
	}, "\n")
}

// Move 游標移動: 單列左右(【營業顯示規格書 | 7、互動規格 | 7.2】); 夾界不迴繞。
func (this *panelEffect) Move(game *cores.Game, key string) {
	this.cursor = moveIndex(this.cursor, key, len(game.Effect))
}

// Item 回游標下的效果項(對排序後順序, 與 View 同序; 空佇列回 nil; 檢視 modal 用)。
func (this *panelEffect) Item(game *cores.Game) any {
	sorted := effectSorted(game)

	if len(sorted) == 0 {
		return nil
	} // if

	return sorted[clampIndex(this.cursor, len(sorted))]
}

// effectSorted 排序後的效果佇列複本(作用順序大者優先、同序效果編號小者優先, 不動引擎佇列;
// View 與 Item 共用同一序, 游標所視即所開)。
func effectSorted(game *cores.Game) []*cores.Effect {
	sorted := append([]*cores.Effect{}, game.Effect...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left, right := effectOrder(game.GetSheet(), sorted[i].GetEffectID()), effectOrder(game.GetSheet(), sorted[j].GetEffectID())

		if left != right {
			return left > right
		} // if

		return sorted[i].GetEffectID() < sorted[j].GetEffectID()
	})
	return sorted
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
		return cores.IdentGuest(sheet, guest.GetGuestID(), cores.NoneID)
	} // if

	if card := self.GetCard(); card != nil {
		return cores.IdentCard(sheet, card.GetCardID(), cores.NoneID)
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
		return textTrigger

	case cores.EffectPersist:
		return "常駐"

	default:
		return "?"
	} // switch
}
