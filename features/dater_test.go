package features

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sheeter "github.com/yinweli/Project_005/sheet"
)

func TestSuiteDater(t *testing.T) {
	suite.Run(t, new(SuiteDater))
}

// SuiteDater 驗證 sheetdata 載入管線：Loader → Sheeter → Dater 端到端串接。
type SuiteDater struct {
	suite.Suite
}

func (this *SuiteDater) newDater() *Dater {
	loader := NewLoader("../sheetdata")
	data := sheeter.NewSheeter(loader)
	this.Require().True(data.FromData())
	this.Require().NoError(loader.Err())
	return NewDater(data)
}

func (this *SuiteDater) TestSetting() {
	dater := this.newDater()
	morale := dater.Setting("Morale")
	this.Require().NotNil(morale)
	this.Equal([]string{"100"}, morale.Value)
	this.NotNil(dater.Setting("RoundMax"))
	this.Nil(dater.Setting("NotExist"))
}

func (this *SuiteDater) TestMissing() {
	dater := this.newDater()
	// 其餘表格目前無資料，查詢應安全回傳 nil（不 panic）
	this.Nil(dater.Card(1))
	this.Nil(dater.Guest(1))
	this.Nil(dater.Skill(1))
	this.Nil(dater.Effect(1))
	this.Nil(dater.Seat(1))
	this.Nil(dater.Award(1))
}
