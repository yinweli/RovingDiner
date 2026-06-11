package cores

import (
	"strconv"
)

// 日誌行發射台(【營業顯示規格書 | 6、畫面規格 | 6.10】): 引擎於發射點直接組最終日誌行(中文敘事),
// 行角色看首字——範圍標題 [ / 操作元 * / 效果 - / 流程直屬 $ / 效果欄位(兩格縮排); 消費端零解析。
// 歸因狀態為平面語意: 最後效果行勝出、範圍標題重置(巢狀派發 effectRun 後外層命令行歸因於內層效果,
// 此為敘事語法而非缺陷)——決定命令行前綴($ / 縮排)與屬性行的 self 省略。
// 發射 = 有話要說: 每拍都有敘事行, 無空拍概念; phase 切換不發射(階段值活在標題前綴與狀態列直讀)。

// EmitTitle 發射範圍標題 + 操作元行(一拍): [R{回合} {階段}] {標題} 與其下 * 行(操作元識別碼由呼叫端組畢);
// 並重置歸因。發射點守「有事才發」: 空名單 / 無事不立題。
func EmitTitle(game *Game, title string, operator ...string) {
	game.logEffect = false
	game.logSelfData = 0
	game.logSelfInstance = NoneID
	line := []string{"[R" + strconv.FormatInt(int64(game.round.GetValue()), 10) + " " + PhaseName(game.phaseCurr) + "] " + title}

	for _, itor := range operator {
		line = append(line, "* "+itor)
	} // for

	game.Emit(line...)
}

// EmitEffect 發射效果頭三行(一拍): - {效果識別碼}、欄 2 對象(空物件顯 空, 空欄佔行位置才穩)、欄 2 階段;
// 並把歸因設為本效果 self(其後命令行縮排掛欄 2)。立即類不入佇列, 效果實例編號傳 NoneID。
func EmitEffect(game *Game, effectID int32, instanceID InstanceID, self Ref, stage EffectStage) {
	dataID, selfID := RefTarget(self)
	target := "空"

	if dataID != 0 {
		target = IdentTarget(game.GetSheet(), dataID, selfID)
	} // if

	game.logEffect = true
	game.logSelfData = dataID
	game.logSelfInstance = selfID
	game.Emit("- "+IdentEffect(game.GetSheet(), effectID, instanceID), "  "+target, "  "+StageText(stage))
}

// EmitProperty 發射屬性命令行(一拍; <屬性> <運算> <值> >> <結果>): 全域屬性(對象零值)用全名、
// 實例屬性作用於歸因 self 時屬性開頭、其他對象識別碼開頭(fan-out 每對象一行);
// 免疫詞條的運算值為群組維度鍵, 行文轉查詢函式形 名稱(群組) 且運算值固定 1(M22 拍板);
// 鎖定 / 解鎖無算術式, 結果為鎖定計數。
func EmitProperty(game *Game, dataID int32, instanceID InstanceID, attr string, op AssignKind, operand, after float64) {
	global := dataID == 0 && instanceID == NoneID
	name := AttrText(attr, global)
	value := NumText(operand)

	if attr == AttrEffectImmune || attr == AttrSkillImmune {
		name += "(" + value + ")"
		value = "1"
	} // if

	body := name

	if global == false && (game.logEffect == false || dataID != game.logSelfData || instanceID != game.logSelfInstance) {
		body = IdentTarget(game.GetSheet(), dataID, instanceID) + " " + name
	} // if

	if op == AssignLock || op == AssignUnlock {
		EmitBody(game, body+" "+AssignText(op)+" >> "+NumText(after))
		return
	} // if

	EmitBody(game, body+" "+AssignText(op)+" "+value+" >> "+NumText(after))
}

// EmitBody 發射命令行(一拍): 依歸因——有當前效果 → 效果欄位(兩格縮排)、否則 → 流程直屬行($)。
func EmitBody(game *Game, body string) {
	if game.logEffect {
		game.Emit("  " + body)
		return
	} // if

	game.Emit("$ " + body)
}
