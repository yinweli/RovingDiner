package infra

import (
	"math/rand"
)

// Rander 單一 seeded PRNG，實作 cores.Rander。
// 同 seed 產生同序列，是回歸測試與 bug 重現的基礎；核心不得碰全域亂數。
type Rander struct {
	rng *rand.Rand
}

// NewRander 以指定種子建立亂數來源。
func NewRander(seed int64) *Rander {
	return &Rander{rng: rand.New(rand.NewSource(seed))}
}

// Intn 回傳 [0, n) 的隨機整數；n <= 0 時回傳 0。
func (this *Rander) Intn(n int) int {
	if n <= 0 {
		return 0
	} // if

	return this.rng.Intn(n)
}

// Shuffle 對 n 個元素隨機洗牌。
func (this *Rander) Shuffle(n int, swap func(i, j int)) {
	this.rng.Shuffle(n, swap)
}

// Weighted 對權重列表做 weighted random，回傳命中索引；
// 總權重 <= 0 或列表為空時回傳 -1（呼叫端據此判定 no-op）。
func (this *Rander) Weighted(weight []int32) int {
	total := int32(0)

	for _, itor := range weight {
		if itor > 0 {
			total += itor
		} // if
	} // for

	if total <= 0 {
		return -1
	} // if

	roll := int32(this.rng.Intn(int(total)))

	for k, v := range weight {
		if v <= 0 {
			continue
		} // if

		roll -= v

		if roll < 0 {
			return k
		} // if
	} // for

	return -1
}
