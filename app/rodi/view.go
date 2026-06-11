package rodi

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// 實例視圖家族: 世界鏡像對單一實例的摺疊狀態。初值查靜態表(出生即欄位複製, 與 NewCard / NewGuest 同源;
// M21 拍板②: 靜態表查得到的查表)、之後靠 property 事件摺疊變化; 查無資料列 → 零值視圖(寬鬆, 比照引擎)。

// cardView 卡牌實例視圖; 屬性 / 鎖定鍵 = 寫側詞條鍵(rules attrRefWrite), property 事件直接以 Attr 入表。
type cardView struct {
	dataID         int32              // 卡牌資料編號
	attr           map[string]float64 // 引用屬性值(cost 等)
	lock           map[string]float64 // 引用屬性鎖定計數(cardSeal / keep / playExile / unplayExile 等)
	bindID         int32              // 卡牌化來源顧客資料編號(0 = 未綁; 容器事件語境欄摺入)
	bindInstanceID cores.InstanceID   // 卡牌化來源顧客實例編號(同上)
}

// newCardView 依靜態表建卡牌視圖初值(費用 + 四旗標鎖; 對應 NewCard 的欄位複製)。
// cardify 的不棄 +1 不發事件, 顯示端以「keep 鎖 > 0 或已綁來源」判不棄旗標(綁定生命週期即該 +1)。
func newCardView(sheet *sheeter.Sheeter, dataID int32) *cardView {
	view := &cardView{dataID: dataID, attr: map[string]float64{}, lock: map[string]float64{}}
	meta := sheet.Card.Get(dataID)

	if meta == nil {
		return view // 查無資料 → 零值視圖
	} // if

	view.attr["cost"] = float64(meta.Cost)
	view.attr["extraRunMin"] = float64(meta.ExtraRunMin)
	view.attr["extraRunMax"] = float64(meta.ExtraRunMax)
	view.lock["cardSeal"] = lockInit(meta.Seal)
	view.lock["keep"] = lockInit(meta.Keep)
	view.lock["playExile"] = lockInit(meta.PlayExile)
	view.lock["unplayExile"] = lockInit(meta.UnplayExile)
	return view
}

// guestView 顧客實例視圖; 鍵規則同 cardView。
type guestView struct {
	dataID int32              // 顧客資料編號
	attr   map[string]float64 // 引用屬性值(sate / calm 等; sate 出生 0)
	lock   map[string]float64 // 引用屬性鎖定計數(sateSeal / calmSeal 等)
}

// newGuestView 依靜態表建顧客視圖初值(七數值 + 兩封印鎖; 對應 NewGuest 的欄位複製)。
func newGuestView(sheet *sheeter.Sheeter, dataID int32) *guestView {
	view := &guestView{dataID: dataID, attr: map[string]float64{}, lock: map[string]float64{}}
	meta := sheet.Guest.Get(dataID)

	if meta == nil {
		return view // 查無資料 → 零值視圖
	} // if

	view.attr["sate"] = 0
	view.attr["sateMax"] = float64(meta.SateMax)
	view.attr["calm"] = float64(meta.Calm)
	view.attr["score"] = float64(meta.Score)
	view.attr["scoreMax"] = float64(meta.ScoreMax)
	view.attr["morale"] = float64(meta.Morale)
	view.attr["moraleMax"] = float64(meta.MoraleMax)
	view.lock["sateSeal"] = lockInit(meta.SateSeal)
	view.lock["calmSeal"] = lockInit(meta.CalmSeal)
	return view
}

// effectView 效果佇列項視圖; 層數 / 結束回合摺自效果事件的快照欄(M21 拍板④)。
type effectView struct {
	effectID       int32            // 效果資料編號
	instanceID     cores.InstanceID // 效果實例編號
	stack          int32            // 當前層數(快照)
	expire         int32            // 結束回合(快照; 0 = 整場保留)
	selfID         int32            // self 資料編號(空物件 0)
	selfInstanceID cores.InstanceID // self 實例編號(空物件 0)
}

// actionView 行動佇列項視圖; 摺自 action 事件(M21 拍板③)。
type actionView struct {
	guestID         int32            // 顧客資料編號
	guestInstanceID cores.InstanceID // 顧客實例編號
	skillID         int32            // 技能資料編號
	task            cores.TaskKind   // 行動類型(飽食 / 耐心)
}

// lockInit 靜態 bool 旗標轉鎖定計數初值(true → 1; 對應 NewValueLock)。
func lockInit(lock bool) float64 {
	if lock {
		return 1
	} // if

	return 0
}
