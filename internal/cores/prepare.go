package cores

import (
	"sort"

	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// CompileCommand 把單一命令字串編譯為效果命令執行器;由 games 注入(cores 無命令解析能力)。語法錯回 error。
// prepareEffect 僅於命令欄非空時呼叫此原語,故空字串處理不在本型別契約內。
type CompileCommand func(source string) (command EffectCommand, err error)

// awardData 抽獎群組的平行候選(cardID 與 weight 一一對應、同序);供 weighted random 直接餵 Rander.Weighted。
type awardData struct {
	cardID []int32
	weight []int32
}

// prepareAward 自 Award 表建「群組編號 → 候選」衍生索引;依 Award.ID 排序確保決定性,僅收 Weight > 0 者(全 0 權重群組自然為空 → roll no-op)。
func prepareAward(data *sheeter.Sheeter) map[int32]awardData {
	result := map[int32]awardData{}

	if data == nil {
		return result
	} // if

	id := data.Award.Keys()
	sort.Slice(id, func(i, j int) bool { return id[i] < id[j] })

	for _, itor := range id {
		award := data.Award.Get(itor)

		if award.Weight <= 0 {
			continue
		} // if

		group := result[award.Group]
		group.cardID = append(group.cardID, award.CardID)
		group.weight = append(group.weight, award.Weight)
		result[award.Group] = group
	} // for

	return result
}

// effectData 效果靜態資料的預編譯形（【營業實作規格書 | 三、套件結構】effect 命令解耦）;cores.prepareEffect 一趟產出、NewEngine 內部建、cores 唯讀消費。
// M11 僅含觸發時機流程所需欄;立即 / 啟動 / 堆疊上限 / 堆疊時間 / 目標 於 M12 擴入。
type effectData struct {
	Kind         EffectKind    // 效果類型（篩選 觸發）
	TriggerKind  TriggerKind   // 觸發時機（篩選 時機）
	TriggerAfter TriggerAfter  // 觸發後行為（保留 / 移除）
	Cond         *exprs.Expr   // 觸發條件（空 → nil，視為恆成立）
	Count        *exprs.Expr   // 觸發次數（空 → nil，視為 1）
	Trigger      EffectCommand // 觸發命令
	End          EffectCommand // 結束命令
}

// prepareEffect 自 Effect 表建「效果編號 → 編譯形」衍生索引(對應【營業實作規格書 | 三、套件結構】effect 命令解耦):
// 命令欄交注入的 compile 編成閉包、cond/count 走 exprs.Parse、Kind/After 0 起算解碼 + 範圍檢查後就地組裝(比照 prepareAward 就地建 awardData)。
// 寬鬆策略(同 prepareAward 不回 error):任一欄編譯失敗 / 列舉越界 → 跳過該效果,嚴格把關交企劃驗證器於載入前。
// compile 僅於命令欄非空時呼叫;無命令資料的測試可傳 nil。data 為 nil 回空索引。
func prepareEffect(data *sheeter.Sheeter, compile CompileCommand) map[int32]effectData {
	result := map[int32]effectData{}

	if data == nil {
		return result
	} // if

	id := data.Effect.Keys()
	sort.Slice(id, func(i, j int) bool { return id[i] < id[j] })

	for _, itor := range id {
		meta := data.Effect.Get(itor)

		trigger, ok := compileCommand(compile, meta.CommandTrigger)

		if ok == false {
			continue
		} // if

		end, ok := compileCommand(compile, meta.CommandEnd)

		if ok == false {
			continue
		} // if

		cond, ok := compileExpr(meta.TriggerCond)

		if ok == false {
			continue
		} // if

		count, ok := compileExpr(meta.TriggerCount)

		if ok == false {
			continue
		} // if

		if meta.Kind < int32(EffectImmed) || meta.Kind > int32(EffectPersist) {
			continue // 效果類型編碼越界 → 跳過
		} // if

		if meta.TriggerAfter < int32(TriggerAfterKeep) || meta.TriggerAfter > int32(TriggerAfterRemove) {
			continue // 觸發後行為編碼越界 → 跳過
		} // if

		result[itor] = effectData{
			Kind:         EffectKind(meta.Kind),
			TriggerKind:  TriggerKind(meta.TriggerKind),
			TriggerAfter: TriggerAfter(meta.TriggerAfter),
			Cond:         cond,
			Count:        count,
			Trigger:      trigger,
			End:          end,
		}
	} // for

	return result
}

// compileCommand 把命令字串編成執行器:空字串 → nil(無命令);非空交注入的 compile,語法錯 → ok=false(跳過該效果)。
func compileCommand(compile CompileCommand, source string) (command EffectCommand, ok bool) {
	if source == "" {
		return nil, true
	} // if

	command, err := compile(source)

	return command, err == nil
}

// compileExpr 把運算式字串編成 *exprs.Expr:空字串 → nil(觸發條件恆成立 / 觸發次數視為 1);語法錯 → ok=false(跳過該效果)。
func compileExpr(source string) (expr *exprs.Expr, ok bool) {
	if source == "" {
		return nil, true
	} // if

	expr, err := exprs.Parse(source)

	return expr, err == nil
}
