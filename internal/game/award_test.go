package game

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteAward(t *testing.T) {
	suite.Run(t, new(SuiteAward))
}

// SuiteAward 驗證抽獎群組衍生索引的聚合與排序決定性。
type SuiteAward struct {
	suite.Suite
}

func (this *SuiteAward) TestBuildAwardIndex() {
	data := &sheeter.Sheeter{}
	data.Award.Data = map[int32]*sheeter.Award{
		1: {ID: 3, Group: 10, CardID: 100, Weight: 1},
		2: {ID: 1, Group: 10, CardID: 101, Weight: 2},
		3: {ID: 2, Group: 10, CardID: 102, Weight: 3},
		4: {ID: 5, Group: 20, CardID: 200, Weight: 1},
	}

	award := BuildAwardIndex(data)

	// 群組 10 依 ID 排序（不受 map 走訪順序影響）
	group := award[10]
	this.Require().Len(group, 3)
	this.Equal(int32(1), group[0].ID)
	this.Equal(int32(2), group[1].ID)
	this.Equal(int32(3), group[2].ID)

	// 群組 20 僅單筆
	this.Require().Len(award[20], 1)
	this.Equal(int32(5), award[20][0].ID)

	// 不存在的群組回傳 nil
	this.Nil(award[99])
}

func (this *SuiteAward) TestEmpty() {
	this.Empty(BuildAwardIndex(&sheeter.Sheeter{})) // 無資料回傳空索引
}
