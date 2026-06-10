package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 卡牌屬性命令（【營業規格書 | 二十五、操作命令清單】cardCost* / cardEffect*）：
// 對命令對象的每張卡牌實例直接改其欄位（非卡牌實例該項 no-op；不限容器位置）。
// cardCost* 重用 writeValue（出牌費用為「寫鎖」屬性：尊重鎖定計數、捨入、夾下限 0），與 self.cost 寫路徑一致。

func commandCardCostAdd(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardCost(eng, target, AssignAdd, arg)
}

func commandCardCostMul(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardCost(eng, target, AssignMul, arg)
}

func commandCardCostSet(eng *Engine, target []InstanceID, arg []exprs.Value) {
	cardCost(eng, target, AssignSet, arg)
}

// cardCost 對命令對象每張卡牌套用賦值符於出牌費用（寫鎖 / 捨入 / 夾下限 0）；N 缺漏 / 非數值整動作 no-op。
func cardCost(eng *Engine, target []InstanceID, op AssignKind, arg []exprs.Value) {
	n, ok := argNum(arg)

	if ok == false {
		return // N 缺漏 / 非數值 → 整動作 no-op
	} // if

	for _, itor := range target {
		card, _, found := eng.locateCard(itor)

		if found == false {
			continue // 非卡牌實例 → 該項 no-op
		} // if

		card.Cost.Apply(op, n)
		card.Cost.Clamp(0)
	} // for
}

func commandCardEffectAdd(eng *Engine, target []InstanceID, arg []exprs.Value) {
	effectID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	for _, itor := range target {
		card, _, found := eng.locateCard(itor)

		if found == false {
			continue
		} // if

		card.EffectID = append(card.EffectID, effectID)
	} // for
}

func commandCardEffectDel(eng *Engine, target []InstanceID, arg []exprs.Value) {
	effectID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	for _, itor := range target {
		card, _, found := eng.locateCard(itor)

		if found == false {
			continue
		} // if

		card.EffectID = removeOneEffect(card.EffectID, effectID)
	} // for
}

func commandCardEffectDelAll(eng *Engine, target []InstanceID, arg []exprs.Value) {
	effectID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	for _, itor := range target {
		card, _, found := eng.locateCard(itor)

		if found == false {
			continue
		} // if

		card.EffectID = removeAllEffect(card.EffectID, effectID)
	} // for
}

// removeOneEffect 自實例效果列表移除第一個 == effectID 者（卡上無此編號則原樣回）。
func removeOneEffect(effect []int32, effectID int32) (result []int32) {
	removed := false

	for _, itor := range effect {
		if removed == false && itor == effectID {
			removed = true
			continue
		} // if

		result = append(result, itor)
	} // for

	return result
}

// removeAllEffect 自實例效果列表移除全部 == effectID 者。
func removeAllEffect(effect []int32, effectID int32) (result []int32) {
	for _, itor := range effect {
		if itor != effectID {
			result = append(result, itor)
		} // if
	} // for

	return result
}
