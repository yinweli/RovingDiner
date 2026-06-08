package infra

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteSheet(t *testing.T) {
	suite.Run(t, new(SuiteSheet))
}

// SuiteSheet 驗證 sheet 載入：對外 Load 的端到端串接，與內部 loader 的讀取／錯誤累積／執行緒安全。
type SuiteSheet struct {
	suite.Suite
}

func (this *SuiteSheet) TestLoad() {
	data, err := Load("../../sheetdata")
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
	this.Require().Error(err) // 目錄無任何 json，loader 累積錯誤
	this.Nil(data)
}

func (this *SuiteSheet) TestLoaderLoad() {
	dir := this.T().TempDir()
	content := []byte(`{"hello":"world"}`)
	this.Require().NoError(os.WriteFile(filepath.Join(dir, "award.json"), content, 0o600))

	load := &loader{dir: dir}
	this.Equal(content, load.Load(sheeter.NewFileName("award", ".json"))) // 讀回原內容
	this.NoError(load.Err())                                              // 成功不記錄錯誤
}

func (this *SuiteSheet) TestLoaderLoadMissing() {
	load := &loader{dir: this.T().TempDir()}
	this.Nil(load.Load(sheeter.NewFileName("nope", ".json"))) // 找不到回傳 nil
	this.Error(load.Err())                                    // 並記錄錯誤
}

func (this *SuiteSheet) TestLoaderErr() {
	load := &loader{dir: this.T().TempDir()}
	this.NoError(load.Err()) // 無錯誤回傳 nil

	load.Error("first", errors.New("e1"))
	load.Error("second", errors.New("e2"))

	err := load.Err()
	this.Require().Error(err)
	this.Contains(err.Error(), "first")     // 回傳累積的第一個錯誤
	this.NotContains(err.Error(), "second") // 後續錯誤不蓋過第一個
}

func (this *SuiteSheet) TestLoaderConcurrent() {
	load := &loader{dir: this.T().TempDir()}

	var wait sync.WaitGroup
	for i := 0; i < 50; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			load.Error("concurrent", errors.New("boom"))
		}()
	} // for
	wait.Wait()

	this.Error(load.Err()) // 併發記錄後仍可安全取得錯誤（搭配 -race 驗證無資料競爭）
}
