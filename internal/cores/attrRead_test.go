package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteAttrRead(t *testing.T) {
	suite.Run(t, new(SuiteAttrRead))
}

// SuiteAttrRead 驗證全域屬性讀取詞彙表(attrRead.go):純值 / 容器 / 衍生 / 物件引用 / self / 鎖 / 靜態 / 查詢函式。
type SuiteAttrRead struct {
	suite.Suite
}

func (this *SuiteAttrRead) TestAttrReadValue() {
	runtime := NewRuntime(0)
	runtime.Game.Morale = Value{Value: 25}
	runtime.Game.MoraleMax = Value{Value: 50}
	runtime.Game.MoraleShield = Value{Value: 9}
	runtime.Game.MoraleBlock = Value{Value: 8}
	runtime.Game.Score = Value{Value: 7}
	runtime.Game.Energy = Value{Value: 3}
	runtime.Game.EnergyMax = Value{Value: 6}
	runtime.Game.HandMax = Value{Value: 10}
	runtime.Game.DrawMax = Value{Value: 5}
	runtime.Game.Round = 4
	runtime.Game.RoundMax = 12
	runtime.Game.DamageValue = 6
	runtime.Game.SeatCount = 2
	runtime.Game.ExitLastSeat = 3
	runtime.Game.ExitCount = 1
	runtime.Game.TaskSkill = 77
	runtime.Game.TaskCount = 4
	runtime.Game.DrawCount = 2
	runtime.Game.DropCount = 1
	runtime.Game.PlayCount = 3
	runtime.Game.ExileCount = 1
	runtime.Game.MorphCount = 2
	runtime.Game.MorphOldID = 100
	runtime.Game.MorphNewID = 200
	runtime.Game.NextPhase = PhasePlayerAction
	eng := &Engine{runtime: runtime}

	this.Equal(float64(25), this.num(eng, "morale"))
	this.Equal(float64(50), this.num(eng, "moraleMax"))
	this.Equal(float64(9), this.num(eng, "moraleShield"))
	this.Equal(float64(8), this.num(eng, "moraleBlock"))
	this.Equal(float64(7), this.num(eng, "score"))
	this.Equal(float64(3), this.num(eng, "energy"))
	this.Equal(float64(6), this.num(eng, "energyMax"))
	this.Equal(float64(0), this.num(eng, "energyKeep")) // 值固定 0
	this.Equal(float64(10), this.num(eng, "handMax"))
	this.Equal(float64(5), this.num(eng, "drawMax"))
	this.Equal(float64(4), this.num(eng, "round"))
	this.Equal(float64(12), this.num(eng, "roundMax"))
	this.Equal(float64(6), this.num(eng, "damageValue"))
	this.Equal(float64(2), this.num(eng, "seatCount"))
	this.Equal(float64(3), this.num(eng, "exitLastSeat"))
	this.Equal(float64(1), this.num(eng, "exitCount"))
	this.Equal(float64(77), this.num(eng, "taskSkill"))
	this.Equal(float64(4), this.num(eng, "taskCount"))
	this.Equal(float64(2), this.num(eng, "drawCount"))
	this.Equal(float64(1), this.num(eng, "dropCount"))
	this.Equal(float64(3), this.num(eng, "playCount"))
	this.Equal(float64(1), this.num(eng, "exileCount"))
	this.Equal(float64(2), this.num(eng, "morphCount"))
	this.Equal(float64(100), this.num(eng, "morphOldID"))
	this.Equal(float64(200), this.num(eng, "morphNewID"))

	value, ok := eng.Attr("nextPhase", nil)
	this.True(ok)
	this.Equal("玩家行動", value.Text())
}

