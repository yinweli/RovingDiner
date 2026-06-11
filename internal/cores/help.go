package cores

// Immune 免疫群組計數(群組編號 → 鎖定計數); 效果免疫 / 技能免疫共用
// (【營業規格書 | 五、實例結構 | 顧客（Guest）實例】)。零值可用(Add 自建表)。
type Immune struct {
	count map[int32]int32 // 群組編號 -> 鎖定計數
}

// NewImmune 建構空免疫計數。
func NewImmune() Immune {
	return Immune{count: map[int32]int32{}}
}

// Add 對群組鎖定計數 +1; 表未建時自建。
func (this *Immune) Add(group int32) {
	if this.count == nil {
		this.count = map[int32]int32{}
	} // if

	this.count[group]++
}

// Del 對群組鎖定計數 -1, 夾 ≥ 0(無鍵 / 空表自然視為 0)。
func (this *Immune) Del(group int32) {
	if this.count[group] > 0 {
		this.count[group]--
	} // if
}

// Get 讀群組鎖定計數; 無鍵回 0(>0 / ==0 判斷由呼叫端比較)。
func (this *Immune) Get(group int32) int32 {
	return this.count[group]
}

// Any 回報任一群組鎖定計數 > 0(顯示端旗標判定用; 歸零鍵與空表回 false)。
func (this *Immune) Any() bool {
	for _, v := range this.count {
		if v > 0 {
			return true
		} // if
	} // for

	return false
}

// Hit 已觸發門檻集合(門檻值 → 已觸發); 飽食 / 耐心門檻共用, 擋同門檻重複觸發
// (【營業規格書 | 二十、獨立流程 | 執行結算】)。零值可用(Add 自建表)。
type Hit struct {
	hit map[int32]bool // 門檻值 -> 已觸發
}

// NewHit 建構空門檻集合。
func NewHit() Hit {
	return Hit{hit: map[int32]bool{}}
}

// Add 標記門檻已觸發(M16 執行結算用); 表未建時自建。
func (this *Hit) Add(threshold int32) {
	if this.hit == nil {
		this.hit = map[int32]bool{}
	} // if

	this.hit[threshold] = true
}

// IsHit 回報門檻是否已觸發。
func (this *Hit) IsHit(threshold int32) bool {
	return this.hit[threshold]
}

// Count 讀已觸發門檻數量(sateHit / calmHit 屬性)。
func (this *Hit) Count() int32 {
	return int32(len(this.hit))
}

// IDList 編號列表(有序多重集合, 允許重複); 卡牌實例效果列表用, 零值可用。
type IDList struct {
	id []int32 // 編號列表(保持加入順序)
}

// NewIDList 以編號集建構列表(複製輸入, 不共享底層)。
func NewIDList(id ...int32) IDList {
	return IDList{id: append([]int32(nil), id...)}
}

// Add 尾端加入(可變參數, 批次附加共用)。
func (this *IDList) Add(id ...int32) {
	this.id = append(this.id, id...)
}

// DelOne 移除第一個 == id 者; 無命中原樣不動。
func (this *IDList) DelOne(id int32) {
	for itor, value := range this.id {
		if value == id {
			this.id = append(this.id[:itor], this.id[itor+1:]...)
			return
		} // if
	} // for
}

// DelAll 移除全部 == id 者。
func (this *IDList) DelAll(id int32) {
	result := []int32{}

	for _, itor := range this.id {
		if itor != id {
			result = append(result, itor)
		} // if
	} // for

	this.id = result
}

// Count 計數 == id 的個數。
func (this *IDList) Count(id int32) int32 {
	result := int32(0)

	for _, itor := range this.id {
		if itor == id {
			result++
		} // if
	} // for

	return result
}

// List 取底層編號列表供迭代(呼叫端唯讀約定)。
func (this *IDList) List() []int32 {
	return this.id
}

// Tally 分組累積計數(群組編號 → 數量); 抽牌 / 棄牌 / 出牌 / 流放的整場累積多重集合共用
// (【營業規格書 | 五、實例結構 | 營業（Game）實例】)。零值可用(Add 自建表)。
type Tally struct {
	count map[int32]int32 // 群組編號 -> 數量
}

// NewTally 建構空累積計數。
func NewTally() Tally {
	return Tally{count: map[int32]int32{}}
}

// Add 對群組數量 +1; 表未建時自建。
func (this *Tally) Add(group int32) {
	if this.count == nil {
		this.count = map[int32]int32{}
	} // if

	this.count[group]++
}

// Get 讀群組數量; 無鍵回 0。
func (this *Tally) Get(group int32) int32 {
	return this.count[group]
}

// Sum 讀全群組數量加總(查詢函式 N == 0 全加總用)。
func (this *Tally) Sum() (sum int32) {
	for _, v := range this.count {
		sum += v
	} // for

	return sum
}

// Reset 清空全部計數(營業開始清空整場累積; 【營業規格書 | 十九、核心流程 | 1. 營業開始階段】)。
func (this *Tally) Reset() {
	this.count = map[int32]int32{}
}
