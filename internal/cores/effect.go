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

// effectOrder 取效果的作用順序（【營業規格書 | 十五、作用順序】）;查無編譯資料回 0（防禦;正常實例其 effectData 必存在）。
func effectOrder(eng *Engine, effect *Effect) int32 {
	meta, ok := eng.effect[effect.EffectID]

	if ok == false {
		return 0
	} // if

	return meta.RunOrder
}

// effectExpire 依作用回合算結束回合（【營業規格書 | 十四、作用回合】:0 → 0 整場保留、N → 當前回合 + N − 1）。供 newEffect 建立與 effectStack 刷新共用。
func effectExpire(eng *Engine, runRound int32) int32 {
	if runRound <= 0 {
		return 0
	} // if

	return eng.runtime.Game.Round + runRound - 1
}
