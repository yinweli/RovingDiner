package rodi

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// keyBind 鍵綁定: 分派與顯示共用同一張表(單一來源, 鍵位列永不與實際行為漂移; M20 拍板)。
// label 空字串 = 有作用但不列示(如 ctrl+c 終端機慣例逃生鍵); row 為【營業顯示規格書 | 6、畫面規格 | 6.11】
// 終態行位(1 導覽 / 2 動作), 鍵一進場就站終態位置、後續不再挪。
type keyBind struct {
	key   string  // 按鍵(對應 tea.KeyMsg.String())
	label string  // 鍵位列顯示標籤(空 = 不列示)
	row   int     // 行位(1 / 2)
	cmd   tea.Cmd // 按鍵行為
}

// barKey 鍵位列組件(【營業顯示規格書 | 6、畫面規格 | 6.11】): 常駐底部、固定 2 行、橫跨全寬、非聚焦;
// 誠實列鍵——只列當前真有行為的鍵, 內容隨里程碑成長、M27 收斂成 §6.11 常態全鍵(M20 拍板)。
// 綁定表為自持 UI 狀態(不讀盤面, View 只吃寬度預算), 父層 Update 經 Find 分派按鍵。
type barKey struct {
	bind []keyBind // 鍵綁定表
}

func newBarKey() barKey {
	return barKey{bind: []keyBind{
		{key: " ", label: "[Space]快/慢/步進", row: 2, cmd: cycle()},
		{key: "n", label: "[N]前進", row: 2, cmd: step()},
		{key: "q", label: "[Q]離開", row: 2, cmd: tea.Quit},
		{key: "ctrl+c", label: "", row: 2, cmd: tea.Quit},
	}}
}

// View 渲染固定 2 行(鍵少的行留白, 高度穩定不抖); 同行鍵以空白分隔, 超寬依預算截斷。
func (this barKey) View(width int) string {
	row1 := []string{}
	row2 := []string{}

	for _, itor := range this.bind {
		if itor.label == "" {
			continue
		} // if

		if itor.row == 1 {
			row1 = append(row1, itor.label)
		} else {
			row2 = append(row2, itor.label)
		} // if
	} // for

	return truncTo(strings.Join(row1, " "), width) + "\n" + truncTo(strings.Join(row2, " "), width)
}

// Find 依按鍵字串查綁定行為; 未綁定回 nil(父層據此不動作)。
func (this barKey) Find(key string) tea.Cmd {
	for _, itor := range this.bind {
		if itor.key == key {
			return itor.cmd
		} // if
	} // for

	return nil
}
