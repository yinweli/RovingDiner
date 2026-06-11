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
// 超寬固定窗截斷、桌號列補右緣 >。
type panelSeat struct{}

// View 渲染標題列 + 5 行(桌號列 1 + 桌內 2 座各 2 行)。
func (this panelSeat) View(game *cores.Game, width int) string {
	row := []string{"", "", "", "", ""}

	for index, itor := range seatTable(game.GetSheet()) {
		col := []string{fmt.Sprintf("桌%v", itor.id), "", "", "", ""}

		if len(itor.seat) > 0 {
			col[1], col[2] = seatGuest(game, itor.seat[0])
		} // if

		if len(itor.seat) > 1 {
			col[3], col[4] = seatGuest(game, itor.seat[1])
		} // if

		for r := range row {
			if index > 0 {
				row[r] += "  "
			} // if

			row[r] += padTo(col[r], seatColumn)
		} // for
	} // for

	text := []string{panelTitle("座位", width), boxMark(strings.TrimRight(row[0], " "), width)}

	for _, itor := range row[1:] {
		text = append(text, boxTrunc(strings.TrimRight(itor, " "), width))
	} // for

	return strings.Join(text, "\n")
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
func seatFlag(game *cores.Game, guest *cores.Guest) string {
	seal := flagNone

	if guest.GetSateSeal().IsLock() || guest.GetCalmSeal().IsLock() {
		seal = "封"
	} // if

	immune := flagNone

	if guest.GetEffectImmune().Any() || guest.GetSkillImmune().Any() {
		immune = "免"
	} // if

	effect := flagNone

	if hasEffect(game, guest) {
		effect = "效"
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
