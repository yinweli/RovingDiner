package game

import (
	"sort"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// BuildAwardIndex 依抽獎群組聚合 Award 資料並依 ID 排序（確保決定性）。
// 此為衍生索引，raw *sheeter.Sheeter 不直接提供；核心於初始化時建一次快取（M4 串接）。
// 對應營業實作規格書【六、Sheet 整合】：一次性預處理移入核心，作為可移植邏輯。
func BuildAwardIndex(data *sheeter.Sheeter) map[int32][]*sheeter.Award {
	award := map[int32][]*sheeter.Award{}

	for _, itor := range data.Award.Values() {
		award[itor.Group] = append(award[itor.Group], itor)
	} // for

	for _, itor := range award {
		sort.Slice(itor, func(i, j int) bool {
			return itor[i].ID < itor[j].ID
		})
	} // for

	return award
}
