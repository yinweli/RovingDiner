package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/games"
	"github.com/yinweli/RovingDiner/internal/tester"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteMirror(t *testing.T) {
	suite.Run(t, new(SuiteMirror))
}

// SuiteMirror 驗證世界鏡像(mirror.go): 座標蓋章與各類事件摺疊。
type SuiteMirror struct {
	suite.Suite
}

// TestNewMirror 驗證建構: 各表就位、座標歸零。
func (this *SuiteMirror) TestNewMirror() {
	target := newMirror(testSheet())
	this.NotNil(target.sheet)
	this.NotNil(target.attr)
	this.NotNil(target.lock)
	this.NotNil(target.card)
	this.NotNil(target.guest)
	this.NotNil(target.zone)
	this.NotNil(target.seat)
	this.Equal(int32(0), target.round)
	this.Equal(cores.PhaseNone, target.phase)
}

// TestMirrorApply 驗證分派: 座標一律更新(無盤面投影的類別亦然)。
func (this *SuiteMirror) TestMirrorApply() {
	target := newMirror(testSheet())
	target.Apply(cores.EventData{Kind: cores.EventPhase, Round: 3, Phase: cores.PhaseRoundStart})
	this.Equal(int32(3), target.round)
	this.Equal(cores.PhaseRoundStart, target.phase)
	this.Empty(target.attr)

	target.Apply(cores.EventData{Kind: cores.EventScope, Round: 4}) // scope / select / phase 無盤面投影
	this.Equal(int32(4), target.round)
}

// TestMirrorApplyProperty 驗證屬性摺疊: 全域 / 實例路由、@ # 分流鎖定計數、未知實例防禦。
func (this *SuiteMirror) TestMirrorApplyProperty() {
	target := newMirror(testSheet())
	target.Apply(cores.EventData{Kind: cores.EventProperty, Attr: "morale", Op: cores.AssignSet, After: 30})
	this.Equal(float64(30), target.attr["morale"]) // 全域值表

	target.Apply(cores.EventData{Kind: cores.EventProperty, Attr: "energyKeep", Op: cores.AssignLock, After: 1})
	this.Equal(float64(1), target.lock["energyKeep"]) // 全域鎖定計數表

	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 11, From: cores.ContainerNone, To: cores.ContainerHand}) // 卡牌出生
	target.Apply(cores.EventData{Kind: cores.EventProperty, InstanceID: 11, Attr: "cost", Op: cores.AssignAdd, After: 5})
	this.Equal(float64(5), target.card[11].attr["cost"]) // 卡牌實例路由

	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 21, From: cores.ContainerNone, To: cores.ContainerWait}) // 顧客出生
	target.Apply(cores.EventData{Kind: cores.EventProperty, InstanceID: 21, Attr: "calm", Op: cores.AssignSub, After: 2})
	this.Equal(float64(2), target.guest[21].attr["calm"]) // 顧客實例路由

	target.Apply(cores.EventData{Kind: cores.EventProperty, InstanceID: 21, Attr: "sateSeal", Op: cores.AssignUnlock, After: 0})
	this.Equal(float64(0), target.guest[21].lock["sateSeal"])

	target.Apply(cores.EventData{Kind: cores.EventProperty, InstanceID: 99, Attr: "calm", After: 9}) // 未知實例 → 不投影
	this.Empty(target.attr["calm"])
}

// TestMirrorApplyContainer 驗證容器摺疊: 出生初值查表與插入端、搬移、座位表、銷毀、重整快照、綁定語境欄。
func (this *SuiteMirror) TestMirrorApplyContainer() {
	target := newMirror(testSheet())
	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 11, From: cores.ContainerNone, To: cores.ContainerDeck})
	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 12, From: cores.ContainerNone, To: cores.ContainerDeck})
	this.Equal([]cores.InstanceID{12, 11}, target.zone[cores.ContainerDeck]) // 牌堆前插 = 新進入者置頂
	this.Equal(float64(2), target.card[11].attr["cost"])                     // 出生初值查表
	this.Equal(float64(1), target.card[11].lock["cardSeal"])

	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 11, From: cores.ContainerDeck, To: cores.ContainerHand}) // 搬移
	this.Equal([]cores.InstanceID{12}, target.zone[cores.ContainerDeck])
	this.Equal([]cores.InstanceID{11}, target.zone[cores.ContainerHand])

	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 21, From: cores.ContainerNone, To: cores.ContainerWait})
	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 22, From: cores.ContainerNone, To: cores.ContainerWait})
	this.Equal([]cores.InstanceID{21, 22}, target.zone[cores.ContainerWait]) // 排隊尾插(隊首在前)
	this.Equal(float64(3), target.guest[21].attr["calm"])                    // 出生初值查表

	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 21, From: cores.ContainerWait, To: cores.ContainerSeat, SeatID: 2}) // 入座
	this.Equal(cores.InstanceID(21), target.seat[2])
	this.Equal([]cores.InstanceID{22}, target.zone[cores.ContainerWait])

	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 21, From: cores.ContainerSeat, To: cores.ContainerNone}) // 離場銷毀
	this.Empty(target.seat)
	this.Nil(target.guest[21])

	target.Apply(cores.EventData{Kind: cores.EventContainer, From: cores.ContainerDeck, To: cores.ContainerDeck, Pick: []cores.PickData{{DataID: 101, InstanceID: 12}, {DataID: 101, InstanceID: 13}}}) // 重整快照 → 整堆置換
	this.Equal([]cores.InstanceID{12, 13}, target.zone[cores.ContainerDeck])

	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 31, From: cores.ContainerNone, To: cores.ContainerCardify}) // 卡牌化來源
	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 32, From: cores.ContainerNone, To: cores.ContainerHand, BindID: 501, BindInstanceID: 31})
	this.Equal(int32(501), target.card[32].bindID) // 綁定語境欄摺入
	this.Equal(cores.InstanceID(31), target.card[32].bindInstanceID)

	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 31, From: cores.ContainerCardify, To: cores.ContainerSeat, SeatID: 1}) // 還原 → 解綁推斷
	this.Equal(int32(0), target.card[32].bindID)
	this.Equal(cores.NoneID, target.card[32].bindInstanceID)
}

