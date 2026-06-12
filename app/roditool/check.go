package roditool

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// 表名與重複欄名常數(Issue.Table / Issue.Column 取值; 表序即 CheckSheet 檢查序)。
const (
	tableAward        = "award"
	tableCard         = "card"
	tableEffect       = "effect"
	tableGuest        = "guest"
	tableSetting      = "setting"
	tableSkill        = "skill"
	tableStage        = "stage"
	columnValue       = "Value"
	columnTriggerKind = "TriggerKind"
)

// Issue 表單檢查的單筆結果: 表 / 列(編號) / 欄 + 中文錯誤(【營業實作規格書 | 附錄：企劃驗證器】表單檢查輸出)。
type Issue struct {
	Table  string // 表名(award / card / effect / guest / setting / skill / stage)
	Row    string // 列編號(主鍵字串形; setting 以鍵名為列)
	Column string // 欄名(表格英文欄位名)
	Msg    string // 錯誤訊息(中文白話; 文法錯誤帶「第 N 字附近」)
}

// String 單筆結果的人類可讀行(CLI 表單檢查逐筆輸出)。
func (this Issue) String() string {
	return fmt.Sprintf("表 %v | 列 %v | 欄 %v | %v", this.Table, this.Row, this.Column, this.Msg)
}

// CheckSheet 表單檢查入口: 逐表(固定表序)逐列(編號升序)逐欄檢查, 彙整全部結果(不發現即空)。
// 通則依【營業實作規格書 | 附錄：企劃驗證器】: 參照欄 0 = 未指派合法跳過、非 0 編號必須存在於目標表;
// 字串欄空 = 未填合法。引擎載入對同批資料採寬鬆跳過, 嚴格把關在此。
func CheckSheet(sheet *sheeter.Sheeter) (result []Issue) {
	if sheet == nil {
		return nil
	} // if

	result = append(result, checkAward(sheet)...)
	result = append(result, checkCard(sheet)...)
	result = append(result, checkEffect(sheet)...)
	result = append(result, checkGuest(sheet)...)
	result = append(result, checkSetting(sheet)...)
	result = append(result, checkSkill(sheet)...)
	result = append(result, checkStage(sheet)...)
	return result
}

// checkAward 檢查抽獎表: CardID → card.ID。
func checkAward(sheet *sheeter.Sheeter) (result []Issue) {
	for _, itor := range sortedKey(sheet.Award.Data) {
		meta := sheet.Award.Data[itor]

		if meta.CardID != 0 && sheet.Card.Get(meta.CardID) == nil {
			result = append(result, Issue{Table: tableAward, Row: num(itor), Column: "CardID", Msg: "查無卡牌編號:" + num(meta.CardID)})
		} // if
	} // for

	return result
}

// checkCard 檢查卡牌表: SkillID → skill.ID(0 = 無技能)。
func checkCard(sheet *sheeter.Sheeter) (result []Issue) {
	for _, itor := range sortedKey(sheet.Card.Data) {
		meta := sheet.Card.Data[itor]

		if meta.SkillID != 0 && sheet.Skill.Get(meta.SkillID) == nil {
			result = append(result, Issue{Table: tableCard, Row: num(itor), Column: "SkillID", Msg: "查無技能編號:" + num(meta.SkillID)})
		} // if
	} // for

	return result
}

// checkGuest 檢查顧客表: SateSkillID / CalmSkillID 門檻配對(格式經 cores.ParseThreshold 單一來源 + 技能編號參照)。
func checkGuest(sheet *sheeter.Sheeter) (result []Issue) {
	for _, itor := range sortedKey(sheet.Guest.Data) {
		meta := sheet.Guest.Data[itor]
		result = append(result, checkThreshold(sheet, itor, "SateSkillID", meta.SateSkillID)...)
		result = append(result, checkThreshold(sheet, itor, "CalmSkillID", meta.CalmSkillID)...)
	} // for

	return result
}

// checkThreshold 檢查單欄門檻配對列表: 空字串 = 未填跳過、格式錯誤回報、技能編號非 0 時須存在於技能表。
func checkThreshold(sheet *sheeter.Sheeter, row int32, column string, source []string) (result []Issue) {
	for _, itor := range source {
		if itor == "" {
			continue // 字串欄空 = 未填(Sheeter 空欄生成 [""])
		} // if

		threshold, err := cores.ParseThreshold(itor)

		if err != nil {
			result = append(result, Issue{Table: tableGuest, Row: num(row), Column: column, Msg: err.Error()})
			continue
		} // if

		if threshold.SkillID != 0 && sheet.Skill.Get(threshold.SkillID) == nil {
			result = append(result, Issue{Table: tableGuest, Row: num(row), Column: column, Msg: "查無技能編號:" + num(threshold.SkillID)})
		} // if
	} // for

	return result
}

