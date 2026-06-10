package cores

import (
	"sort"

	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// Data 一份遊戲資料：原始靜態表 + 載入期衍生索引（抽獎 / 預編譯效果）。
// 衍生索引跟資料走、不跟營業走：同一份 Data 可供多場營業共用；組裝期建立、營業期唯讀。
type Data struct {
	sheet  *sheeter.Sheeter     // 原始靜態表（Sheeter 資料 port）
	award  map[int32]AwardData  // 抽獎衍生索引（群組 → 候選）;prepareAward 建,供 Game.RollCard 用
	effect map[int32]EffectData // 預編譯效果索引（效果編號 → 編譯形）;prepareEffect 建,供效果流程查 Kind / 命令 / 條件
}

// NewData 以原始表組裝遊戲資料，建構時整理衍生索引（prepareAward / prepareEffect）;
// 效果編譯需命令解析，由 games 經 Compiler 注入（無命令資料可傳 nil）;sheet nil → 空表空索引。
func NewData(sheet *sheeter.Sheeter, compile Compiler) *Data {
	return &Data{
		sheet:  sheet,
		award:  prepareAward(sheet),
		effect: prepareEffect(sheet, compile),
	}
}

// GetSheet 取原始靜態表（卡牌 / 顧客 / 座位 / 技能 / 設定查詢）。
func (this *Data) GetSheet() *sheeter.Sheeter {
	return this.sheet
}

// GetEffect 查預編譯效果;查無回 ok=false。
func (this *Data) GetEffect(effectID int32) (meta EffectData, ok bool) {
	meta, ok = this.effect[effectID]
	return meta, ok
}

// SetEffect 組裝期補登 / 覆寫預編譯效果（測試注入自訂編譯閉包用;營業開始後不應再呼叫）。
func (this *Data) SetEffect(effectID int32, meta EffectData) {
	this.effect[effectID] = meta
}

// AwardData 抽獎群組的平行候選(cardID 與 weight 一一對應、同序);供 weighted random 直接餵 Rander.Weighted。
type AwardData struct {
	cardID []int32
	weight []int32
}

// prepareAward 自 Award 表建「群組編號 → 候選」衍生索引;依 Award.ID 排序確保決定性,僅收 Weight > 0 者(全 0 權重群組自然為空 → roll no-op)。
func prepareAward(data *sheeter.Sheeter) map[int32]AwardData {
	result := map[int32]AwardData{}

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

// EffectData 效果靜態資料的預編譯形（【營業實作規格書 | 三、套件結構】effect 命令解耦）;prepareEffect 一趟產出、營業期唯讀消費，為效果靜態欄的唯一視圖。
// 堆疊上限 / 堆疊時間 / 目標 / 立即·啟動命令 隨 M12.1–M12.3 各自讀者擴入。
type EffectData struct {
	Kind         EffectKind   // 效果類型（篩選 觸發）
	TriggerKind  TriggerKind  // 觸發時機（篩選 時機）
	TriggerAfter TriggerAfter // 觸發後行為（保留 / 移除）
	Group        int32        // 效果群組編號（effectGroup 查詢 / 免疫閘門）
	RunRound     int32        // 作用回合（結束回合計算）
	RunOrder     int32        // 作用順序（佇列排序）
	Stack        int32        // 堆疊層數（每次堆疊增量;0 / 1 → 1）
	StackMax     int32        // 堆疊上限（0 = 無上限）
	StackTime    StackTime    // 堆疊時間（不變 / 刷新）
	TargetKind   TargetKind   // 目標類型（self 選取方式）
	TargetCount  int32        // 目標數量（新選 / 隨機 須選數量）
	Cond         *exprs.Expr  // 觸發條件（空 → nil，視為恆成立）
	Count        *exprs.Expr  // 觸發次數（空 → nil，視為 1）
	Immed        EffectExec   // 立即命令（立即類型）
	Trigger      EffectExec   // 觸發命令（觸發類型）
	Start        EffectExec   // 啟動命令（常駐類型）
	End          EffectExec   // 結束命令（觸發 / 常駐）
}

// prepareEffect 自 Effect 表建「效果編號 → 編譯形」衍生索引(對應【營業實作規格書 | 三、套件結構】effect 命令解耦):
// 命令欄交注入的 compile 編成閉包、cond/count 走 exprs.Parse、Kind/After 0 起算解碼 + 範圍檢查後就地組裝(比照 prepareAward 就地建 AwardData)。
// 寬鬆策略(同 prepareAward 不回 error):任一欄編譯失敗 / 列舉越界 → 跳過該效果,嚴格把關交企劃驗證器於載入前。
// compile 僅於命令欄非空時呼叫;無命令資料的測試可傳 nil。data 為 nil 回空索引。
func prepareEffect(data *sheeter.Sheeter, compile Compiler) map[int32]EffectData {
	result := map[int32]EffectData{}

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

		immed, ok := compileCommand(compile, meta.CommandImmed)

		if ok == false {
			continue
		} // if

		start, ok := compileCommand(compile, meta.CommandStart)

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

		if meta.StackTime < int32(StackTimeStay) || meta.StackTime > int32(StackTimeRefresh) {
			continue // 堆疊時間編碼越界 → 跳過
		} // if

		if meta.TargetKind < int32(TargetNone) || meta.TargetKind > int32(TargetCardRand) {
			continue // 目標類型編碼越界 → 跳過
		} // if

		result[itor] = EffectData{
			Kind:         EffectKind(meta.Kind),
			TriggerKind:  TriggerKind(meta.TriggerKind),
			TriggerAfter: TriggerAfter(meta.TriggerAfter),
			Group:        meta.Group,
			RunRound:     meta.RunRound,
			RunOrder:     meta.RunOrder,
			Stack:        meta.Stack,
			StackMax:     meta.StackMax,
			StackTime:    StackTime(meta.StackTime),
			TargetKind:   TargetKind(meta.TargetKind),
			TargetCount:  meta.TargetCount,
			Cond:         cond,
			Count:        count,
			Immed:        immed,
			Trigger:      trigger,
			Start:        start,
			End:          end,
		}
	} // for

	return result
}

// compileCommand 把命令字串編成執行器:空字串 → nil(無命令);非空交注入的 compile,語法錯 → ok=false(跳過該效果)。
func compileCommand(compile Compiler, source string) (command EffectExec, ok bool) {
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
