package rodi

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// keyMode 鍵盤模式(【營業顯示規格書 | 7、互動規格 | 7.5】): 鍵位分常態 / 選取 / 出牌 / modal 態四張表,
// 分派與鍵位列顯示都隨當前模式查表(M26 R1 立框架)。模式由 model 自狀態導出(modal 堆疊 / 選取等待 /
// 玩家行動等待 x 聚焦手牌), 不另存欄位——與速率模式(mode)是兩回事。
type keyMode int

const (
	keyModeNormal keyMode = iota // 常態(自由切區 / 移動 / 檢視)
	keyModePick                  // 選取模式(引擎暫停等選; M27 R4)
	keyModePlay                  // 出牌模式(玩家行動等待且聚焦手牌; M27 拍板——Space 出牌 / E 結束只在此亮)
	keyModeModal                 // modal 態(檢視 / 計數 / 說明 modal 開啟中)
)

// keyQuit 終端機慣例逃生鍵(四模式皆有作用、不列示)。
const keyQuit = "ctrl+c"

// keyBind 鍵綁定: 分派與顯示共用同一張表(單一來源, 鍵位列永不與實際行為漂移; M20 拍板)。
// label 空字串 = 有作用但不列示(如 ctrl+c 終端機慣例逃生鍵、併入 tab 標籤的 shift+tab);
// row 為【營業顯示規格書 | 6、畫面規格 | 6.11】終態行位(1 導覽 / 2 動作), 鍵一進場就站終態位置、後續不再挪。
type keyBind struct {
	key   string  // 按鍵(對應 tea.KeyMsg.String())
	label string  // 鍵位列顯示標籤(空 = 不列示)
	row   int     // 行位(1 / 2)
	cmd   tea.Cmd // 按鍵行為
}

// barKey 鍵位列組件(【營業顯示規格書 | 6、畫面規格 | 6.11】): 常駐底部、固定 2 行、橫跨全寬、非聚焦
// (不入 Tab 循環); 誠實列鍵——只列當前真有行為的鍵, M27 收斂成 §6.11 全鍵終態(M20 拍板)。
// 綁定表為自持 UI 狀態(不讀盤面, View 只吃模式 / 寬度預算 / 提示文字), 父層 Update 經 Find 分派按鍵。
type barKey struct {
	bind map[keyMode][]keyBind // 鍵綁定表(per-mode; M26 R1)
}

func newBarKey() barKey {
	return barKey{bind: map[keyMode][]keyBind{
		keyModeNormal: {
			{key: "tab", label: "[Tab/Shift+Tab]切區", row: 1, cmd: tab(1)},
			{key: "shift+tab", label: "", row: 1, cmd: tab(-1)},
			{key: keyUp, label: "[Arrow]移動游標", row: 1, cmd: move(keyUp)},
			{key: keyDown, label: "", row: 1, cmd: move(keyDown)},
			{key: keyLeft, label: "", row: 1, cmd: move(keyLeft)},
			{key: keyRight, label: "", row: 1, cmd: move(keyRight)},
			{key: "enter", label: "[Enter]檢視", row: 1, cmd: enter()},
			{key: "f1", label: "[F1]說明", row: 1, cmd: help()},
			{key: "f2", label: "[F2]計數", row: 1, cmd: count()},
			{key: " ", label: "[Space]快/慢/步進", row: 2, cmd: cycle()},
			{key: "n", label: "[N]前進", row: 2, cmd: step()},
			{key: "q", label: "[Q]離開", row: 2, cmd: tea.Quit},
			{key: keyQuit, label: "", row: 2, cmd: tea.Quit},
		},
		keyModePick: {
			{key: "tab", label: "[Tab/Shift+Tab]切區", row: 1, cmd: tab(1)},
			{key: "shift+tab", label: "", row: 1, cmd: tab(-1)},
			{key: keyUp, label: "[Arrow]移動游標", row: 1, cmd: move(keyUp)},
			{key: keyDown, label: "", row: 1, cmd: move(keyDown)},
			{key: keyLeft, label: "", row: 1, cmd: move(keyLeft)},
			{key: keyRight, label: "", row: 1, cmd: move(keyRight)},
			{key: " ", label: "[Space]加選/取消", row: 1, cmd: toggle()},
			{key: "enter", label: "[Enter]確認", row: 1, cmd: confirm()},
			{key: keyQuit, label: "", row: 2, cmd: tea.Quit},
		},
		keyModePlay: {
			{key: "tab", label: "[Tab/Shift+Tab]切區", row: 1, cmd: tab(1)},
			{key: "shift+tab", label: "", row: 1, cmd: tab(-1)},
			{key: keyUp, label: "[Arrow]移動游標", row: 1, cmd: move(keyUp)},
			{key: keyDown, label: "", row: 1, cmd: move(keyDown)},
			{key: keyLeft, label: "", row: 1, cmd: move(keyLeft)},
			{key: keyRight, label: "", row: 1, cmd: move(keyRight)},
			{key: " ", label: "[Space]出牌", row: 1, cmd: play()},
			{key: "enter", label: "[Enter]檢視", row: 1, cmd: enter()},
			{key: "e", label: "[E]結束", row: 1, cmd: end()},
			{key: keyQuit, label: "", row: 2, cmd: tea.Quit},
		},
		keyModeModal: {
			{key: keyUp, label: "[Up/Down]欄位捲動", row: 1, cmd: move(keyUp)},
			{key: keyDown, label: "", row: 1, cmd: move(keyDown)},
			{key: "esc", label: "[Esc]關閉", row: 1, cmd: pop()},
			{key: keyQuit, label: "", row: 2, cmd: tea.Quit},
		},
	}}
}

// View 渲染當前模式表固定 2 行(鍵少的行留白, 高度穩定不抖); 同行鍵以空白分隔, 超寬依預算截斷;
// hint 非空時取代行 2(選取模式的選取提示; 【營業顯示規格書 | 6、畫面規格 | 6.11】行 1 鍵位、行 2 提示)。
func (this barKey) View(keymode keyMode, width int, hint string) string {
	row1 := []string{}
	row2 := []string{}

	for _, itor := range this.bind[keymode] {
		if itor.label == "" {
			continue
		} // if

		if itor.row == 1 {
			row1 = append(row1, itor.label)
		} else {
			row2 = append(row2, itor.label)
		} // if
	} // for

	text := strings.Join(row2, " ")

	if hint != "" {
		text = hint
	} // if

	return truncTo(strings.Join(row1, " "), width) + "\n" + truncTo(text, width)
}

// Find 依當前模式表查按鍵綁定行為; 未綁定回 nil(父層據此不動作)。
func (this barKey) Find(keymode keyMode, key string) tea.Cmd {
	for _, itor := range this.bind[keymode] {
		if itor.key == key {
			return itor.cmd
		} // if
	} // for

	return nil
}
