package rodi

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// effectPanel 效果佇列組件(區 4; 【營業顯示規格書 | 6、畫面規格 | 6.6】): 依作用順序排序(大者優先、
// 同序效果編號小者優先, 同【營業規格書 | 十五、作用順序】); 每項 2 行垂直區塊橫向並排——
// 第 1 行 效果識別碼xStack (剩餘回合)(層數 1 省 xN、整場保留顯 永), 第 2 行 self 識別碼 + 類型(觸發 / 常駐);
// 欄寬 content-fit、欄距 2; 超寬固定窗截斷補右緣 >(M21 拍板⑦)。
type effectPanel struct{}

// View 渲染標題列(含佇列數) + 2 行; 空佇列兩行留白(高度穩定)。
func (this effectPanel) View(world *mirror, width int) string {
	sorted := append([]*effectView{}, world.effect...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left, right := effectOrder(world.sheet, sorted[i].effectID), effectOrder(world.sheet, sorted[j].effectID)

		if left != right {
			return left > right
		} // if

		return sorted[i].effectID < sorted[j].effectID
	})

	row1 := []string{}
	row2 := []string{}

	for _, itor := range sorted {
		text1 := identEffect(world.sheet, itor.effectID, cores.NoneID)

		if itor.stack > 1 {
			text1 += "x" + num(float64(itor.stack))
		} // if

		if itor.expire > 0 {
			text1 += " (" + num(float64(itor.expire-world.round)) + ")"
		} else {
			text1 += " (永)"
		} // if

		text2 := effectSelf(world, itor) + " " + effectKindName(world.sheet, itor.effectID)
		size := max(lipgloss.Width(text1), lipgloss.Width(text2))
		row1 = append(row1, padTo(text1, size))
		row2 = append(row2, padTo(text2, size))
	} // for

	return strings.Join([]string{
		panelTitle(fmt.Sprintf("效果佇列(%v)", len(world.effect)), width),
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

// effectSelf 佇列項 self 的主畫面投影: 依鏡像視圖辨型(卡牌 / 顧客); 空物件或實例已不在鏡像顯「空」。
func effectSelf(world *mirror, view *effectView) string {
	if view.selfInstanceID == cores.NoneID {
		return "空"
	} // if

	if world.guest[view.selfInstanceID] != nil {
		return identGuest(world.sheet, view.selfID, cores.NoneID)
	} // if

	if world.card[view.selfInstanceID] != nil {
		return identCard(world.sheet, view.selfID, cores.NoneID)
	} // if

	return "空" // 防禦: self 實例已不在鏡像
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
