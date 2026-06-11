package cores

import (
	"strconv"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// 識別碼工具(【營業顯示規格書 | 5、顯示慣例：識別碼格式】): 模板「資料編號@名稱#實例編號」。
// 名稱憑靜態表自查; 主畫面省實例段 = 呼叫端傳 NoneID; 技能為靜態、任何層都無實例段;
// 空物件由呼叫端顯「空」, 不套模板。查無資料列名稱顯 ?(寬鬆)。
// 日誌行合成(發射端)與盤面組件(顯示端)共用同一份, 故居 cores。

// IdentCard 卡牌識別碼。
func IdentCard(sheet *sheeter.Sheeter, dataID int32, instanceID InstanceID) string {
	name := "?"

	if meta := sheet.Card.Get(dataID); meta != nil {
		name = meta.Name
	} // if

	return identText(dataID, name, instanceID)
}

// IdentGuest 顧客識別碼。
func IdentGuest(sheet *sheeter.Sheeter, dataID int32, instanceID InstanceID) string {
	name := "?"

	if meta := sheet.Guest.Get(dataID); meta != nil {
		name = meta.Name
	} // if

	return identText(dataID, name, instanceID)
}

// IdentSkill 技能識別碼(靜態無實例段)。
func IdentSkill(sheet *sheeter.Sheeter, dataID int32) string {
	name := "?"

	if meta := sheet.Skill.Get(dataID); meta != nil {
		name = meta.Name
	} // if

	return identText(dataID, name, NoneID)
}

// IdentEffect 效果識別碼。
func IdentEffect(sheet *sheeter.Sheeter, dataID int32, instanceID InstanceID) string {
	name := "?"

	if meta := sheet.Effect.Get(dataID); meta != nil {
		name = meta.Name
	} // if

	return identText(dataID, name, instanceID)
}

// IdentTarget 辨型識別碼: 對象可能是卡牌或顧客(容器移動者 / 效果 self / 選中清單元素)時,
// 依資料編號先查卡牌表、再查顧客表辨型(實務上編號區段不相撞); 兩表皆查無顯 ?。
func IdentTarget(sheet *sheeter.Sheeter, dataID int32, instanceID InstanceID) string {
	if sheet.Card.Get(dataID) != nil {
		return IdentCard(sheet, dataID, instanceID)
	} // if

	return IdentGuest(sheet, dataID, instanceID)
}

// identText 模板組裝: 資料編號@名稱 + 可省略的 #實例編號(NoneID 即省略)。
func identText(dataID int32, name string, instanceID InstanceID) string {
	text := strconv.FormatInt(int64(dataID), 10) + "@" + name

	if instanceID != NoneID {
		text += "#" + strconv.FormatInt(int64(instanceID), 10)
	} // if

	return text
}