func (this *SuiteAttrRead) TestAttrReadContainer() {
	runtime := NewRuntime(0)
	runtime.Hand = []*Card{{}, {}}
	runtime.Wait = []*Guest{{}}
	runtime.Roam = []*Guest{{}, {}, {}}
	runtime.Seat[1] = &Guest{}
	runtime.Cardify = []*Guest{{}}
	runtime.Action = []*Action{{}}
	eng := &Engine{runtime: runtime}

	this.Equal(float64(1), this.num(eng, "seatSize"))
	this.Equal(float64(1), this.num(eng, "waitSize"))
	this.Equal(float64(3), this.num(eng, "roamSize"))
	this.Equal(float64(1), this.num(eng, "cardifySize"))
	this.Equal(float64(1), this.num(eng, "taskSize"))
	this.Equal(float64(6), this.num(eng, "guestSize")) // 1+1+3+1
}

func (this *SuiteAttrRead) TestAttrReadRoundLeft() {
	runtime := NewRuntime(0)
	runtime.Game.Round = 8
	runtime.Game.RoundMax = 10
	eng := &Engine{runtime: runtime}
	this.Equal(float64(2), this.num(eng, "roundLeft"))

	runtime.Game.Round = 12 // 超過上限 → 夾 0
	this.Equal(float64(0), this.num(eng, "roundLeft"))
}

func (this *SuiteAttrRead) TestAttrReadObject() {
	runtime := NewRuntime(0)
	card := &Card{InstanceID: 11}
	guest := &Guest{InstanceID: 21}
	runtime.Game.DrawLast = card
	runtime.Game.DropLast = card
	runtime.Game.ExileLast = card
	runtime.Game.MorphLast = card
	runtime.Game.SeatLast = guest
	runtime.Game.ExitLast = guest
	runtime.Game.TaskGuest = guest
	runtime.Game.DamageGuest = guest
	eng := &Engine{runtime: runtime}

	for _, name := range []string{"drawLast", "dropLast", "exileLast", "morphLast"} {
		value, ok := eng.Attr(name, nil)
		this.True(ok)
		this.True(value.Ref().Same(cardRef{card: card}), name)
	} // for

	for _, name := range []string{"seatLast", "exitLast", "taskGuest", "damageGuest"} {
		value, ok := eng.Attr(name, nil)
		this.True(ok)
		this.True(value.Ref().Same(guestRef{guest: guest}), name)
	} // for

	value, ok := eng.Attr("playLast", nil) // 尚無 → 空物件
	this.True(ok)
	this.True(value.IsNone())
}

func (this *SuiteAttrRead) TestAttrReadSelf() {
	eng := &Engine{runtime: NewRuntime(0)}

	_, ok := eng.Attr("self", nil) // 未綁定 → 失敗
	this.False(ok)

	eng.self = &Self{} // 綁定空物件 → none
	value, ok := eng.Attr("self", nil)
	this.True(ok)
	this.True(value.IsNone())

	guest := &Guest{InstanceID: 9} // 綁定顧客 → 引用
	eng.self = &Self{Guest: guest}
	value, ok = eng.Attr("self", nil)
	this.True(ok)
	this.True(value.Ref().Same(guestRef{guest: guest}))
}

func (this *SuiteAttrRead) TestAttrReadLock() {
	runtime := NewRuntime(0)
	runtime.Game.Morale = Value{Lock: 1}
	runtime.Game.MoraleMax = Value{Lock: 2}
	runtime.Game.MoraleShield = Value{Lock: 3}
	runtime.Game.MoraleBlock = Value{Lock: 4}
	runtime.Game.Score = Value{Lock: 5}
	runtime.Game.Energy = Value{Value: 3, Lock: 6}
	runtime.Game.EnergyMax = Value{Lock: 7}
	runtime.Game.EnergyKeep = Value{Lock: 8}
	runtime.Game.HandMax = Value{Lock: 9}
	runtime.Game.DrawMax = Value{Lock: 10}
	eng := &Engine{runtime: runtime}

	this.Equal(float64(1), this.num(eng, "moraleLock"))
	this.Equal(float64(2), this.num(eng, "moraleMaxLock"))
	this.Equal(float64(3), this.num(eng, "moraleShieldLock"))
	this.Equal(float64(4), this.num(eng, "moraleBlockLock"))
	this.Equal(float64(5), this.num(eng, "scoreLock"))
	this.Equal(float64(6), this.num(eng, "energyLock"))
	this.Equal(float64(7), this.num(eng, "energyMaxLock"))
	this.Equal(float64(8), this.num(eng, "energyKeepLock"))
	this.Equal(float64(9), this.num(eng, "handMaxLock"))
	this.Equal(float64(10), this.num(eng, "drawMaxLock"))
}

