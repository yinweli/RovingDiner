package rodi

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// seatColumn 桌欄固定寬(識別碼 14 + 空格 1 + 飽999耐999 10; 【營業顯示規格書 | 6、畫面規格 | 6.3】),
// 位數 / 名長變動時欄位不左右跳動。
const seatColumn = 25

// flagNone 旗標空槽(2 格; 未命中該槽留白)。
const flagNone = "  "

// panelSeat 座位組件(區 1; 【營業顯示規格書 | 6、畫面規格 | 6.3】): 桌子排成水平 strip(左右相鄰即鄰桌拓樸),
// 每桌 2 座位上下疊、每位顧客 2 行摘要(識別碼 + 飽耐 / 旗標列)、空位顯「空」; 桌欄固定寬 25、欄距 2。
// 游標 = 桌 x 座(左右換桌、上下切座; M26 R2), 游標態反白整個顧客格 2 行(【營業顯示規格書 | 7、互動規格 | 7.4】
// 粒度); 桌窗格跟游標捲、左緣 < 全行縮排對齊、桌號列補右緣 >。
// 選取模式(M27 R4; 顧客候選限座位區): 非候選(含空位)變暗、已選選取色底、游標只在候選間吸附步進、
// 標題加註 (請選顧客)。
type panelSeat struct {
	curTable int        // 游標桌索引(自持 UI 狀態; 讀取時夾界)
	curSeat  int        // 游標桌內座索引(0 上 / 1 下; 讀取時夾界)
	pick     *pickState // 選取模式共享狀態(newModel 注入, 唯讀)
}

// View 渲染標題列 + 5 行(桌號列 1 + 桌內 2 座各 2 行)。
func (this *panelSeat) View(game *cores.Game, width int, focus bool) string {
	table := seatTable(game.GetSheet())
	picking := this.pickSnap(game, table)
	cursor := clampIndex(this.curTable, len(table))
	cell := [][]string{}
	size := []int{}

	for index, itor := range table {
		col := []string{fmt.Sprintf("桌%v", itor.id), "", "", "", ""}

		if len(itor.seat) > 0 {
			col[1], col[2] = seatGuest(game, itor.seat[0])
		} // if

		if len(itor.seat) > 1 {
			col[3], col[4] = seatGuest(game, itor.seat[1])
		} // if

		for r := range col {
			col[r] = padTo(col[r], seatColumn)
		} // for

		if picking { // 三視覺態: 非候選(含空位)變暗、已選選取色底(先著色後游標, 樣式不動寬度)
			for s, seatID := range itor.seat {
				at := this.pick.guestIndex(game.Seat[seatID])
				base := 1 + s*2

				if at < 0 {
					col[base] = restyle(&styleDim, col[base])
					col[base+1] = restyle(&styleDim, col[base+1])
					continue
				} // if

				if this.pick.chosen(at) {
					col[base] = restyle(&styleChosen, col[base])
					col[base+1] = restyle(&styleChosen, col[base+1])
				} // if
			} // for
		} // if

		if focus && index == cursor {
			base := 1 + clampIndex(this.curSeat, len(itor.seat))*2
			col[base] = restyle(&styleCursor, col[base])
			col[base+1] = restyle(&styleCursor, col[base+1])
		} // if

		cell = append(cell, col)
		size = append(size, seatColumn)
	} // for

	first := stripFirst(size, 2, width-4, cursor)
	row := []string{"", "", "", "", ""}

	for r := range row {
		part := []string{}

		for _, itor := range cell[first:] {
			part = append(part, itor[r])
		} // for

		head := ""

		if first > 0 {
			head = markIndent // 左緣記號佔位: 各行同縮排, 桌欄上下對齊
		} // if

		if first > 0 && r == 0 {
			head = markHead
		} // if

		row[r] = head + strings.Join(part, "  ")
	} // for

	title := "座位"

	if picking {
		title += styleNote.Render(noteGuest) // 標題提醒: 候選在本區, Tab 切走仍見等待落點(聚焦反白時由 focusView 剝色讓位)
	} // if

	text := []string{panelTitle(title, width), boxMark(row[0], width)}

	for _, itor := range row[1:] {
		text = append(text, boxTrunc(itor, width))
	} // for

	return strings.Join(text, "\n")
}

