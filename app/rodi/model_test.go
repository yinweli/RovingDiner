package rodi

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteModel(t *testing.T) {
	suite.Run(t, new(SuiteModel))
}

// SuiteModel 驗證 Bubble Tea 殼(model.go): Update 純函式分派、組件 View、waitEvent 等待點。
// Run 為組裝入口(需要 TTY), 比照 cmd 瘦組裝不納白箱測試。
type SuiteModel struct {
	suite.Suite
}

// TestNewModel 驗證建構: 欄位就位、鏡像 / 組件 / 寬度預算就緒、計數歸零。
func (this *SuiteModel) TestNewModel() {
	target := newModel(nil, 7, 601, "sheetdata")
	this.Nil(target.adapter)
	this.Equal(int64(7), target.seed)
	this.Equal(int32(601), target.stageID)
	this.Equal("sheetdata", target.dataDir)
	this.NotNil(target.world)
	this.Len(target.keybar.bind, 2)
	this.Len(target.comp, 2) // 狀態列 + 鍵位列
	this.Equal(100, target.width)
	this.Equal(0, target.serial)
}

// TestModelInit 驗證起跑 Cmd 存在(header + 首次等待)。
func (this *SuiteModel) TestModelInit() {
	this.NotNil(newModel(nil, 1, 601, "sheetdata").Init())
}

// TestModelUpdate 驗證訊息分派: 事件遞增序號、摺疊鏡像並續等; 終局印終局行、停止消費; 視窗尺寸更新寬度預算;
// 按鍵查綁定表分派(q / ctrl+c 離開、未綁定鍵不動作)。
func (this *SuiteModel) TestModelUpdate() {
	result, cmd := newModel(nil, 1, 601, "sheetdata").Update(eventMsg(cores.EventData{Kind: cores.EventProperty, Round: 2, Attr: "morale", After: 30}))
	this.Equal(1, result.(model).serial)
	this.Equal(int32(2), result.(model).world.round) // 事件已摺疊進鏡像
	this.Equal(float64(30), result.(model).world.attr["morale"])
	this.NotNil(cmd)

	result, cmd = result.(model).Update(eventMsg(cores.EventData{Kind: cores.EventScope}))
	this.Equal(2, result.(model).serial)
	this.NotNil(cmd)

	_, cmd = newModel(nil, 1, 601, "sheetdata").Update(doneMsg(true))
	this.NotNil(cmd) // 終局行印進 scrollback(成敗常駐顯示歸狀態列階段欄)

	_, cmd = newModel(nil, 1, 601, "sheetdata").Update(doneMsg(false))
	this.NotNil(cmd)

	result, cmd = newModel(nil, 1, 601, "sheetdata").Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	this.Equal(120, result.(model).width)
	this.Nil(cmd)

	_, cmd = newModel(nil, 1, 601, "sheetdata").Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	this.NotNil(cmd)
	this.Equal(tea.QuitMsg{}, cmd())

	_, cmd = newModel(nil, 1, 601, "sheetdata").Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	this.NotNil(cmd)
	this.Equal(tea.QuitMsg{}, cmd())

	result, cmd = newModel(nil, 1, 601, "sheetdata").Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	this.Equal(0, result.(model).serial)
	this.Nil(cmd)

	result, cmd = newModel(nil, 1, 601, "sheetdata").Update(struct{}{}) // 未知訊息型別不動作
	this.Equal(0, result.(model).serial)
	this.Nil(cmd)
}

// TestModelView 驗證 View: 組件 footer(狀態列 + 鍵位列)、寬度預算生效; 事件本文不歸 View 管。
func (this *SuiteModel) TestModelView() {
	target := newModel(nil, 7, 601, "sheetdata")
	this.Contains(target.View(), "回合")
	this.Contains(target.View(), "快速") // 模式欄 M25 前固定快速
	this.Contains(target.View(), "[Q]離開")

	target.width = 4 // 寬度預算生效: 狀態列截到首欄
	this.Contains(target.View(), "回合\n")
	this.NotContains(target.View(), "士氣")
}

// TestWaitEvent 驗證等待點: 事件到回事件訊息、終局到回成敗訊息。
func (this *SuiteModel) TestWaitEvent() {
	target := &adapter{event: make(chan cores.EventData, 1), done: make(chan bool, 1)}
	target.event <- cores.EventData{Kind: cores.EventPhase, Round: 2}
	this.Equal(eventMsg(cores.EventData{Kind: cores.EventPhase, Round: 2}), waitEvent(target)())

	target.done <- true
	this.Equal(doneMsg(true), waitEvent(target)())
}
