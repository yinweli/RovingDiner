package infra

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteSheet(t *testing.T) {
	suite.Run(t, new(SuiteSheet))
}

// SuiteSheet 驗證 sheetdata 載入管線：Loader → Sheeter 端到端串接。
type SuiteSheet struct {
	suite.Suite
}

func (this *SuiteSheet) TestLoad() {
	data, err := Load("../sheetdata")
	this.Require().NoError(err)
	this.Require().NotNil(data)

	morale := data.Setting.Get("Morale")
	this.Require().NotNil(morale)
	this.Equal([]string{"100"}, morale.Value)
	this.NotNil(data.Setting.Get("RoundMax"))
	this.Nil(data.Setting.Get("NotExist"))

	// 其餘表格目前無資料，核心直接以 reader.Get 查詢應安全回傳 nil（不 panic）
	this.Nil(data.Card.Get(1))
	this.Nil(data.Guest.Get(1))
	this.Nil(data.Skill.Get(1))
	this.Nil(data.Effect.Get(1))
	this.Nil(data.Seat.Get(1))
	this.Nil(data.Award.Get(1))
}

func (this *SuiteSheet) TestLoadMissing() {
	data, err := Load(this.T().TempDir())
	this.Require().Error(err) // 目錄無任何 json，Loader 累積錯誤
	this.Nil(data)
}
