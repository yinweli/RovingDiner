package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteView(t *testing.T) {
	suite.Run(t, new(SuiteView))
}

// SuiteView 驗證實例視圖家族(view.go): 出生初值的靜態欄複製。
type SuiteView struct {
	suite.Suite
}

// TestNewCardView 驗證卡牌視圖初值: 費用 / 額外發動 / 四旗標鎖照表複製; 查無資料 → 零值視圖。
func (this *SuiteView) TestNewCardView() {
	target := newCardView(testSheet(), 101)
	this.Equal(int32(101), target.dataID)
	this.Equal(float64(2), target.attr["cost"])
	this.Equal(float64(1), target.attr["extraRunMin"])
	this.Equal(float64(3), target.attr["extraRunMax"])
	this.Equal(float64(1), target.lock["cardSeal"])
	this.Equal(float64(0), target.lock["keep"])
	this.Equal(float64(0), target.lock["playExile"])
	this.Equal(float64(0), target.lock["unplayExile"])

	missing := newCardView(testSheet(), 999) // 查無資料 → 零值視圖
	this.Equal(int32(999), missing.dataID)
	this.Empty(missing.attr)
	this.NotNil(missing.attr)
}

// TestNewGuestView 驗證顧客視圖初值: 七數值 / 兩封印鎖照表複製(飽食出生 0); 查無資料 → 零值視圖。
func (this *SuiteView) TestNewGuestView() {
	target := newGuestView(testSheet(), 501)
	this.Equal(int32(501), target.dataID)
	this.Equal(float64(0), target.attr["sate"])
	this.Equal(float64(6), target.attr["sateMax"])
	this.Equal(float64(3), target.attr["calm"])
	this.Equal(float64(4), target.attr["score"])
	this.Equal(float64(10), target.attr["scoreMax"])
	this.Equal(float64(5), target.attr["morale"])
	this.Equal(float64(8), target.attr["moraleMax"])
	this.Equal(float64(1), target.lock["sateSeal"])
	this.Equal(float64(0), target.lock["calmSeal"])

	missing := newGuestView(testSheet(), 999) // 查無資料 → 零值視圖
	this.Equal(int32(999), missing.dataID)
	this.Empty(missing.attr)
	this.NotNil(missing.effectImmune) // 免疫表出生恆空但可用
}

// TestGuestViewHasImmune 驗證免疫旗標判定: 任一類任一群組計數 > 0 即命中; 歸零不命中。
func (this *SuiteView) TestGuestViewHasImmune() {
	target := newGuestView(testSheet(), 501)
	this.False(target.hasImmune())

	target.effectImmune[5] = 1
	this.True(target.hasImmune())

	target.effectImmune[5] = 0
	target.skillImmune[9] = 2
	this.True(target.hasImmune())

	target.skillImmune[9] = 0
	this.False(target.hasImmune())
}

// TestLockInit 驗證靜態旗標轉鎖定計數初值。
func (this *SuiteView) TestLockInit() {
	this.Equal(float64(1), lockInit(true))
	this.Equal(float64(0), lockInit(false))
}
