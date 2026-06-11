package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteCommandGuest(t *testing.T) {
	suite.Run(t, new(SuiteCommandGuest))
}

// SuiteCommandGuest 驗證顧客免疫 / 行動命令(commandGuest.go): effectImmune± / skillImmune±(varargs、夾 ≥0)、taskAdd。
type SuiteCommandGuest struct {
	suite.Suite
}

func (this *SuiteCommandGuest) TestEffectImmune() {
	game := newGame()
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(1, guest)
	id := []cores.InstanceID{guest.GetInstanceID()}

	commandEffectImmuneAdd(game, id, nums(5, 7, 5)) // 群組 5 +2、群組 7 +1
	this.Equal(int32(2), guest.GetEffectImmune().Get(5))
	this.Equal(int32(1), guest.GetEffectImmune().Get(7))

	commandEffectImmuneAdd(game, id, nums(7)) // 群組 7 → 2
	this.Equal(int32(2), guest.GetEffectImmune().Get(7))

	commandEffectImmuneAdd(game, id, []exprs.Value{exprs.NewText("x")}) // 非數值 → 略過
	this.Equal(int32(2), guest.GetEffectImmune().Get(5))

	commandEffectImmuneDel(game, id, nums(5, 5, 5)) // 群組 5:2 → 1 → 0 → 夾 0
	this.Equal(int32(0), guest.GetEffectImmune().Get(5))

	commandEffectImmuneDel(game, id, []exprs.Value{exprs.NewText("x"), exprs.NewNum(7)}) // 非數值略過、群組 7 -1
	this.Equal(int32(1), guest.GetEffectImmune().Get(7))

	commandEffectImmuneAdd(game, []cores.InstanceID{99}, nums(5)) // 非顧客實例 → no-op(Add)
	commandEffectImmuneDel(game, []cores.InstanceID{99}, nums(7)) // 非顧客實例 → no-op(Del)
	this.Equal(int32(0), guest.GetEffectImmune().Get(5))
	this.Equal(int32(1), guest.GetEffectImmune().Get(7))
}

func (this *SuiteCommandGuest) TestSkillImmune() {
	game := newGame()
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(1, guest)
	id := []cores.InstanceID{guest.GetInstanceID()}

	commandSkillImmuneAdd(game, id, nums(9))
	this.Equal(int32(1), guest.GetSkillImmune().Get(9))

	commandSkillImmuneDel(game, id, nums(9))
	this.Equal(int32(0), guest.GetSkillImmune().Get(9))
}

// TestImmuneEmit 驗證免疫命令的屬性事件(群組維度投影; M22 拍板): Operand 載群組編號、前後值載該群組計數、夾 0 不動以 Before == After 表達。
func (this *SuiteCommandGuest) TestImmuneEmit() {
	game, record := newGameRecord()
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(1, guest)
	id := []cores.InstanceID{guest.GetInstanceID()}

	commandEffectImmuneAdd(game, id, nums(5))
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventProperty, DataID: 501, InstanceID: guest.GetInstanceID(), Attr: "effectImmune", Op: cores.AssignAdd, Operand: 5, Before: 0, After: 1}, record.Event[0])

	commandSkillImmuneDel(game, id, nums(9)) // 夾 0 不動 → Before == After、Op 為 Sub
	this.Require().Len(record.Event, 2)
	this.Equal(cores.EventData{Kind: cores.EventProperty, DataID: 501, InstanceID: guest.GetInstanceID(), Attr: "skillImmune", Op: cores.AssignSub, Operand: 9, Before: 0, After: 0}, record.Event[1])
}

func (this *SuiteCommandGuest) TestTaskAdd() {
	game := newGame()
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(1, guest)
	id := []cores.InstanceID{guest.GetInstanceID()}

	commandTaskAdd(game, id, nums(1, 2005)) // 行動類型 1(耐心)、技能 2005
	this.Require().Len(game.Action, 1)
	this.Equal(guest, game.Action[0].GetGuest())
	this.Equal(cores.TaskCalm, game.Action[0].GetKind())
	this.Equal(int32(2005), game.Action[0].GetSkillID())

	commandTaskAdd(game, id, nums(1)) // 缺技能編號 → no-op
	this.Len(game.Action, 1)

	commandTaskAdd(game, id, nil) // 缺行動類型 → no-op
	this.Len(game.Action, 1)

	commandTaskAdd(game, []cores.InstanceID{99}, nums(0, 1)) // 非顧客實例 → no-op
	this.Len(game.Action, 1)
}
