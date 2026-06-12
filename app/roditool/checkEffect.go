package roditool

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
	"github.com/yinweli/RovingDiner/internal/games"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// effectColumn 適用類型矩陣單欄: 欄名 + 各類型適用(索引 = 效果類型編碼: 0 立即 / 1 觸發 / 2 常駐)+ 已填判定。
type effectColumn struct {
	column string                          // 欄名(表格英文欄位名)
	apply  [3]bool                         // 各類型是否使用此欄(對齊矩陣的 v 記號)
	filled func(meta *sheeter.Effect) bool // 已填判定(數值非 0 / 字串非空)
}

// effectMatrix 適用類型矩陣(【營業規格書 | 四、表格結構】效果表格): 檢查「填了不適用欄位」——
// 當前類型不適用且已填即報(2026-06-12 拍板要驗)。全類型適用欄(TargetKind / TargetCount)無此檢查、不列。
var effectMatrix = []effectColumn{
	{"Group", [3]bool{false, true, true}, func(meta *sheeter.Effect) bool { return meta.Group != 0 }},
	{columnTriggerKind, [3]bool{false, true, false}, func(meta *sheeter.Effect) bool { return meta.TriggerKind != "" }},
	{"TriggerCond", [3]bool{true, true, false}, func(meta *sheeter.Effect) bool { return meta.TriggerCond != "" }},
	{"TriggerCount", [3]bool{true, true, false}, func(meta *sheeter.Effect) bool { return meta.TriggerCount != "" }},
	{"TriggerAfter", [3]bool{false, true, false}, func(meta *sheeter.Effect) bool { return meta.TriggerAfter != 0 }},
	{"RunRound", [3]bool{false, true, true}, func(meta *sheeter.Effect) bool { return meta.RunRound != 0 }},
	{"Stack", [3]bool{false, true, true}, func(meta *sheeter.Effect) bool { return meta.Stack != 0 }},
	{"StackMax", [3]bool{false, true, true}, func(meta *sheeter.Effect) bool { return meta.StackMax != 0 }},
	{"StackTime", [3]bool{false, true, true}, func(meta *sheeter.Effect) bool { return meta.StackTime != 0 }},
	{"CommandImmed", [3]bool{true, false, false}, func(meta *sheeter.Effect) bool { return meta.CommandImmed != "" }},
	{"CommandTrigger", [3]bool{false, true, false}, func(meta *sheeter.Effect) bool { return meta.CommandTrigger != "" }},
	{"CommandStart", [3]bool{false, false, true}, func(meta *sheeter.Effect) bool { return meta.CommandStart != "" }},
	{"CommandEnd", [3]bool{false, true, true}, func(meta *sheeter.Effect) bool { return meta.CommandEnd != "" }},
}

