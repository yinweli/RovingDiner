package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteHelp(t *testing.T) {
	suite.Run(t, new(SuiteHelp))
}

// SuiteHelp 驗證 help.go 的組件與無狀態輔助:免疫計數 / 門檻集合 / 編號列表 / 累積計數 / 參數取值 / 分組計數與比較 / 引用鎖定取值。
type SuiteHelp struct {
	suite.Suite
}

// TestNewImmune 驗證 NewImmune 建構空計數。
func (this *SuiteHelp) TestNewImmune() {
	immune := NewImmune()
	this.Equal(int32(0), immune.Get(1)) // 空表 → 0
}

// TestImmuneAdd 驗證 Add 計數遞增、零值自建表。
func (this *SuiteHelp) TestImmuneAdd() {
	immune := Immune{} // 零值可用

	immune.Add(5)
	immune.Add(5)
	this.Equal(int32(2), immune.Get(5))
}

// TestImmuneDel 驗證 Del 計數遞減、夾 ≥ 0、空表安全。
func (this *SuiteHelp) TestImmuneDel() {
	immune := NewImmune()
	immune.Add(5)

	immune.Del(5)
	this.Equal(int32(0), immune.Get(5))

	immune.Del(5) // 已 0 再減 → 夾 ≥ 0
	this.Equal(int32(0), immune.Get(5))

	zero := Immune{}
	zero.Del(9) // 零值（nil 表）→ 視為 0,不爆
	this.Equal(int32(0), zero.Get(9))
}

// TestImmuneGet 驗證 Get 讀計數、無鍵回 0。
func (this *SuiteHelp) TestImmuneGet() {
	immune := Immune{count: map[int32]int32{7: 2}}
	this.Equal(int32(2), immune.Get(7))
	this.Equal(int32(0), immune.Get(99)) // 無鍵 → 0
}

// TestNewHit 驗證 NewHit 建構空集合。
func (this *SuiteHelp) TestNewHit() {
	hit := NewHit()
	this.Equal(int32(0), hit.Count())
}

// TestHitAdd 驗證 Add 標記門檻、零值自建表、重複標記不增量。
func (this *SuiteHelp) TestHitAdd() {
	hit := Hit{} // 零值可用

	hit.Add(10)
	hit.Add(10) // 重複標記 → 集合語意不增量
	this.True(hit.IsHit(10))
	this.Equal(int32(1), hit.Count())
}

// TestHitIsHit 驗證 IsHit 已觸發判定。
func (this *SuiteHelp) TestHitIsHit() {
	hit := Hit{hit: map[int32]bool{10: true}}
	this.True(hit.IsHit(10))
	this.False(hit.IsHit(20)) // 未標記 → false

	zero := Hit{}
	this.False(zero.IsHit(10)) // 零值（nil 表）→ false,不爆
}

// TestHitCount 驗證 Count 已觸發數量。
func (this *SuiteHelp) TestHitCount() {
	hit := Hit{hit: map[int32]bool{10: true, 20: true}}
	this.Equal(int32(2), hit.Count())

	zero := Hit{}
	this.Equal(int32(0), zero.Count()) // 零值 → 0
}

// TestNewIDList 驗證 NewIDList 複製輸入建構（不共享底層）。
func (this *SuiteHelp) TestNewIDList() {
	source := []int32{1, 2}
	list := NewIDList(source...)
	source[0] = 9 // 改輸入不影響列表

	this.Equal([]int32{1, 2}, list.List())

	empty := NewIDList() // 無輸入 → 空列表
	this.Empty(empty.List())
}

// TestIDListAdd 驗證 Add 尾端加入（可變參數、零值可用、允許重複）。
func (this *SuiteHelp) TestIDListAdd() {
	list := IDList{} // 零值可用

	list.Add(5)
	list.Add(7, 5) // 批次 + 重複
	this.Equal([]int32{5, 7, 5}, list.List())
}

