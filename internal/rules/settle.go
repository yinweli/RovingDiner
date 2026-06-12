package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// gameEnd 終止哨兵: 終止判定命中時自結算 panic 拋出、由 RunPhase 單點 recover 轉為對應終止站
// (【營業規格書 | 二十、獨立流程 | [終止判定]】「立即跳出結算進入對應階段」; M16 拍板: 全鏈簽章零改動,
// self 綁定靠各層 defer restore 於解棧時自動還原)。
type gameEnd struct {
	phase cores.PhaseKind // PhaseGameSucc / PhaseGameFail
}

// Settle 執行結算(【營業規格書 | 二十、獨立流程 | 執行結算】): 結算旗標已立則 no-op(執行命令結算尾的不重入保證);
// 無事結算(settleBusy 預判全不中)靜默 no-op、不發題(M18 拍板: 命令結算尾高頻呼叫, 空題灌爆日誌)。
// 流程: 立旗標 → 發範圍標題 → 範圍壓回 → 終止判定 → 飽食門檻 → 飽食離場 → 終止判定 → 耐心門檻 → 生氣離場 → 終止判定 → 手牌上限棄牌 → 除旗標。
// 終止判定命中以 panic(gameEnd 哨兵)跳出、RunPhase recover。供 games 的編譯命令結算尾與 phaseRoundEnd 呼叫。
func Settle(game *cores.Game) {
	if game.Settling {
		return // 結算旗標已立 → 不重入
	} // if

	if settleBusy(game) == false {
		return // 無事結算(判定不會中、壓回門檻離場棄牌全無)→ 靜默 no-op, 不發題
	} // if

	game.Settling = true
	cores.EmitTitle(game, "執行結算") // 範圍標題: 執行結算(有事才發; M18 拍板)
	clampOver(game)
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

// settleBusy 廉價預判本次結算是否會有動作(有事才發範圍標題; M18 拍板):
// 終止判定會中 / 範圍待壓回 / 門檻待標記 / 離場線已達 / 手牌超上限, 任一成立即有事; 全不中 → Settle 靜默 no-op(各步皆 no-op, 行為等價)。
// 判定點在進場、狀態未變: 前段預測必準; 前段真有動作則題已該發, 後段預測失準無影響——「預判有事 ↔ 實際有事」一致。
// 判式與正式迴圈共用同一組述詞(overMax / reach* / exitable* / end*), 避免兩份邏輯漂移。
func settleBusy(game *cores.Game) bool {
	if endFail(game) || endSucc(game) {
		return true
	} // if

	if overMax(game.GetMorale(), game.GetMoraleMax()) || overMax(game.GetEnergy(), game.GetEnergyMax()) {
		return true
	} // if

	for _, itor := range allGuest(game) {
		if overMax(itor.GetMorale(), itor.GetMoraleMax()) || pendingSate(game, itor) || pendingCalm(game, itor) || exitableSate(itor) || exitableCalm(itor) {
			return true
		} // if
	} // for

	return int32(len(game.Hand)) > game.GetHandMax().GetValue()
}

// overMax 範圍壓回述詞: 當前值超過上限且未鎖定(鎖定 = 拒寫, 解鎖後下次結算壓回); clampOver 與 settleBusy 預判共用。
func overMax(value, high *cores.Value) bool {
	return value.IsLock() == false && value.GetValue() > high.GetValue()
}

// clampOver 範圍壓回(【營業規格書 | 二十、獨立流程 | 執行結算】首步): 上限縮減使當前值超標時壓回至上限並發屬性行
// (全域士氣 / 出牌點數、顧客士氣); 下限與寫入當下的範圍在詞條夾(【營業規格書 | 二十三、屬性清單 | 寫入範圍】), 此處僅收上限連動。
// 置於終止判定之前: moraleMax 歸零 → morale 壓 0 → 緊接的失敗判定即收場; 離場扣士氣亦用壓回後的顧客 morale。
func clampOver(game *cores.Game) {
	if overMax(game.GetMorale(), game.GetMoraleMax()) {
		game.GetMorale().Set(float64(game.GetMoraleMax().GetValue()))
		cores.EmitProperty(game, 0, cores.NoneID, "morale", cores.AssignSet, float64(game.GetMorale().GetValue()), float64(game.GetMorale().GetValue()))
	} // if

	if overMax(game.GetEnergy(), game.GetEnergyMax()) {
		game.GetEnergy().Set(float64(game.GetEnergyMax().GetValue()))
		cores.EmitProperty(game, 0, cores.NoneID, "energy", cores.AssignSet, float64(game.GetEnergy().GetValue()), float64(game.GetEnergy().GetValue()))
	} // if

	for _, itor := range allGuest(game) {
		if overMax(itor.GetMorale(), itor.GetMoraleMax()) {
			itor.GetMorale().Set(float64(itor.GetMoraleMax().GetValue()))
			cores.EmitProperty(game, itor.GetGuestID(), itor.GetInstanceID(), "morale", cores.AssignSet, float64(itor.GetMorale().GetValue()), float64(itor.GetMorale().GetValue()))
		} // if
	} // for
}

// judgeEnd [終止判定](【營業規格書 | 二十、獨立流程 | [終止判定]】): 失敗優先於成功; 命中 → 清結算旗標、panic 哨兵跳出。
func judgeEnd(game *cores.Game) {
	if endFail(game) {
		game.Settling = false
		panic(gameEnd{phase: cores.PhaseGameFail})
	} // if

	if endSucc(game) {
		game.Settling = false
		panic(gameEnd{phase: cores.PhaseGameSucc})
	} // if
}

// endFail 終止判定的失敗條件(回合達上限 / 士氣耗盡; 失敗優先於成功); judgeEnd 與 settleBusy 預判共用。
func endFail(game *cores.Game) bool {
	return game.GetRound().GetValue() >= game.GetRoundMax().GetValue() || game.GetMorale().GetValue() <= 0
}

// endSucc 終止判定的成功條件(全場顧客清空); judgeEnd 與 settleBusy 預判共用。
func endSucc(game *cores.Game) bool {
	return len(game.Wait) == 0 && len(game.Seat.Sorted()) == 0 && len(game.Roam) == 0 && len(game.Cardify) == 0
}

// hitSate 飽食門檻: 對所有顧客由低到高檢查, 未觸發者標記並以對應門檻技能入行動佇列(行動類型 = 飽食);
// sate 鎖定中跳過(sate 鎖定 = 飽食線暫停, 遊蕩顧客即此態; M16 拍板)。
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
			if reachSate(itor, threshold.Value) {
				itor.GetSateHit().Add(threshold.Value)
				action := cores.NewAction(itor, cores.TaskSate, threshold.SkillID)
				game.Action.Push(action)
				emitAction(game, action) // 入列事件(M21 拍板)
			} // if
		} // for
	} // for
}

