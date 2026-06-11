package rodi

import (
	"strconv"
	"strings"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// journal 事件日誌轉寫器(【營業顯示規格書 | 6、畫面規格 | 6.10】): 把事件流逐筆轉寫為日誌行,
// 行歷史與歸因狀態自持於日誌側(M22 拍板)。行角色看首字: 範圍標題 [ / 操作元 * /
// 效果 - / 流程直屬 $ / 效果欄位(欄 2)。歸因狀態機: 範圍標題重置「當前效果」、效果行設定之,
// 其後的命令行有當前效果掛欄 2、否則為流程直屬行; phase 切換與行動出列不立行
// (前者階段值併入下一標題前綴、後者緊接的顧客行動標題已承載; M22 拍板)。
type journal struct {
	sheet          *sheeter.Sheeter // 靜態表(識別碼名稱自查)
	line           []string         // 全量行歷史(渲染端取尾段)
	effect         bool             // 歸因: 當前效果存在
	selfDataID     int32            // 歸因: 當前效果 self 資料編號(屬性開頭判定)
	selfInstanceID cores.InstanceID // 歸因: 當前效果 self 實例編號(同上)
	morph          string           // 變身配對: 暫存的銷毀舊實例識別碼(空字串 = 無)
}

func newJournal(sheet *sheeter.Sheeter) *journal {
	return &journal{sheet: sheet}
}

// Append 轉寫一筆事件(0..n 行)。
func (this *journal) Append(eventData cores.EventData) {
	switch eventData.Kind {
	case cores.EventScope:
		this.appendScope(eventData)

	case cores.EventEffect:
		this.appendEffect(eventData)

	case cores.EventProperty:
		this.appendBody(this.propertyBody(eventData))

	case cores.EventContainer:
		this.appendBody(this.containerBody(eventData))

	case cores.EventInstance:
		this.appendInstance(eventData)

	case cores.EventSelect:
		this.appendSelect(eventData)

	case cores.EventAction:
		if eventData.Alive {
			this.appendBody(identGuest(this.sheet, eventData.DataID, eventData.InstanceID) + " >> 行動佇列(" + taskText(eventData.Task) + ")")
		} // if

	case cores.EventPhase: // 不立行(M22 拍板)
	} // switch
}

// appendScope 範圍標題 + 操作元行: [R{回合} {階段}] {事件} 與其下 * 行(動作類依序為卡牌(或顧客)、技能;
// 技能編號零值防禦略過), 並重置歸因。
func (this *journal) appendScope(eventData cores.EventData) {
	this.effect = false
	this.selfDataID = 0
	this.selfInstanceID = cores.NoneID
	this.line = append(this.line, "[R"+strconv.FormatInt(int64(eventData.Round), 10)+" "+phaseName(eventData.Phase)+"] "+scopeText(eventData.Scope, eventData.Trigger))

	switch eventData.Scope {
	case cores.ScopePlay:
		this.line = append(this.line, "* "+identCard(this.sheet, eventData.DataID, eventData.InstanceID))

	case cores.ScopeGuest:
		this.line = append(this.line, "* "+identGuest(this.sheet, eventData.DataID, eventData.InstanceID))

	default: // 其餘範圍無對象操作元
	} // switch

	if eventData.SkillID != 0 {
		this.line = append(this.line, "* "+identSkill(this.sheet, eventData.SkillID))
	} // if
}

// appendEffect 效果行 + 效果欄位行(對象 / 階段; 命令行隨後逐筆抵達): - {效果識別碼}、欄 2 對象
// (無目標顯 空, 空欄佔行位置才穩)、欄 2 階段, 並把歸因設為本效果。
func (this *journal) appendEffect(eventData cores.EventData) {
	target := "空"

	if eventData.DataID != 0 {
		target = identTarget(this.sheet, eventData.DataID, eventData.InstanceID)
	} // if

	this.line = append(this.line, "- "+identEffect(this.sheet, eventData.EffectID, eventData.EffectInstanceID), "  "+target, "  "+stageText(eventData.Stage))
	this.effect = true
	this.selfDataID = eventData.DataID
	this.selfInstanceID = eventData.InstanceID
}

// appendBody 依歸因狀態機收命令行: 有當前效果 → 效果欄位(欄 2)、否則 → 流程直屬行($)。
func (this *journal) appendBody(body string) {
	if this.effect {
		this.line = append(this.line, "  "+body)
		return
	} // if

	this.line = append(this.line, "$ "+body)
}

// propertyBody 屬性命令行本文(<屬性> <運算> <值> >> <結果>): 全域屬性用全名、實例屬性作用於當前效果
// self 時屬性開頭、其他對象識別碼開頭(fan-out 每對象一行); 免疫詞條的 Operand 為群組維度鍵,
// 行文轉查詢函式形 名稱(群組) 且運算值固定 1(M22 拍板); 鎖定 / 解鎖無算術式, 結果為鎖定計數。
func (this *journal) propertyBody(eventData cores.EventData) string {
	global := eventData.DataID == 0 && eventData.InstanceID == cores.NoneID
	name := attrText(eventData.Attr, global)
	value := numText(eventData.Operand)

	if eventData.Attr == attrEffectImmune || eventData.Attr == attrSkillImmune {
		name += "(" + value + ")"
		value = "1"
	} // if

	body := name

	if global == false && (this.effect == false || eventData.DataID != this.selfDataID || eventData.InstanceID != this.selfInstanceID) {
		body = identTarget(this.sheet, eventData.DataID, eventData.InstanceID) + " " + name
	} // if

	if eventData.Op == cores.AssignLock || eventData.Op == cores.AssignUnlock {
		return body + " " + assignText(eventData.Op) + " >> " + numText(eventData.After)
	} // if

	return body + " " + assignText(eventData.Op) + " " + value + " >> " + numText(eventData.After)
}

// containerBody 容器命令行本文: 重整(From == To)為洗牌行(全序不印, 順序直讀盤面即見; M22 拍板)、
// 搬移為 <識別碼> >> <去向>(入座帶座位編號、To == None 即離場)。
func (this *journal) containerBody(eventData cores.EventData) string {
	if eventData.From == eventData.To {
		return containerText(eventData.From) + " 洗牌"
	} // if

	dest := containerText(eventData.To)

	if eventData.To == cores.ContainerSeat && eventData.SeatID > 0 {
		dest += numText(float64(eventData.SeatID))
	} // if

	return identTarget(this.sheet, eventData.DataID, eventData.InstanceID) + " >> " + dest
}

// appendInstance 變身配對(instance 唯一真身 = morph 銷毀舊 + 建立新, 必相鄰; M18 拍板):
// 銷毀暫存舊識別碼、建立時合併為一行 <舊> >> <新>(M22 拍板); 防禦: 未見銷毀的建立以 ? 起頭。
func (this *journal) appendInstance(eventData cores.EventData) {
	ident := identTarget(this.sheet, eventData.DataID, eventData.InstanceID)

	if eventData.Alive == false {
		this.morph = ident
		return
	} // if

	old := this.morph

	if old == "" {
		old = "?"
	} // if

	this.morph = ""
	this.appendBody(old + " >> " + ident)
}

// appendSelect 玩家選取紀錄($ 選取 <來源> -> <選中…>; 恆為流程直屬行): 效果目標選取(EffectID 非零)
// 來源顯效果識別碼、其餘查詞彙對照; 選中清單空白時防禦顯 空。
func (this *journal) appendSelect(eventData cores.EventData) {
	source := sourceText(eventData.Source)

	if eventData.EffectID != 0 {
		source = identEffect(this.sheet, eventData.EffectID, cores.NoneID)
	} // if

	pick := []string{}

	for _, itor := range eventData.Pick {
		pick = append(pick, identTarget(this.sheet, itor.DataID, itor.InstanceID))
	} // for

	chosen := "空"

	if len(pick) > 0 {
		chosen = strings.Join(pick, " ")
	} // if

	this.line = append(this.line, "$ 選取 "+source+" -> "+chosen)
}

// numText 數值轉顯示字串(整數去小數位、運算值可為小數)。
func numText(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
