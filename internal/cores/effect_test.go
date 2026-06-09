package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteEffect(t *testing.T) {
	suite.Run(t, new(SuiteEffect))
}

// SuiteEffect 驗證效果佇列操作原語（effect.go）:入列 / 移除 / 作用順序排序。
type SuiteEffect struct {
	suite.Suite
}

func (this *SuiteEffect) TestEffectPush() {
	runtime := NewRuntime(0)
	eng := this.engine(runtime)

	effectPush(eng, &Effect{InstanceID: 1, EffectID: 401})
	effectPush(eng, &Effect{InstanceID: 2, EffectID: 402})
	this.Require().Len(runtime.Effect, 2)
	this.Equal(InstanceID(1), runtime.Effect[0].InstanceID)
	this.Equal(InstanceID(2), runtime.Effect[1].InstanceID)
}

func (this *SuiteEffect) TestEffectRemove() {
	runtime := NewRuntime(0)
	keep := &Effect{InstanceID: 1, EffectID: 401}
	drop := &Effect{InstanceID: 2, EffectID: 402}
	runtime.Effect = []*Effect{keep, drop}
	eng := this.engine(runtime)

	effectRemove(eng, drop) // 命中 → 移除
	this.Require().Len(runtime.Effect, 1)
	this.Equal(InstanceID(1), runtime.Effect[0].InstanceID)

	effectRemove(eng, &Effect{InstanceID: 9}) // 未命中 → 不變
	this.Len(runtime.Effect, 1)
}

func (this *SuiteEffect) TestEffectSort() {
	runtime := NewRuntime(0)
	eng := this.engine(runtime)

	effect := []*Effect{
		{InstanceID: 1, EffectID: 402}, // RunOrder 10
		{InstanceID: 2, EffectID: 403}, // RunOrder 20（最大 → 最先）
		{InstanceID: 3, EffectID: 999}, // 查無資料 → RunOrder 0（最後）
		{InstanceID: 4, EffectID: 401}, // RunOrder 10、同序 EffectID 401 < 402 → 先於 402
	}
	effectSort(eng, effect)

	order := []int32{effect[0].EffectID, effect[1].EffectID, effect[2].EffectID, effect[3].EffectID}
	this.Equal([]int32{403, 401, 402, 999}, order)
}

// === 測試輔助（置尾） ===

func (this *SuiteEffect) engine(runtime *Runtime) *Engine {
	return NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{}, nil)
}