// Move 游標移動: 左右換桌、上下切桌內 2 座(【營業顯示規格書 | 7、互動規格 | 7.2】); 夾界不迴繞。
// 選取模式: 方向語意保留, 沿方向掃描至下一個候選、無則不動(【7.4】自動略過非候選; M27 拍板——
// 左右不跨座位列、上下不跨桌)。
func (this *panelSeat) Move(game *cores.Game, key string) {
	table := seatTable(game.GetSheet())

	if this.pickSnap(game, table) {
		this.pickMove(game, table, key)
		return
	} // if

	switch key {
	case keyLeft:
		this.curTable--

	case keyRight:
		this.curTable++

	case keyUp:
		this.curSeat--

	case keyDown:
		this.curSeat++
	} // switch

	this.curTable = clampIndex(this.curTable, len(table))
	this.curSeat = clampIndex(this.curSeat, 2)

	if len(table) > 0 {
		this.curSeat = clampIndex(this.curSeat, len(table[this.curTable].seat))
	} // if
}

// Item 回游標座位上的顧客(空位 / 無桌回 nil; 檢視 modal 用)。
func (this *panelSeat) Item(game *cores.Game) any {
	table := seatTable(game.GetSheet())

	if len(table) == 0 {
		return nil
	} // if

	itor := table[clampIndex(this.curTable, len(table))]
	guest := game.Seat[itor.seat[clampIndex(this.curSeat, len(itor.seat))]]

	if guest == nil {
		return nil // 空位(型別斷言要的是無項目, 不回帶 nil 的具型指標)
	} // if

	return guest
}

// pickSnap 選取模式吸附判定: 本區含候選時若游標不在候選上, 吸到第一個候選(游標只在候選間;
// 【營業顯示規格書 | 7、互動規格 | 7.4】); 回報本區是否處於選取著色狀態(無候選即非本區選取, 照常渲染)。
func (this *panelSeat) pickSnap(game *cores.Game, table []tableInfo) bool {
	if this.pick.active() == false {
		return false
	} // if

	spot := this.pickSpot(game, table)

	if len(spot) == 0 {
		return false
	} // if

	if this.pickAt(game, table) < 0 {
		this.curTable, this.curSeat = spot[0][0], spot[0][1]
	} // if

	return true
}

// pickSpot 候選位置序列(桌索引 x 桌內座索引, 桌序為主序; 候選身分 = 實例指標)。
func (this *panelSeat) pickSpot(game *cores.Game, table []tableInfo) (result [][2]int) {
	for index, itor := range table {
		for s, seatID := range itor.seat {
			if this.pick.guestIndex(game.Seat[seatID]) >= 0 {
				result = append(result, [2]int{index, s})
			} // if
		} // for
	} // for

	return result
}

// pickAt 游標在候選位置序列的索引(不在候選上回 -1)。
func (this *panelSeat) pickAt(game *cores.Game, table []tableInfo) int {
	for index, itor := range this.pickSpot(game, table) {
		if itor[0] == clampIndex(this.curTable, len(table)) && itor[1] == this.curSeat {
			return index
		} // if
	} // for

	return -1
}

