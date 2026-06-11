package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// effectStack 套用 [堆疊處理](【營業規格書 | 二十、獨立流程 | 堆疊處理】): 把效果以 self 加入效果佇列, 或對佇列既有同份疊層。
// override > 0 為 effectRun 命令指定的增量; 否則用效果堆疊層數(0 / 1 → 1)。回佇列實例與實際增加層數(供常駐 啟動命令判斷)。
// 效果事件: 新建入列 / 既有實際增層 / 刷新改變結束回合發 加入(載層數與結束回合快照; M21 拍板);
// 免疫 / 堆疊滿且結束回合不變 不發(無事不發; M18)。
func effectStack(game *cores.Game, self cores.Ref, effectID, override int32) (effect *cores.Effect, added int32) {
	meta, ok := game.EffectData(effectID)

	if ok == false {
		return nil, 0 // 查無編譯資料 → no-op(防禦)
	} // if

	if self.GetGuest() != nil && self.GetGuest().GetEffectImmune().Get(meta.Group) > 0 {
		return nil, 0 // 顧客免疫該效果群組 → 不建立 / 不堆疊 / 不執行啟動命令
	} // if

	add := override

	if add <= 0 {
		add = max(meta.Stack, 1) // 堆疊層數 0 / 1 視為 1
	} // if

	existing := game.Effect.Find(self, effectID)

	if existing == nil {
		effect = cores.NewEffect(game, effectID, self, add) // 建構即依堆疊上限夾制
		game.Effect.Push(effect)
		emitEffectState(game, effect, cores.EffectStageJoin, true)
		added = effect.GetStack()
		return effect, added
	} // if

	added = existing.StackAdd(add, meta.StackMax)
	expire := existing.GetExpire()

	if meta.StackTime == cores.StackTimeRefresh {
		existing.Refresh(game, meta.RunRound) // 刷新時重算結束回合, 不變則維持原值(先刷新再發, 快照載刷後值)
	} // if

	if added > 0 || existing.GetExpire() != expire {
		emitEffectState(game, existing, cores.EffectStageJoin, true) // 實際增層或刷新改變結束回合才發(堆疊滿且結束回合不變 = 無事不發; M21 拍板)
	} // if

	return existing, added
}