// exitSate 飽食離場: 飽食值達離場線者依序離場(給滿意值、不扣士氣; 重用 guestExitOne); sate 鎖定中跳過。
// 候選為快照, 離場觸發可能已搬動他人 → 逐位再定位(不在座位 / 遊蕩 → 略過)。
func exitSate(game *cores.Game) {
	for _, itor := range allGuest(game) {
		if exitableSate(itor) == false {
			continue // sate 鎖定 = 飽食線暫停 / 未達離場線
		} // if

		if _, where, ok := game.LocateGuest(itor.GetInstanceID()); ok && (where == cores.ContainerSeat || where == cores.ContainerRoam) {
			guestExitOne(game, itor, where, true, false) // 飽食離場: 給滿意值、不扣士氣
		} // if
	} // for
}

// hitCalm 耐心門檻: 對所有顧客由高到低檢查, 未觸發者標記並以對應門檻技能入行動佇列(行動類型 = 耐心); 遊蕩顧客照常運作。
func hitCalm(game *cores.Game) {
	for _, itor := range allGuest(game) {
		meta, ok := game.GuestData(itor.GetGuestID())

		if ok == false {
			continue // 無門檻資料
		} // if

		for _, threshold := range meta.Calm {
			if reachCalm(itor, threshold.Value) {
				itor.GetCalmHit().Add(threshold.Value)
				action := cores.NewAction(itor, cores.TaskCalm, threshold.SkillID)
				game.Action.Push(action)
				emitAction(game, action) // 入列事件(M21 拍板)
			} // if
		} // for
	} // for
}

