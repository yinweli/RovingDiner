package roditool

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteCheckEffect(t *testing.T) {
	suite.Run(t, new(SuiteCheckEffect))
}

// SuiteCheckEffect 驗證效果表檢查(checkEffect.go): 編碼範圍 / 觸發時機名稱 / 運算式與命令文法詞彙 / 適用類型矩陣。
type SuiteCheckEffect struct {
	suite.Suite
}

// TestCheckEffect 驗證逐列檢查: 每列一種違規(輸出序 = 列升序 x 欄檢查序), 全合法列零筆。
func (this *SuiteCheckEffect) TestCheckEffect() {
	sheet := &sheeter.Sheeter{}
	sheet.Effect.Data = map[int32]*sheeter.Effect{
		401: {ID: 401, Kind: 9},                                                                  // Kind 超界(矩陣因編碼不合法略過)
		402: {ID: 402, Kind: 1, TriggerKind: "guestSeat", TargetKind: 7},                         // TargetKind 超界
		403: {ID: 403, Kind: 1, TriggerKind: "guestSeat", TriggerAfter: 2},                       // TriggerAfter 超界
		404: {ID: 404, Kind: 1, TriggerKind: "guestSeat", StackTime: 5},                          // StackTime 超界
		405: {ID: 405, Kind: 1, TriggerKind: "nope"},                                             // 未知觸發時機名稱
		406: {ID: 406, Kind: 1, TriggerKind: "guestSeat", TriggerCond: "morale >"},               // 運算式文法錯
		407: {ID: 407, Kind: 1, TriggerKind: "guestSeat", CommandTrigger: "noSuchCommand(none)"}, // 命令詞彙錯(引擎編譯只 Parse 不攔)
		408: {ID: 408, Kind: 0, TriggerKind: "guestSeat"},                                        // 矩陣: 立即類型不適用觸發時機
		409: {ID: 409, Kind: 1, TriggerKind: "guestSeat", TriggerCond: "morale > 0", TriggerCount: "2",
			TriggerAfter: 1, RunRound: 2, Stack: 1, StackMax: 3, StackTime: 1,
			CommandTrigger: "handAdd(none, 1, 1)", CommandEnd: "morale += 1"}, // 全合法 → 零筆
		410: {ID: 410, Kind: 1}, // 反向必填: 觸發類型缺觸發時機
	}
	issue := checkEffect(sheet)
	this.Require().Len(issue, 9)
	this.Equal(Issue{Table: "effect", Row: "401", Column: "Kind", Msg: "效果類型編碼超出範圍(0 立即 / 1 觸發 / 2 常駐):9"}, issue[0])
	this.Equal(Issue{Table: "effect", Row: "402", Column: "TargetKind", Msg: "目標類型編碼超出範圍(0~6):7"}, issue[1])
	this.Equal(Issue{Table: "effect", Row: "403", Column: "TriggerAfter", Msg: "觸發後行為編碼超出範圍(0 保留 / 1 移除):2"}, issue[2])
	this.Equal(Issue{Table: "effect", Row: "404", Column: "StackTime", Msg: "堆疊時間編碼超出範圍(0 不變 / 1 刷新):5"}, issue[3])
	this.Equal(Issue{Table: "effect", Row: "405", Column: "TriggerKind", Msg: "未知的觸發時機名稱:nope"}, issue[4])
	this.Equal("406", issue[5].Row)
	this.Equal("TriggerCond", issue[5].Column)
	this.Contains(issue[5].Msg, "第") // 文法錯帶「第 N 字附近」(訊息細節由 exprs 釘住)
	this.Equal("407", issue[6].Row)
	this.Equal("CommandTrigger", issue[6].Column)
	this.Contains(issue[6].Msg, "未知的操作命令:noSuchCommand")
	this.Equal(Issue{Table: "effect", Row: "408", Column: "TriggerKind", Msg: "欄位不適用於立即類型, 不應填值"}, issue[7])
	this.Equal(Issue{Table: "effect", Row: "410", Column: "TriggerKind", Msg: "觸發類型缺觸發時機, 效果永不觸發"}, issue[8])
}

// TestCheckEffectMatrix 驗證適用類型矩陣全欄掃描: 常駐類型填觸發族欄位逐欄回報、觸發類型填立即 / 啟動命令回報。
func (this *SuiteCheckEffect) TestCheckEffectMatrix() {
	sheet := &sheeter.Sheeter{}
	sheet.Effect.Data = map[int32]*sheeter.Effect{
		401: {ID: 401, Kind: 2, TriggerKind: "guestSeat", TriggerCond: "1", TriggerCount: "1", TriggerAfter: 1,
			CommandImmed: "morale += 1", CommandTrigger: "morale += 1"}, // 常駐: 觸發族 + 立即 / 觸發命令全不適用
		402: {ID: 402, Kind: 1, TriggerKind: "guestSeat", CommandImmed: "morale += 1", CommandStart: "morale += 1"}, // 觸發: 立即 / 啟動命令不適用
		403: {ID: 403, Kind: 0, Stack: 2, StackMax: 1, StackTime: 1, RunRound: 3, Group: 5},                         // 立即: 佇列族全不適用
	}
	issue := checkEffect(sheet)
	column := []string{}

	for _, itor := range issue {
		column = append(column, itor.Row+"/"+itor.Column)
	} // for

	this.Equal([]string{
		"401/TriggerKind", "401/TriggerCond", "401/TriggerCount", "401/TriggerAfter", "401/CommandImmed", "401/CommandTrigger",
		"402/CommandImmed", "402/CommandStart",
		"403/Group", "403/RunRound", "403/Stack", "403/StackMax", "403/StackTime",
	}, column)
}

// TestCheckExprField 驗證運算式欄: 空欄未填通過、文法錯回報(識別子詞彙查名屬 R4A)。
func (this *SuiteCheckEffect) TestCheckExprField() {
	this.Empty(checkExprField("401", "TriggerCond", ""))
	this.Empty(checkExprField("401", "TriggerCond", "morale > 0"))
	issue := checkExprField("401", "TriggerCond", "1 +")
	this.Require().Len(issue, 1)
	this.Equal("TriggerCond", issue[0].Column)
}

// TestCheckCommandField 驗證命令欄: 空欄未填通過、文法錯與詞彙 / arity 錯(games.Validate)各自回報。
func (this *SuiteCheckEffect) TestCheckCommandField() {
	this.Empty(checkCommandField("401", "CommandImmed", ""))
	this.Empty(checkCommandField("401", "CommandImmed", "handAdd(none, 1, 1)"))

	issue := checkCommandField("401", "CommandImmed", "handAdd(none") // 文法錯
	this.Require().Len(issue, 1)
	this.Equal("CommandImmed", issue[0].Column)

	issue = checkCommandField("401", "CommandImmed", "guestExit(guestPick, 1, 1)") // arity 錯(M28 R1)
	this.Require().Len(issue, 1)
	this.Contains(issue[0].Msg, "參數數量不符")

	issue = checkCommandField("401", "CommandImmed", "morale.cost = 1") // 引用基底錯(M28 R2)
	this.Require().Len(issue, 1)
	this.Contains(issue[0].Msg, "不是合法的物件引用")
}

// TestKindText 驗證效果類型中文名(含防禦分支)。
func (this *SuiteCheckEffect) TestKindText() {
	this.Equal("立即", kindText(0))
	this.Equal("觸發", kindText(1))
	this.Equal("常駐", kindText(2))
	this.Equal("?", kindText(9)) // 防禦: 呼叫端已保證合法, 仍回 ? 不爆
}
