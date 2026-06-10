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
	game := NewGame(0, nil, nil, nil, nil)
	game.morale = NewValue(25, 0)
	game.moraleMax = NewValue(50, 0)
	game.moraleShield = NewValue(9, 0)
	game.moraleBlock = NewValue(8, 0)
	game.score = NewValue(7, 0)
	game.energy = NewValue(3, 0)
	game.energyMax = NewValue(6, 0)
	game.handMax = NewValue(10, 0)
	game.drawMax = NewValue(5, 0)
	game.round = NewValue(4, 0)
	game.roundMax = NewValue(12, 0)
	game.damageValue = 6
	game.seatCount = 2
	game.exitLastSeat = 3
	game.exitCount = 1
	game.taskSkill = 77
	game.taskCount = 4
	game.drawCount = 2
	game.dropCount = 1
	game.playCount = 3
	game.exileCount = 1
	game.morphCount = 2
	game.morphOldID = 100
	game.morphNewID = 200
	game.nextPhase = PhasePlayerAction

	this.Equal(float64(25), this.num(game, "morale"))
	this.Equal(float64(50), this.num(game, "moraleMax"))
	this.Equal(float64(9), this.num(game, "moraleShield"))
	this.Equal(float64(8), this.num(game, "moraleBlock"))
	this.Equal(float64(7), this.num(game, "score"))
	this.Equal(float64(3), this.num(game, "energy"))
	this.Equal(float64(6), this.num(game, "energyMax"))
	this.Equal(float64(0), this.num(game, "energyKeep")) // 值固定 0
	this.Equal(float64(10), this.num(game, "handMax"))
	this.Equal(float64(5), this.num(game, "drawMax"))
	this.Equal(float64(4), this.num(game, "round"))
	this.Equal(float64(12), this.num(game, "roundMax"))
	this.Equal(float64(6), this.num(game, "damageValue"))
	this.Equal(float64(2), this.num(game, "seatCount"))
	this.Equal(float64(3), this.num(game, "exitLastSeat"))
	this.Equal(float64(1), this.num(game, "exitCount"))
	this.Equal(float64(77), this.num(game, "taskSkill"))
	this.Equal(float64(4), this.num(game, "taskCount"))
	this.Equal(float64(2), this.num(game, "drawCount"))
	this.Equal(float64(1), this.num(game, "dropCount"))
	this.Equal(float64(3), this.num(game, "playCount"))
	this.Equal(float64(1), this.num(game, "exileCount"))
	this.Equal(float64(2), this.num(game, "morphCount"))
	this.Equal(float64(100), this.num(game, "morphOldID"))
	this.Equal(float64(200), this.num(game, "morphNewID"))

	value, ok := game.Attr("nextPhase", nil)
	this.True(ok)
	this.Equal("玩家行動", value.Text())
}

func (this *SuiteAttrRead) TestAttrReadContainer() {
	game := NewGame(0, nil, nil, nil, nil)
	game.Hand = []*Card{{}, {}}
	game.Wait = []*Guest{{}}
	game.Roam = []*Guest{{}, {}, {}}
	game.Seat[1] = &Guest{}
	game.Cardify = []*Guest{{}}
	game.Action = ActionList{{}}

	this.Equal(float64(1), this.num(game, "seatSize"))
	this.Equal(float64(1), this.num(game, "waitSize"))
	this.Equal(float64(3), this.num(game, "roamSize"))
	this.Equal(float64(1), this.num(game, "cardifySize"))
	this.Equal(float64(1), this.num(game, "taskSize"))
	this.Equal(float64(6), this.num(game, "guestSize")) // 1+1+3+1
}

func (this *SuiteAttrRead) TestAttrReadRoundLeft() {
	game := NewGame(0, nil, nil, nil, nil)
	game.round = NewValue(8, 0)
	game.roundMax = NewValue(10, 0)
	this.Equal(float64(2), this.num(game, "roundLeft"))

	game.round = NewValue(12, 0) // 超過上限 → 夾 0
	this.Equal(float64(0), this.num(game, "roundLeft"))
}

