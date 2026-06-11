package cores

import (
	"sort"
)

// Guest 顧客實例; 對應【營業規格書 | 五、實例結構 | 顧客（Guest）實例】。
//
// SateMax(飽食值離場線)為【營業規格書 | 二十三、屬性清單 | 顧客引用屬性】登記的
// 「寫鎖」屬性 sateMax, 故以 Value 儲存; §五 顧客實例表目前漏列此欄(待規格補正)。
type Guest struct {
	instanceID   InstanceID // 實例唯一識別碼
	guestID      int32      // 引用對應顧客資料
	seatID       int32      // 占用的座位編號(非座位列表時為 0; 成對寫入收於 SeatList.Place / Remove)
	score        Value      // 滿意值
	scoreMax     Value      // 滿意值上限
	morale       Value      // 士氣值
	moraleMax    Value      // 士氣值上限
	sate         Value      // 飽食值
	sateMax      Value      // 飽食值離場線
	calm         Value      // 耐心值
	sateSeal     Value      // 封印飽食技能(Value 固定 0, 狀態於 Lock)
	calmSeal     Value      // 封印耐心技能(Value 固定 0, 狀態於 Lock)
	sateHit      Hit        // 已觸發飽食門檻集合(避免重複觸發)
	calmHit      Hit        // 已觸發耐心門檻集合
	effectImmune Immune     // 效果免疫群組(效果群組編號 -> 鎖定計數)
	skillImmune  Immune     // 技能免疫群組(技能群組編號 -> 鎖定計數)
	freeze       int32      // 凍結起始回合(卡牌化時記錄; 解凍後重置 0)
}

// NewGuest 依顧客編號實例化新顧客(載顧客資料初始值: Score / ScoreMax / Morale / MoraleMax / Calm / SateMax 數值、封印 bool → 鎖定計數);
// Sate 初值 0(顧客資料無此欄、隨服務累積至飽食值離場線); Hit / Immune 初始化空表。資料不存在回 nil。
func NewGuest(game *Game, guestID int32) *Guest {
	meta := game.GetSheet().Guest.Get(guestID)

	if meta == nil {
		return nil
	} // if

	return &Guest{
		instanceID:   game.NextID(),
		guestID:      guestID,
		score:        NewValue(meta.Score, 0),
		scoreMax:     NewValue(meta.ScoreMax, 0),
		morale:       NewValue(meta.Morale, 0),
		moraleMax:    NewValue(meta.MoraleMax, 0),
		calm:         NewValue(meta.Calm, 0),
		sateMax:      NewValue(meta.SateMax, 0),
		sateSeal:     NewValueLock(meta.SateSeal),
		calmSeal:     NewValueLock(meta.CalmSeal),
		sateHit:      NewHit(),
		calmHit:      NewHit(),
		effectImmune: NewImmune(),
		skillImmune:  NewImmune(),
	}
}

// GetInstanceID 讀實例編號。
func (this *Guest) GetInstanceID() InstanceID {
	return this.instanceID
}

// GetGuestID 讀顧客編號。
func (this *Guest) GetGuestID() int32 {
	return this.guestID
}

// GetSeatID 讀占用的座位編號; 無 setter, 寫入收於 SeatList.Place / Remove 成對處理。
func (this *Guest) GetSeatID() int32 {
	return this.seatID
}

// GetScore 取滿意值; 讀寫紀律(鎖定守衛)由 Value 把關。
func (this *Guest) GetScore() *Value {
	return &this.score
}

// GetScoreMax 取滿意值上限。
func (this *Guest) GetScoreMax() *Value {
	return &this.scoreMax
}

// GetMorale 取士氣值。
func (this *Guest) GetMorale() *Value {
	return &this.morale
}

// GetMoraleMax 取士氣值上限。
func (this *Guest) GetMoraleMax() *Value {
	return &this.moraleMax
}

// GetSate 取飽食值。
func (this *Guest) GetSate() *Value {
	return &this.sate
}

// GetSateMax 取飽食值離場線。
func (this *Guest) GetSateMax() *Value {
	return &this.sateMax
}

// GetCalm 取耐心值。
func (this *Guest) GetCalm() *Value {
	return &this.calm
}

// GetSateSeal 取封印飽食技能。
func (this *Guest) GetSateSeal() *Value {
	return &this.sateSeal
}

// GetCalmSeal 取封印耐心技能。
func (this *Guest) GetCalmSeal() *Value {
	return &this.calmSeal
}

// GetSateHit 取已觸發飽食門檻集合。
func (this *Guest) GetSateHit() *Hit {
	return &this.sateHit
}

// GetCalmHit 取已觸發耐心門檻集合。
func (this *Guest) GetCalmHit() *Hit {
	return &this.calmHit
}

// GetEffectImmune 取效果免疫群組計數。
func (this *Guest) GetEffectImmune() *Immune {
	return &this.effectImmune
}

// GetSkillImmune 取技能免疫群組計數。
func (this *Guest) GetSkillImmune() *Immune {
	return &this.skillImmune
}

