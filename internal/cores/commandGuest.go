package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 顧客免疫 / 行動命令（【營業規格書 | 二十五、操作命令清單】effectImmune* / skillImmune* / taskAdd）：
// 對命令對象的每位顧客操作免疫群組鎖定計數，或加入行動（非顧客實例該項 no-op；不限容器位置）。
// 免疫群組編號為命令尾端的 varargs（trailing 參數，逐個套用）。

func commandEffectImmuneAdd(eng *Engine, target []InstanceID, arg []exprs.Value) {
	immuneAdd(eng, target, arg, (*Guest).GetEffectImmune)
}

func commandEffectImmuneDel(eng *Engine, target []InstanceID, arg []exprs.Value) {
	immuneDel(eng, target, arg, (*Guest).GetEffectImmune)
}

func commandSkillImmuneAdd(eng *Engine, target []InstanceID, arg []exprs.Value) {
	immuneAdd(eng, target, arg, (*Guest).GetSkillImmune)
}

func commandSkillImmuneDel(eng *Engine, target []InstanceID, arg []exprs.Value) {
	immuneDel(eng, target, arg, (*Guest).GetSkillImmune)
}

// immuneAdd 對命令對象每位顧客、每個 varargs 群組編號，其免疫群組鎖定計數 + 1；pick 取顧客的免疫計數組件（增減紀律由 Immune 把關）。
func immuneAdd(eng *Engine, target []InstanceID, arg []exprs.Value, pick func(guest *Guest) *Immune) {
	for _, itor := range target {
		guest, _, ok := eng.locateGuest(itor)

		if ok == false {
			continue // 非顧客實例 → 該項 no-op
		} // if

		for _, value := range arg {
			if value.IsNum() {
				pick(guest).Add(int32(value.Num()))
			} // if
		} // for
	} // for
}

// immuneDel 對命令對象每位顧客、每個 varargs 群組編號，其免疫群組鎖定計數 - 1（夾 ≥ 0,由 Immune.Del 把關）。
func immuneDel(eng *Engine, target []InstanceID, arg []exprs.Value, pick func(guest *Guest) *Immune) {
	for _, itor := range target {
		guest, _, ok := eng.locateGuest(itor)

		if ok == false {
			continue
		} // if

		for _, value := range arg {
			if value.IsNum() {
				pick(guest).Del(int32(value.Num()))
			} // if
		} // for
	} // for
}

// commandTaskAdd 對命令對象每位顧客，將「顧客 + 行動類型 + 技能編號」加入行動佇列尾端（【二十五 | taskAdd】）。
// 參數：行動類型（0 飽食 / 1 耐心，對齊 TaskKind）、技能編號；任一缺漏 / 非數值整動作 no-op。
func commandTaskAdd(eng *Engine, target []InstanceID, arg []exprs.Value) {
	kind, ok := argInt(arg)

	if ok == false {
		return
	} // if

	skillID, ok := argInt(arg[1:])

	if ok == false {
		return
	} // if

	for _, itor := range target {
		guest, _, found := eng.locateGuest(itor)

		if found == false {
			continue
		} // if

		eng.runtime.Action.Push(NewAction(guest, TaskKind(kind), skillID))
	} // for
}
