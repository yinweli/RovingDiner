package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// fireTrigger 派發一個觸發時機（【營業規格書 | 二十、獨立流程 | 觸發時機】）;自效果佇列篩選「效果類型 == 觸發 && 觸發時機 == 時機」者、依作用順序排序、逐一處理。
// 觸發時機僅為時間訊號、不攜帶資料,各效果自持 self（啟動技能時已固定）。
func fireTrigger(game *Game, timing TriggerKind) {
	fire := EffectList{}

	for _, itor := range game.Effect {
		meta, ok := game.effectData[itor.GetEffectID()]

		if ok == false {
			continue // 查無編譯資料（防禦;正常實例必有對應效果）→ 略過
		} // if

		if meta.Kind == EffectTrigger && meta.TriggerKind == timing {
			fire = append(fire, itor)
		} // if
	} // for

	fire.Sort(game) // 快照排序;不擾動 game.Effect（順序無關）

	for _, itor := range fire {
		fireOne(game, itor)
	} // for
}

// fireOne 處理觸發列表中的單一效果:綁定 self → 觸發條件閘門 → 觸發次數（§十二）→ 連續執行觸發命令 → 觸發後行為 == 移除則執行結束命令並出佇列。
func fireOne(game *Game, effect *Effect) {
	meta := game.effectData[effect.GetEffectID()] // fireTrigger 已確認存在

	prev := game.self
	self := effect.GetSelf()
	game.self = &self // self / selfSame / selfNear 取效果建立時固定值（self 建立後不變,區域拷貝等價）

	defer func() { game.self = prev }() // 結算重入時逐層 save / restore

	if condPass(game, meta.Cond) == false {
		return // 觸發條件不成立（【十一】評估失敗即不成立）
	} // if

	count, ok := triggerCount(game, meta.Count)

	if ok == false {
		return // 觸發次數 <= 0 / 失敗 → 該效果中止（不觸發、不移除、留佇列）
	} // if

	runEffectCommand(game, meta.Trigger, effect.GetStack()*count) // 重複（堆疊層數 × M）次

	if meta.TriggerAfter == TriggerAfterRemove {
		runEffectCommand(game, meta.End, effect.GetStack()) // 重複 堆疊層數 次
		game.Effect.Remove(effect.GetInstanceID())
	} // if
}

// condPass 評估觸發條件;nil（空欄）→ 恆成立;評估失敗 / 非真值 → 不成立（【營業規格書 | 十一、觸發條件】）。以 §六 Value.Truthy 判真假。
func condPass(game *Game, cond *exprs.Expr) bool {
	if cond == nil {
		return true
	} // if

	result, ok := cond.Eval(game.env())

	if ok == false {
		return false
	} // if

	pass, ok := result.Truthy()

	return ok && pass
}

// triggerCount 評估觸發次數（【營業規格書 | 十二、觸發次數】）:nil（留空）→ 1;評估失敗 / 非數值 / 結果 < 0 / = 0 → ok == false（該效果中止）。小數四捨五入。
func triggerCount(game *Game, count *exprs.Expr) (n int32, ok bool) {
	if count == nil {
		return 1, true
	} // if

	result, valid := count.Eval(game.env())

	if valid == false || result.IsNum() == false {
		return 0, false
	} // if

	n = exprs.Round(result.Num())

	if n <= 0 {
		return 0, false
	} // if

	return n, true
}

// runEffectCommand 連續執行效果命令 times 次;nil 命令（空欄）整體略過、times <= 0 不執行。
func runEffectCommand(game *Game, command EffectCommand, times int32) {
	if command == nil {
		return
	} // if

	for itor := int32(0); itor < times; itor++ {
		command(game)
		// TODO(M16)：執行命令 結算重入尾 —— 結算旗標 == false → 執行結算（【營業規格書 | 二十、獨立流程 | 執行命令】）;結算為 M16 seam。
	} // for
}
