// Package features 是 Go-only 的基礎設施層（不移植 C#）：
// sheet 載入、Dater 組裝、PRNG 來源。與 sheet/ 同層、非 internal。
//
// 載入流程：Loader 讀 sheetdata/*.json → sheeter.NewSheeter(loader).FromData()
// 建立 *sheeter.Sheeter → NewDater(...) 包成 game.Dater（含衍生索引）。
package features

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	sheeter "github.com/yinweli/Project_005/sheet"
)

// Loader 從檔案系統讀取 sheetdata 的裝載器；實作 sheeter.Loader。
// Load / Error 會被 Sheeter 於多個 goroutine 併發呼叫，故須維持執行緒安全。
type Loader struct {
	dir  string     // sheetdata 目錄
	lock sync.Mutex // 保護 errs
	errs []error    // 載入過程累積的錯誤
}

// NewLoader 建立指向指定 sheetdata 目錄的裝載器。
func NewLoader(dir string) *Loader {
	return &Loader{dir: dir}
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
