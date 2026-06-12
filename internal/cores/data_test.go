package cores

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteData(t *testing.T) {
	suite.Run(t, new(SuiteData))
}

// SuiteData 驗證遊戲資料(data.go): 聚合建構 / 原始表與預編譯效果存取 / 衍生索引整理(含寬鬆跳過)。
type SuiteData struct {
	suite.Suite
}

// TestNewData 驗證 NewData 聚合原始表並整理衍生索引; sheet nil → 空表空索引。
func (this *SuiteData) TestNewData() {
	sheet := buildSheet()
	data := NewData(sheet, nil)
	this.Require().NotNil(data)
	this.Same(sheet, data.GetSheet())
	this.NotEmpty(data.award)  // 抽獎索引已建
	this.NotEmpty(data.effect) // 預編譯效果索引已建

	empty := NewData(nil, nil)
	this.Equal(&sheeter.Sheeter{}, empty.GetSheet()) // sheet nil 正規化為空表(查詢與識別碼查名不爆)
	this.Empty(empty.award)
	this.Empty(empty.effect)
}

// TestDataGetSheet 驗證 GetSheet 取回原始靜態表。
func (this *SuiteData) TestDataGetSheet() {
	sheet := buildSheet()
	this.Same(sheet, NewData(sheet, nil).GetSheet())
}

// TestDataGetAward 驗證 GetAward 查抽獎群組候選、查無回 ok=false。
func (this *SuiteData) TestDataGetAward() {
	data := NewData(buildSheet(), nil)

	meta, ok := data.GetAward(7)
	this.True(ok)
	this.Equal([]int32{101, 102}, meta.cardID)

	_, ok = data.GetAward(999) // 查無 → 失敗
	this.False(ok)
}

// TestDataGetEffect 驗證 GetEffect 查預編譯效果、查無回 ok=false。
func (this *SuiteData) TestDataGetEffect() {
	data := NewData(buildSheet(), nil)

	meta, ok := data.GetEffect(401)
	this.True(ok)
	this.Equal(int32(5), meta.Group)

	_, ok = data.GetEffect(999) // 查無 → 失敗
	this.False(ok)
}

// TestDataSetEffect 驗證 SetEffect 組裝期補登 / 覆寫預編譯效果(測試注入自訂閉包用)。
func (this *SuiteData) TestDataSetEffect() {
	data := NewData(nil, nil)

	data.SetEffect(800, EffectData{Group: 9})
	meta, ok := data.GetEffect(800)
	this.True(ok)
	this.Equal(int32(9), meta.Group)

	data.SetEffect(800, EffectData{Group: 5}) // 覆寫
	meta, _ = data.GetEffect(800)
	this.Equal(int32(5), meta.Group)
}

// TestDataGetGuest 驗證 GetGuest 查顧客門檻配對、查無回 ok=false(即無門檻)。
func (this *SuiteData) TestDataGetGuest() {
	sheet := &sheeter.Sheeter{}
	sheet.Guest.Data = map[int32]*sheeter.Guest{
		501: {ID: 501, SateSkillID: []string{"6^301"}},
	}
	data := NewData(sheet, nil)

	meta, ok := data.GetGuest(501)
	this.True(ok)
	this.Equal([]Threshold{{Value: 6, SkillID: 301}}, meta.Sate)

	_, ok = data.GetGuest(999) // 查無 → 失敗
	this.False(ok)
}

func (this *SuiteData) TestPrepareAward() {
	award := prepareAward(buildSheet())
	this.Equal([]int32{101, 102}, award[7].cardID) // 群組 7:101(w3) / 102(w1)
	this.Equal([]int32{3, 1}, award[7].weight)
	this.NotContains(award, int32(8)) // 群組 8 全 0 權重 → 不收
	this.Equal([]int32{999}, award[9].cardID)

	this.Empty(prepareAward(nil)) // data nil → 空
}