func (this *SuiteAttrRead) TestAttrReadObject() {
	game := NewGame(0, nil, nil, nil, nil)
	card := &Card{instanceID: 11}
	guest := &Guest{instanceID: 21}
	game.drawLast = card
	game.dropLast = card
	game.exileLast = card
	game.morphLast = card
	game.seatLast = guest
	game.exitLast = guest
	game.taskGuest = guest
	game.damageGuest = guest

	for _, name := range []string{"drawLast", "dropLast", "exileLast", "morphLast"} {
		value, ok := game.Attr(name, nil)
		this.True(ok)
		this.True(NewRefCard(card).IsSame(value.Ref()), name)
	} // for

	for _, name := range []string{"seatLast", "exitLast", "taskGuest", "damageGuest"} {
		value, ok := game.Attr(name, nil)
		this.True(ok)
		this.True(NewRefGuest(guest).IsSame(value.Ref()), name)
	} // for

	value, ok := game.Attr("playLast", nil) // 尚無 → 空物件
	this.True(ok)
	this.True(value.IsNone())
}

func (this *SuiteAttrRead) TestAttrReadSelf() {
	game := NewGame(0, nil, nil, nil, nil)

	_, ok := game.Attr("self", nil) // 未綁定 → 失敗
	this.False(ok)

	game.self = &Ref{} // 綁定空物件 → none
	value, ok := game.Attr("self", nil)
	this.True(ok)
	this.True(value.IsNone())

	guest := &Guest{instanceID: 9} // 綁定顧客 → 引用
	game.self = &Ref{guest: guest}
	value, ok = game.Attr("self", nil)
	this.True(ok)
	this.True(NewRefGuest(guest).IsSame(value.Ref()))
}

func (this *SuiteAttrRead) TestAttrReadLock() {
	game := NewGame(0, nil, nil, nil, nil)
	game.morale = NewValue(0, 1)
	game.moraleMax = NewValue(0, 2)
	game.moraleShield = NewValue(0, 3)
	game.moraleBlock = NewValue(0, 4)
	game.score = NewValue(0, 5)
	game.energy = NewValue(3, 6)
	game.energyMax = NewValue(0, 7)
	game.energyKeep = NewValue(0, 8)
	game.handMax = NewValue(0, 9)
	game.drawMax = NewValue(0, 10)

	this.Equal(float64(1), this.num(game, "moraleLock"))
	this.Equal(float64(2), this.num(game, "moraleMaxLock"))
	this.Equal(float64(3), this.num(game, "moraleShieldLock"))
	this.Equal(float64(4), this.num(game, "moraleBlockLock"))
	this.Equal(float64(5), this.num(game, "scoreLock"))
	this.Equal(float64(6), this.num(game, "energyLock"))
	this.Equal(float64(7), this.num(game, "energyMaxLock"))
	this.Equal(float64(8), this.num(game, "energyKeepLock"))
	this.Equal(float64(9), this.num(game, "handMaxLock"))
	this.Equal(float64(10), this.num(game, "drawMaxLock"))
}

func (this *SuiteAttrRead) TestAttrReadStatic() {
	game := NewGame(0, nil, nil, nil, nil)
	game.Seat[1] = &Guest{}
	game.Seat[2] = &Guest{}
	game.data = buildSheet()

	this.Equal(float64(1), this.num(game, "seatLeft"))  // 3 座位 - 占用 2
	this.Equal(float64(2), this.num(game, "tableSize")) // 桌 1、2
	this.Equal(float64(2), this.num(game, "tableGuest", exprs.NewNum(1)))
	this.Equal(float64(0), this.num(game, "tableGuest", exprs.NewNum(2)))
	this.Equal(float64(0), this.num(game, "tableGuest", exprs.NewNum(0))) // N==0 不命中

	_, ok := game.Attr("tableGuest", nil) // 缺參數 → 失敗
	this.False(ok)
}

func (this *SuiteAttrRead) TestAttrReadTableCount() {
	game := NewGame(0, nil, nil, nil, nil)
	game.Seat[1] = &Guest{}
	game.Seat[2] = &Guest{} // 桌1 = 2 人;桌2 = 0 人
	game.data = buildSheet()

	this.Equal(float64(1), this.num(game, "tableCount", exprs.NewText(">="), exprs.NewNum(2)))
	this.Equal(float64(1), this.num(game, "tableCount", exprs.NewText("=="), exprs.NewNum(0)))
	this.Equal(float64(1), this.num(game, "tableCount", exprs.NewText(">"), exprs.NewNum(0)))
	this.Equal(float64(2), this.num(game, "tableCount", exprs.NewText("<="), exprs.NewNum(2)))
	this.Equal(float64(0), this.num(game, "tableCount", exprs.NewText("<"), exprs.NewNum(0)))
	this.Equal(float64(2), this.num(game, "tableCount", exprs.NewText("!="), exprs.NewNum(1)))

	_, ok := game.Attr("tableCount", []exprs.Value{exprs.NewText("~="), exprs.NewNum(0)}) // 未知運算符
	this.False(ok)
	_, ok = game.Attr("tableCount", []exprs.Value{exprs.NewNum(1)}) // 參數數量不符
	this.False(ok)
	_, ok = game.Attr("tableCount", []exprs.Value{exprs.NewNum(1), exprs.NewNum(2)}) // 型別不符
	this.False(ok)
}

