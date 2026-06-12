package infra

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteRander(t *testing.T) {
	suite.Run(t, new(SuiteRander))
}

// SuiteRander 驗證單一 seeded PRNG 的決定性與邊界行為。
type SuiteRander struct {
	suite.Suite
}

func (this *SuiteRander) TestRanderIntn() {
	rander := NewRander(1)
	this.Equal(0, rander.Intn(0))  // n <= 0 回傳 0
	this.Equal(0, rander.Intn(-5)) // n <= 0 回傳 0

	for i := 0; i < 100; i++ {
		v := rander.Intn(10)
		this.GreaterOrEqual(v, 0)
		this.Less(v, 10)
	} // for
}

func (this *SuiteRander) TestRanderIntnDeterministic() {
	a := NewRander(42)
	b := NewRander(42)

	for i := 0; i < 100; i++ {
		this.Equal(a.Intn(1000), b.Intn(1000))
	} // for
}

func (this *SuiteRander) TestRanderShuffle() {
	rander := NewRander(1)

	// n = 0:不呼叫 swap、不 panic
	swapped := false
	rander.Shuffle(0, func(i, j int) { swapped = true })
	this.False(swapped)

	// n = 1:單一元素洗牌後不變
	single := []int{99}
	rander.Shuffle(len(single), func(i, j int) { single[i], single[j] = single[j], single[i] })
	this.Equal([]int{99}, single)

	// 洗牌為排列: 元素集合不增不減不重複, 且確實重排(非 no-op)
	origin := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	data := make([]int, len(origin))
	copy(data, origin)
	rander.Shuffle(len(data), func(i, j int) { data[i], data[j] = data[j], data[i] })
	this.ElementsMatch(origin, data) // 仍是同一組元素
	this.NotEqual(origin, data)      // 順序已改變(確認 Shuffle 真的有作用)
}

func (this *SuiteRander) TestRanderShuffleDeterministic() {
	a := NewRander(42)
	b := NewRander(42)

	dataA := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	dataB := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	a.Shuffle(len(dataA), func(i, j int) { dataA[i], dataA[j] = dataA[j], dataA[i] })
	b.Shuffle(len(dataB), func(i, j int) { dataB[i], dataB[j] = dataB[j], dataB[i] })
	this.Equal(dataA, dataB) // 同 seed 產生同一洗牌序
}

func (this *SuiteRander) TestRanderWeighted() {
	rander := NewRander(7)

	// 空 / 零權重回傳 -1(呼叫端據此判定 no-op)
	this.Equal(-1, rander.Weighted(nil))
	this.Equal(-1, rander.Weighted([]int32{0, 0}))

	// 單一非零權重必命中該索引
	this.Equal(1, rander.Weighted([]int32{0, 5, 0}))

	// 命中索引恆落在有效權重位置上
	for i := 0; i < 100; i++ {
		idx := rander.Weighted([]int32{3, 0, 7})
		this.Contains([]int{0, 2}, idx)
	} // for
}