// settingKey 設定表必備鍵(【營業實作規格書 | 附錄：企劃驗證器】setting 檢查項目; 引擎讀法見 rules.settingNum)。
var settingKey = []string{"RoundMax", "Morale", "MoraleMax", "EnergyMax", "HandMax", "DrawMax"}

// checkSetting 檢查設定表: 必備鍵存在且設定值為數字(引擎以 ParseFloat 讀首值, 同其寬鬆面的格式知識)。
func checkSetting(sheet *sheeter.Sheeter) (result []Issue) {
	for _, itor := range settingKey {
		meta := sheet.Setting.Get(itor)

		if meta == nil {
			result = append(result, Issue{Table: tableSetting, Row: itor, Column: "ID", Msg: "缺少設定鍵:" + itor})
			continue
		} // if

		if len(meta.Value) == 0 || meta.Value[0] == "" {
			result = append(result, Issue{Table: tableSetting, Row: itor, Column: columnValue, Msg: "設定值未填"})
			continue
		} // if

		if _, err := strconv.ParseFloat(meta.Value[0], 64); err != nil {
			result = append(result, Issue{Table: tableSetting, Row: itor, Column: columnValue, Msg: "設定值需為數字:" + meta.Value[0]})
		} // if
	} // for

	return result
}

// checkSkill 檢查技能表: EffectID 列表逐項 → effect.ID。
func checkSkill(sheet *sheeter.Sheeter) (result []Issue) {
	for _, itor := range sortedKey(sheet.Skill.Data) {
		meta := sheet.Skill.Data[itor]

		for _, effectID := range meta.EffectID {
			if effectID != 0 && sheet.Effect.Get(effectID) == nil {
				result = append(result, Issue{Table: tableSkill, Row: num(itor), Column: "EffectID", Msg: "查無效果編號:" + num(effectID)})
			} // if
		} // for
	} // for

	return result
}

// checkStage 檢查關卡表: 四牌堆列表 → card.ID、排隊顧客列表 → guest.ID、前置技能列表 → skill.ID。
func checkStage(sheet *sheeter.Sheeter) (result []Issue) {
	for _, itor := range sortedKey(sheet.Stage.Data) {
		meta := sheet.Stage.Data[itor]
		result = append(result, checkIDList(itor, "HandID", meta.HandID, "查無卡牌編號", func(id int32) bool { return sheet.Card.Get(id) != nil })...)
		result = append(result, checkIDList(itor, "DeckID", meta.DeckID, "查無卡牌編號", func(id int32) bool { return sheet.Card.Get(id) != nil })...)
		result = append(result, checkIDList(itor, "DropID", meta.DropID, "查無卡牌編號", func(id int32) bool { return sheet.Card.Get(id) != nil })...)
		result = append(result, checkIDList(itor, "ExileID", meta.ExileID, "查無卡牌編號", func(id int32) bool { return sheet.Card.Get(id) != nil })...)
		result = append(result, checkIDList(itor, "WaitID", meta.WaitID, "查無顧客編號", func(id int32) bool { return sheet.Guest.Get(id) != nil })...)
		result = append(result, checkIDList(itor, "PrefixSkillID", meta.PrefixSkillID, "查無技能編號", func(id int32) bool { return sheet.Skill.Get(id) != nil })...)
	} // for

	return result
}

// checkIDList 檢查編號列表參照欄: 逐項套通則(0 = 未指派跳過、非 0 須通過 exist 述詞)。
func checkIDList(row int32, column string, id []int32, msg string, exist func(id int32) bool) (result []Issue) {
	for _, itor := range id {
		if itor != 0 && exist(itor) == false {
			result = append(result, Issue{Table: tableStage, Row: num(row), Column: column, Msg: msg + ":" + num(itor)})
		} // if
	} // for

	return result
}

// sortedKey 取表格主鍵升序列表(map 走訪去隨機化, 使輸出順序決定性)。
func sortedKey[T any](data map[int32]*T) (result []int32) {
	for k := range data {
		result = append(result, k)
	} // for

	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

// num 編號轉字串(列編號 / 錯誤訊息共用)。
func num(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}
