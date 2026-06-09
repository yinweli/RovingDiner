package cores

import (
	"sort"
)

// effectPush 把效果加入效果佇列;效果佇列順序無關（【營業規格書 | 六、容器結構】），故 append，處理時才依作用順序排序。
func effectPush(eng *Engine, effect *Effect) {
	eng.runtime.Effect = append(eng.runtime.Effect, effect)
}

// effectRemove 依實例編號自效果佇列移除指定效果;供觸發後移除 / 推進效果 / 清理效果共用（【營業規格書 | 二十、獨立流程】）。
func effectRemove(eng *Engine, effect *Effect) {
	eng.runtime.Effect = removeEffect(eng.runtime.Effect, effect)
}

// effectSort 依【營業規格書 | 十五、作用順序】就地排序效果列表:作用順序大者優先，同作用順序時效果編號小者優先;
// 觸發時機 / 推進效果 / 清理效果三流程共用此排序。
func effectSort(eng *Engine, effect []*Effect) {
	sort.SliceStable(effect, func(i, j int) bool {
		left, right := effectOrder(eng, effect[i]), effectOrder(eng, effect[j])

		if left != right {
			return left > right
		} // if

		return effect[i].EffectID < effect[j].EffectID
	})
}

// effectOrder 取效果的作用順序（【營業規格書 | 十五、作用順序】）;查無效果資料回 0（防禦;正常實例由 newEffect 保證資料存在）。
func effectOrder(eng *Engine, effect *Effect) int32 {
	meta := eng.data.Effect.Get(effect.EffectID)

	if meta == nil {
		return 0
	} // if

	return meta.RunOrder
}
