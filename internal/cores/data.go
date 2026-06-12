package cores

import (
	"sort"
	"strconv"
	"strings"

	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// Data 一份遊戲資料: 原始靜態表 + 載入期衍生索引(抽獎 / 預編譯效果 / 顧客門檻)。
// 衍生索引跟資料走、不跟營業走: 同一份 Data 可供多場營業共用; 組裝期建立、營業期唯讀。
type Data struct {
	sheet  *sheeter.Sheeter     // 原始靜態表(Sheeter 資料 port)
	award  map[int32]AwardData  // 抽獎衍生索引(群組 → 候選); prepareAward 建, 供 Game.RollCard 用
	effect map[int32]EffectData // 預編譯效果索引(效果編號 → 編譯形); prepareEffect 建, 供效果流程查 Kind / 命令 / 條件
	guest  map[int32]GuestData  // 顧客門檻衍生索引(顧客編號 → 飽食 / 耐心門檻配對); prepareGuest 建, 供執行結算用
}

// NewData 以原始表組裝遊戲資料, 建構時整理衍生索引(prepareAward / prepareEffect / prepareGuest);
// 效果編譯需命令解析, 由 games 經 Compiler 注入(無命令資料可傳 nil); sheet nil → 空表空索引
// (正規化為空 Sheeter, 發射台識別碼查名免判 nil)。
func NewData(sheet *sheeter.Sheeter, compile Compiler) *Data {
	if sheet == nil {
		sheet = &sheeter.Sheeter{}
	} // if

	return &Data{
		sheet:  sheet,
		award:  prepareAward(sheet),
		effect: prepareEffect(sheet, compile),
		guest:  prepareGuest(sheet),
	}
}

// GetSheet 取原始靜態表(卡牌 / 顧客 / 座位 / 技能 / 設定查詢)。
func (this *Data) GetSheet() *sheeter.Sheeter {
	return this.sheet
}

// GetAward 查抽獎群組候選; 查無回 ok=false(供 Game.RollCard 做 weighted random)。
func (this *Data) GetAward(group int32) (meta AwardData, ok bool) {
	meta, ok = this.award[group]
	return meta, ok
}

// GetEffect 查預編譯效果; 查無回 ok=false。
func (this *Data) GetEffect(effectID int32) (meta EffectData, ok bool) {
	meta, ok = this.effect[effectID]
	return meta, ok
}

// SetEffect 組裝期補登 / 覆寫預編譯效果(測試注入自訂編譯閉包用; 營業開始後不應再呼叫)。
func (this *Data) SetEffect(effectID int32, meta EffectData) {
	this.effect[effectID] = meta
}

// GetGuest 查顧客門檻配對; 查無回 ok=false(無門檻資料的顧客不建項, 查無即無門檻)。
func (this *Data) GetGuest(guestID int32) (meta GuestData, ok bool) {
	meta, ok = this.guest[guestID]
	return meta, ok
}

// AwardData 抽獎群組的平行候選(cardID 與 weight 一一對應、同序); 供 weighted random 直接餵 Rander.Weighted。
type AwardData struct {
	cardID []int32
	weight []int32
}

// prepareAward 自 Award 表建「群組編號 → 候選」衍生索引; 依 Award.ID 排序確保決定性, 僅收 Weight > 0 者(全 0 權重群組自然為空 → roll no-op)。
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

// EffectData 效果靜態資料的預編譯形(【營業實作規格書 | 三、套件結構】effect 命令解耦); prepareEffect 一趟產出、營業期唯讀消費, 為效果靜態欄的唯一視圖。
// 堆疊上限 / 堆疊時間 / 目標 / 立即·啟動命令 隨 M12.1–M12.3 各自讀者擴入。
type EffectData struct {
	Kind         EffectKind   // 效果類型(篩選 觸發)
	TriggerKind  TriggerKind  // 觸發時機(篩選 時機)
	TriggerAfter TriggerAfter // 觸發後行為(保留 / 移除)
	Group        int32        // 效果群組編號(effectGroup 查詢 / 免疫閘門)
	RunRound     int32        // 作用回合(結束回合計算)
	RunOrder     int32        // 作用順序(佇列排序)
	Stack        int32        // 堆疊層數(每次堆疊增量; 0 / 1 → 1)
	StackMax     int32        // 堆疊上限(0 = 無上限)
	StackTime    StackTime    // 堆疊時間(不變 / 刷新)
	TargetKind   TargetKind   // 目標類型(self 選取方式)
	TargetCount  int32        // 目標數量(新選 / 隨機 須選數量)
	Cond         *exprs.Expr  // 觸發條件(空 → nil, 視為恆成立)
	Count        *exprs.Expr  // 觸發次數(空 → nil, 視為 1)
	Immed        EffectExec   // 立即命令(立即類型)
	Trigger      EffectExec   // 觸發命令(觸發類型)
	Start        EffectExec   // 啟動命令(常駐類型)
	End          EffectExec   // 結束命令(觸發 / 常駐)
}

// prepareEffect 自 Effect 表建「效果編號 → 編譯形」衍生索引(對應【營業實作規格書 | 三、套件結構】effect 命令解耦):
// 命令欄交注入的 compile 編成閉包、cond/count 走 exprs.Parse、Kind/After 0 起算解碼 + 範圍檢查後就地組裝(比照 prepareAward 就地建 AwardData)。
// 寬鬆策略(同 prepareAward 不回 error): 任一欄編譯失敗 / 列舉越界 → 跳過該效果, 嚴格把關交企劃驗證器於載入前。
// compile 僅於命令欄非空時呼叫; 無命令資料的測試可傳 nil。data 為 nil 回空索引。
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

// GuestData 顧客門檻的解析形: 飽食門檻升序、耐心門檻降序(對齊【營業規格書 | 二十、獨立流程 | 執行結算】迭代序)。
type GuestData struct {
	Sate []Threshold // 飽食門檻配對(門檻值由低到高)
	Calm []Threshold // 耐心門檻配對(門檻值由高到低)
}

// Threshold 門檻配對: 門檻值與對應門檻技能(顧客表格「門檻值^技能編號」的解析形; 【營業規格書 | 四、表格結構 | 顧客（Guest）表格】)。
type Threshold struct {
	Value   int32 // 門檻值
	SkillID int32 // 門檻技能編號
}

// ParseThreshold 解析單筆「門檻值^技能編號」配對; 壞格式(缺 ^ / 多段 / 兩段非整數)回帶位置錯誤
// (exprs.SyntaxError, 與運算式 / 命令兩文法同型)。格式知識的單一定義點(M28 R3):
// prepareGuest 寬鬆載入(錯誤即跳筆)與企劃驗證器嚴格檢查(錯誤即報告)共用, 不雙寫。
func ParseThreshold(source string) (result Threshold, err error) {
	part := strings.Split(source, "^")

	if len(part) != 2 {
		return Threshold{}, &exprs.SyntaxError{Pos: 0, Msg: "門檻配對需為「門檻值^技能編號」兩段:" + source}
	} // if

	value, errValue := strconv.ParseInt(part[0], 10, 32)

	if errValue != nil {
		return Threshold{}, &exprs.SyntaxError{Pos: 0, Msg: "門檻值需為整數:" + part[0]}
	} // if

	skillID, errSkill := strconv.ParseInt(part[1], 10, 32)

	if errSkill != nil {
		return Threshold{}, &exprs.SyntaxError{Pos: len([]rune(part[0])) + 1, Msg: "技能編號需為整數:" + part[1]}
	} // if

	return Threshold{Value: int32(value), SkillID: int32(skillID)}, nil
}

// prepareGuest 自 Guest 表建「顧客編號 → 門檻配對」衍生索引: 解析 SateSkillID / CalmSkillID 的「門檻值^技能編號」字串並排序;
// 壞格式(缺 ^ / 非數字)跳過該筆(寬鬆, 比照 prepareAward / prepareEffect, 嚴格把關交企劃驗證器); 無門檻的顧客不建項。
func prepareGuest(data *sheeter.Sheeter) map[int32]GuestData {
	result := map[int32]GuestData{}

	if data == nil {
		return result
	} // if

	for _, itor := range data.Guest.Keys() {
		guest := data.Guest.Get(itor)
		sate := parseThreshold(guest.SateSkillID)
		calm := parseThreshold(guest.CalmSkillID)

		if len(sate) == 0 && len(calm) == 0 {
			continue // 無門檻 → 不建項
		} // if

		sort.SliceStable(sate, func(i, j int) bool { return sate[i].Value < sate[j].Value })
		sort.SliceStable(calm, func(i, j int) bool { return calm[i].Value > calm[j].Value })
		result[itor] = GuestData{Sate: sate, Calm: calm}
	} // for

	return result
}

// parseThreshold 解析「門檻值^技能編號」配對列表; 壞格式跳過該筆(寬鬆面; 格式知識在 ParseThreshold)。
// 供 prepareGuest 的飽食 / 耐心兩欄共用。
func parseThreshold(source []string) (result []Threshold) {
	for _, itor := range source {
		threshold, err := ParseThreshold(itor)

		if err != nil {
			continue // 壞格式 → 跳過該筆
		} // if

		result = append(result, threshold)
	} // for

	return result
}

// compileCommand 把命令字串編成執行器: 空字串 → nil(無命令); 非空交注入的 compile, 語法錯 → ok=false(跳過該效果);
// 編譯器未注入(NewData「無命令資料可傳 nil」)時帶命令的效果無從編譯, 同語法錯跳過。
func compileCommand(compile Compiler, source string) (command EffectExec, ok bool) {
	if source == "" {
		return nil, true
	} // if

	if compile == nil {
		return nil, false // 無編譯器 → 帶命令的效果跳過
	} // if

	command, err := compile(source)
	return command, err == nil
}

// compileExpr 把運算式字串編成 *exprs.Expr: 空字串 → nil(觸發條件恆成立 / 觸發次數視為 1); 語法錯 → ok=false(跳過該效果)。
func compileExpr(source string) (expr *exprs.Expr, ok bool) {
	if source == "" {
		return nil, true
	} // if

	expr, err := exprs.Parse(source)
	return expr, err == nil
}
