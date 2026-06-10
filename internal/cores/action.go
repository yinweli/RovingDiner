package cores

// Action 行動實例;「顧客 + 行動類型 + 技能編號」三元組,建立後不可變。
// 對應【營業規格書 | 五、實例結構 | 行動（Action）實例】。
type Action struct {
	guest   *Guest   // 顧客實例
	kind    TaskKind // 行動類型（飽食 / 耐心）
	skillID int32    // 顧客行動階段彈出後啟動的技能編號
}

// NewAction 建構行動實例。
func NewAction(guest *Guest, kind TaskKind, skillID int32) *Action {
	return &Action{
		guest:   guest,
		kind:    kind,
		skillID: skillID,
	}
}

// GetGuest 讀顧客實例。
func (this *Action) GetGuest() *Guest {
	return this.guest
}

// GetKind 讀行動類型。
func (this *Action) GetKind() TaskKind {
	return this.kind
}

// GetSkillID 讀技能編號。
func (this *Action) GetSkillID() int32 {
	return this.skillID
}

// ActionList 行動佇列(先進先出);對應【營業規格書 | 六、容器結構 | 行動佇列】。
// 具名 slice:唯讀操作(長度 / 迭代)直接用語言內建,僅佇列紀律(尾入 / 首出)收為方法。
type ActionList []*Action

// Push 加入佇列尾端。
func (this *ActionList) Push(action *Action) {
	*this = append(*this, action)
}

// Pop 彈出佇列首位;空佇列回 nil。
func (this *ActionList) Pop() *Action {
	if len(*this) == 0 {
		return nil
	} // if

	result := (*this)[0]
	*this = (*this)[1:]
	return result
}
