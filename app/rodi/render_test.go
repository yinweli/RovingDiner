package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteRender(t *testing.T) {
	suite.Run(t, new(SuiteRender))
}

// SuiteRender 驗證渲染工具(render.go): 顯示寬補白 / 截斷 / 兩行對齊表, 全形 2 格半形 1 格混排不歪。
type SuiteRender struct {
	suite.Suite
}

// TestPadTo 驗證右補空白至顯示寬: 全形以 2 格計; 已達 / 超出寬度原樣回傳。
func (this *SuiteRender) TestPadTo() {
	this.Equal("abc  ", padTo("abc", 5))
	this.Equal("中文 ", padTo("中文", 5)) // 全形 2 格: 中文 = 4 格 + 補 1
	this.Equal("abc", padTo("abc", 3))
	this.Equal("abcd", padTo("abcd", 3)) // 超寬不截斷(截斷歸 truncTo)
}

// TestTruncTo 驗證截斷至顯示寬: 全形不切半字; 未超寬原樣回傳。
func (this *SuiteRender) TestTruncTo() {
	this.Equal("abc", truncTo("abc", 5))
	this.Equal("ab", truncTo("abcd", 2))
	this.Equal("中", truncTo("中文字", 3)) // 3 格只容 1 個全形, 不切半字
}

// TestTruncMark 驗證截斷補右緣 > 記號: 未超寬原樣、超寬截斷後 > 貼右緣、過窄純截斷。
func (this *SuiteRender) TestTruncMark() {
	this.Equal("abc", truncMark("abc", 5))
	this.Equal("abcd >", truncMark("abcdefgh", 6))
	this.Equal("中文 >", truncMark("中文字串", 6)) // 全形不切半字
	this.Equal("ab", truncMark("abcdef", 2)) // 過窄(防禦) → 純截斷
}

// TestPanelTitle 驗證面板標題列: 標題前後留 1 空白、餘寬補橫線、收尾 +、超寬截斷。
func (this *SuiteRender) TestPanelTitle() {
	this.Equal("+- 座位 ---+", panelTitle("座位", 12))
	this.Equal("+- 座位+", panelTitle("座位", 8)) // 餘寬 0 → 不補線(防禦; 實際區寬不會這麼窄)
	this.Equal("+- 座+", panelTitle("座位", 6))  // 超寬截斷
}

// TestPanelTitleSeam 驗證接縫版標題列: 無左端 +(由左欄中線供應)、其餘同 panelTitle。
func (this *SuiteRender) TestPanelTitleSeam() {
	this.Equal("- 事件日誌 -+", panelTitleSeam("事件日誌", 13))
}

// TestBoxRow 驗證帶框內容行: 內容補白至內容寬(區寬 - 4)後包框與 cell padding。
func (this *SuiteRender) TestBoxRow() {
	this.Equal("| abc    |", boxRow("abc", 10))
	this.Equal("|  |", boxRow("", 4)) // 空內容: 框與 padding 仍在(墊高行用)
}

// TestBoxMark 驗證帶框截斷行(> 版): 超寬截至內容寬、> 站最後內容格(padding 與框保留)。
func (this *SuiteRender) TestBoxMark() {
	this.Equal("| abc    |", boxMark("abc", 10))
	this.Equal("| abcd > |", boxMark("abcdefgh", 10))
}

// TestBoxTrunc 驗證帶框截斷行(純版): 超寬截至內容寬、不補記號。
func (this *SuiteRender) TestBoxTrunc() {
	this.Equal("| abc    |", boxTrunc("abc", 10))
	this.Equal("| abcdef |", boxTrunc("abcdefgh", 10))
}

// TestBoxRowSeam 驗證接縫版帶框內容行: 無左框(左緣由左欄中線供應)、內容寬 = 區寬 - 3。
func (this *SuiteRender) TestBoxRowSeam() {
	this.Equal(" abc     |", boxRowSeam("abc", 10))
}

// TestNum 驗證整數屬性值轉字串。
func (this *SuiteRender) TestNum() {
	this.Equal("0", num(0))
	this.Equal("1250", num(1250))
	this.Equal("-3", num(-3))
}