// pickMove 候選間方向移動: 左右沿同座位列掃描換桌、上下沿桌內座序掃描切座, 掃到候選即停、
// 掃不到不動(方向語意與非選取模式一致, 僅略過非候選)。
func (this *panelSeat) pickMove(game *cores.Game, table []tableInfo, key string) {
	at := clampIndex(this.curTable, len(table))

	switch key {
	case keyLeft:
		for i := at - 1; i >= 0; i-- {
			if this.pickHit(game, table, i, this.curSeat) {
				this.curTable = i
				return
			} // if
		} // for

	case keyRight:
		for i := at + 1; i < len(table); i++ {
			if this.pickHit(game, table, i, this.curSeat) {
				this.curTable = i
				return
			} // if
		} // for

	case keyUp:
		for i := this.curSeat - 1; i >= 0; i-- {
			if this.pickHit(game, table, at, i) {
				this.curSeat = i
				return
			} // if
		} // for

	case keyDown:
		for i := this.curSeat + 1; i < len(table[at].seat); i++ {
			if this.pickHit(game, table, at, i) {
				this.curSeat = i
				return
			} // if
		} // for
	} // switch
}

// pickHit 回報桌 t 座 s 存在且坐著候選顧客。
func (this *panelSeat) pickHit(game *cores.Game, table []tableInfo, t, s int) bool {
	return s < len(table[t].seat) && this.pick.guestIndex(game.Seat[table[t].seat[s]]) >= 0
}

// tableInfo 單一桌次(桌次編號 + 桌內座位編號, 升序)。
type tableInfo struct {
	id   int32   // 桌次編號
	seat []int32 // 桌內座位編號(升序; 上排在前)
}

// seatTable 自座位表組桌次清單(桌次升序、桌內座位升序); 拓樸 = 桌次序即水平相鄰。
func seatTable(sheet *sheeter.Sheeter) (result []tableInfo) {
	group := map[int32][]int32{}

	for k, v := range sheet.Seat.Data {
		group[v.TableID] = append(group[v.TableID], k)
	} // for

	for k, v := range group {
		sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
		result = append(result, tableInfo{id: k, seat: v})
	} // for

	sort.Slice(result, func(i, j int) bool { return result[i].id < result[j].id })
	return result
}

// seatGuest 單一座位的 2 行(顧客摘要 + 旗標列); 空位顯「空」、旗標列全空留白。
func seatGuest(game *cores.Game, seatID int32) (row1, row2 string) {
	guest := game.Seat[seatID]

	if guest == nil {
		return "空", ""
	} // if

	ident := cores.IdentGuest(game.GetSheet(), guest.GetGuestID(), cores.NoneID)
	row1 = ident + " 飽" + num(guest.GetSate().GetValue()) + "耐" + num(guest.GetCalm().GetValue())
	flag := seatFlag(game, guest)

	if flag != "" {
		row2 = strings.Repeat(" ", lipgloss.Width(ident)+1) + flag // 對齊到 飽 起始欄
	} // if

	return row1, row2
}

// seatFlag 旗標列(封 免 效 三槽、各 2 格, 命中才顯、未命中該槽留白、全空回空字串):
// 封 = 任一封印技能鎖定中; 免 = 任一免疫群組計數 > 0(M22 拍板); 效 = 顧客身上有 active 效果。
// 狀態語意色(M27 配色): 封 紅(負面)/ 免 青(保護)/ 效 綠(同日誌效果綠); 先排序定槽、樣式最後上,
// 顧客格被變暗 / 選取色底 / 游標反白時經 restyle 讓位。
func seatFlag(game *cores.Game, guest *cores.Guest) string {
	seal := flagNone

	if guest.GetSateSeal().IsLock() || guest.GetCalmSeal().IsLock() {
		seal = styleBad.Render("封")
	} // if

	immune := flagNone

	if guest.GetEffectImmune().Any() || guest.GetSkillImmune().Any() {
		immune = styleWard.Render("免")
	} // if

	effect := flagNone

	if hasEffect(game, guest) {
		effect = styleGood.Render("效")
	} // if

	return strings.TrimRight(seal+immune+effect, " ")
}

// hasEffect 回報顧客身上是否有 active 效果(效 旗標用; 以效果佇列項 self 比對實例指標)。
func hasEffect(game *cores.Game, guest *cores.Guest) bool {
	for _, itor := range game.Effect {
		if itor.GetSelf().GetGuest() == guest {
			return true
		} // if
	} // for

	return false
}