func (this *SuiteAttrRead) TestAttrReadStatic() {
	runtime := NewRuntime(0)
	runtime.Seat[1] = &Guest{}
	runtime.Seat[2] = &Guest{}
	eng := &Engine{runtime: runtime, data: buildSheet()}

	this.Equal(float64(1), this.num(eng, "seatLeft"))  // 3 座位 - 占用 2
	this.Equal(float64(2), this.num(eng, "tableSize")) // 桌 1、2
	this.Equal(float64(2), this.num(eng, "tableGuest", exprs.NewNum(1)))
	this.Equal(float64(0), this.num(eng, "tableGuest", exprs.NewNum(2)))
	this.Equal(float64(0), this.num(eng, "tableGuest", exprs.NewNum(0))) // N==0 不命中

	_, ok := eng.Attr("tableGuest", nil) // 缺參數 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRead) TestAttrReadTableCount() {
	runtime := NewRuntime(0)
	runtime.Seat[1] = &Guest{}
	runtime.Seat[2] = &Guest{} // 桌1 = 2 人;桌2 = 0 人
	eng := &Engine{runtime: runtime, data: buildSheet()}

	this.Equal(float64(1), this.num(eng, "tableCount", exprs.NewText(">="), exprs.NewNum(2)))
	this.Equal(float64(1), this.num(eng, "tableCount", exprs.NewText("=="), exprs.NewNum(0)))
	this.Equal(float64(1), this.num(eng, "tableCount", exprs.NewText(">"), exprs.NewNum(0)))
	this.Equal(float64(2), this.num(eng, "tableCount", exprs.NewText("<="), exprs.NewNum(2)))
	this.Equal(float64(0), this.num(eng, "tableCount", exprs.NewText("<"), exprs.NewNum(0)))
	this.Equal(float64(2), this.num(eng, "tableCount", exprs.NewText("!="), exprs.NewNum(1)))

	_, ok := eng.Attr("tableCount", []exprs.Value{exprs.NewText("~="), exprs.NewNum(0)}) // 未知運算符
	this.False(ok)
	_, ok = eng.Attr("tableCount", []exprs.Value{exprs.NewNum(1)}) // 參數數量不符
	this.False(ok)
	_, ok = eng.Attr("tableCount", []exprs.Value{exprs.NewNum(1), exprs.NewNum(2)}) // 型別不符
	this.False(ok)
}

