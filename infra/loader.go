package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// NewLoader 建立指向指定 sheetdata 目錄的裝載器。
func NewLoader(dir string) *Loader {
	return &Loader{dir: dir}
}

// Loader 從檔案系統讀取 sheetdata 的裝載器；實作 sheeter.Loader。
// Load / Error 會被 Sheeter 於多個 goroutine 併發呼叫，故須維持執行緒安全。
type Loader struct {
	dir  string     // sheetdata 目錄
	lock sync.Mutex // 保護 errs
	errs []error    // 載入過程累積的錯誤
}

// Load 讀取檔案內容；找不到 / 讀取失敗時回傳 nil 並記錄錯誤。
func (this *Loader) Load(filename sheeter.FileName) []byte {
	path := filepath.Join(this.dir, filename.File())
	data, err := os.ReadFile(path)

	if err != nil {
		this.Error(filename.File(), err)
		return nil
	} // if

	return data
}

// Error 記錄載入過程中發生的錯誤。
func (this *Loader) Error(name string, err error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.errs = append(this.errs, fmt.Errorf("loader: %s: %w", name, err))
}

// Err 回傳載入過程累積的第一個錯誤；無錯誤時回傳 nil。
func (this *Loader) Err() error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if len(this.errs) == 0 {
		return nil
	} // if

	return this.errs[0]
}