// exitCalm 生氣離場: 耐心值 <= 0 者依序離場(扣士氣、不給滿意值; 重用 guestExitOne); 遊蕩顧客照常運作。
// 候選為快照, 離場觸發可能已搬動他人 → 逐位再定位(不在座位 / 遊蕩 → 略過)。
func exitCalm(game *cores.Game) {
	for _, itor := range allGuest(game) {
		if exitableCalm(itor) == false {
			continue // 未達離場線
		} // if

		if _, where, ok := game.LocateGuest(itor.GetInstanceID()); ok && (where == cores.ContainerSeat || where == cores.ContainerRoam) {
			guestExitOne(game, itor, where, false, true) // 生氣離場: 扣士氣、不給滿意值
		} // if
	} // for
}

// reachSate 飽食門檻判式: 飽食值達門檻且未標記; hitSate 迴圈與 settleBusy 預判(pendingSate)共用。
func reachSate(guest *cores.Guest, value int32) bool {
	return guest.GetSate().GetValue() >= value && guest.GetSateHit().IsHit(value) == false
}

// reachCalm 耐心門檻判式: 耐心值低於等於門檻且未標記; hitCalm 迴圈與 settleBusy 預判(pendingCalm)共用。
func reachCalm(guest *cores.Guest, value int32) bool {
	return guest.GetCalm().GetValue() <= value && guest.GetCalmHit().IsHit(value) == false
}

// exitableSate 飽食離場線判式: sate 未鎖定且飽食值達離場線; exitSate 與 settleBusy 預判共用。
func exitableSate(guest *cores.Guest) bool {
	return guest.GetSate().IsLock() == false && guest.GetSate().GetValue() >= guest.GetSateMax().GetValue()
}

// exitableCalm 生氣離場線判式: 耐心值 <= 0; exitCalm 與 settleBusy 預判共用。
func exitableCalm(guest *cores.Guest) bool {
	return guest.GetCalm().GetValue() <= 0
}

// pendingSate 顧客是否有待標記的飽食門檻(settleBusy 預判; 與 hitSate 同一跳過條件 + reachSate 判式)。
func pendingSate(game *cores.Game, guest *cores.Guest) bool {
	if guest.GetSate().IsLock() {
		return false // sate 鎖定 = 飽食線暫停
	} // if

	meta, ok := game.GuestData(guest.GetGuestID())

	if ok == false {
		return false // 無門檻資料
	} // if

	for _, threshold := range meta.Sate {
		if reachSate(guest, threshold.Value) {
			return true
		} // if
	} // for

	return false
}

// pendingCalm 顧客是否有待標記的耐心門檻(settleBusy 預判; 與 hitCalm 同一 reachCalm 判式, 遊蕩照常)。
func pendingCalm(game *cores.Game, guest *cores.Guest) bool {
	meta, ok := game.GuestData(guest.GetGuestID())

	if ok == false {
		return false // 無門檻資料
	} // if

	for _, threshold := range meta.Calm {
		if reachCalm(guest, threshold.Value) {
			return true
		} // if
	} // for

	return false
}

// discardOver 手牌上限棄牌(【營業規格書 | 二十、獨立流程 | 執行結算】尾段): 手牌超過上限時暫停流程交 Operator 逐張棄置;
// 選取結果發玩家輸入紀錄(流程名 discardOver, 前端映「手牌上限」; M18 拍板);
// 回傳逐張驗證位於手牌(走 placeCard, 照觸發 cardDrop); 一輪無進展(Operator 行為不良)→ 防禦跳出不掛死(M16 拍板)。
func discardOver(game *cores.Game) {
	for int32(len(game.Hand)) > game.GetHandMax().GetValue() {
		over := len(game.Hand) - int(game.GetHandMax().GetValue())
		before := len(game.Hand)
		chosen := game.GetOperator().PickDiscard(promptText(cores.SourceText("discardOver"), "手牌"), append(cores.CardList{}, game.Hand...), over)
		emitSelect(game, "discardOver", 0, cardIdentList(game, chosen))

		for _, itor := range chosen {
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

// allGuest 結算的「所有顧客」快照: 座位列表(座位編號序)+ 遊蕩列表(列表序), 確保決定性;
// 迭代期間容器可變, 離場類消費端逐位再定位把關。
func allGuest(game *cores.Game) (result []*cores.Guest) {
	result = append(result, game.Seat.Sorted()...)
	result = append(result, game.Roam...)
	return result
}
