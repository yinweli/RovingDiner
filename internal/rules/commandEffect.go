package rules

import (
	"slices"

	"github.com/yinweli/RovingDiner/internal/cores"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

// 效果佇列命令(【營業規格書 | 二十五、操作命令清單】effectClear / effectDel / effectRun): 核心即效果佇列的三詞條, M9 延後、M13 接回。

// commandEffectClear 清除群組效果: 自佇列清除「self ∈ 命令對象 且 效果群組編號 ∈ 指定群組」的效果(退場: 結束命令 × 層數 → 出佇列)。
// target == nil(對象 none)為全域掃描、非 nil 空集合 → no-op(ExecOperate 正規化約定, 見 cores.SelectorNone);
// 群組 0 不命中任何效果; 同一效果命中多個指定群組只清一次(單一群組欄、自然滿足);
// 凍結中顧客的效果不清(M13 拍板: 效果凍結擴及全域掃描)。參數: 效果群組編號 1..N(varargs)。
func commandEffectClear(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	group := intList(arg)
	hit := cores.EffectList{}

	for _, itor := range game.Effect {
		meta, ok := game.EffectData(itor.GetEffectID())

		if ok == false || frozenSelf(game, itor) {
			continue // 查無編譯資料(防禦)/ 效果凍結 → 不清
		} // if

		if meta.Group == 0 || slices.Contains(group, meta.Group) == false {
			continue // 群組 0 不命中 / 不具指定群組 → 不清
		} // if

		if target != nil && selfIn(itor, target) == false {
			continue // 非全域掃描時, self 不屬於命令對象集合 → 不清
		} // if

		hit = append(hit, itor)
	} // for

	hit.Sort(game) // 快照排序; 不擾動 game.Effect(順序無關)

	for _, itor := range hit {
		retireEffect(game, itor)
	} // for
}

// commandEffectDel 移除效果: 對每個命令對象元素以其為 self 查同份效果, 層數 -= min(N, 層數)、結束命令 × 實際扣減數、
// 層數歸 0 → 出佇列; 同份效果不存在 / N <= 0 → no-op。參數: 效果編號、N。
func commandEffectDel(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	effectID, okID := argInt(arg)
	n, okN := argInt(argTail(arg, 1))

	if okID == false || okN == false || n <= 0 {
		return // 缺參數 / N <= 0 → no-op
	} // if

	meta, ok := game.EffectData(effectID)

	if ok == false {
		return // 查無編譯資料(防禦)→ no-op
	} // if

	for _, itor := range target {
		self, found := instanceRef(game, itor)

		if found == false {
			continue // 實例不存在 → 該項 no-op
		} // if

		effect := game.Effect.Find(self, effectID)

		if effect == nil {
			continue // 同份效果不存在 → 該項 no-op
		} // if

		sub := effect.StackSub(n)
		runEffectEnd(game, effect, meta.End, sub, effect.GetStack() > 0) // 退層留佇列 / 歸零出佇列(快照載退後層數; M21 拍板)

		if effect.GetStack() <= 0 {
			game.Effect.Remove(effect.GetInstanceID())
		} // if
	} // for
}

// commandEffectRun 啟動效果: 對每個命令對象元素以其為 self 依效果類型派發(M13 拍板: dispatchEffect 帶 N override——
// 立即類型過觸發條件後跑一次立即命令、忽略 N; 觸發 / 常駐套 [堆疊處理], N > 0 覆寫本次增加層數、N <= 0 用效果堆疊層數;
// 常駐每實際增一層跑一次啟動命令)。只忽略目標類型 + 目標數量, 其他設定全部保留。參數: 效果編號、N。
func commandEffectRun(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	effectID, okID := argInt(arg)
	n, okN := argInt(argTail(arg, 1))

	if okID == false || okN == false {
		return // 缺參數 → no-op
	} // if

	meta, ok := game.EffectData(effectID)

	if ok == false {
		return // 查無編譯資料(防禦)→ no-op
	} // if

	for _, itor := range target {
		self, found := instanceRef(game, itor)

		if found == false {
			continue // 實例不存在 → 該項 no-op
		} // if

		dispatchEffect(game, meta, effectID, self, n)
	} // for
}
