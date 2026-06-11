package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 卡牌屬性命令(【營業規格書 | 二十五、操作命令清單】cardCost* / cardEffect*):
// 對命令對象的每張卡牌實例直接改其欄位(非卡牌實例該項 no-op; 不限容器位置)。
// cardCost* 經 Value.Apply 套用賦值符(出牌費用為「寫鎖」屬性: 尊重鎖定計數、捨入)後夾下限 0。

func commandCardCostAdd(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardCost(game, target, cores.AssignAdd, arg)
}

func commandCardCostMul(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardCost(game, target, cores.AssignMul, arg)
}

func commandCardCostSet(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	cardCost(game, target, cores.AssignSet, arg)
}

// cardCost 對命令對象每張卡牌套用賦值符於出牌費用(寫鎖 / 捨入 / 夾下限 0); N 缺漏 / 非數值整動作 no-op。
// 數值型操作命令事件(M21 拍板): 逐卡比照 ExecAssign 收口包前後值發屬性事件, 鎖定拒寫以 Before == After 表達。
func cardCost(game *cores.Game, target []cores.InstanceID, op cores.AssignKind, arg []exprs.Value) {
	n, ok := argNum(arg)

	if ok == false {
		return // N 缺漏 / 非數值 → 整動作 no-op
	} // if

	for _, itor := range target {
		card, _, found := game.LocateCard(itor)

		if found == false {
			continue // 非卡牌實例 → 該項 no-op
		} // if

		card.GetCost().Apply(op, n)
		card.GetCost().Clamp(0)
		cores.EmitProperty(game, card.GetCardID(), card.GetInstanceID(), "cost", op, n, float64(card.GetCost().GetValue()))
	} // for
}

func commandCardEffectAdd(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	effectID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	for _, itor := range target {
		card, _, found := game.LocateCard(itor)

		if found == false {
			continue
		} // if

		card.GetEffectID().Add(effectID)
	} // for
}

func commandCardEffectDel(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	effectID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	for _, itor := range target {
		card, _, found := game.LocateCard(itor)

		if found == false {
			continue
		} // if

		card.GetEffectID().DelOne(effectID)
	} // for
}

func commandCardEffectDelAll(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	effectID, ok := argInt(arg)

	if ok == false {
		return
	} // if

	for _, itor := range target {
		card, _, found := game.LocateCard(itor)

		if found == false {
			continue
		} // if

		card.GetEffectID().DelAll(effectID)
	} // for
}
