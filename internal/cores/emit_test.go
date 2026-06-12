package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteEmit(t *testing.T) {
	suite.Run(t, new(SuiteEmit))
}

// SuiteEmit 驗證日誌行發射台(emit.go): 組行模板與歸因語意(平面: 最後效果行勝出、範圍標題重置)。
type SuiteEmit struct {
	suite.Suite
}

// TestEmitTitle 驗證範圍標題: [R{回合} {階段}] 前綴 + 操作元 * 行一拍多行, 並重置歸因。
func (this *SuiteEmit) TestEmitTitle() {
	record := &fakePresenter{}
	game := NewGame(0, 0, nil, nil, nil, record)
	game.GetRound().Set(3)
	game.SetPhase(PhaseRoundStart)
	EmitTitle(game, "執行結算")
	this.Require().Len(record.line, 1)
	this.Equal([]string{"[R3 回合開始] 執行結算"}, record.line[0])

	EmitTitle(game, "玩家出牌", "101@#1", "301@") // 操作元識別碼由呼叫端組畢
	this.Equal([]string{"[R3 回合開始] 玩家出牌", "* 101@#1", "* 301@"}, record.line[1])

	EmitEffect(game, 401, 9, Ref{}, EffectStageImmed) // 先立歸因
	EmitTitle(game, "時機:回合開始")                        // 標題重置歸因
	EmitBody(game, "x")
	this.Equal([]string{"$ x"}, record.line[4])
}

// TestEmitEffect 驗證效果頭三行: 效果識別碼 / 對象(空物件顯 空)/ 階段, 並把歸因設為本效果 self。
func (this *SuiteEmit) TestEmitEffect() {
	record := &fakePresenter{}
	game := NewGame(0, 0, NewData(buildSheet(), nil), nil, nil, record)
	EmitEffect(game, 401, 9, NewRefGuest(&Guest{instanceID: 6, guestID: 501}), EffectStageJoin)
	this.Require().Len(record.line, 1)
	this.Equal([]string{"- 401@#9", "  501@#6", "  加入"}, record.line[0])

	EmitBody(game, "y") // 歸因已設 → 命令行縮排掛欄 2
	this.Equal([]string{"  y"}, record.line[1])

	EmitEffect(game, 403, NoneID, Ref{}, EffectStageImmed) // 空物件 self 顯 空; 立即類無實例段
	this.Equal([]string{"- 403@", "  空", "  立即"}, record.line[2])
}

// TestEmitProperty 驗證屬性行: 全域全名 / 無歸因識別碼開頭 / 歸因 self 省略 / 異對象識別碼開頭 /
// 免疫詞條查詢函式形 / 鎖定無算術式。
func (this *SuiteEmit) TestEmitProperty() {
	record := &fakePresenter{}
	game := NewGame(0, 0, NewData(buildSheet(), nil), nil, nil, record)

	EmitProperty(game, 0, NoneID, "morale", AssignAdd, 2, 7) // 全域屬性: 全名、無對象識別碼
	this.Equal([]string{"$ 餐廳士氣值 += 2 >> 7"}, record.line[0])

	EmitProperty(game, 501, 6, "calm", AssignSub, 1, 2) // 實例屬性無歸因: 識別碼開頭
	this.Equal([]string{"$ 501@#6 耐心值 -= 1 >> 2"}, record.line[1])

	EmitProperty(game, 501, 6, AttrEffectImmune, AssignAdd, 5, 1) // 免疫詞條: 名稱(群組)、運算值固定 1
	this.Equal([]string{"$ 501@#6 效果免疫群組(5) += 1 >> 1"}, record.line[2])

	EmitProperty(game, 0, NoneID, "energy", AssignLock, 0, 1) // 鎖定: 無算術式、結果為鎖定計數
	this.Equal([]string{"$ 出牌點數 @ >> 1"}, record.line[3])

	EmitEffect(game, 401, 9, NewRefGuest(&Guest{instanceID: 6, guestID: 501}), EffectStageStart)
	EmitProperty(game, 501, 6, "calm", AssignAdd, 3, 5) // 歸因 self 命中 → 屬性開頭
	this.Equal([]string{"  耐心值 += 3 >> 5"}, record.line[5])

	EmitProperty(game, 501, 7, "calm", AssignAdd, 3, 5) // 歸因存在但對象不同 → 識別碼開頭
	this.Equal([]string{"  501@#7 耐心值 += 3 >> 5"}, record.line[6])
}

// TestEmitBody 驗證命令行前綴: 無歸因 $ 流程直屬、有歸因縮排掛欄 2。
func (this *SuiteEmit) TestEmitBody() {
	record := &fakePresenter{}
	game := NewGame(0, 0, nil, nil, nil, record)
	EmitBody(game, "x")
	this.Equal([]string{"$ x"}, record.line[0])

	game.logEffect = true
	EmitBody(game, "y")
	this.Equal([]string{"  y"}, record.line[1])
}
