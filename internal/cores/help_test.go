package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteHelp(t *testing.T) {
	suite.Run(t, new(SuiteHelp))
}

// SuiteHelp 驗證 help.go 的無狀態輔助:參數取值 / 分組計數與比較 / 引用鎖定取值 / 容器掃描與歸屬。
type SuiteHelp struct {
	suite.Suite
}

func (this *SuiteHelp) TestOneInt() {
	num, ok := oneInt([]exprs.Value{exprs.NewNum(5)})
	this.True(ok)
	this.Equal(int32(5), num)

	num, ok = oneInt([]exprs.Value{exprs.NewNum(3.9)}) // 浮點截斷為整數
	this.True(ok)
	this.Equal(int32(3), num)

	_, ok = oneInt(nil) // 數量不符(0 個)
	this.False(ok)

	_, ok = oneInt([]exprs.Value{exprs.NewNum(1), exprs.NewNum(2)}) // 數量不符(2 個)
	this.False(ok)

	_, ok = oneInt([]exprs.Value{exprs.NewText("x")}) // 型別不符
	this.False(ok)
}

func (this *SuiteHelp) TestTwoInt() {
	a, b, ok := twoInt([]exprs.Value{exprs.NewNum(2), exprs.NewNum(5)})
	this.True(ok)
	this.Equal(int32(2), a)
	this.Equal(int32(5), b)

	_, _, ok = twoInt([]exprs.Value{exprs.NewNum(1)}) // 數量不符(1 個)
	this.False(ok)

	_, _, ok = twoInt([]exprs.Value{exprs.NewText("x"), exprs.NewNum(1)}) // 第一參型別不符
	this.False(ok)

	_, _, ok = twoInt([]exprs.Value{exprs.NewNum(1), exprs.NewText("x")}) // 第二參型別不符
	this.False(ok)
}

func (this *SuiteHelp) TestArgInt() {
	n, ok := argInt([]exprs.Value{exprs.NewNum(3.9)}) // 浮點截斷
	this.True(ok)
	this.Equal(int32(3), n)

	_, ok = argInt(nil) // 缺漏
	this.False(ok)

	_, ok = argInt([]exprs.Value{exprs.NewText("x")}) // 非數值
	this.False(ok)
}

func (this *SuiteHelp) TestArgNum() {
	n, ok := argNum([]exprs.Value{exprs.NewNum(2.5)})
	this.True(ok)
	this.Equal(float64(2.5), n)

	_, ok = argNum(nil) // 缺漏
	this.False(ok)

	_, ok = argNum([]exprs.Value{exprs.NewBool(true)}) // 非數值
	this.False(ok)
}

func (this *SuiteHelp) TestArgBool() {
	this.True(argBool([]exprs.Value{exprs.NewBool(true)}))
	this.False(argBool([]exprs.Value{exprs.NewBool(false)}))
	this.False(argBool(nil))                            // 缺漏
	this.False(argBool([]exprs.Value{exprs.NewNum(1)})) // 非布林
	// 讀第 k 個參數由呼叫端傳 arg[k:]:此處讀 index 1(M9.3 copy/clone 的洗牌位於 index 1)
	this.True(argBool([]exprs.Value{exprs.NewNum(1), exprs.NewBool(true)}[1:]))
}

func (this *SuiteHelp) TestIntList() {
	this.Equal([]int32{1, 2}, intList([]exprs.Value{exprs.NewNum(1), exprs.NewText("x"), exprs.NewNum(2)})) // 略過非數值
	this.Nil(intList(nil))
}

func (this *SuiteHelp) TestArgTail() {
	full := []exprs.Value{exprs.NewNum(1), exprs.NewNum(2)}
	this.Len(argTail(full, 1), 1) // 自 index 1
	this.Nil(argTail(full, 2))    // from == len → nil
	this.Nil(argTail(full, 5))    // from > len → nil
}

func (this *SuiteHelp) TestGroupSize() {
	data := buildSheet()
	card := []*Card{{CardID: 101}, {CardID: 101}, {CardID: 102}}

	result, ok := groupSize(card, data, []exprs.Value{exprs.NewNum(0)}) // N==0 全量
	this.True(ok)
	this.Equal(float64(3), result.Num())

	result, ok = groupSize(card, data, []exprs.Value{exprs.NewNum(1)}) // 群組 1 = 卡 101 兩張
	this.True(ok)
	this.Equal(float64(2), result.Num())

	_, ok = groupSize(card, data, nil) // 參數不符 → 失敗
	this.False(ok)
}

func (this *SuiteHelp) TestGroupTotal() {
	total := map[int32]int32{1: 3, 2: 5}

	result, ok := groupTotal(total, []exprs.Value{exprs.NewNum(0)}) // N==0 全加總
	this.True(ok)
	this.Equal(float64(8), result.Num())

	result, ok = groupTotal(total, []exprs.Value{exprs.NewNum(1)})
	this.True(ok)
	this.Equal(float64(3), result.Num())

	_, ok = groupTotal(total, nil) // 參數不符 → 失敗
	this.False(ok)
}