// checkEffect 檢查效果表: 編碼範圍(Kind / TargetKind / TriggerAfter / StackTime)、TriggerKind 非空時的
// 成員資格(cores.TriggerKindLegal)與觸發類型缺漏(反向必填單格)、運算式欄文法(exprs.Parse)、
// 命令欄文法 + 詞彙(games.Parse + Validate; 引擎編譯只 Parse, 詞彙錯僅在此攔)、適用類型矩陣(僅 Kind 合法時)。
func checkEffect(sheet *sheeter.Sheeter) (result []Issue) {
	for _, itor := range sortedKey(sheet.Effect.Data) {
		meta := sheet.Effect.Data[itor]
		row := num(itor)

		if meta.Kind < 0 || meta.Kind > int32(cores.EffectPersist) {
			result = append(result, Issue{Table: tableEffect, Row: row, Column: "Kind", Msg: "效果類型編碼超出範圍(0 立即 / 1 觸發 / 2 常駐):" + num(meta.Kind)})
		} // if

		if meta.TargetKind < 0 || meta.TargetKind > int32(cores.TargetCardRand) {
			result = append(result, Issue{Table: tableEffect, Row: row, Column: "TargetKind", Msg: "目標類型編碼超出範圍(0~6):" + num(meta.TargetKind)})
		} // if

		if meta.TriggerAfter < 0 || meta.TriggerAfter > int32(cores.TriggerAfterRemove) {
			result = append(result, Issue{Table: tableEffect, Row: row, Column: "TriggerAfter", Msg: "觸發後行為編碼超出範圍(0 保留 / 1 移除):" + num(meta.TriggerAfter)})
		} // if

		if meta.StackTime < 0 || meta.StackTime > int32(cores.StackTimeRefresh) {
			result = append(result, Issue{Table: tableEffect, Row: row, Column: "StackTime", Msg: "堆疊時間編碼超出範圍(0 不變 / 1 刷新):" + num(meta.StackTime)})
		} // if

		if meta.TriggerKind != "" && cores.TriggerKindLegal[cores.TriggerKind(meta.TriggerKind)] == false {
			result = append(result, Issue{Table: tableEffect, Row: row, Column: columnTriggerKind, Msg: "未知的觸發時機名稱:" + meta.TriggerKind})
		} // if

		// 反向必填僅驗此一格(2026-06-12 拍板): 觸發類型缺觸發時機無任何合法解釋(進佇列但永不觸發), 必為漏填;
		// 其餘「適用但空著」各有合法預設語意(條件恆成立 / 次數 1 / 保留 / 永久)或可為標記效果, 不驗。
		if meta.Kind == int32(cores.EffectTrigger) && meta.TriggerKind == "" {
			result = append(result, Issue{Table: tableEffect, Row: row, Column: columnTriggerKind, Msg: "觸發類型缺觸發時機, 效果永不觸發"})
		} // if

		result = append(result, checkExprField(row, "TriggerCond", meta.TriggerCond)...)
		result = append(result, checkExprField(row, "TriggerCount", meta.TriggerCount)...)
		result = append(result, checkCommandField(row, "CommandImmed", meta.CommandImmed)...)
		result = append(result, checkCommandField(row, "CommandTrigger", meta.CommandTrigger)...)
		result = append(result, checkCommandField(row, "CommandStart", meta.CommandStart)...)
		result = append(result, checkCommandField(row, "CommandEnd", meta.CommandEnd)...)

		if meta.Kind >= 0 && meta.Kind <= int32(cores.EffectPersist) {
			for _, column := range effectMatrix {
				if column.apply[meta.Kind] == false && column.filled(meta) {
					result = append(result, Issue{Table: tableEffect, Row: row, Column: column.column, Msg: "欄位不適用於" + kindText(meta.Kind) + "類型, 不應填值"})
				} // if
			} // for
		} // if
	} // for

	return result
}

// checkExprField 檢查運算式欄: 空 = 未填合法; 非空驗文法(exprs.Parse)。
// 識別子詞彙查名待 R4A(運算式詞彙走訪), 本站先攔文法錯。
func checkExprField(row, column, source string) (result []Issue) {
	if source == "" {
		return nil
	} // if

	if _, err := exprs.Parse(source); err != nil {
		result = append(result, Issue{Table: tableEffect, Row: row, Column: column, Msg: err.Error()})
	} // if

	return result
}

// checkCommandField 檢查命令欄: 空 = 未填合法; 非空驗文法(games.Parse)+ 詞彙與參數
// (games.Validate: verb / 命令對象 / arity / 左值可寫性 / 引用基底)。
func checkCommandField(row, column, source string) (result []Issue) {
	if source == "" {
		return nil
	} // if

	command, err := games.Parse(source)

	if err != nil {
		return append(result, Issue{Table: tableEffect, Row: row, Column: column, Msg: err.Error()})
	} // if

	if err = games.Validate(command); err != nil {
		result = append(result, Issue{Table: tableEffect, Row: row, Column: column, Msg: err.Error()})
	} // if

	return result
}

// kindText 效果類型中文名(矩陣錯誤訊息用; 呼叫端已保證編碼合法)。
func kindText(kind int32) string {
	switch cores.EffectKind(kind) {
	case cores.EffectImmed:
		return "立即"

	case cores.EffectTrigger:
		return "觸發"

	case cores.EffectPersist:
		return "常駐"

	default:
		return "?"
	} // switch
}
