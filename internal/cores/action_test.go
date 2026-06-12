package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteAction(t *testing.T) {
	suite.Run(t, new(SuiteAction))
}

// SuiteAction 驗證行動實例與行動佇列(action.go): 三元組建構 / 取值, 佇列尾入 / 首出。
type SuiteAction struct {
	suite.Suite
}

// TestNewAction 驗證 NewAction 建構行動三元組。
func (this *SuiteAction) TestNewAction() {
	guest := &Guest{instanceID: 1}
	action := NewAction(guest, TaskCalm, 2005)
	this.Same(guest, action.GetGuest())
	this.Equal(TaskCalm, action.GetKind())
	this.Equal(int32(2005), action.GetSkillID())
}

// TestActionGetGuest 驗證 GetGuest 取回顧客實例。
func (this *SuiteAction) TestActionGetGuest() {
	guest := &Guest{instanceID: 2}
	this.Same(guest, NewAction(guest, TaskSate, 0).GetGuest())
}

// TestActionGetKind 驗證 GetKind 取回行動類型。
func (this *SuiteAction) TestActionGetKind() {
	this.Equal(TaskSate, NewAction(nil, TaskSate, 0).GetKind())
	this.Equal(TaskCalm, NewAction(nil, TaskCalm, 0).GetKind())
}

// TestActionGetSkillID 驗證 GetSkillID 取回技能編號。
func (this *SuiteAction) TestActionGetSkillID() {
	this.Equal(int32(30001), NewAction(nil, TaskSate, 30001).GetSkillID())
}

// TestActionListPush 驗證 Push 加入佇列尾端(先入者居首)。
func (this *SuiteAction) TestActionListPush() {
	first := NewAction(&Guest{instanceID: 1}, TaskSate, 101)
	second := NewAction(&Guest{instanceID: 2}, TaskCalm, 102)
	list := ActionList{}

	list.Push(first)
	list.Push(second)
	this.Require().Len(list, 2)
	this.Same(first, list[0]) // 先入者居首
	this.Same(second, list[1])
}

// TestActionListPop 驗證 Pop 先進先出彈出、空佇列回 nil。
func (this *SuiteAction) TestActionListPop() {
	first := NewAction(&Guest{instanceID: 1}, TaskSate, 101)
	second := NewAction(&Guest{instanceID: 2}, TaskCalm, 102)
	list := ActionList{first, second}

	this.Same(first, list.Pop()) // 先進先出
	this.Same(second, list.Pop())
	this.Empty(list)
	this.Nil(list.Pop()) // 空佇列 → nil
}
