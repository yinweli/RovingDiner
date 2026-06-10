package cores

// Card 卡牌實例;對應【營業規格書 | 五、實例結構 | 卡牌（Card）實例】。
type Card struct {
	instanceID  InstanceID // 實例唯一識別碼（變身會重分配;改寫入口僅 NewCard / Morph）
	cardID      int32      // 引用對應卡牌資料
	cost        Value      // 出牌費用
	extraRunMin Value      // 額外發動次數下限
	extraRunMax Value      // 額外發動次數上限
	keep        Value      // 不棄卡牌（Value 固定 0，狀態於 Lock）
	seal        Value      // 封印卡牌（Value 固定 0，狀態於 Lock）
	playExile   Value      // 出牌後流放（Value 固定 0，狀態於 Lock）
	unplayExile Value      // 未出牌流放（Value 固定 0，狀態於 Lock）
	effectID    IDList     // 實例效果列表（效果編號多重集合;允許重複）
	cardify     *Guest     // 卡牌化來源顧客（初值 nil;與不棄鎖成對,改寫入口僅 CardifyBind / CardifyFree）
}

// NewCard 依卡牌編號實例化新卡（載卡牌資料初始值;bool 欄 → 鎖定計數、SkillID → Skill.EffectID）;資料不存在回 nil。
func NewCard(eng *Engine, cardID int32) *Card {
	meta := eng.data.Card.Get(cardID)

	if meta == nil {
		return nil
	} // if

	return &Card{
		instanceID:  eng.runtime.NextID(),
		cardID:      cardID,
		cost:        NewValue(meta.Cost, 0),
		extraRunMin: NewValue(meta.ExtraRunMin, 0),
		extraRunMax: NewValue(meta.ExtraRunMax, 0),
		keep:        NewValuel(meta.Keep),
		seal:        NewValuel(meta.Seal),
		playExile:   NewValuel(meta.PlayExile),
		unplayExile: NewValuel(meta.UnplayExile),
		effectID:    NewIDList(skillEffect(eng, meta.SkillID)...),
	}
}

// CopyCard 複製卡牌：淺複製依 source 的卡牌編號載入卡牌資料初始值、深複製複製 source 當前狀態（效果列表深複製）;
// 卡牌化來源皆 none、實例編號重生。淺複製於卡牌資料不存在時回 nil。
func CopyCard(eng *Engine, source *Card, deep bool) *Card {
	if deep == false {
		return NewCard(eng, source.cardID)
	} // if

	return &Card{
		instanceID:  eng.runtime.NextID(),
		cardID:      source.cardID,
		cost:        source.cost,
		extraRunMin: source.extraRunMin,
		extraRunMax: source.extraRunMax,
		keep:        source.keep,
		seal:        source.seal,
		playExile:   source.playExile,
		unplayExile: source.unplayExile,
		effectID:    NewIDList(source.effectID.List()...),
	}
}

// GetInstanceID 讀實例編號。
func (this *Card) GetInstanceID() InstanceID {
	return this.instanceID
}

// GetCardID 讀卡牌編號。
func (this *Card) GetCardID() int32 {
	return this.cardID
}

// GetCost 取出牌費用;讀寫紀律（鎖定守衛）由 Value 把關。
func (this *Card) GetCost() *Value {
	return &this.cost
}

// GetExtraRunMin 取額外發動次數下限。
func (this *Card) GetExtraRunMin() *Value {
	return &this.extraRunMin
}

// GetExtraRunMax 取額外發動次數上限。
func (this *Card) GetExtraRunMax() *Value {
	return &this.extraRunMax
}

// GetKeep 取不棄卡牌。
func (this *Card) GetKeep() *Value {
	return &this.keep
}

// GetSeal 取封印卡牌。
func (this *Card) GetSeal() *Value {
	return &this.seal
}

// GetPlayExile 取出牌後流放。
func (this *Card) GetPlayExile() *Value {
	return &this.playExile
}

// GetUnplayExile 取未出牌流放。
func (this *Card) GetUnplayExile() *Value {
	return &this.unplayExile
}

// GetEffectID 取實例效果列表。
func (this *Card) GetEffectID() *IDList {
	return &this.effectID
}

// GetCardify 讀卡牌化來源顧客（未綁定回 nil）。
func (this *Card) GetCardify() *Guest {
	return this.cardify
}

// CardifyBind 綁定卡牌化來源並鎖不棄（cardify 處理流程 step 5+6 成對紀律:
// 不棄鎖即為保護綁定卡而存在,【營業規格書 | 二十五、操作命令清單 | cardify】）。
func (this *Card) CardifyBind(guest *Guest) {
	this.cardify = guest
	this.keep.Lock()
}

// CardifyFree 解綁卡牌化來源並解鎖不棄（restore 處理流程 step 6+7;Value.Unlock 夾 ≥ 0）。
func (this *Card) CardifyFree() {
	this.cardify = nil
	this.keep.Unlock()
}

// Morph 變身重設（【營業規格書 | 二十五、操作命令清單 | *Morph】step 4）:重分配實例編號（舊編號失效）、
// 換卡牌編號、依新卡資料僅載 出牌費用 / 不棄鎖 / 封印鎖 / 實例效果列表;查無卡牌資料回 false 不動。
// 實例身分的改寫入口僅 NewCard 與此;抽獎 / 事件 / 觸發留 morph 流程。
func (this *Card) Morph(eng *Engine, cardID int32) bool {
	meta := eng.data.Card.Get(cardID)

	if meta == nil {
		return false
	} // if

	this.instanceID = eng.runtime.NextID()
	this.cardID = cardID
	this.cost = NewValue(meta.Cost, 0)
	this.keep = NewValuel(meta.Keep)
	this.seal = NewValuel(meta.Seal)
	this.effectID = NewIDList(skillEffect(eng, meta.SkillID)...)
	return true
}

// CardList 卡牌容器;四牌堆共用（先進後出,新進入者置頂;手牌「玩家檢視序」現行同為前端插入）。
// 對應【營業規格書 | 六、容器結構】手牌 / 抽牌牌堆 / 棄牌牌堆 / 流放牌堆。
// 洗牌（要 Rander）與 deckTop auto-shuffle 為 eng 依賴流程,維持自由函式 / selector。
type CardList []*Card

// Push 加入牌堆頂端（前端）。
func (this *CardList) Push(card *Card) {
	*this = append(CardList{card}, *this...)
}

// Remove 依實例編號移除指定卡牌,其餘元素保持原序;未命中不變。
func (this *CardList) Remove(instanceID InstanceID) {
	result := CardList{}

	for _, itor := range *this {
		if itor.instanceID != instanceID {
			result = append(result, itor)
		} // if
	} // for

	*this = result
}

// Find 依實例編號找出卡牌;未命中回 nil。供 locateCard 逐牌堆掃描。
func (this *CardList) Find(instanceID InstanceID) *Card {
	for _, itor := range *this {
		if itor.instanceID == instanceID {
			return itor
		} // if
	} // for

	return nil
}
