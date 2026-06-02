package infra

import (
	"errors"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// Load 從指定目錄裝載全部 sheetdata，回傳 ready-to-use 的 *sheeter.Sheeter。
// 這份由 Sheeter 雙語言生成的資料聚合即核心的靜態資料 port，核心直接使用；
// 衍生索引（如抽獎依群組聚合）由核心自行預建，不在本層。
func Load(dir string) (*sheeter.Sheeter, error) {
	loader := NewLoader(dir)
	data := sheeter.NewSheeter(loader)
	ok := data.FromData()

	if err := loader.Err(); err != nil {
		return nil, err
	} // if

	if ok == false {
		return nil, errors.New("infra: load sheet failed: " + dir)
	} // if

	return data, nil
}
