package infra

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// Load 從指定目錄裝載全部 sheetdata，回傳 ready-to-use 的 *sheeter.Sheeter。
// 這份由 Sheeter 雙語言生成的資料聚合即核心的靜態資料 port，核心直接使用；
// 衍生索引（如抽獎依群組聚合）由核心自行預建，不在本層。
func Load(dir string) (data *sheeter.Sheeter, err error) {
	load := &loader{dir: dir}
	data = sheeter.NewSheeter(load)
	ok := data.FromData()

	if err = load.Err(); err != nil {
		return nil, err
	} // if

	if ok == false {
		return nil, errors.New("infra: load sheet failed: " + dir)
	} // if

	return data, nil
}

// loader 從檔案系統讀取 sheetdata 的裝載器；實作 sheeter.Loader，僅供 Load 內部使用。
// Load / Error 會被 Sheeter 於多個 goroutine 併發呼叫，故須維持執行緒安全。
type loader struct {
	dir  string     // sheetdata 目錄
	lock sync.Mutex // 保護 errs
	errs []error    // 載入過程累積的錯誤
}

// Load 讀取檔案內容；找不到 / 讀取失敗時回傳 nil 並記錄錯誤。
func (this *loader) Load(filename sheeter.FileName) []byte {
	path := filepath.Join(this.dir, filename.File())
	data, err := os.ReadFile(path)

	if err != nil {
		this.Error(filename.File(), err)
		return nil
	} // if

	return data
}

// Error 記錄載入過程中發生的錯誤。
func (this *loader) Error(name string, err error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.errs = append(this.errs, fmt.Errorf("loader: %s: %w", name, err))
}

// Err 回傳載入過程累積的第一個錯誤；無錯誤時回傳 nil。
func (this *loader) Err() error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if len(this.errs) == 0 {
		return nil
	} // if

	return this.errs[0]
}
