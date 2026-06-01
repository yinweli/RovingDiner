package features

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteLoader(t *testing.T) {
	suite.Run(t, new(SuiteLoader))
}

// SuiteLoader 驗證 sheetdata 裝載器的讀取、錯誤累積與執行緒安全。
type SuiteLoader struct {
	suite.Suite
}

func (this *SuiteLoader) TestLoad() {
	dir := this.T().TempDir()
	content := []byte(`{"hello":"world"}`)
	this.Require().NoError(os.WriteFile(filepath.Join(dir, "award.json"), content, 0o600))

	loader := NewLoader(dir)
	this.Equal(content, loader.Load(sheeter.NewFileName("award", ".json"))) // 讀回原內容
	this.NoError(loader.Err())                                              // 成功不記錄錯誤
}

func (this *SuiteLoader) TestLoadMissing() {
	loader := NewLoader(this.T().TempDir())
	this.Nil(loader.Load(sheeter.NewFileName("nope", ".json"))) // 找不到回傳 nil
	this.Error(loader.Err())                                    // 並記錄錯誤
}

func (this *SuiteLoader) TestErr() {
	loader := NewLoader(this.T().TempDir())
	this.NoError(loader.Err()) // 無錯誤回傳 nil

	loader.Error("first", errors.New("e1"))
	loader.Error("second", errors.New("e2"))

	err := loader.Err()
	this.Require().Error(err)
	this.Contains(err.Error(), "first")     // 回傳累積的第一個錯誤
	this.NotContains(err.Error(), "second") // 後續錯誤不蓋過第一個
}

func (this *SuiteLoader) TestErrorConcurrent() {
	loader := NewLoader(this.T().TempDir())

	var wait sync.WaitGroup
	for i := 0; i < 50; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			loader.Error("concurrent", errors.New("boom"))
		}()
	} // for
	wait.Wait()

	this.Error(loader.Err()) // 併發記錄後仍可安全取得錯誤（搭配 -race 驗證無資料競爭）
}
