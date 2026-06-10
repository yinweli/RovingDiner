package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// gameEnd 終止哨兵:終止判定命中時自結算 panic 拋出、由 RunPhase 單點 recover 轉為對應終止站
// （【營業規格書 | 二十、獨立流程 | [終止判定]】「立即跳出結算進入對應階段」;M16 拍板:全鏈簽章零改動,
// self 綁定靠各層 defer restore 於解棧時自動還原）。
type gameEnd struct {
	phase cores.PhaseKind // PhaseGameSucc / PhaseGameFail
}

// Settle 執行結算（【營業規格書 | 二十、獨立流程 | 執行結算】）:結算旗標已立則 no-op（執行命令結算尾的不重入保證）;
// 流程:立旗標 → 終止判定 → 飽食門檻 → 飽食離場 → 終止判定 → 耐心門檻 → 生氣離場 → 終止判定 → 手牌上限棄牌 → 除旗標。
// 終止判定命中以 panic（gameEnd 哨兵）跳出、RunPhase recover。供 games 的編譯命令結算尾與 phaseRoundEnd 呼叫。
func Settle(game *cores.Game) {
	if game.Settling {
		return // 結算旗標已立 → 不重入
	} // if

	game.Settling = true
	judgeEnd(game)
	hitSate(game)
	exitSate(game)
	judgeEnd(game)
	hitCalm(game)
	exitCalm(game)
	judgeEnd(game)
	discardOver(game)
	game.Settling = false
}

// judgeEnd [終止判定]（【營業規格書 | 二十、獨立流程 | [終止判定]】）:失敗優先於成功;命中 → 清結算旗標、panic 哨兵跳出。
func judgeEnd(game *cores.Game) {
	if game.GetRound().GetValue() >= game.GetRoundMax().GetValue() || game.GetMorale().GetValue() <= 0 {
		game.Settling = false
		panic(gameEnd{phase: cores.PhaseGameFail})
	} // if

	if len(game.Wait) == 0 && len(game.Seat.Sorted()) == 0 && len(game.Roam) == 0 && len(game.Cardify) == 0 {
		game.Settling = false
		panic(gameEnd{phase: cores.PhaseGameSucc})
	} // if
}

// hitSate 飽食門檻:對所有顧客由低到高檢查,未觸發者標記並以對應門檻技能入行動佇列（行動類型 = 飽食）;
// sate 鎖定中跳過（sate 鎖定 = 飽食線暫停,遊蕩顧客即此態;M16 拍板）。
func hitSate(game *cores.Game) {
	for _, itor := range allGuest(game) {
		if itor.GetSate().IsLock() {
			continue // sate 鎖定 = 飽食線暫停
		} // if

		meta, ok := game.GuestData(itor.GetGuestID())

		if ok == false {
			continue // 無門檻資料
		} // if

		for _, threshold := range meta.Sate {
			if itor.GetSate().GetValue() >= threshold.Value && itor.GetSateHit().IsHit(threshold.Value) == false {
				itor.GetSateHit().Add(threshold.Value)
				game.Action.Push(cores.NewAction(itor, cores.TaskSate, threshold.SkillID))
			} // if
		} // for
	} // for
}

// exitSate 飽食離場:飽食值達離場線者依序離場（給滿意值、不扣士氣;重用 guestExitOne）;sate 鎖定中跳過。
// 候選為快照,離場觸發可能已搬動他人 → 逐位再定位（不在座位 / 遊蕩 → 略過）。
func exitSate(game *cores.Game) {
	for _, itor := range allGuest(game) {
		if itor.GetSate().IsLock() {
			continue // sate 鎖定 = 飽食線暫停
		} // if

		if itor.GetSate().GetValue() < itor.GetSateMax().GetValue() {
			continue // 未達離場線
		} // if

		if _, where, ok := game.LocateGuest(itor.GetInstanceID()); ok && (where == cores.ContainerSeat || where == cores.ContainerRoam) {
			guestExitOne(game, itor, where, true, false) // 飽食離場:給滿意值、不扣士氣
		} // if
	} // for
}

// hitCalm 耐心門檻:對所有顧客由高到低檢查,未觸發者標記並以對應門檻技能入行動佇列（行動類型 = 耐心）;遊蕩顧客照常運作。
func hitCalm(game *cores.Game) {
	for _, itor := range allGuest(game) {
		meta, ok := game.GuestData(itor.GetGuestID())

		if ok == false {
			continue // 無門檻資料
		} // if

		for _, threshold := range meta.Calm {
			if itor.GetCalm().GetValue() <= threshold.Value && itor.GetCalmHit().IsHit(threshold.Value) == false {
				itor.GetCalmHit().Add(threshold.Value)
				game.Action.Push(cores.NewAction(itor, cores.TaskCalm, threshold.SkillID))
			} // if
		} // for
	} // for
}

// exitCalm 生氣離場:耐心值 <= 0 者依序離場（扣士氣、不給滿意值;重用 guestExitOne）;遊蕩顧客照常運作。
// 候選為快照,離場觸發可能已搬動他人 → 逐位再定位（不在座位 / 遊蕩 → 略過）。
func exitCalm(game *cores.Game) {
	for _, itor := range allGuest(game) {
		if itor.GetCalm().GetValue() > 0 {
			continue // 未達離場線
		} // if

		if _, where, ok := game.LocateGuest(itor.GetInstanceID()); ok && (where == cores.ContainerSeat || where == cores.ContainerRoam) {
			guestExitOne(game, itor, where, false, true) // 生氣離場:扣士氣、不給滿意值
		} // if
	} // for
}

// discardOver 手牌上限棄牌（【營業規格書 | 二十、獨立流程 | 執行結算】尾段）:手牌超過上限時暫停流程交 Operator 逐張棄置;
// 回傳逐張驗證位於手牌（走 placeCard,照觸發 cardDrop）;一輪無進展（Operator 行為不良）→ 防禦跳出不掛死（M16 拍板）。
func discardOver(game *cores.Game) {
	for int32(len(game.Hand)) > game.GetHandMax().GetValue() {
		over := len(game.Hand) - int(game.GetHandMax().GetValue())
		before := len(game.Hand)

		for _, itor := range game.GetOperator().PickDiscard(append(cores.CardList{}, game.Hand...), over) {
			if _, where, ok := game.LocateCard(itor.GetInstanceID()); ok && where == cores.ContainerHand {
				removeCard(game, cores.ContainerHand, itor)
				placeCard(game, cores.ContainerDrop, itor)
			} // if
		} // for

		if len(game.Hand) >= before {
			return // 一輪無進展 → 防禦跳出
		} // if
	} // for
}

// allGuest 結算的「所有顧客」快照:座位列表（座位編號序）+ 遊蕩列表（列表序）,確保決定性;
// 迭代期間容器可變,離場類消費端逐位再定位把關。
func allGuest(game *cores.Game) (result []*cores.Guest) {
	result = append(result, game.Seat.Sorted()...)
	result = append(result, game.Roam...)
	return result
}
