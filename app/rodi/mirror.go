package rodi

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// mirror 世界鏡像(顯示側狀態; 【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 四之二】單向事件投影):
// 事件摺疊一次、各組件唯讀共用(M20 拍板)。摺疊總原則(M21 拍板): 靜態表查得到的查表(名稱 / 實例初值 /
// 容器插入端【營業規格書 | 六、容器結構】), 執行期才產生的必須有事件(實例編號 / 容器位置 / 佇列 / 層數 / 牌序)。
// 規則狀態以鏡像為單一來源, 組件只自持 UI 狀態(游標 / 捲動 / 模式)。
type mirror struct {
	sheet  *sheeter.Sheeter                           // 靜態表(名稱與初值查表)
	round  int32                                      // 當前回合(事件座標)
	phase  cores.PhaseKind                            // 當前階段(事件座標)
	attr   map[string]float64                         // 全域屬性值(寫側詞條鍵 → 後值)
	lock   map[string]float64                         // 全域屬性鎖定計數(寫側詞條鍵 → 後值; @ / # 事件載鎖定計數)
	card   map[cores.InstanceID]*cardView             // 卡牌實例視圖
	guest  map[cores.InstanceID]*guestView            // 顧客實例視圖
	zone   map[cores.ContainerKind][]cores.InstanceID // 有序容器成員(座位除外); 第 1 個 = 頂端 / 隊首
	seat   map[int32]cores.InstanceID                 // 座位編號 → 在座顧客
	effect []*effectView                              // 效果佇列視圖(佇列序; 顯示時依作用順序另排)
	action []*actionView                              // 行動佇列視圖(先進先出)
	morph  cores.ContainerKind                        // morph 配對暫存: 舊實例所在容器(ContainerNone = 無暫存)
	slot   int                                        // morph 配對暫存: 舊實例容器索引
}

func newMirror(sheet *sheeter.Sheeter) *mirror {
	return &mirror{
		sheet: sheet,
		attr:  map[string]float64{},
		lock:  map[string]float64{},
		card:  map[cores.InstanceID]*cardView{},
		guest: map[cores.InstanceID]*guestView{},
		zone:  map[cores.ContainerKind][]cores.InstanceID{},
		seat:  map[int32]cores.InstanceID{},
	}
}

// Apply 摺疊一筆事件: 座標一律更新, 之後依類別分派(scope / select / phase 無盤面投影, 歸 M22 日誌)。
func (this *mirror) Apply(eventData cores.EventData) {
	this.round = eventData.Round
	this.phase = eventData.Phase

	switch eventData.Kind {
	case cores.EventProperty:
		this.applyProperty(eventData)

	case cores.EventContainer:
		this.applyContainer(eventData)

	case cores.EventInstance:
		this.applyInstance(eventData)

	case cores.EventEffect:
		this.applyEffect(eventData)

	case cores.EventAction:
		this.applyAction(eventData)

	default:
		// scope / select / phase: 無盤面投影
	} // switch
}

// applyProperty 屬性摺疊: 對象欄零值 = 全域、否則路由至實例視圖; @ / # 入鎖定計數表、其餘入值表。
func (this *mirror) applyProperty(eventData cores.EventData) {
	attr, lock := this.attr, this.lock

	if eventData.InstanceID != cores.NoneID {
		if view, ok := this.card[eventData.InstanceID]; ok {
			attr, lock = view.attr, view.lock
		} else if view, ok := this.guest[eventData.InstanceID]; ok {
			attr, lock = view.attr, view.lock
		} else {
			return // 未知實例(防禦) → 不投影
		} // if
	} // if

	switch eventData.Op {
	case cores.AssignLock, cores.AssignUnlock:
		lock[eventData.Attr] = eventData.After

	default:
		attr[eventData.Attr] = eventData.After
	} // switch
}

// applyContainer 容器摺疊: From == To 為重整(Pick 載全序, 整堆置換); 其餘為搬移——
// 出生(From None)建視圖(初值查表)、銷毀(To None)刪視圖; 插入端照【營業規格書 | 六、容器結構】編死
// (卡牌四牌堆前插 = 新進入者置頂、排隊 / 遊蕩 / 卡牌化尾插、座位走座位表)。
func (this *mirror) applyContainer(eventData cores.EventData) {
	if eventData.From == eventData.To {
		this.zone[eventData.To] = pickIDList(eventData.Pick) // 重整快照(M21 拍板⑤)
		return
	} // if

	if eventData.From == cores.ContainerNone { // 出生: 建視圖
		if cardZone(eventData.To) {
			this.card[eventData.InstanceID] = newCardView(this.sheet, eventData.DataID)
		} else {
			this.guest[eventData.InstanceID] = newGuestView(this.sheet, eventData.DataID)
		} // if
	} // if

	this.zoneRemove(eventData.From, eventData.InstanceID)

	if eventData.From == cores.ContainerCardify { // 離開卡牌化列表 = 綁定生命週期結束; 還原解綁無事件, 以此推斷
		for _, itor := range this.card {
			if itor.bindInstanceID == eventData.InstanceID {
				itor.bindID, itor.bindInstanceID = 0, cores.NoneID
			} // if
		} // for
	} // if

	switch {
	case eventData.To == cores.ContainerNone: // 銷毀: 刪視圖
		delete(this.card, eventData.InstanceID)
		delete(this.guest, eventData.InstanceID)

	case eventData.To == cores.ContainerSeat:
		this.seat[eventData.SeatID] = eventData.InstanceID

	case cardZone(eventData.To): // 卡牌四牌堆: 前插, 新進入者置頂
		this.zone[eventData.To] = append([]cores.InstanceID{eventData.InstanceID}, this.zone[eventData.To]...)

	default: // 排隊 / 遊蕩 / 卡牌化: 尾插(隊尾 / 集合追加)
		this.zone[eventData.To] = append(this.zone[eventData.To], eventData.InstanceID)
	} // switch

	if view, ok := this.card[eventData.InstanceID]; ok { // 綁定語境欄摺入; 無綁定事件載零值, 等價未綁
		view.bindID, view.bindInstanceID = eventData.BindID, eventData.BindInstanceID
	} // if
}

