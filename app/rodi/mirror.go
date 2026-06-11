package rodi

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// mirror 世界鏡像(顯示側狀態; 【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 四之二】單向事件投影):
// 事件摺疊一次、各組件唯讀共用(M20 拍板)——回合 / 階段取自每筆事件的座標蓋章, 全域屬性取自 property 事件
// (對象欄零值 = 全域; 開局六鍵快照使開局值亦可摺疊, 鏡像免自查設定表格)。
// 規則狀態以鏡像為單一來源, 組件只自持 UI 狀態(游標 / 捲動 / 模式)。
type mirror struct {
	round int32              // 當前回合(事件座標)
	phase cores.PhaseKind    // 當前階段(事件座標)
	attr  map[string]float64 // 全域屬性值(寫側詞條鍵 → 後值)
	lock  map[string]float64 // 全域屬性鎖定計數(寫側詞條鍵 → 後值; @ / # 事件載鎖定計數)
}

func newMirror() *mirror {
	return &mirror{attr: map[string]float64{}, lock: map[string]float64{}}
}

// Apply 摺疊一筆事件: 座標一律更新; 全域 property 依賦值符分流——@ / # 入鎖定計數表、其餘入值表。
// 引用屬性(對象欄非零)與其他事件類別在 M20 不投影(六區 / 日誌的摺疊隨 M21 / M22 增補)。
func (this *mirror) Apply(eventData cores.EventData) {
	this.round = eventData.Round
	this.phase = eventData.Phase

	if eventData.Kind != cores.EventProperty || eventData.InstanceID != cores.NoneID {
		return
	} // if

	switch eventData.Op {
	case cores.AssignLock, cores.AssignUnlock:
		this.lock[eventData.Attr] = eventData.After

	default:
		this.attr[eventData.Attr] = eventData.After
	} // switch
}