func (this *SuiteData) TestPrepareEffect() {
	stub := func(source string) (EffectExec, error) {
		if source == "bad" {
			return nil, errors.New("compile failed")
		} // if

		return func(*Game) {}, nil
	}

	data := &sheeter.Sheeter{}
	data.Effect.Data = map[int32]*sheeter.Effect{
		700: {
			ID: 700, Kind: int32(EffectTrigger), TriggerKind: "cardPlay",
			Group: 5, RunRound: 1, RunOrder: 10,
			Stack: 2, StackMax: 3, StackTime: int32(StackTimeRefresh),
			TargetKind: int32(TargetGuestRand), TargetCount: 2,
			CommandImmed: "run", CommandTrigger: "run", CommandStart: "run",
			TriggerCond: "morale > 5", TriggerCount: "2",
		}, // 成功
		701: {ID: 701, CommandTrigger: "bad"},                    // 觸發命令編譯失敗 → 跳過
		702: {ID: 702, CommandTrigger: "run", CommandEnd: "bad"}, // 結束命令編譯失敗 → 跳過
		703: {ID: 703, TriggerCond: "1 +"},                       // 觸發條件語法錯 → 跳過
		704: {ID: 704, TriggerCount: "1 +"},                      // 觸發次數語法錯 → 跳過
		705: {ID: 705, Kind: 99},                                 // 效果類型編碼越界 → 跳過
		706: {ID: 706, TriggerAfter: 99},                         // 觸發後行為編碼越界 → 跳過
		709: {ID: 709, StackTime: 99},                            // 堆疊時間編碼越界 → 跳過
		710: {ID: 710, TargetKind: 99},                           // 目標類型編碼越界 → 跳過
		707: {ID: 707, CommandImmed: "bad"},                      // 立即命令編譯失敗 → 跳過
		708: {ID: 708, CommandStart: "bad"},                      // 啟動命令編譯失敗 → 跳過
	}

	result := prepareEffect(data, stub)
	this.Require().Len(result, 1) // 僅 700 成功
	this.Require().Contains(result, int32(700))
	this.Equal(EffectTrigger, result[700].Kind)
	this.Equal(TriggerCardPlay, result[700].TriggerKind)
	this.NotNil(result[700].Trigger) // 命令編成執行器
	this.Nil(result[700].End)        // 空欄 → nil
	this.NotNil(result[700].Cond)
	this.NotNil(result[700].Count)
	this.Equal(int32(5), result[700].Group)
	this.Equal(int32(1), result[700].RunRound)
	this.Equal(int32(10), result[700].RunOrder)
	this.Equal(int32(2), result[700].Stack)
	this.Equal(int32(3), result[700].StackMax)
	this.Equal(StackTimeRefresh, result[700].StackTime)
	this.Equal(TargetGuestRand, result[700].TargetKind)
	this.Equal(int32(2), result[700].TargetCount)
	this.NotNil(result[700].Immed)
	this.NotNil(result[700].Start)

	this.Empty(prepareEffect(nil, stub)) // data nil → 空
}

// TestParseThreshold 驗證單筆門檻配對解析: 好格式回配對、壞格式(缺 ^ / 多段 / 非整數 / 空字串)回帶位置錯誤。
func (this *SuiteData) TestParseThreshold() {
	threshold, err := ParseThreshold("6^301")
	this.NoError(err)
	this.Equal(Threshold{Value: 6, SkillID: 301}, threshold)

	_, err = ParseThreshold("3") // 缺 ^
	this.Error(err)

	_, err = ParseThreshold("4^5^6") // 多段
	this.Error(err)

	_, err = ParseThreshold("x^1") // 門檻值非整數
	this.Error(err)

	_, err = ParseThreshold("") // 空字串: 未填即跳筆的通則由呼叫端把關
	this.Error(err)

	_, err = ParseThreshold("12^y") // 技能編號非整數: 位置指向第二段起點
	this.Require().Error(err)
	syntaxError := &exprs.SyntaxError{}
	this.Require().True(errors.As(err, &syntaxError))
	this.Equal(3, syntaxError.Pos)
}

