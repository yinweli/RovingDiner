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
	game := NewGame(0, nil, nil, nil, nil)
	guest := &Guest{instanceID: 11, seatID: 1}
	game.Seat[1] = guest
	injectPort(game)

	commandEffectImmuneAdd(game, []InstanceID{11}, nums(5, 7, 5)) // nil 表初始化;群組 5 +2、群組 7 +1
	this.Equal(int32(2), guest.GetEffectImmune().Get(5))
	this.Equal(int32(1), guest.GetEffectImmune().Get(7))

	commandEffectImmuneAdd(game, []InstanceID{11}, nums(7)) // 表已存在(非 nil 分支);群組 7 → 2
	this.Equal(int32(2), guest.GetEffectImmune().Get(7))

	commandEffectImmuneAdd(game, []InstanceID{11}, []exprs.Value{exprs.NewText("x")}) // 非數值 → 略過
	this.Equal(int32(2), guest.GetEffectImmune().Get(5))

	commandEffectImmuneDel(game, []InstanceID{11}, nums(5, 5, 5)) // 群組 5：2 → 1 → 0 → 夾 0
	this.Equal(int32(0), guest.GetEffectImmune().Get(5))

	commandEffectImmuneDel(game, []InstanceID{11}, []exprs.Value{exprs.NewText("x"), exprs.NewNum(7)}) // 非數值略過、群組 7 -1
	this.Equal(int32(1), guest.GetEffectImmune().Get(7))

	commandEffectImmuneAdd(game, []InstanceID{99}, nums(5)) // 非顧客實例 → no-op（Add）
	commandEffectImmuneDel(game, []InstanceID{99}, nums(7)) // 非顧客實例 → no-op（Del）
	this.Equal(int32(0), guest.GetEffectImmune().Get(5))
	this.Equal(int32(1), guest.GetEffectImmune().Get(7))
}

func (this *SuiteCommandGuest) TestSkillImmune() {
	game := NewGame(0, nil, nil, nil, nil)
	guest := &Guest{instanceID: 11, seatID: 1}
	game.Seat[1] = guest
	injectPort(game)

	commandSkillImmuneAdd(game, []InstanceID{11}, nums(9))
	this.Equal(int32(1), guest.GetSkillImmune().Get(9))

	commandSkillImmuneDel(game, []InstanceID{11}, nums(9))
	this.Equal(int32(0), guest.GetSkillImmune().Get(9))
}

func (this *SuiteCommandGuest) TestTaskAdd() {
	game := NewGame(0, nil, nil, nil, nil)
	guest := &Guest{instanceID: 11, seatID: 1}
	game.Seat[1] = guest
	injectPort(game)

	commandTaskAdd(game, []InstanceID{11}, nums(1, 2005)) // 行動類型 1(耐心)、技能 2005
	this.Require().Len(game.Action, 1)
	this.Equal(guest, game.Action[0].GetGuest())
	this.Equal(TaskCalm, game.Action[0].GetKind())
	this.Equal(int32(2005), game.Action[0].GetSkillID())

	commandTaskAdd(game, []InstanceID{11}, nums(1)) // 缺技能編號 → no-op
	this.Len(game.Action, 1)

	commandTaskAdd(game, []InstanceID{11}, nil) // 缺行動類型 → no-op
	this.Len(game.Action, 1)

	commandTaskAdd(game, []InstanceID{99}, nums(0, 1)) // 非顧客實例 → no-op
	this.Len(game.Action, 1)
}

func (this *SuiteCommandGuest) TestLocateGuestContainers() {
	game := NewGame(0, nil, nil, nil, nil)
	game.Seat[1] = &Guest{instanceID: 1, seatID: 1}
	game.Wait = []*Guest{{instanceID: 2}}
	game.Roam = []*Guest{{instanceID: 3}}
	game.Cardify = []*Guest{{instanceID: 4}}
	injectPort(game)

	for id, want := range map[InstanceID]ContainerKind{
		1: ContainerSeat, 2: ContainerWait, 3: ContainerRoam, 4: ContainerCardify,
	} {
		guest, where, ok := game.locateGuest(id)
		this.Require().True(ok)
		this.Equal(id, guest.GetInstanceID())
		this.Equal(want, where)
	} // for

	_, where, ok := game.locateGuest(99) // 不存在
	this.False(ok)
	this.Equal(ContainerNone, where)
}

// === 測試輔助(置尾) ===