// applyInstance 實例摺疊(morph 唯一真身: 銷毀舊 + 建立新兩發、位置不變): 銷毀記下容器槽位並移除,
// 建立以同槽位插回(初值查表); 無暫存槽位的建立(防禦)只建視圖不入容器。
func (this *mirror) applyInstance(eventData cores.EventData) {
	if eventData.Alive == false {
		this.morph, this.slot = this.zoneFind(eventData.InstanceID)
		this.zoneRemove(this.morph, eventData.InstanceID)
		delete(this.card, eventData.InstanceID)
		return
	} // if

	this.card[eventData.InstanceID] = newCardView(this.sheet, eventData.DataID)

	if this.morph != cores.ContainerNone {
		zone := this.zone[this.morph]
		index := min(this.slot, len(zone))
		this.zone[this.morph] = append(zone[:index], append([]cores.InstanceID{eventData.InstanceID}, zone[index:]...)...)
		this.morph, this.slot = cores.ContainerNone, 0
	} // if
}

// applyEffect 效果佇列摺疊: 加入 → upsert 快照(層數 / 結束回合 / self); 結束 → 退場(Alive false)出列、
// 退層(Alive true)更新快照; 其餘階段(立即 / 觸發 / 啟動 / 條件不成立)不動佇列。
func (this *mirror) applyEffect(eventData cores.EventData) {
	switch eventData.Stage {
	case cores.EffectStageJoin:
		for _, itor := range this.effect {
			if itor.instanceID == eventData.EffectInstanceID {
				itor.stack, itor.expire = eventData.Stack, eventData.Expire
				return
			} // if
		} // for

		this.effect = append(this.effect, &effectView{
			effectID:       eventData.EffectID,
			instanceID:     eventData.EffectInstanceID,
			stack:          eventData.Stack,
			expire:         eventData.Expire,
			selfID:         eventData.DataID,
			selfInstanceID: eventData.InstanceID,
		})

	case cores.EffectStageEnd:
		if eventData.Alive {
			for _, itor := range this.effect {
				if itor.instanceID == eventData.EffectInstanceID {
					itor.stack, itor.expire = eventData.Stack, eventData.Expire
				} // if
			} // for

			return
		} // if

		result := []*effectView{}

		for _, itor := range this.effect {
			if itor.instanceID != eventData.EffectInstanceID {
				result = append(result, itor)
			} // if
		} // for

		this.effect = result

	default:
		// 立即 / 觸發 / 啟動 / 條件不成立: 不動佇列
	} // switch
}

// applyAction 行動佇列摺疊: 入列尾插、出列移除首個匹配項(彈出為先進先出, 必為隊首; 掃描為防禦)。
func (this *mirror) applyAction(eventData cores.EventData) {
	if eventData.Alive {
		this.action = append(this.action, &actionView{
			guestID:         eventData.DataID,
			guestInstanceID: eventData.InstanceID,
			skillID:         eventData.SkillID,
			task:            eventData.Task,
		})
		return
	} // if

	for itor := range this.action {
		if this.action[itor].guestInstanceID == eventData.InstanceID && this.action[itor].skillID == eventData.SkillID && this.action[itor].task == eventData.Task {
			this.action = append(this.action[:itor], this.action[itor+1:]...)
			return
		} // if
	} // for
}

// hasEffect 回報實例身上是否有 active 效果(座位區 效 旗標用; 以效果佇列項 self 比對)。
func (this *mirror) hasEffect(instanceID cores.InstanceID) bool {
	for _, itor := range this.effect {
		if itor.selfInstanceID == instanceID {
			return true
		} // if
	} // for

	return false
}

// zoneFind 找實例所在的有序容器與索引(morph 限卡牌, 不查座位表); 未命中回 ContainerNone。
func (this *mirror) zoneFind(instanceID cores.InstanceID) (kind cores.ContainerKind, index int) {
	for k, v := range this.zone {
		for itor := range v {
			if v[itor] == instanceID {
				return k, itor
			} // if
		} // for
	} // for

	return cores.ContainerNone, 0
}

// zoneRemove 自容器移除實例(座位清空座位表項、有序容器去元素、ContainerNone 不動作)。
func (this *mirror) zoneRemove(kind cores.ContainerKind, instanceID cores.InstanceID) {
	if kind == cores.ContainerNone {
		return
	} // if

	if kind == cores.ContainerSeat {
		for k, v := range this.seat {
			if v == instanceID {
				delete(this.seat, k)
			} // if
		} // for

		return
	} // if

	result := []cores.InstanceID{}

	for _, itor := range this.zone[kind] {
		if itor != instanceID {
			result = append(result, itor)
		} // if
	} // for

	this.zone[kind] = result
}

// cardZone 回報容器是否屬卡牌四牌堆(出生視圖型別與插入端判定用)。
func cardZone(kind cores.ContainerKind) bool {
	return kind == cores.ContainerHand || kind == cores.ContainerDeck || kind == cores.ContainerDrop || kind == cores.ContainerExile
}

// pickIDList 自重整快照取出實例編號序列。
func pickIDList(pick []cores.PickData) (result []cores.InstanceID) {
	result = make([]cores.InstanceID, 0, len(pick))

	for _, itor := range pick {
		result = append(result, itor.InstanceID)
	} // for

	return result
}