func (this *SuiteAttrRead) TestAttrReadGroupQuery() {
	runtime := NewRuntime(0)
	runtime.Hand = []*Card{{CardID: 101}, {CardID: 101}, {CardID: 102}}
	runtime.Deck = []*Card{{CardID: 101}}
	runtime.Drop = []*Card{{CardID: 102}, {CardID: 102}}
	runtime.Exile = []*Card{{CardID: 101}}
	runtime.Game.DrawTotal = map[int32]int32{1: 3, 2: 5}
	runtime.Game.DropTotal = map[int32]int32{1: 1}
	runtime.Game.PlayTotal = map[int32]int32{2: 2}
	runtime.Game.ExileTotal = map[int32]int32{1: 4}
	eng := &Engine{runtime: runtime, data: buildSheet()}

	this.Equal(float64(2), this.num(eng, "handSize", exprs.NewNum(1))) // 群組 1 = 卡 101 兩張
	this.Equal(float64(1), this.num(eng, "handSize", exprs.NewNum(2))) // 群組 2 = 卡 102 一張
	this.Equal(float64(3), this.num(eng, "handSize", exprs.NewNum(0))) // N==0 全量
	this.Equal(float64(1), this.num(eng, "deckSize", exprs.NewNum(1)))
	this.Equal(float64(2), this.num(eng, "dropSize", exprs.NewNum(0)))
	this.Equal(float64(1), this.num(eng, "exileSize", exprs.NewNum(1)))

	this.Equal(float64(3), this.num(eng, "drawTotal", exprs.NewNum(1)))
	this.Equal(float64(8), this.num(eng, "drawTotal", exprs.NewNum(0))) // 全加總
	this.Equal(float64(1), this.num(eng, "dropTotal", exprs.NewNum(1)))
	this.Equal(float64(2), this.num(eng, "playTotal", exprs.NewNum(2)))
	this.Equal(float64(4), this.num(eng, "exileTotal", exprs.NewNum(1)))

	_, ok := eng.Attr("handSize", []exprs.Value{exprs.NewText("x")}) // 參數型別不符
	this.False(ok)
	_, ok = eng.Attr("drawTotal", nil) // 缺參數
	this.False(ok)
}

// num 取全域屬性求值結果的數字;斷言命中且為數值。
func (this *SuiteAttrRead) num(eng *Engine, name string, arg ...exprs.Value) float64 {
	value, ok := eng.Attr(name, arg)
	this.Require().True(ok)
	this.Require().True(value.IsNum())
	return value.Num()
}

// buildSheet 組裝測試用 in-memory 靜態表:3 座位(桌1:座1,2;桌2:座3)、2 卡牌(群組 1 / 2)、1 效果(群組 5)。
// 由 attrRead_test.go 與 attrRefRead_test.go 共用。
func buildSheet() *sheeter.Sheeter {
	data := &sheeter.Sheeter{}
	data.Seat.Data = map[int32]*sheeter.Seat{
		1: {ID: 1, TableID: 1, SameSeatID: []int32{1, 2}, NearSeatID: []int32{3}},
		2: {ID: 2, TableID: 1, SameSeatID: []int32{1, 2}, NearSeatID: []int32{3}},
		3: {ID: 3, TableID: 2, SameSeatID: []int32{3}, NearSeatID: []int32{1, 2}},
	}
	data.Card.Data = map[int32]*sheeter.Card{
		101: {ID: 101, Group: 1},
		102: {ID: 102, Group: 2},
		103: {ID: 103, Group: 1, Cost: 2, Keep: true, Seal: true, SkillID: 301}, // M9.3 newCard 用（bool 欄 → 鎖、SkillID → 效果列表）
	}
	data.Effect.Data = map[int32]*sheeter.Effect{
		201: {ID: 201, Group: 5},
	}
	data.Skill.Data = map[int32]*sheeter.Skill{
		301: {ID: 301, EffectID: []int32{401, 402}}, // 卡 103 的技能效果列表
	}
	data.Award.Data = map[int32]*sheeter.Award{
		1: {ID: 1, Group: 7, CardID: 101, Weight: 3}, // 群組 7：候選 101(w3) / 102(w1)
		2: {ID: 2, Group: 7, CardID: 102, Weight: 1},
		3: {ID: 3, Group: 8, CardID: 101, Weight: 0}, // 群組 8：權重 0 → roll no-op
		4: {ID: 4, Group: 9, CardID: 999, Weight: 1}, // 群組 9：抽中編號 999 無卡牌資料 → 跳過該張
	}
	data.Guest.Data = map[int32]*sheeter.Guest{
		501: {ID: 501, Score: 0, ScoreMax: 10, Morale: 5, MoraleMax: 8, Calm: 3, SateSeal: true},
	}
	return data
}