// TestMirrorApplyInstance 驗證實例摺疊(morph): 銷毀記槽位、建立同槽位插回; 無暫存槽位只建視圖。
func (this *SuiteMirror) TestMirrorApplyInstance() {
	target := newMirror(testSheet())
	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 11, From: cores.ContainerNone, To: cores.ContainerHand})
	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 12, From: cores.ContainerNone, To: cores.ContainerHand})
	target.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 13, From: cores.ContainerNone, To: cores.ContainerHand}) // 手牌: [13 12 11]

	target.Apply(cores.EventData{Kind: cores.EventInstance, DataID: 101, InstanceID: 12, Alive: false}) // morph 銷毀舊
	this.Nil(target.card[12])
	this.Equal([]cores.InstanceID{13, 11}, target.zone[cores.ContainerHand])

	target.Apply(cores.EventData{Kind: cores.EventInstance, DataID: 103, InstanceID: 14, Alive: true}) // morph 建立新 → 同槽位插回
	this.Equal([]cores.InstanceID{13, 14, 11}, target.zone[cores.ContainerHand])
	this.Equal(int32(103), target.card[14].dataID)

	target.Apply(cores.EventData{Kind: cores.EventInstance, DataID: 103, InstanceID: 15, Alive: true}) // 無暫存槽位(防禦) → 只建視圖
	this.NotNil(target.card[15])
	this.Equal([]cores.InstanceID{13, 14, 11}, target.zone[cores.ContainerHand])

	target.Apply(cores.EventData{Kind: cores.EventInstance, InstanceID: 99, Alive: false}) // 未知實例銷毀(防禦) → 無槽位
	target.Apply(cores.EventData{Kind: cores.EventInstance, DataID: 103, InstanceID: 16, Alive: true})
	this.NotNil(target.card[16])
	this.Equal([]cores.InstanceID{13, 14, 11}, target.zone[cores.ContainerHand])
}

// TestMirrorApplyEffect 驗證效果佇列摺疊: 加入 upsert 快照、結束分退層 / 退場、其餘階段不動佇列。
func (this *SuiteMirror) TestMirrorApplyEffect() {
	target := newMirror(testSheet())
	target.Apply(cores.EventData{Kind: cores.EventEffect, DataID: 501, InstanceID: 21, EffectID: 401, EffectInstanceID: 41, Stage: cores.EffectStageJoin, Stack: 2, Expire: 5, Alive: true})
	this.Require().Len(target.effect, 1)
	this.Equal(&effectView{effectID: 401, instanceID: 41, stack: 2, expire: 5, selfID: 501, selfInstanceID: 21}, target.effect[0])

	target.Apply(cores.EventData{Kind: cores.EventEffect, EffectID: 401, EffectInstanceID: 41, Stage: cores.EffectStageJoin, Stack: 3, Expire: 7, Alive: true}) // 再疊 upsert
	this.Require().Len(target.effect, 1)
	this.Equal(int32(3), target.effect[0].stack)
	this.Equal(int32(7), target.effect[0].expire)

	target.Apply(cores.EventData{Kind: cores.EventEffect, EffectID: 401, EffectInstanceID: 41, Stage: cores.EffectStageTrigger}) // 其餘階段不動佇列
	this.Len(target.effect, 1)

	target.Apply(cores.EventData{Kind: cores.EventEffect, EffectID: 401, EffectInstanceID: 41, Stage: cores.EffectStageEnd, Stack: 1, Expire: 7, Alive: true}) // 退層
	this.Equal(int32(1), target.effect[0].stack)

	target.Apply(cores.EventData{Kind: cores.EventEffect, EffectID: 402, EffectInstanceID: 42, Stage: cores.EffectStageJoin, Stack: 1, Alive: true})
	target.Apply(cores.EventData{Kind: cores.EventEffect, EffectID: 401, EffectInstanceID: 41, Stage: cores.EffectStageEnd, Stack: 1, Alive: false}) // 退場(他項保留)
	this.Require().Len(target.effect, 1)
	this.Equal(cores.InstanceID(42), target.effect[0].instanceID)
}

