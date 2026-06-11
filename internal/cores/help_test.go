package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteHelp(t *testing.T) {
	suite.Run(t, new(SuiteHelp))
}

// SuiteHelp 驗證 help.go 的組件: 免疫計數 / 門檻集合 / 編號列表 / 累積計數。
type SuiteHelp struct {
	suite.Suite
}

// TestNewImmune 驗證 NewImmune 建構空計數。
func (this *SuiteHelp) TestNewImmune() {
	immune := NewImmune()
	this.Equal(int32(0), immune.Get(1)) // 空表 → 0
}

// TestImmuneAdd 驗證 Add 計數遞增、零值自建表。
func (this *SuiteHelp) TestImmuneAdd() {
	immune := Immune{} // 零值可用

	immune.Add(5)
	immune.Add(5)
	this.Equal(int32(2), immune.Get(5))
}

// TestImmuneDel 驗證 Del 計數遞減、夾 ≥ 0、空表安全。
func (this *SuiteHelp) TestImmuneDel() {
	immune := NewImmune()
	immune.Add(5)

	immune.Del(5)
	this.Equal(int32(0), immune.Get(5))

	immune.Del(5) // 已 0 再減 → 夾 ≥ 0
	this.Equal(int32(0), immune.Get(5))

	zero := Immune{}
	zero.Del(9) // 零值(nil 表)→ 視為 0, 不爆
	this.Equal(int32(0), zero.Get(9))
}

// TestImmuneGet 驗證 Get 讀計數、無鍵回 0。
func (this *SuiteHelp) TestImmuneGet() {
	immune := Immune{count: map[int32]int32{7: 2}}
	this.Equal(int32(2), immune.Get(7))
	this.Equal(int32(0), immune.Get(99)) // 無鍵 → 0
}

// TestImmuneAny 驗證 Any 任一群組計數 > 0; 歸零鍵與空表回 false。
func (this *SuiteHelp) TestImmuneAny() {
	immune := Immune{count: map[int32]int32{7: 1}}
	this.True(immune.Any())

	immune.Del(7) // 歸零鍵仍在表 → false
	this.False(immune.Any())

	zero := Immune{}
	this.False(zero.Any())
}

// TestNewHit 驗證 NewHit 建構空集合。
func (this *SuiteHelp) TestNewHit() {
	hit := NewHit()
	this.Equal(int32(0), hit.Count())
}

// TestHitAdd 驗證 Add 標記門檻、零值自建表、重複標記不增量。
func (this *SuiteHelp) TestHitAdd() {
	hit := Hit{} // 零值可用

	hit.Add(10)
	hit.Add(10) // 重複標記 → 集合語意不增量
	this.True(hit.IsHit(10))
	this.Equal(int32(1), hit.Count())
}

// TestHitIsHit 驗證 IsHit 已觸發判定。
func (this *SuiteHelp) TestHitIsHit() {
	hit := Hit{hit: map[int32]bool{10: true}}
	this.True(hit.IsHit(10))
	this.False(hit.IsHit(20)) // 未標記 → false

	zero := Hit{}
	this.False(zero.IsHit(10)) // 零值(nil 表)→ false, 不爆
}

// TestHitCount 驗證 Count 已觸發數量。
func (this *SuiteHelp) TestHitCount() {
	hit := Hit{hit: map[int32]bool{10: true, 20: true}}
	this.Equal(int32(2), hit.Count())

	zero := Hit{}
	this.Equal(int32(0), zero.Count()) // 零值 → 0
}

// TestNewIDList 驗證 NewIDList 複製輸入建構(不共享底層)。
func (this *SuiteHelp) TestNewIDList() {
	source := []int32{1, 2}
	list := NewIDList(source...)
	source[0] = 9 // 改輸入不影響列表

	this.Equal([]int32{1, 2}, list.List())

	empty := NewIDList() // 無輸入 → 空列表
	this.Empty(empty.List())
}

// TestIDListAdd 驗證 Add 尾端加入(可變參數、零值可用、允許重複)。
func (this *SuiteHelp) TestIDListAdd() {
	list := IDList{} // 零值可用

	list.Add(5)
	list.Add(7, 5) // 批次 + 重複
	this.Equal([]int32{5, 7, 5}, list.List())
}

// TestIDListDelOne 驗證 DelOne 移除第一個命中、無命中不動。
func (this *SuiteHelp) TestIDListDelOne() {
	list := NewIDList(5, 7, 5)

	list.DelOne(5)
	this.Equal([]int32{7, 5}, list.List()) // 僅移第一個

	list.DelOne(9) // 無命中 → 不動
	this.Equal([]int32{7, 5}, list.List())
}

// TestIDListDelAll 驗證 DelAll 移除全部命中。
func (this *SuiteHelp) TestIDListDelAll() {
	list := NewIDList(5, 7, 5)

	list.DelAll(5)
	this.Equal([]int32{7}, list.List())

	list.DelAll(9) // 無命中 → 不動
	this.Equal([]int32{7}, list.List())
}

// TestIDListCount 驗證 Count 計數指定編號個數。
func (this *SuiteHelp) TestIDListCount() {
	list := NewIDList(5, 7, 5)
	this.Equal(int32(2), list.Count(5))
	this.Equal(int32(0), list.Count(9)) // 無命中 → 0
}

// TestIDListList 驗證 List 取底層編號列表(保持加入順序)。
func (this *SuiteHelp) TestIDListList() {
	list := NewIDList(3, 1, 2)
	this.Equal([]int32{3, 1, 2}, list.List())

	zero := IDList{}
	this.Empty(zero.List()) // 零值 → 空
}

// TestNewTally 驗證 NewTally 建構空累積計數。
func (this *SuiteHelp) TestNewTally() {
	total := NewTally()
	this.Equal(int32(0), total.Get(1))
	this.Equal(int32(0), total.Sum())
}

// TestTallyAdd 驗證 Add 對群組數量 +1; 零值未建表時自建。
func (this *SuiteHelp) TestTallyAdd() {
	total := Tally{} // 零值可用

	total.Add(1)
	total.Add(1)
	total.Add(2)
	this.Equal(int32(2), total.Get(1))
	this.Equal(int32(1), total.Get(2))
}

// TestTallyGet 驗證 Get 讀群組數量、無鍵回 0。
func (this *SuiteHelp) TestTallyGet() {
	total := Tally{count: map[int32]int32{1: 3}}
	this.Equal(int32(3), total.Get(1))
	this.Equal(int32(0), total.Get(9)) // 無鍵 → 0
}

// TestTallySum 驗證 Sum 全群組加總; 零值回 0。
func (this *SuiteHelp) TestTallySum() {
	total := Tally{count: map[int32]int32{1: 3, 2: 5}}
	this.Equal(int32(8), total.Sum())

	zero := Tally{}
	this.Equal(int32(0), zero.Sum()) // 零值 → 0
}

// TestTallyReset 驗證 Reset 清空全部計數(清空後可繼續累計)。
func (this *SuiteHelp) TestTallyReset() {
	total := Tally{count: map[int32]int32{1: 3, 2: 5}}

	total.Reset()
	this.Equal(int32(0), total.Sum())

	total.Add(1) // 清空後可繼續累計
	this.Equal(int32(1), total.Get(1))
}
