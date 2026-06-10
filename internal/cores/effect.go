package cores

import (
	"sort"
)

// Effect 效果實例;對應【營業規格書 | 五、實例結構 | 效果（Effect）實例】。
// 效果的靜態類型（立即 / 觸發 / 常駐）見靜態表格 Effect.Kind 與 define.go 的 EffectKind。
type Effect struct {
	instanceID InstanceID // 實例唯一識別碼
	effectID   int32      // 引用對應效果資料
	expire     int32      // 結束回合（作用回合 = 0 時為 0，代表整場保留）
	stack      int32      // 當前堆疊層數
	self       Ref        // self 物件（空物件 / 顧客 / 卡牌）;建立時固定,之後不變
}

// NewEffect 依效果編號建構效果實例;結束回合依【營業規格書 | 十四、作用回合】（0 → 0 整場保留、N → 當前回合 + N − 1）;
// 當前層數由呼叫端決定（堆疊處理【營業規格書 | 十六、堆疊規則】）;查無編譯資料回 nil。
func NewEffect(eng *Engine, effectID int32, self Ref, stack int32) *Effect {
	meta, ok := eng.effect[effectID]

	if ok == false {
		return nil
	} // if

	return &Effect{
		instanceID: eng.runtime.NextID(),
		effectID:   effectID,
		expire:     effectExpire(eng, meta.RunRound),
		stack:      stack,
		self:       self,
	}
}

// GetInstanceID 讀實例編號。
func (this *Effect) GetInstanceID() InstanceID {
	return this.instanceID
}

// GetEffectID 讀效果編號。
func (this *Effect) GetEffectID() int32 {
	return this.effectID
}

// GetExpire 讀結束回合。
func (this *Effect) GetExpire() int32 {
	return this.expire
}

// GetStack 讀當前堆疊層數。
func (this *Effect) GetStack() int32 {
	return this.stack
}

// GetSelf 讀 self 物件。
func (this *Effect) GetSelf() Ref {
	return this.self
}

// SetExpire 覆寫結束回合;供堆疊刷新（絕對值）與凍結補回（GetExpire() + 差額）共用。
func (this *Effect) SetExpire(expire int32) {
	this.expire = expire
}

// StackAdd 增加堆疊層數並依堆疊上限夾制（0 = 無上限）;回實際增加層數（【營業規格書 | 十六、堆疊規則】）。
func (this *Effect) StackAdd(add, stackMax int32) int32 {
	before := this.stack
	this.stack = capStack(this.stack+add, stackMax)
	return this.stack - before
}

// EffectList 效果佇列(順序無關);對應【營業規格書 | 六、容器結構 | 效果佇列】。
// 處理時依作用順序排序快照(effectSort)、不就地排佇列;唯讀操作(長度 / 迭代)直接用語言內建。
type EffectList []*Effect

// Push 加入佇列（順序無關 → 尾端 append）。
func (this *EffectList) Push(effect *Effect) {
	*this = append(*this, effect)
}

// Remove 依實例編號移除指定效果,其餘元素保持原序;未命中不變。
// 供觸發後移除 / 推進效果 / 清理效果共用（【營業規格書 | 二十、獨立流程】）。
func (this *EffectList) Remove(instanceID InstanceID) {
	result := EffectList{}

	for _, itor := range *this {
		if itor.instanceID != instanceID {
			result = append(result, itor)
		} // if
	} // for

	*this = result
}

// Find 同份查找（效果編號同 && self 同,依 Ref.IsSame 比對,無目標的 self 視為彼此相同）;無回 nil。
// 供堆疊處理判定新建 / 疊層（【營業規格書 | 十六、堆疊規則】）。
func (this *EffectList) Find(self Ref, effectID int32) *Effect {
	for _, itor := range *this {
		if itor.effectID == effectID && itor.self.IsSame(self) {
			return itor
		} // if
	} // for

	return nil
}

// === eng 依賴的效果輔助（自由函式） ===

// effectSort 依【營業規格書 | 十五、作用順序】就地排序效果列表:作用順序大者優先，同作用順序時效果編號小者優先;
// 觸發時機 / 推進效果 / 清理效果三流程共用此排序。
func effectSort(eng *Engine, effect []*Effect) {
	sort.SliceStable(effect, func(i, j int) bool {
		left, right := effectOrder(eng, effect[i]), effectOrder(eng, effect[j])

		if left != right {
			return left > right
		} // if

		return effect[i].effectID < effect[j].effectID
	})
}

// effectOrder 取效果的作用順序（【營業規格書 | 十五、作用順序】）;查無編譯資料回 0（防禦;正常實例其 effectData 必存在）。
func effectOrder(eng *Engine, effect *Effect) int32 {
	meta, ok := eng.effect[effect.effectID]

	if ok == false {
		return 0
	} // if

	return meta.RunOrder
}

// effectExpire 依作用回合算結束回合（【營業規格書 | 十四、作用回合】:0 → 0 整場保留、N → 當前回合 + N − 1）。供 NewEffect 建立與 effectStack 刷新共用。
func effectExpire(eng *Engine, runRound int32) int32 {
	if runRound <= 0 {
		return 0
	} // if

	return eng.runtime.Game.GetRound().GetValue() + runRound - 1
}

// capStack 依堆疊上限夾層數（堆疊上限 0 = 無上限,不夾）。供 Effect.StackAdd 與堆疊處理新建路徑共用。
func capStack(layer, stackMax int32) int32 {
	if stackMax > 0 && layer > stackMax {
		return stackMax
	} // if

	return layer
}
