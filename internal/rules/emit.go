package rules

import (
	"strings"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// 日誌行發射輔助(流程與效果系統的本文組裝收口; 組行原語見 cores 發射台 emit.go):
// 標題 / 效果頭 / 屬性行直呼 cores.EmitTitle / EmitEffect / EmitProperty, 本檔只組搬移 / 行動 / 選取的本文。

// emitGuestMove 發射顧客容器搬移行: <識別碼> >> <去向>(入座帶座位編號、to == ContainerNone 即離場)。
// 卡牌搬移由 placeCard 收口, 顧客無收口、各搬移點逐點發。
func emitGuestMove(game *cores.Game, guest *cores.Guest, to cores.ContainerKind, seatID int32) {
	dest := cores.ContainerText(to)

	if to == cores.ContainerSeat && seatID > 0 {
		dest += cores.NumText(float64(seatID))
	} // if

	cores.EmitBody(game, cores.IdentGuest(game.GetSheet(), guest.GetGuestID(), guest.GetInstanceID())+" >> "+dest)
}

// emitCardMove 發射卡牌容器搬移行: <識別碼> >> <去向>(placeCard 收口與開局 / 洗回共用)。
func emitCardMove(game *cores.Game, card *cores.Card, to cores.ContainerKind) {
	cores.EmitBody(game, cores.IdentCard(game.GetSheet(), card.GetCardID(), card.GetInstanceID())+" >> "+cores.ContainerText(to))
}

// emitDeckOrder 發射抽牌牌堆重整行(洗牌 / 洗回後收口): 全序不印, 順序直讀盤面即見(M22 拍板)。
func emitDeckOrder(game *cores.Game) {
	cores.EmitBody(game, cores.ContainerText(cores.ContainerDeck)+" 洗牌")
}

// emitAction 發射行動佇列入列行: <顧客識別碼> >> 行動佇列(<行動類型>)。
// 出列不發(緊接的顧客行動標題已承載; M22 拍板)。
func emitAction(game *cores.Game, action *cores.Action) {
	guest := action.GetGuest()
	cores.EmitBody(game, cores.IdentGuest(game.GetSheet(), guest.GetGuestID(), guest.GetInstanceID())+" >> 行動佇列("+cores.TaskText(action.GetKind())+")")
}

// emitMorph 發射變身行: <舊識別碼> >> <新識別碼>(位置不變的身分變更, 一行收口; M18 拍板的雙發協定作廢)。
func emitMorph(game *cores.Game, oldID int32, oldInstance cores.InstanceID, card *cores.Card) {
	cores.EmitBody(game, cores.IdentCard(game.GetSheet(), oldID, oldInstance)+" >> "+cores.IdentCard(game.GetSheet(), card.GetCardID(), card.GetInstanceID()))
}

// emitPlayTitle 發射玩家出牌範圍標題(操作元 = 卡牌 + 技能)。
func emitPlayTitle(game *cores.Game, card *cores.Card) {
	cores.EmitTitle(game, "玩家出牌", actionOperator(game, cores.IdentCard(game.GetSheet(), card.GetCardID(), card.GetInstanceID()), cardSkill(game, card.GetCardID()))...)
}

// emitGuestTitle 發射顧客行動範圍標題(操作元 = 顧客 + 技能)。
func emitGuestTitle(game *cores.Game, action *cores.Action) {
	guest := action.GetGuest()
	cores.EmitTitle(game, "顧客行動", actionOperator(game, cores.IdentGuest(game.GetSheet(), guest.GetGuestID(), guest.GetInstanceID()), action.GetSkillID())...)
}

// actionOperator 組動作類範圍標題的操作元清單: 主體識別碼 + 技能識別碼(技能編號零值防禦略過)。
func actionOperator(game *cores.Game, subject string, skillID int32) []string {
	result := []string{subject}

	if skillID != 0 {
		result = append(result, cores.IdentSkill(game.GetSheet(), skillID))
	} // if

	return result
}

// emitSelect 發射玩家選取紀錄行($ 選取 <來源> -> <選中…>; 恆為流程直屬行, 搭 seed 重現用, 只記 Operator 真選取):
// 效果目標選取(effectID 非零)來源顯效果識別碼、其餘查詞彙對照; 選中清單空白時防禦顯 空。
func emitSelect(game *cores.Game, source string, effectID int32, ident []string) {
	text := cores.SourceText(source)

	if effectID != 0 {
		text = cores.IdentEffect(game.GetSheet(), effectID, cores.NoneID)
	} // if

	chosen := "空"

	if len(ident) > 0 {
		chosen = strings.Join(ident, " ")
	} // if

	game.Emit("$ 選取 " + text + " -> " + chosen)
}

// cardIdentList 把卡牌清單組成識別碼清單(選取紀錄行的選中段)。
func cardIdentList(game *cores.Game, card []*cores.Card) (result []string) {
	for _, itor := range card {
		result = append(result, cores.IdentCard(game.GetSheet(), itor.GetCardID(), itor.GetInstanceID()))
	} // for

	return result
}

// guestIdentList 把顧客清單組成識別碼清單(選取紀錄行的選中段)。
func guestIdentList(game *cores.Game, guest []*cores.Guest) (result []string) {
	for _, itor := range guest {
		result = append(result, cores.IdentGuest(game.GetSheet(), itor.GetGuestID(), itor.GetInstanceID()))
	} // for

	return result
}

// promptText 組選取提示前文(`<來源名> 要求選<對象>`; 【營業顯示規格書 | 6、畫面規格 | 6.11】):
// 引擎直出、顯示端只附加已選進度(M27 拍板); 來源名 = 效果名稱(目標選取)或命令對象 / 流程中文名,
// 名稱不帶識別碼(與 emitSelect 的識別碼形式受眾不同、不衝突)。
func promptText(source, object string) string {
	return source + " 要求選" + object
}

// effectName 效果名稱(選取提示的來源名; 查無資料顯 ?)。
func effectName(game *cores.Game, effectID int32) string {
	if meta := game.GetSheet().Effect.Get(effectID); meta != nil {
		return meta.Name
	} // if

	return "?"
}

// pickObject Pick 詞條鍵轉提示對象詞: 手牌類選手牌、牌堆類選牌堆卡。
func pickObject(source string) string {
	if source == "handPick" {
		return "手牌"
	} // if

	return "牌堆卡"
}