// GetFreeze 讀凍結起始回合。
func (this *Guest) GetFreeze() int32 {
	return this.freeze
}

// SetFreeze 寫凍結起始回合(卡牌化記錄 / 還原歸零)。
func (this *Guest) SetFreeze(round int32) {
	this.freeze = round
}

// RoamLock 遊蕩入列自動鎖: sate / sateSeal / calmSeal 鎖定計數各 +1(guestRoam / guestSpawn 入遊蕩共用,
// 【營業規格書 | 二十五、操作命令清單 | guestRoam】)。
func (this *Guest) RoamLock() {
	this.sate.Lock()
	this.sateSeal.Lock()
	this.calmSeal.Lock()
}

// RoamUnlock 解遊蕩自動鎖: 各 -1(Value.Unlock 夾 ≥ 0; guestReturn 回座)。
func (this *Guest) RoamUnlock() {
	this.sate.Unlock()
	this.sateSeal.Unlock()
	this.calmSeal.Unlock()
}

// WaitList 排隊佇列(先進先出 + 優先插隊); 對應【營業規格書 | 六、容器結構 | 排隊佇列】。
type WaitList []*Guest

// Insert 插入佇列前端(waitAdd 優先入座; 【營業規格書 | 二十五、操作命令清單 | waitAdd】)。
func (this *WaitList) Insert(guest *Guest) {
	*this = append(WaitList{guest}, *this...)
}

// Pop 彈出佇列首位; 空佇列回 nil。
func (this *WaitList) Pop() *Guest {
	if len(*this) == 0 {
		return nil
	} // if

	result := (*this)[0]
	*this = (*this)[1:]
	return result
}

// Remove 依實例編號移除指定顧客, 其餘元素保持原序; 未命中不變。
func (this *WaitList) Remove(instanceID InstanceID) {
	result := WaitList{}

	for _, itor := range *this {
		if itor.instanceID != instanceID {
			result = append(result, itor)
		} // if
	} // for

	*this = result
}

// Find 依實例編號查找顧客; 未命中回 nil。供 LocateGuest 掃描排隊佇列。
func (this *WaitList) Find(instanceID InstanceID) *Guest {
	for _, itor := range *this {
		if itor.instanceID == instanceID {
			return itor
		} // if
	} // for

	return nil
}

// SeatList 座位列表(座位編號 → 顧客; map 就地增刪故值接收器); 對應【營業規格書 | 六、容器結構 | 座位列表】。
type SeatList map[int32]*Guest

// Place 顧客入座: 設定顧客座位編號並登記至座位(成對寫入收斂)。
func (this SeatList) Place(seatID int32, guest *Guest) {
	guest.seatID = seatID
	this[seatID] = guest
}

// Remove 顧客離座: 自座位列表移除並歸零顧客座位編號(成對寫入收斂; 未入座者 seatID 0 無鍵, no-op)。
func (this SeatList) Remove(guest *Guest) {
	delete(this, guest.seatID)
	guest.seatID = 0
}

// Find 依實例編號查找在座顧客; 未命中回 nil。供 LocateGuest 掃描座位列表(編號唯一, map 迭代無序無礙)。
func (this SeatList) Find(instanceID InstanceID) *Guest {
	for _, itor := range this {
		if itor != nil && itor.instanceID == instanceID {
			return itor
		} // if
	} // for

	return nil
}

// Sorted 取全部入座顧客, 依座位編號升序(消除 map 迭代無序, 確保 Pick / Rand 候選決定性)。
func (this SeatList) Sorted() (result []*Guest) {
	seatID := make([]int32, 0, len(this))

	for k := range this {
		seatID = append(seatID, k)
	} // for

	sort.Slice(seatID, func(i, j int) bool { return seatID[i] < seatID[j] })

	for _, itor := range seatID {
		result = append(result, this[itor])
	} // for

	return result
}

// Occupied 計座位編號列表中當下有顧客占用的座位數。供 sameSize / nearSize 查同桌 / 鄰桌人數。
func (this SeatList) Occupied(seatID []int32) (count int32) {
	for _, itor := range seatID {
		if this[itor] != nil {
			count++
		} // if
	} // for

	return count
}

// GuestList 順序無關顧客列表; 遊蕩 / 卡牌化列表共用(【營業規格書 | 六、容器結構】)。
type GuestList []*Guest

// Push 加入列表尾端。
func (this *GuestList) Push(guest *Guest) {
	*this = append(*this, guest)
}

// Remove 依實例編號移除指定顧客, 其餘元素保持原序; 未命中不變。
func (this *GuestList) Remove(instanceID InstanceID) {
	result := GuestList{}

	for _, itor := range *this {
		if itor.instanceID != instanceID {
			result = append(result, itor)
		} // if
	} // for

	*this = result
}

// Find 依實例編號查找顧客; 未命中回 nil。供 LocateGuest 掃描遊蕩 / 卡牌化列表。
func (this *GuestList) Find(instanceID InstanceID) *Guest {
	for _, itor := range *this {
		if itor.instanceID == instanceID {
			return itor
		} // if
	} // for

	return nil
}
