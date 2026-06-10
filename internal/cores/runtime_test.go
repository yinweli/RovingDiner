package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
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

	// 容器初始化：座位 map 必須可用、累積計數零值可用（Sum 回 0）
	this.NotNil(runtime.Seat)
	this.Equal(int32(0), runtime.Game.GetDrawTotal().Sum())
	this.Equal(int32(0), runtime.Game.GetDropTotal().Sum())
	this.Equal(int32(0), runtime.Game.GetPlayTotal().Sum())
	this.Equal(int32(0), runtime.Game.GetExileTotal().Sum())

	// 初始階段為空、容器為空
	this.Equal(PhaseNone, runtime.Game.GetNextPhase())
	this.Empty(runtime.Hand)
	this.Empty(runtime.Deck)
	this.Empty(runtime.Effect)
	this.Empty(runtime.Action)
}

func (this *SuiteRuntime) TestRuntimeNextID() {
	runtime := NewRuntime(0)
	this.Equal(InstanceID(1), runtime.NextID())
	this.Equal(InstanceID(2), runtime.NextID())
	this.Equal(InstanceID(3), runtime.NextID())
}