// TestMirrorApplyAction 驗證行動佇列摺疊: 入列尾插、出列移除首個匹配項、未匹配防禦不動。
func (this *SuiteMirror) TestMirrorApplyAction() {
	target := newMirror(testSheet())
	target.Apply(cores.EventData{Kind: cores.EventAction, DataID: 501, InstanceID: 21, SkillID: 301, Task: cores.TaskCalm, Alive: true})
	target.Apply(cores.EventData{Kind: cores.EventAction, DataID: 501, InstanceID: 22, SkillID: 301, Task: cores.TaskSate, Alive: true})
	this.Require().Len(target.action, 2)
	this.Equal(&actionView{guestID: 501, guestInstanceID: 21, skillID: 301, task: cores.TaskCalm}, target.action[0])

	target.Apply(cores.EventData{Kind: cores.EventAction, DataID: 501, InstanceID: 21, SkillID: 301, Task: cores.TaskCalm}) // 出列
	this.Require().Len(target.action, 1)
	this.Equal(cores.InstanceID(22), target.action[0].guestInstanceID)

	target.Apply(cores.EventData{Kind: cores.EventAction, DataID: 501, InstanceID: 99, SkillID: 301, Task: cores.TaskCalm}) // 未匹配(防禦) → 不動
	this.Len(target.action, 1)
}

// TestMirrorHasEffect 驗證 active 效果查詢: 佇列項 self 命中 / 未命中。
func (this *SuiteMirror) TestMirrorHasEffect() {
	target := newMirror(testSheet())
	target.Apply(cores.EventData{Kind: cores.EventEffect, DataID: 501, InstanceID: 21, EffectID: 401, EffectInstanceID: 41, Stage: cores.EffectStageJoin, Stack: 1, Alive: true})
	this.True(target.hasEffect(21))
	this.False(target.hasEffect(99))
}

// TestMirrorGame 驗證鏡像對完整一局真實事件流的摺疊(端到端): 終局座標為終止站、容器成員皆有視圖(無懸空編號)、
// 六區渲染不噴錯——投影合約與摺疊規則的整合釘。
func (this *SuiteMirror) TestMirrorGame() {
	record := &tester.RecordPresenter{}
	games.Run(0, 601, tester.BuildSheet(), tester.FakeOperator{}, record)
	target := newMirror(tester.BuildSheet())

	for itor := range record.Event {
		target.Apply(record.Event[itor])
	} // for

	this.Contains([]cores.PhaseKind{cores.PhaseGameSucc, cores.PhaseGameFail}, target.phase)

	for _, zone := range target.zone {
		for _, itor := range zone {
			this.True(target.card[itor] != nil || target.guest[itor] != nil) // 容器成員必有視圖
		} // for
	} // for

	for _, itor := range target.seat {
		this.NotNil(target.guest[itor])
	} // for

	for _, itor := range []component{seatPanel{}, poolPanel{}, actionPanel{}, effectPanel{}, handPanel{}, pilePanel{}, statusBar{}} {
		this.NotEmpty(itor.View(target, 100))
	} // for
}

// === 測試輔助(置尾) ===

// testSheet 迷你靜態表(視圖初值與識別碼查表用): 座位 桌1(1,2)/ 桌2(3)、卡 101 / 103、顧客 501、技能 301、效果 401。
func testSheet() *sheeter.Sheeter {
	sheet := &sheeter.Sheeter{}
	sheet.Seat.Data = map[int32]*sheeter.Seat{
		1: {ID: 1, TableID: 1, SameSeatID: []int32{2}, NearSeatID: []int32{3}},
		2: {ID: 2, TableID: 1, SameSeatID: []int32{1}, NearSeatID: []int32{3}},
		3: {ID: 3, TableID: 2, SameSeatID: []int32{3}, NearSeatID: []int32{1, 2}},
	}
	sheet.Card.Data = map[int32]*sheeter.Card{
		101: {ID: 101, Name: "上菜", Cost: 2, ExtraRunMin: 1, ExtraRunMax: 3, Seal: true},
		103: {ID: 103, Name: "結帳", Cost: 1, Keep: true},
	}
	sheet.Guest.Data = map[int32]*sheeter.Guest{
		501: {ID: 501, Name: "老饕", Score: 4, ScoreMax: 10, Morale: 5, MoraleMax: 8, Calm: 3, SateMax: 6, SateSeal: true},
	}
	sheet.Skill.Data = map[int32]*sheeter.Skill{
		301: {ID: 301, Name: "開朗"},
	}
	sheet.Effect.Data = map[int32]*sheeter.Effect{
		401: {ID: 401, Name: "加耐", Kind: 1, RunOrder: 5},
		402: {ID: 402, Name: "護盾", Kind: 2, RunOrder: 9},
		403: {ID: 403, Name: "立即", Kind: 0, RunOrder: 9},
	}
	return sheet
}
