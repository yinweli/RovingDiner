package rodi

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteModel(t *testing.T) {
	suite.Run(t, new(SuiteModel))
}

// SuiteModel 驗證 Bubble Tea 殼(model.go): Update 純函式分派、全畫面 layout、最低尺寸守門、waitEvent 等待點。
// Run 為組裝入口(需要 TTY), 比照 cmd 瘦組裝不納白箱測試。
type SuiteModel struct {
	suite.Suite
}

// TestNewModel 驗證建構: 欄位就位、鏡像 / 日誌 / 組件 / 寬高預算就緒。
func (this *SuiteModel) TestNewModel() {
	target := newModel(nil, testSheet())
	this.Nil(target.adapter)
	this.NotNil(target.world)
	this.NotNil(target.log)
	this.Len(target.keybar.bind, 2)
	this.Len(target.comp, 6) // 六區(座位 / 場外 / 行動 / 效果 / 手牌 / 牌堆); 狀態列 / 日誌 / 鍵位列為 layout 角色專屬掛點
	this.Equal(minWidth, target.width)
	this.Equal(minHeight, target.height)
}

// TestModelInit 驗證起跑 Cmd 存在(首次等待)。
func (this *SuiteModel) TestModelInit() {
	this.NotNil(newModel(nil, testSheet()).Init())
}

// TestModelUpdate 驗證訊息分派: 事件摺疊鏡像 + 轉寫日誌並續等; 終局停止消費; 視窗尺寸更新寬高預算;
// 按鍵查綁定表分派(q / ctrl+c 離開、未綁定鍵不動作)。
func (this *SuiteModel) TestModelUpdate() {
	result, cmd := newModel(nil, testSheet()).Update(eventMsg(cores.EventData{Kind: cores.EventProperty, Round: 2, Attr: "morale", After: 30}))
	this.Equal(int32(2), result.(model).world.round) // 事件已摺疊進鏡像
	this.Equal(float64(30), result.(model).world.attr["morale"])
	this.Equal([]string{"$ 餐廳士氣值 = 0 >> 30"}, result.(model).log.journal.line) // 事件已轉寫進日誌
	this.NotNil(cmd)

	_, cmd = newModel(nil, testSheet()).Update(doneMsg(true))
	this.Nil(cmd) // 終局: 停止消費(成敗已由終局 phase 事件投影)

	result, cmd = newModel(nil, testSheet()).Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	this.Equal(120, result.(model).width)
	this.Equal(40, result.(model).height)
	this.Nil(cmd)

	_, cmd = newModel(nil, testSheet()).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	this.NotNil(cmd)
	this.Equal(tea.QuitMsg{}, cmd())

	_, cmd = newModel(nil, testSheet()).Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	this.NotNil(cmd)
	this.Equal(tea.QuitMsg{}, cmd())

	_, cmd = newModel(nil, testSheet()).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	this.Nil(cmd) // 未綁定鍵不動作

	_, cmd = newModel(nil, testSheet()).Update(struct{}{}) // 未知訊息型別不動作
	this.Nil(cmd)
}

// TestModelView 驗證全畫面 layout: 共 height 行、左欄六區與日誌同列起頭、狀態列釘底、鍵位列收尾;
// 低於最低尺寸守門顯提示。
func (this *SuiteModel) TestModelView() {
	target := newModel(nil, testSheet())
	row := strings.Split(target.View(), "\n")
	this.Len(row, minHeight) // 預設 100x30 → 滿版 30 行
	this.Contains(row[0], "+- 座位")
	this.Contains(row[0], "+- 事件日誌")         // 右欄與左欄同列起頭
	this.Contains(row[minHeight-3], "快速")    // 狀態列釘底(鍵位列上方)
	this.Contains(row[minHeight-1], "[Q]離開") // 鍵位列固定 2 行, 現有鍵全在第 2 行(第 1 行留白)

	target.height = 40 // 更高: 留白墊在六區與狀態列之間, 狀態列仍釘底
	row = strings.Split(target.View(), "\n")
	this.Len(row, 40)
	this.Contains(row[37], "快速")

	this.Len(strings.Split(target.leftView(70, 5), "\n"), 26) // 內容超高(防禦): 留白歸零、不裁內容

	target.width = 80 // 低於最低尺寸 → 守門提示
	this.Equal("請放大終端機 (現 80x40, 最低 100x30)", target.View())

	target.width = 100
	target.height = 20
	this.Equal("請放大終端機 (現 100x20, 最低 100x30)", target.View())
}

// TestWaitEvent 驗證等待點: 事件到回事件訊息、終局到回成敗訊息。
func (this *SuiteModel) TestWaitEvent() {
	target := &adapter{event: make(chan cores.EventData, 1), done: make(chan bool, 1)}
	target.event <- cores.EventData{Kind: cores.EventPhase, Round: 2}
	this.Equal(eventMsg(cores.EventData{Kind: cores.EventPhase, Round: 2}), waitEvent(target)())

	target.done <- true
	this.Equal(doneMsg(true), waitEvent(target)())
}