func (this *SuiteHelp) TestCountByGroup() {
	data := buildSheet()
	card := []*Card{{CardID: 101}, {CardID: 101}, {CardID: 102}, {CardID: 999}} // 999 無靜態資料

	this.Equal(int32(4), countByGroup(card, 0, data)) // group==0 全量(不過濾)
	this.Equal(int32(2), countByGroup(card, 1, data)) // 群組 1 = 卡 101 兩張
	this.Equal(int32(1), countByGroup(card, 2, data)) // 群組 2 = 卡 102 一張
	this.Equal(int32(0), countByGroup(card, 9, data)) // 無此群組(含資料缺失卡牌)
}

func (this *SuiteHelp) TestTotalByGroup() {
	total := map[int32]int32{1: 3, 2: 5}

	this.Equal(int32(8), totalByGroup(total, 0)) // n==0 全加總
	this.Equal(int32(3), totalByGroup(total, 1))
	this.Equal(int32(0), totalByGroup(total, 9)) // 無此鍵 → 0
}

func (this *SuiteHelp) TestCompareOp() {
	for _, itor := range []struct {
		op   string
		a, b int32
		want bool
	}{
		{"<", 1, 2, true}, {"<", 2, 1, false},
		{">", 2, 1, true}, {">", 1, 2, false},
		{"<=", 2, 2, true}, {"<=", 3, 2, false},
		{">=", 2, 2, true}, {">=", 1, 2, false},
		{"==", 2, 2, true}, {"==", 1, 2, false},
		{"!=", 1, 2, true}, {"!=", 2, 2, false},
	} {
		result, ok := compareOp(itor.op, itor.a, itor.b)
		this.True(ok, itor.op)
		this.Equal(itor.want, result, itor.op)
	} // for

	_, ok := compareOp("~=", 1, 1) // 未知運算符 → 失敗
	this.False(ok)
}

func (this *SuiteHelp) TestCardLock() {
	card := &Card{Cost: Value{Lock: 4}}
	pick := func(c *Card) int32 { return c.Cost.Lock }

	result, ok := cardLock(cardRef{card: card}, pick)
	this.True(ok)
	this.Equal(float64(4), result.Num())

	_, ok = cardLock(guestRef{guest: &Guest{}}, pick) // 非卡牌引用 → 失敗
	this.False(ok)
}

func (this *SuiteHelp) TestGuestLock() {
	guest := &Guest{Calm: Value{Lock: 7}}
	pick := func(g *Guest) int32 { return g.Calm.Lock }

	result, ok := guestLock(guestRef{guest: guest}, pick)
	this.True(ok)
	this.Equal(float64(7), result.Num())

	_, ok = guestLock(cardRef{card: &Card{}}, pick) // 非顧客引用 → 失敗
	this.False(ok)
}

func (this *SuiteHelp) TestInContainer() {
	card := &Card{InstanceID: 5}
	container := []*Card{{InstanceID: 1}, {InstanceID: 5}}

	this.True(inContainer(card, container))                  // 同實例編號(不同指標)
	this.False(inContainer(&Card{InstanceID: 9}, container)) // 不在容器
	this.False(inContainer(card, nil))                       // 空容器
}

func (this *SuiteHelp) TestOccupiedAmong() {
	runtime := NewRuntime(0)
	runtime.Seat[1] = &Guest{}
	runtime.Seat[3] = &Guest{}
	eng := &Engine{runtime: runtime}

	this.Equal(int32(2), occupiedAmong(eng, []int32{1, 2, 3})) // 座 1、3 占用,座 2 空
	this.Equal(int32(0), occupiedAmong(eng, []int32{2, 4}))    // 皆空
	this.Equal(int32(0), occupiedAmong(eng, nil))              // 空列表
}

func (this *SuiteHelp) TestEffectSelfIs() {
	card := &Card{InstanceID: 1}
	guest := &Guest{InstanceID: 2}

	cardEffect := &Effect{Self: Self{Card: card}}
	this.True(effectSelfIs(cardEffect, cardRef{card: card}))                  // 卡牌 self 命中
	this.False(effectSelfIs(cardEffect, cardRef{card: &Card{InstanceID: 9}})) // 卡牌 self 不同實例
	this.False(effectSelfIs(cardEffect, guestRef{guest: guest}))              // 卡牌 self 對顧客引用

	guestEffect := &Effect{Self: Self{Guest: guest}}
	this.True(effectSelfIs(guestEffect, guestRef{guest: guest}))                  // 顧客 self 命中
	this.False(effectSelfIs(guestEffect, guestRef{guest: &Guest{InstanceID: 9}})) // 顧客 self 不同實例
	this.False(effectSelfIs(guestEffect, cardRef{card: card}))                    // 顧客 self 對卡牌引用

	this.False(effectSelfIs(cardEffect, fakeRef{})) // 既非卡牌也非顧客引用 → false
}

func (this *SuiteHelp) TestCardSkillGroup() {
	eng := &Engine{data: buildSheet()}
	this.Equal(int32(3), cardSkillGroup(eng, 103)) // 卡 103 → 技能 301 → 群組 3
	this.Equal(int32(0), cardSkillGroup(eng, 101)) // 卡 101 無技能（SkillID 0）→ 0
	this.Equal(int32(0), cardSkillGroup(eng, 999)) // 卡牌資料不存在 → 0
}

// === 測試輔助（置尾） ===

// fakeRef 是既非卡牌也非顧客的第三方引用,用以驗證 effectSelfIs 對未知引用型別回 false。
type fakeRef struct{}

// Same 永不相等;僅為滿足 exprs.Ref 介面而存在。
func (this fakeRef) Same(other exprs.Ref) bool {
	return false
}
