package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// 投影事件發射輔助（M18 Emit 接線;【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 四之二】）:
// 流程與效果系統的事件組裝收口,發射一律經 Game.Emit 蓋章座標（回合 / 階段）,發射點只填本體欄位。

// emitEffect 發射效果事件（日誌效果行）:對象欄載 self 本體（空物件留零值）、效果編號 / 效果實例編號 / 階段為本體欄;
// 立即類不入佇列,效果實例編號傳 cores.NoneID。
func emitEffect(game *cores.Game, effectID int32, instanceID cores.InstanceID, self cores.Ref, stage cores.EffectStage) {
	dataID, selfID := cores.RefTarget(self)
	game.Emit(cores.EventData{Kind: cores.EventEffect, DataID: dataID, InstanceID: selfID, EffectID: effectID, EffectInstanceID: instanceID, Stage: stage})
}

// emitGuestMove 發射顧客容器搬移事件:新建直入 from == ContainerNone、銷毀離場 to == ContainerNone;
// to == ContainerSeat 時帶座位編號,其餘傳 0。卡牌搬移由 placeCard 收口,顧客無收口、各搬移點逐點發。
func emitGuestMove(game *cores.Game, guest *cores.Guest, from, to cores.ContainerKind, seatID int32) {
	game.Emit(cores.EventData{Kind: cores.EventContainer, DataID: guest.GetGuestID(), InstanceID: guest.GetInstanceID(), From: from, To: to, SeatID: seatID})
}

// emitProperty 發射流程直寫的屬性事件（白名單;M18 拍板）:流程呼叫點自包前後值;全域屬性對象欄傳零值。
// 命令路徑的屬性事件歸 ExecAssign 收口,不經此。
func emitProperty(game *cores.Game, dataID int32, instanceID cores.InstanceID, attr string, op cores.AssignKind, operand, before, after float64) {
	game.Emit(cores.EventData{Kind: cores.EventProperty, DataID: dataID, InstanceID: instanceID, Attr: attr, Op: op, Operand: operand, Before: before, After: after})
}

// emitSelect 發射玩家選取紀錄（日誌 $ 選取行;搭 seed 重現用,只記 Operator 真選取）:
// 來源為詞條鍵 / 流程名（前端轉中文）;effectID 供效果目標選取掛效果編號（前端查表顯示來源名）,其餘傳 0。
// 選中清單每發新建切片、發後不改（EventData 值複製合約的唯一共享點）。
func emitSelect(game *cores.Game, source string, effectID int32, pick []cores.PickData) {
	game.Emit(cores.EventData{Kind: cores.EventSelect, Source: source, EffectID: effectID, Pick: pick})
}

// cardPickData 把卡牌清單組成選中清單（每發新建切片,不共享來源底層）。
func cardPickData(card []*cores.Card) (result []cores.PickData) {
	result = make([]cores.PickData, 0, len(card))

	for _, itor := range card {
		result = append(result, cores.PickData{DataID: itor.GetCardID(), InstanceID: itor.GetInstanceID()})
	} // for

	return result
}

// guestPickData 把顧客清單組成選中清單（每發新建切片,不共享來源底層）。
func guestPickData(guest []*cores.Guest) (result []cores.PickData) {
	result = make([]cores.PickData, 0, len(guest))

	for _, itor := range guest {
		result = append(result, cores.PickData{DataID: itor.GetGuestID(), InstanceID: itor.GetInstanceID()})
	} // for

	return result
}