// TestNumFloor 驗證護盾 / 格擋顯示下限: <= 0 顯 0。
func (this *SuiteRender) TestNumFloor() {
	this.Equal("5", numFloor(5))
	this.Equal("0", numFloor(0))
	this.Equal("0", numFloor(-2))
}

// TestAlignTable 驗證多行對齊表: 欄寬 = 該欄各列最大顯示寬、欄距 2 空白、行尾不留補白; 空表回空。
func (this *SuiteRender) TestAlignTable() {
	this.Equal([]string{
		"項目  次數  對象",
		"入座  12    501@老饕",
		"離場  3",
	}, alignTable([][]string{{"項目", "次數", "對象"}, {"入座", "12", "501@老饕"}, {"離場", "3", ""}}))
	this.Empty(alignTable(nil))
}

// TestAlignPair 驗證兩行對齊表: 欄寬 = 標籤 / 數值較寬者、欄距 2 空白、行尾不留補白。
func (this *SuiteRender) TestAlignPair() {
	row1, row2 := alignPair([]string{"回合", "士氣值", "x"}, []string{"3/10", "25", "1250"})
	this.Equal("回合  士氣值  x", row1)
	this.Equal("3/10  25      1250", row2)
}

// === 測試輔助(置尾) ===

// testSheet 迷你靜態表(識別碼與初值查表用): 座位 桌1(1,2)/ 桌2(3)、卡 101 / 103、顧客 501、技能 301、效果 401。
func testSheet() *sheeter.Sheeter {
	sheet := &sheeter.Sheeter{}
	sheet.Seat.Data = map[int32]*sheeter.Seat{
		1: {ID: 1, TableID: 1, SameSeatID: []int32{2}, NearSeatID: []int32{3}},
		2: {ID: 2, TableID: 1, SameSeatID: []int32{1}, NearSeatID: []int32{3}},
		3: {ID: 3, TableID: 2, SameSeatID: []int32{3}, NearSeatID: []int32{1, 2}},
	}
	sheet.Card.Data = map[int32]*sheeter.Card{
		101: {ID: 101, Name: "上菜", Group: 3, SkillID: 301, Cost: 2, ExtraRunMin: 1, ExtraRunMax: 3, Seal: true},
		103: {ID: 103, Name: "結帳", Cost: 1, Keep: true},
	}
	sheet.Guest.Data = map[int32]*sheeter.Guest{
		501: {ID: 501, Name: "老饕", Score: 4, ScoreMax: 10, Morale: 5, MoraleMax: 8, Calm: 3, SateMax: 6, SateSeal: true,
			SateSkillID: []string{"5^301", "7^301"}, CalmSkillID: []string{"9^301", "7^301"}},
	}
	sheet.Skill.Data = map[int32]*sheeter.Skill{
		301: {ID: 301, Name: "開朗", EffectID: []int32{401, 401, 402}},
	}
	sheet.Effect.Data = map[int32]*sheeter.Effect{ // 401 帶滿安全靜態欄(命令欄留空——testGame 無 compiler, 非空會炸 prepareEffect)
		401: {ID: 401, Name: "加耐", Kind: 1, RunOrder: 5, TargetKind: 2, TargetCount: 1, RunRound: 3,
			Stack: 1, StackMax: 3, StackTime: 1, TriggerKind: "cardPlay", TriggerCond: "self.sate > 0", TriggerCount: "self.calm"},
		402: {ID: 402, Name: "護盾", Kind: 2, RunOrder: 9},
		403: {ID: 403, Name: "立即", Kind: 0, RunOrder: 9},
	}
	return sheet
}

// testGame 空白營業實例(盤面直讀渲染用): 吃 testSheet、不裝詞彙、不驅動,
// 容器與屬性由各測試自行擺盤(port 全 nil; 渲染唯讀, 不會觸碰)。
func testGame() *cores.Game {
	return cores.NewGame(0, 0, cores.NewData(testSheet(), nil), nil, nil, nil)
}
