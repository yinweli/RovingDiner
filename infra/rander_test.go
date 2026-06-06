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

// TODO: Shuffle 也要有單元測試

func (this *SuiteRander) TestRanderWeighted() {
	rander := NewRander(7)

	// 空 / 零權重回傳 -1（呼叫端據此判定 no-op）
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