// TestPrepareGuest 驗證顧客門檻衍生索引: 解析「門檻值^技能編號」、飽食升序 / 耐心降序、壞格式跳過該筆、無門檻不建項。
func (this *SuiteData) TestPrepareGuest() {
	data := &sheeter.Sheeter{}
	data.Guest.Data = map[int32]*sheeter.Guest{
		501: {ID: 501, SateSkillID: []string{"9^301", "6^302"}, CalmSkillID: []string{"1^303", "2^304"}},
		502: {ID: 502, SateSkillID: []string{"x^1", "2^y", "3", "4^5^6"}}, // 全壞格式(非數字 / 缺 ^ / 多段)→ 不建項
		503: {ID: 503, CalmSkillID: []string{"bad", "5^301"}},             // 壞筆跳過、好筆保留
		504: {ID: 504},                                                    // 無門檻 → 不建項
	}

	result := prepareGuest(data)
	this.Require().Contains(result, int32(501))
	this.Equal([]Threshold{{Value: 6, SkillID: 302}, {Value: 9, SkillID: 301}}, result[501].Sate) // 升序
	this.Equal([]Threshold{{Value: 2, SkillID: 304}, {Value: 1, SkillID: 303}}, result[501].Calm) // 降序
	this.NotContains(result, int32(502))
	this.Equal([]Threshold{{Value: 5, SkillID: 301}}, result[503].Calm)
	this.NotContains(result, int32(504))

	this.Empty(prepareGuest(nil)) // data nil → 空
}

// === 測試輔助(置尾) ===

// buildSheet 組裝測試用 in-memory 靜態表; 與 tester.BuildSheet 同內容的 cores 白箱測試私有複本
// (cores 的同套件測試不能 import tester——tester import cores 會成環)。
func buildSheet() *sheeter.Sheeter {
	data := &sheeter.Sheeter{}
	data.Seat.Data = map[int32]*sheeter.Seat{
		1: {ID: 1, TableID: 1, SameSeatID: []int32{1, 2}, NearSeatID: []int32{3}},
		2: {ID: 2, TableID: 1, SameSeatID: []int32{1, 2}, NearSeatID: []int32{3}},
		3: {ID: 3, TableID: 2, SameSeatID: []int32{3}, NearSeatID: []int32{1, 2}},
	}
	data.Card.Data = map[int32]*sheeter.Card{
		101: {ID: 101, Group: 1},
		102: {ID: 102, Group: 2},
		103: {ID: 103, Group: 1, Cost: 2, Keep: true, Seal: true, SkillID: 301}, // bool 欄 → 鎖、SkillID → 效果列表
	}
	data.Effect.Data = map[int32]*sheeter.Effect{
		201: {ID: 201, Group: 5},
		401: {ID: 401, Group: 5, RunOrder: 10, RunRound: 2}, // RunRound>0 → Expire = 建立回合 + 1; RunOrder 10
		402: {ID: 402, Group: 5, RunOrder: 10, RunRound: 0}, // RunRound 0 → Expire 0; 與 401 同序、EffectID 402 > 401
		403: {ID: 403, Group: 6, RunOrder: 20, RunRound: 1}, // RunOrder 20 最大 → 排序最先
	}
	data.Skill.Data = map[int32]*sheeter.Skill{
		301: {ID: 301, Group: 3, EffectID: []int32{401, 402}}, // 卡 103 的技能(群組 3)效果列表
	}
	data.Award.Data = map[int32]*sheeter.Award{
		1: {ID: 1, Group: 7, CardID: 101, Weight: 3}, // 群組 7:候選 101(w3) / 102(w1)
		2: {ID: 2, Group: 7, CardID: 102, Weight: 1},
		3: {ID: 3, Group: 8, CardID: 101, Weight: 0}, // 群組 8:權重 0 → roll no-op
		4: {ID: 4, Group: 9, CardID: 999, Weight: 1}, // 群組 9:抽中編號 999 無卡牌資料 → 跳過該張
	}
	data.Guest.Data = map[int32]*sheeter.Guest{
		501: {ID: 501, Score: 0, ScoreMax: 10, Morale: 5, MoraleMax: 8, Calm: 3, SateMax: 12, SateSeal: true},
	}
	return data
}
