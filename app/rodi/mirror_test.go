package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteMirror(t *testing.T) {
	suite.Run(t, new(SuiteMirror))
}

// SuiteMirror 驗證世界鏡像(mirror.go): 座標蓋章與全域 property 摺疊。
type SuiteMirror struct {
	suite.Suite
}

// TestNewMirror 驗證建構: 兩表就位、座標歸零。
func (this *SuiteMirror) TestNewMirror() {
	target := newMirror()
	this.NotNil(target.attr)
	this.NotNil(target.lock)
	this.Equal(int32(0), target.round)
	this.Equal(cores.PhaseNone, target.phase)
}

// TestMirrorApply 驗證摺疊: 座標一律更新; 全域 property 依賦值符分流值表 / 鎖定計數表; 引用屬性不投影。
func (this *SuiteMirror) TestMirrorApply() {
	target := newMirror()
	target.Apply(cores.EventData{Kind: cores.EventPhase, Round: 3, Phase: cores.PhaseRoundStart})
	this.Equal(int32(3), target.round) // 座標一律更新(非 property 亦然)
	this.Equal(cores.PhaseRoundStart, target.phase)
	this.Empty(target.attr)

	target.Apply(cores.EventData{Kind: cores.EventProperty, Attr: "morale", Op: cores.AssignSet, After: 30})
	this.Equal(float64(30), target.attr["morale"]) // 全域 property → 值表

	target.Apply(cores.EventData{Kind: cores.EventProperty, Attr: "morale", Op: cores.AssignSub, After: 28})
	this.Equal(float64(28), target.attr["morale"])

	target.Apply(cores.EventData{Kind: cores.EventProperty, Attr: "energyKeep", Op: cores.AssignLock, After: 1})
	this.Equal(float64(1), target.lock["energyKeep"]) // @ / # → 鎖定計數表, 不污染值表
	this.Empty(target.attr["energyKeep"])

	target.Apply(cores.EventData{Kind: cores.EventProperty, InstanceID: 5, Attr: "sate", After: 9})
	this.Empty(target.attr["sate"]) // 引用屬性(對象欄非零)不投影
}
