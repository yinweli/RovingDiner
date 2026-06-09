package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteCommandGuest(t *testing.T) {
	suite.Run(t, new(SuiteCommandGuest))
}

// SuiteCommandGuest 驗證顧客免疫 / 行動命令(commandGuest.go):effectImmune± / skillImmune±(varargs、夾 ≥0)、taskAdd、locateGuest 四容器。
type SuiteCommandGuest struct {
	suite.Suite
}

func (this *SuiteCommandGuest) TestEffectImmune() {
	runtime := NewRuntime(0)
	guest := &Guest{InstanceID: 11, SeatID: 1}
	runtime.Seat[1] = guest
	eng := this.engine(runtime)

	commandEffectImmuneAdd(eng, []InstanceID{11}, nums(5, 7, 5)) // nil 表初始化;群組 5 +2、群組 7 +1
	this.Equal(int32(2), guest.EffectImmune[5])
	this.Equal(int32(1), guest.EffectImmune[7])

	commandEffectImmuneAdd(eng, []InstanceID{11}, nums(7)) // 表已存在(非 nil 分支);群組 7 → 2
	this.Equal(int32(2), guest.EffectImmune[7])

	commandEffectImmuneAdd(eng, []InstanceID{11}, []exprs.Value{exprs.NewText("x")}) // 非數值 → 略過
	this.Equal(int32(2), guest.EffectImmune[5])

	commandEffectImmuneDel(eng, []InstanceID{11}, nums(5, 5, 5)) // 群組 5：2 → 1 → 0 → 夾 0
	this.Equal(int32(0), guest.EffectImmune[5])

	commandEffectImmuneDel(eng, []InstanceID{11}, []exprs.Value{exprs.NewText("x"), exprs.NewNum(7)}) // 非數值略過、群組 7 -1
	this.Equal(int32(1), guest.EffectImmune[7])

	commandEffectImmuneAdd(eng, []InstanceID{99}, nums(5)) // 非顧客實例 → no-op（Add）
	commandEffectImmuneDel(eng, []InstanceID{99}, nums(7)) // 非顧客實例 → no-op（Del）
	this.Equal(int32(0), guest.EffectImmune[5])
	this.Equal(int32(1), guest.EffectImmune[7])
}

func (this *SuiteCommandGuest) TestSkillImmune() {
	runtime := NewRuntime(0)
	guest := &Guest{InstanceID: 11, SeatID: 1}
	runtime.Seat[1] = guest
	eng := this.engine(runtime)

	commandSkillImmuneAdd(eng, []InstanceID{11}, nums(9))
	this.Equal(int32(1), guest.SkillImmune[9])

	commandSkillImmuneDel(eng, []InstanceID{11}, nums(9))
	this.Equal(int32(0), guest.SkillImmune[9])
}

func (this *SuiteCommandGuest) TestTaskAdd() {
	runtime := NewRuntime(0)
	guest := &Guest{InstanceID: 11, SeatID: 1}
	runtime.Seat[1] = guest
	eng := this.engine(runtime)

	commandTaskAdd(eng, []InstanceID{11}, nums(1, 2005)) // 行動類型 1(耐心)、技能 2005
	this.Require().Len(runtime.Action, 1)
	this.Equal(guest, runtime.Action[0].Guest)
	this.Equal(TaskCalm, runtime.Action[0].Kind)
	this.Equal(int32(2005), runtime.Action[0].SkillID)

	commandTaskAdd(eng, []InstanceID{11}, nums(1)) // 缺技能編號 → no-op
	this.Len(runtime.Action, 1)

	commandTaskAdd(eng, []InstanceID{11}, nil) // 缺行動類型 → no-op
	this.Len(runtime.Action, 1)

	commandTaskAdd(eng, []InstanceID{99}, nums(0, 1)) // 非顧客實例 → no-op
	this.Len(runtime.Action, 1)
}

func (this *SuiteCommandGuest) TestLocateGuestContainers() {
	runtime := NewRuntime(0)
	runtime.Seat[1] = &Guest{InstanceID: 1, SeatID: 1}
	runtime.Wait = []*Guest{{InstanceID: 2}}
	runtime.Roam = []*Guest{{InstanceID: 3}}
	runtime.Cardify = []*Guest{{InstanceID: 4}}
	eng := this.engine(runtime)

	for id, want := range map[InstanceID]ContainerKind{
		1: ContainerSeat, 2: ContainerWait, 3: ContainerRoam, 4: ContainerCardify,
	} {
		guest, where, ok := eng.locateGuest(id)
		this.Require().True(ok)
		this.Equal(id, guest.InstanceID)
		this.Equal(want, where)
	} // for

	_, where, ok := eng.locateGuest(99) // 不存在
	this.False(ok)
	this.Equal(ContainerNone, where)
}

// === 測試輔助(置尾) ===

func (this *SuiteCommandGuest) engine(runtime *Runtime) *Engine {
	return NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{}, nil)
}
