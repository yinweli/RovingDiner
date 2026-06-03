package game

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/defines"
)

func TestSuiteRuntime(t *testing.T) {
	suite.Run(t, new(SuiteRuntime))
}

// SuiteRuntime 驗證 Runtime 聚合狀態的初始化與實例編號配發。
type SuiteRuntime struct {
	suite.Suite
}

func (this *SuiteRuntime) TestNewRuntime() {
	runtime := NewRuntime(123)
	this.Require().NotNil(runtime.Game)
	this.Equal(int64(123), runtime.Seed)

	// 容器初始化：map 必須可用、多重集合 map 已就緒
	this.NotNil(runtime.Seat)
	this.NotNil(runtime.Game.DrawTotal)
	this.NotNil(runtime.Game.DropTotal)
	this.NotNil(runtime.Game.PlayTotal)
	this.NotNil(runtime.Game.ExileTotal)

	// 初始階段為空、容器為空
	this.Equal(defines.PhaseNone, runtime.Game.NextPhase)
	this.Empty(runtime.Hand)
	this.Empty(runtime.Deck)
	this.Empty(runtime.Effect)
	this.Empty(runtime.Action)
}

func (this *SuiteRuntime) TestRuntimeNextID() {
	runtime := NewRuntime(0)
	this.Equal(defines.InstanceID(1), runtime.NextID())
	this.Equal(defines.InstanceID(2), runtime.NextID())
	this.Equal(defines.InstanceID(3), runtime.NextID())
}