// TestIDListDelOne 驗證 DelOne 移除第一個命中、無命中不動。
func (this *SuiteHelp) TestIDListDelOne() {
	list := NewIDList(5, 7, 5)

	list.DelOne(5)
	this.Equal([]int32{7, 5}, list.List()) // 僅移第一個

	list.DelOne(9) // 無命中 → 不動
	this.Equal([]int32{7, 5}, list.List())
}

// TestIDListDelAll 驗證 DelAll 移除全部命中。
func (this *SuiteHelp) TestIDListDelAll() {
	list := NewIDList(5, 7, 5)

	list.DelAll(5)
	this.Equal([]int32{7}, list.List())

	list.DelAll(9) // 無命中 → 不動
	this.Equal([]int32{7}, list.List())
}

// TestIDListCount 驗證 Count 計數指定編號個數。
func (this *SuiteHelp) TestIDListCount() {
	list := NewIDList(5, 7, 5)
	this.Equal(int32(2), list.Count(5))
	this.Equal(int32(0), list.Count(9)) // 無命中 → 0
}

// TestIDListList 驗證 List 取底層編號列表（保持加入順序）。
func (this *SuiteHelp) TestIDListList() {
	list := NewIDList(3, 1, 2)
	this.Equal([]int32{3, 1, 2}, list.List())

	zero := IDList{}
	this.Empty(zero.List()) // 零值 → 空
}

// TestNewTally 驗證 NewTally 建構空累積計數。
func (this *SuiteHelp) TestNewTally() {
	total := NewTally()
	this.Equal(int32(0), total.Get(1))
	this.Equal(int32(0), total.Sum())
}

// TestTallyAdd 驗證 Add 對群組數量 +1;零值未建表時自建。
func (this *SuiteHelp) TestTallyAdd() {
	total := Tally{} // 零值可用

	total.Add(1)
	total.Add(1)
	total.Add(2)
	this.Equal(int32(2), total.Get(1))
	this.Equal(int32(1), total.Get(2))
}

// TestTallyGet 驗證 Get 讀群組數量、無鍵回 0。
func (this *SuiteHelp) TestTallyGet() {
	total := Tally{count: map[int32]int32{1: 3}}
	this.Equal(int32(3), total.Get(1))
	this.Equal(int32(0), total.Get(9)) // 無鍵 → 0
}

// TestTallySum 驗證 Sum 全群組加總;零值回 0。
func (this *SuiteHelp) TestTallySum() {
	total := Tally{count: map[int32]int32{1: 3, 2: 5}}
	this.Equal(int32(8), total.Sum())

	zero := Tally{}
	this.Equal(int32(0), zero.Sum()) // 零值 → 0
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
	card := []*Card{{cardID: 101}, {cardID: 101}, {cardID: 102}}

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
	total := Tally{count: map[int32]int32{1: 3, 2: 5}}

	result, ok := groupTotal(&total, []exprs.Value{exprs.NewNum(0)}) // N==0 全加總
	this.True(ok)
	this.Equal(float64(8), result.Num())

	result, ok = groupTotal(&total, []exprs.Value{exprs.NewNum(1)})
	this.True(ok)
	this.Equal(float64(3), result.Num())

	_, ok = groupTotal(&total, nil) // 參數不符 → 失敗
	this.False(ok)
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
	card := &Card{cost: NewValue(0, 4)}
	pick := func(c *Card) int32 { return c.GetCost().GetLock() }

	result, ok := cardLock(NewRefCard(card), pick)
	this.True(ok)
	this.Equal(float64(4), result.Num())

	_, ok = cardLock(NewRefGuest(&Guest{}), pick) // 非卡牌引用 → 失敗
	this.False(ok)
}

func (this *SuiteHelp) TestGuestLock() {
	guest := &Guest{calm: NewValue(0, 7)}
	pick := func(g *Guest) int32 { return g.GetCalm().GetLock() }

	result, ok := guestLock(NewRefGuest(guest), pick)
	this.True(ok)
	this.Equal(float64(7), result.Num())

	_, ok = guestLock(NewRefCard(&Card{}), pick) // 非顧客引用 → 失敗
	this.False(ok)
}
