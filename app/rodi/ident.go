package rodi

import (
	"strconv"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// 識別碼工具(【營業顯示規格書 | 5、顯示慣例：識別碼格式】): 模板「資料編號@名稱#實例編號」。
// 名稱由前端憑同一份靜態表自查(M17 拍板: 事件只帶編號); 主畫面投影省實例段 = 呼叫端傳 cores.NoneID;
// 技能為靜態、任何層都無實例段; 空物件由呼叫端顯「空」, 不套模板。查無資料列名稱顯 ?(寬鬆)。

// identCard 卡牌識別碼。
func identCard(sheet *sheeter.Sheeter, dataID int32, instanceID cores.InstanceID) string {
	name := "?"

	if meta := sheet.Card.Get(dataID); meta != nil {
		name = meta.Name
	} // if

	return identText(dataID, name, instanceID)
}

// identGuest 顧客識別碼。
func identGuest(sheet *sheeter.Sheeter, dataID int32, instanceID cores.InstanceID) string {
	name := "?"

	if meta := sheet.Guest.Get(dataID); meta != nil {
		name = meta.Name
	} // if

	return identText(dataID, name, instanceID)
}

// identSkill 技能識別碼(靜態無實例段)。
func identSkill(sheet *sheeter.Sheeter, dataID int32) string {
	name := "?"

	if meta := sheet.Skill.Get(dataID); meta != nil {
		name = meta.Name
	} // if

	return identText(dataID, name, cores.NoneID)
}

// identEffect 效果識別碼。
func identEffect(sheet *sheeter.Sheeter, dataID int32, instanceID cores.InstanceID) string {
	name := "?"

	if meta := sheet.Effect.Get(dataID); meta != nil {
		name = meta.Name
	} // if

	return identText(dataID, name, instanceID)
}

// identText 模板組裝: 資料編號@名稱 + 可省略的 #實例編號(NoneID 即省略)。
func identText(dataID int32, name string, instanceID cores.InstanceID) string {
	text := strconv.FormatInt(int64(dataID), 10) + "@" + name

	if instanceID != cores.NoneID {
		text += "#" + strconv.FormatInt(int64(instanceID), 10)
	} // if

	return text
}
