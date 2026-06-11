package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 顧客免疫 / 行動命令(【營業規格書 | 二十五、操作命令清單】effectImmune* / skillImmune* / taskAdd):
// 對命令對象的每位顧客操作免疫群組鎖定計數, 或加入行動(非顧客實例該項 no-op; 不限容器位置)。
// 免疫群組編號為命令尾端的 varargs(trailing 參數, 逐個套用)。

func commandEffectImmuneAdd(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	immuneAdd(game, target, arg, "effectImmune", (*cores.Guest).GetEffectImmune)
}

func commandEffectImmuneDel(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	immuneDel(game, target, arg, "effectImmune", (*cores.Guest).GetEffectImmune)
}

func commandSkillImmuneAdd(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	immuneAdd(game, target, arg, "skillImmune", (*cores.Guest).GetSkillImmune)
}

func commandSkillImmuneDel(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	immuneDel(game, target, arg, "skillImmune", (*cores.Guest).GetSkillImmune)
}

// immuneAdd 對命令對象每位顧客、每個 varargs 群組編號, 其免疫群組鎖定計數 + 1; pick 取顧客的免疫計數組件(增減紀律由 Immune 把關)。
// 每次套用發一筆 property 事件: Operand 載群組編號(維度鍵重用右值欄)、前後值載該群組計數(M22 拍板)。
func immuneAdd(game *cores.Game, target []cores.InstanceID, arg []exprs.Value, attr string, pick func(guest *cores.Guest) *cores.Immune) {
	for _, itor := range target {
		guest, _, ok := game.LocateGuest(itor)

		if ok == false {
			continue // 非顧客實例 → 該項 no-op
		} // if

		for _, value := range arg {
			if value.IsNum() {
				group := int32(value.Num())
				before := pick(guest).Get(group)
				pick(guest).Add(group)
				emitProperty(game, guest.GetGuestID(), guest.GetInstanceID(), attr, cores.AssignAdd, float64(group), float64(before), float64(pick(guest).Get(group)))
			} // if
		} // for
	} // for
}

// immuneDel 對命令對象每位顧客、每個 varargs 群組編號, 其免疫群組鎖定計數 - 1(夾 ≥ 0, 由 Immune.Del 把關)。
// 投影同 immuneAdd(Op 為 Sub); 夾 0 不動時照發、Before == After 表達無變化(比照鎖定拒寫; M22 拍板)。
func immuneDel(game *cores.Game, target []cores.InstanceID, arg []exprs.Value, attr string, pick func(guest *cores.Guest) *cores.Immune) {
	for _, itor := range target {
		guest, _, ok := game.LocateGuest(itor)

		if ok == false {
			continue
		} // if

		for _, value := range arg {
			if value.IsNum() {
				group := int32(value.Num())
				before := pick(guest).Get(group)
				pick(guest).Del(group)
				emitProperty(game, guest.GetGuestID(), guest.GetInstanceID(), attr, cores.AssignSub, float64(group), float64(before), float64(pick(guest).Get(group)))
			} // if
		} // for
	} // for
}

// commandTaskAdd 對命令對象每位顧客, 將「顧客 + 行動類型 + 技能編號」加入行動佇列尾端(【二十五 | taskAdd】)。
// 參數: 行動類型(0 飽食 / 1 耐心, 對齊 TaskKind)、技能編號; 任一缺漏 / 非數值整動作 no-op。
func commandTaskAdd(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	kind, ok := argInt(arg)

	if ok == false {
		return
	} // if

	skillID, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	for _, itor := range target {
		guest, _, found := game.LocateGuest(itor)

		if found == false {
			continue
		} // if

		action := cores.NewAction(guest, cores.TaskKind(kind), skillID)
		game.Action.Push(action)
		emitAction(game, action, true) // 入列投影(M21 拍板)
	} // for
}
