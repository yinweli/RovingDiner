package cores

// effectStack 套用 [堆疊處理]（【營業規格書 | 二十、獨立流程 | 堆疊處理】）:把效果以 self 加入效果佇列,或對佇列既有同份疊層。
// override > 0 為 effectRun 命令指定的增量;否則用效果堆疊層數（0 / 1 → 1）。回實際增加層數（供常駐 啟動命令判斷）。
func effectStack(eng *Engine, self Self, effectID, override int32) (added int32) {
	meta, ok := eng.effect[effectID]

	if ok == false {
		return 0 // 查無編譯資料 → no-op（防禦）
	} // if

	if self.Guest != nil && self.Guest.EffectImmune[meta.Group] > 0 {
		return 0 // 顧客免疫該效果群組 → 不建立 / 不堆疊 / 不執行啟動命令
	} // if

	add := override

	if add <= 0 {
		add = max(meta.Stack, 1) // 堆疊層數 0 / 1 視為 1
	} // if

	existing := findStack(eng, self, effectID)

	if existing == nil {
		layer := capStack(add, meta.StackMax)
		effectPush(eng, newEffect(eng, effectID, self, layer))

		return layer
	} // if

	before := existing.Stack
	existing.Stack = capStack(existing.Stack+add, meta.StackMax)

	if meta.StackTime == StackTimeRefresh {
		existing.Expire = effectExpire(eng, meta.RunRound) // 刷新時重算結束回合（不變則維持原值）
	} // if

	return existing.Stack - before
}

// findStack 自效果佇列找與 (effectID, self) 同份的效果（效果編號同 && self 同;無目標的 self 視為彼此相同）;無回 nil。
func findStack(eng *Engine, self Self, effectID int32) *Effect {
	for _, itor := range eng.runtime.Effect {
		if itor.EffectID == effectID && sameSelf(itor.Self, self) {
			return itor
		} // if
	} // for

	return nil
}

// sameSelf 比對兩個 self 是否同份:同為顧客 / 卡牌時比實例編號,同為無目標（空物件）時相等,型別不符不相等。
func sameSelf(a, b Self) bool {
	if a.Guest != nil && b.Guest != nil {
		return a.Guest.InstanceID == b.Guest.InstanceID
	} // if

	if a.Card != nil && b.Card != nil {
		return a.Card.InstanceID == b.Card.InstanceID
	} // if

	return a.IsNone() && b.IsNone()
}

// capStack 依堆疊上限夾層數（堆疊上限 0 = 無上限,不夾）。
func capStack(layer, stackMax int32) int32 {
	if stackMax > 0 && layer > stackMax {
		return stackMax
	} // if

	return layer
}
