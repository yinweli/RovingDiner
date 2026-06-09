package cores

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuitePrepare(t *testing.T) {
	suite.Run(t, new(SuitePrepare))
}

// SuitePrepare 驗證自表格整理衍生資料（prepare.go）:抽獎候選索引、預編譯效果索引（含寬鬆跳過）。
type SuitePrepare struct {
	suite.Suite
}

func (this *SuitePrepare) TestPrepareAward() {
	award := prepareAward(buildSheet())
	this.Equal([]int32{101, 102}, award[7].cardID) // 群組 7:101(w3) / 102(w1)
	this.Equal([]int32{3, 1}, award[7].weight)
	this.NotContains(award, int32(8)) // 群組 8 全 0 權重 → 不收
	this.Equal([]int32{999}, award[9].cardID)

	this.Empty(prepareAward(nil)) // data nil → 空
}

func (this *SuitePrepare) TestPrepareEffect() {
	stub := func(source string) (EffectCommand, error) {
		if source == "bad" {
			return nil, errors.New("compile failed")
		} // if

		return func(*Engine) {}, nil
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
