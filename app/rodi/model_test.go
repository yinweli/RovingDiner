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

// SuiteModel 驗證 Bubble Tea 殼(model.go): Update 純函式分派、View footer、waitEvent 等待點。
// Run 為組裝入口(需要 TTY), 比照 cmd 瘦組裝不納白箱測試。
type SuiteModel struct {
	suite.Suite
}

// TestNewModel 驗證建構: 欄位就位、計數歸零。
func (this *SuiteModel) TestNewModel() {
	target := newModel(nil, 7, 601, "sheetdata")
	this.Nil(target.adapter)
	this.Equal(int64(7), target.seed)
	this.Equal(int32(601), target.stageID)
	this.Equal("sheetdata", target.dataDir)
	this.Equal(0, target.serial)
	this.False(target.finish)
}

// TestModelInit 驗證起跑 Cmd 存在(header + 首次等待)。
func (this *SuiteModel) TestModelInit() {
	this.NotNil(newModel(nil, 1, 601, "sheetdata").Init())
}

// TestModelUpdate 驗證訊息分派: 事件遞增序號並續等; 終局記成敗、停止消費; q / ctrl+c 離開、其他鍵不動作。
func (this *SuiteModel) TestModelUpdate() {
	result, cmd := newModel(nil, 1, 601, "sheetdata").Update(eventMsg(cores.EventData{Kind: cores.EventPhase}))
	this.Equal(1, result.(model).serial)
	this.NotNil(cmd)

	result, cmd = result.(model).Update(eventMsg(cores.EventData{Kind: cores.EventScope}))
	this.Equal(2, result.(model).serial)
	this.NotNil(cmd)

	result, cmd = result.(model).Update(doneMsg(true))
	this.True(result.(model).finish)
	this.True(result.(model).succ)
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
}

// TestModelView 驗證 footer: 跑動中顯營業中、終局顯成敗, 事件本文不歸 View 管。
func (this *SuiteModel) TestModelView() {
	target := newModel(nil, 7, 601, "sheetdata")
	this.Contains(target.View(), "營業中")
	this.Contains(target.View(), "seed 7")

	target.finish = true
	target.succ = true
	this.Contains(target.View(), "成功")

	target.succ = false
	this.Contains(target.View(), "失敗")
}

// TestWaitEvent 驗證等待點: 事件到回事件訊息、終局到回成敗訊息。
func (this *SuiteModel) TestWaitEvent() {
	target := &adapter{event: make(chan cores.EventData, 1), done: make(chan bool, 1)}
	target.event <- cores.EventData{Kind: cores.EventPhase, Round: 2}
	this.Equal(eventMsg(cores.EventData{Kind: cores.EventPhase, Round: 2}), waitEvent(target)())

	target.done <- true
	this.Equal(doneMsg(true), waitEvent(target)())
}
