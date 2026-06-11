package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// 事件發射輔助(M18 Emit 接線; 【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 四之二】):
// 流程與效果系統的事件組裝收口, 發射一律經 Game.Emit 蓋章座標(回合 / 階段), 發射點只填本體欄位。

// emitEffect 發射效果事件(日誌效果行): 對象欄載 self 本體(空物件留零值)、效果編號 / 效果實例編號 / 階段為本體欄;
// 立即類不入佇列, 效果實例編號傳 cores.NoneID。
func emitEffect(game *cores.Game, effectID int32, instanceID cores.InstanceID, self cores.Ref, stage cores.EffectStage) {
	dataID, selfID := cores.RefTarget(self)
	game.Emit(cores.EventData{Kind: cores.EventEffect, DataID: dataID, InstanceID: selfID, EffectID: effectID, EffectInstanceID: instanceID, Stage: stage})
}

// emitGuestMove 發射顧客容器搬移事件: 新建直入 from == ContainerNone、銷毀離場 to == ContainerNone;
// to == ContainerSeat 時帶座位編號, 其餘傳 0。卡牌搬移由 placeCard 收口, 顧客無收口、各搬移點逐點發。
func emitGuestMove(game *cores.Game, guest *cores.Guest, from, to cores.ContainerKind, seatID int32) {
	game.Emit(cores.EventData{Kind: cores.EventContainer, DataID: guest.GetGuestID(), InstanceID: guest.GetInstanceID(), From: from, To: to, SeatID: seatID})
}

// emitCardMove 發射卡牌容器搬移事件(placeCard 收口與開局 / 洗回事件共用; M21 拍板):
// 已綁卡牌化來源者帶語境欄(來源顧客資料 + 實例編號, 比照 SeatID 先例), 供前端標注 cardify 卡。
func emitCardMove(game *cores.Game, card *cores.Card, from, to cores.ContainerKind) {
	eventData := cores.EventData{Kind: cores.EventContainer, DataID: card.GetCardID(), InstanceID: card.GetInstanceID(), From: from, To: to}

	if guest := card.GetCardify(); guest != nil {
		eventData.BindID = guest.GetGuestID()
		eventData.BindInstanceID = guest.GetInstanceID()
	} // if

	game.Emit(eventData)
}

// emitDeckOrder 發射抽牌牌堆重整快照(From == To == Deck、Pick 載重整後全序; M21 拍板):
// 洗牌 / 洗回後收口順序——逐卡移動事件載成員真相、快照載順序真相, 前端整堆置換。
func emitDeckOrder(game *cores.Game) {
	game.Emit(cores.EventData{Kind: cores.EventContainer, From: cores.ContainerDeck, To: cores.ContainerDeck, Pick: cardPickData(game.Deck)})
}

// emitAction 發射行動佇列事件(入列 alive == true / 出列 alive == false; M21 拍板):
// 對象欄載顧客、SkillID 載技能、Task 載行動類型; 入列發於門檻命中與 taskAdd、出列發於顧客行動彈出(含封印防禦路徑)。
func emitAction(game *cores.Game, action *cores.Action, alive bool) {
	game.Emit(cores.EventData{Kind: cores.EventAction, DataID: action.GetGuest().GetGuestID(), InstanceID: action.GetGuest().GetInstanceID(), SkillID: action.GetSkillID(), Task: action.GetKind(), Alive: alive})
}

// emitEffectState 發射載佇列狀態快照的效果事件(加入 / 結束; M21 拍板): Stack / Expire 為事件後絕對值,
// Alive 區分留佇列(true; 加入與退層)/ 出佇列(false; 退場, Stack 為退場層數)。
// 不載佇列狀態的階段(立即 / 觸發 / 啟動 / 條件不成立)仍走 emitEffect。
func emitEffectState(game *cores.Game, effect *cores.Effect, stage cores.EffectStage, alive bool) {
	dataID, selfID := cores.RefTarget(effect.GetSelf())
	game.Emit(cores.EventData{Kind: cores.EventEffect, DataID: dataID, InstanceID: selfID, EffectID: effect.GetEffectID(), EffectInstanceID: effect.GetInstanceID(), Stage: stage, Stack: effect.GetStack(), Expire: effect.GetExpire(), Alive: alive})
}

// emitProperty 發射流程直寫的屬性事件(白名單; M18 拍板): 流程呼叫點自包前後值; 全域屬性對象欄傳零值。
// 命令路徑的屬性事件歸 ExecAssign 收口, 不經此。
func emitProperty(game *cores.Game, dataID int32, instanceID cores.InstanceID, attr string, op cores.AssignKind, operand, before, after float64) {
	game.Emit(cores.EventData{Kind: cores.EventProperty, DataID: dataID, InstanceID: instanceID, Attr: attr, Op: op, Operand: operand, Before: before, After: after})
}

// emitSelect 發射玩家選取紀錄(日誌 $ 選取行; 搭 seed 重現用, 只記 Operator 真選取):
// 來源為詞條鍵 / 流程名(前端轉中文); effectID 供效果目標選取掛效果編號(前端查表顯示來源名), 其餘傳 0。
// 選中清單每發新建切片、發後不改(EventData 值複製合約的唯一共享點)。
func emitSelect(game *cores.Game, source string, effectID int32, pick []cores.PickData) {
	game.Emit(cores.EventData{Kind: cores.EventSelect, Source: source, EffectID: effectID, Pick: pick})
}

// cardPickData 把卡牌清單組成選中清單(每發新建切片, 不共享來源底層)。
func cardPickData(card []*cores.Card) (result []cores.PickData) {
	result = make([]cores.PickData, 0, len(card))

	for _, itor := range card {
		result = append(result, cores.PickData{DataID: itor.GetCardID(), InstanceID: itor.GetInstanceID()})
	} // for

	return result
}

// guestPickData 把顧客清單組成選中清單(每發新建切片, 不共享來源底層)。
func guestPickData(guest []*cores.Guest) (result []cores.PickData) {
	result = make([]cores.PickData, 0, len(guest))

	for _, itor := range guest {
		result = append(result, cores.PickData{DataID: itor.GetGuestID(), InstanceID: itor.GetInstanceID()})
	} // for

	return result
}