func (this *SuiteAttrRead) TestAttrReadGroupQuery() {
	game := NewGame(0, nil, nil, nil, nil)
	game.Hand = []*Card{{cardID: 101}, {cardID: 101}, {cardID: 102}}
	game.Deck = []*Card{{cardID: 101}}
	game.Drop = []*Card{{cardID: 102}, {cardID: 102}}
	game.Exile = []*Card{{cardID: 101}}
	game.drawTotal = Tally{count: map[int32]int32{1: 3, 2: 5}}
	game.dropTotal = Tally{count: map[int32]int32{1: 1}}
	game.playTotal = Tally{count: map[int32]int32{2: 2}}
	game.exileTotal = Tally{count: map[int32]int32{1: 4}}
	game.data = buildSheet()

	this.Equal(float64(2), this.num(game, "handSize", exprs.NewNum(1))) // 群組 1 = 卡 101 兩張
	this.Equal(float64(1), this.num(game, "handSize", exprs.NewNum(2))) // 群組 2 = 卡 102 一張
	this.Equal(float64(3), this.num(game, "handSize", exprs.NewNum(0))) // N==0 全量
	this.Equal(float64(1), this.num(game, "deckSize", exprs.NewNum(1)))
	this.Equal(float64(2), this.num(game, "dropSize", exprs.NewNum(0)))
	this.Equal(float64(1), this.num(game, "exileSize", exprs.NewNum(1)))

	this.Equal(float64(3), this.num(game, "drawTotal", exprs.NewNum(1)))
	this.Equal(float64(8), this.num(game, "drawTotal", exprs.NewNum(0))) // 全加總
	this.Equal(float64(1), this.num(game, "dropTotal", exprs.NewNum(1)))
	this.Equal(float64(2), this.num(game, "playTotal", exprs.NewNum(2)))
	this.Equal(float64(4), this.num(game, "exileTotal", exprs.NewNum(1)))

	_, ok := game.Attr("handSize", []exprs.Value{exprs.NewText("x")}) // 參數型別不符
	this.False(ok)
	_, ok = game.Attr("drawTotal", nil) // 缺參數
	this.False(ok)
}

// num 取全域屬性求值結果的數字;斷言命中且為數值。
func (this *SuiteAttrRead) num(game *Game, name string, arg ...exprs.Value) float64 {
	value, ok := game.Attr(name, arg)
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
		401: {ID: 401, Group: 5, RunOrder: 10, RunRound: 2}, // M10：RunRound>0 → Expire = 建立回合 + 1；RunOrder 10
		402: {ID: 402, Group: 5, RunOrder: 10, RunRound: 0}, // M10：RunRound 0 → Expire 0；與 401 同序、EffectID 402 > 401
		403: {ID: 403, Group: 6, RunOrder: 20, RunRound: 1}, // M10：RunOrder 20 最大 → 排序最先
	}
	data.Skill.Data = map[int32]*sheeter.Skill{
		301: {ID: 301, Group: 3, EffectID: []int32{401, 402}}, // 卡 103 的技能（群組 3）效果列表
	}
	data.Award.Data = map[int32]*sheeter.Award{
		1: {ID: 1, Group: 7, CardID: 101, Weight: 3}, // 群組 7：候選 101(w3) / 102(w1)
		2: {ID: 2, Group: 7, CardID: 102, Weight: 1},
		3: {ID: 3, Group: 8, CardID: 101, Weight: 0}, // 群組 8：權重 0 → roll no-op
		4: {ID: 4, Group: 9, CardID: 999, Weight: 1}, // 群組 9：抽中編號 999 無卡牌資料 → 跳過該張
	}
	data.Guest.Data = map[int32]*sheeter.Guest{
		501: {ID: 501, Score: 0, ScoreMax: 10, Morale: 5, MoraleMax: 8, Calm: 3, SateMax: 12, SateSeal: true},
	}
	return data
}
