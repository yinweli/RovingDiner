package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuitePhaseGameStart(t *testing.T) {
	suite.Run(t, new(SuitePhaseGameStart))
}

// SuitePhaseGameStart 驗證營業開始階段(phaseGameStart.go): 設定載入 / 歸零設置 / 累積表清空 / 開局建置(buildStage)/
// 前置技能 / gameStart 觸發。
type SuitePhaseGameStart struct {
	suite.Suite
}

func (this *SuitePhaseGameStart) TestPhaseGameStart() {
	count := 0
	fired := 0
	data := tester.BuildData()
	game := cores.NewGame(0, 601, data, tester.FakeOperator{}, tester.FakeRander{}, nil) // 關卡 601:前置技能 301
	Register(game)
	data.SetEffect(401, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) { count++ }})
	data.SetEffect(402, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) { count++ }})
	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerGameStart, Trigger: func(game *cores.Game) { fired++ }})
	game.Effect.Push(cores.NewEffect(game, 901, cores.Ref{}, 1))

	game.GetRound().Set(7) // 佈置髒狀態 → 驗證歸零
	game.Settling = true
	game.GetDrawTotal().Add(1)
	game.GetDropTotal().Add(1)
	game.GetPlayTotal().Add(1)
	game.GetExileTotal().Add(1)

	this.Equal(cores.PhaseRoundStart, phaseGameStart(game))
	this.Equal(int32(12), game.GetRoundMax().GetValue()) // 設定六鍵載入
	this.Equal(int32(30), game.GetMorale().GetValue())
	this.Equal(int32(50), game.GetMoraleMax().GetValue())
	this.Equal(int32(3), game.GetEnergyMax().GetValue())
	this.Equal(int32(10), game.GetHandMax().GetValue())
	this.Equal(int32(5), game.GetDrawMax().GetValue())

	this.Equal(int32(0), game.GetRound().GetValue()) // 回合 / 結算旗標歸零
	this.False(game.Settling)
	this.Equal(int32(3), game.GetEnergy().GetValue()) // 點數補滿
	this.Equal(int32(0), game.GetDrawTotal().Sum())   // 四累積表清空
	this.Equal(int32(0), game.GetDropTotal().Sum())
	this.Equal(int32(0), game.GetPlayTotal().Sum())
	this.Equal(int32(0), game.GetExileTotal().Sum())

	this.Len(game.Hand, 1) // 開局建置(關卡 601)
	this.Len(game.Deck, 2)
	this.Len(game.Wait, 2)
	this.Equal([]int32{301}, game.PrefixSkill)
	this.Equal(2, count) // 前置技能 301 → 效果 401 / 402 各啟動一次
	this.Equal(1, fired) // 營業開始觸發
}

// TestPhaseGameStartEmit 驗證營業開始的發射接線: 設定六鍵 + 點數補滿逐筆屬性行(流程寫入白名單; M20 拍板);
// 開局建置逐實例搬移行(發射序 = 容器序; M21 拍板); 前置技能發範圍標題(操作元 = 技能)。
func (this *SuitePhaseGameStart) TestPhaseGameStartEmit() {
	record := &tester.RecordPresenter{}
	game := cores.NewGame(0, 601, tester.BuildData(), tester.FakeOperator{}, tester.FakeRander{}, record)
	Register(game)
	phaseGameStart(game)

	expect := []string{
		// 開局快照(六鍵 + 點數補滿)位居行流最前端、序同 loadSetting
		"$ 回合上限 = 12 >> 12",
		"$ 餐廳士氣值 = 30 >> 30",
		"$ 餐廳士氣值上限 = 50 >> 50",
		"$ 出牌點數上限 = 3 >> 3",
		"$ 手牌張數上限 = 10 >> 10",
		"$ 補牌張數上限 = 5 >> 5",
		"$ 出牌點數 = 3 >> 3",
		// 開局建置: 逐實例搬移行依容器序(關卡 601; 原設置靜默作廢), 行即帶實例編號供消費端對照
		"$ 101@#1 >> 手牌",
		"$ 101@#2 >> 抽牌堆",
		"$ 102@#3 >> 抽牌堆",
		"$ 102@#4 >> 棄牌堆",
		"$ 101@#5 >> 流放堆",
		"$ 501@#6 >> 排隊",
		"$ 501@#7 >> 排隊",
		// 前置技能範圍標題(操作元 = 技能 301; 直呼站函式未踏站 → 前綴階段顯 -)
		"[R0 -] 前置技能",
		"* 301@",
	}
	flat := record.Flat()
	this.Require().GreaterOrEqual(len(flat), len(expect))
	this.Equal(expect, flat[:len(expect)])
}

// TestBuildStage 驗證開局建置: 五容器順序語意(第 1 個 = 頂端 / 隊首)、設置不觸發時機、壞引用逐筆跳過、查無關卡空盤面。
func (this *SuitePhaseGameStart) TestBuildStage() {
	data := tester.BuildData()
	game := cores.NewGame(0, 601, data, tester.FakeOperator{}, tester.FakeRander{}, nil)

	buildStage(game)
	this.Require().Len(game.Hand, 1) // 手牌列表
	this.Equal(int32(101), game.Hand[0].GetCardID())
	this.Require().Len(game.Deck, 2) // 第 1 個 = 牌堆頂
	this.Equal(int32(101), game.Deck[0].GetCardID())
	this.Equal(int32(102), game.Deck[1].GetCardID())
	this.Require().Len(game.Drop, 1)
	this.Equal(int32(102), game.Drop[0].GetCardID())
	this.Require().Len(game.Exile, 1)
	this.Equal(int32(101), game.Exile[0].GetCardID())
	this.Require().Len(game.Wait, 2) // 第 1 個 = 隊首
	this.Equal(int32(501), game.Wait[0].GetGuestID())
	this.Equal([]int32{301}, game.PrefixSkill)

	this.Equal(int32(0), game.GetDrawCount()) // 設置非命令: 不觸發 / 不動事件屬性
	this.Nil(game.GetDrawLast())
	this.Nil(game.GetDropLast())
	this.Nil(game.GetExileLast())

	bad := cores.NewGame(0, 602, data, tester.FakeOperator{}, tester.FakeRander{}, nil) // 全列壞引用 → 逐筆跳過
	buildStage(bad)
	this.Empty(bad.Hand)
	this.Empty(bad.Deck)
	this.Empty(bad.Wait)
	this.Equal([]int32{999}, bad.PrefixSkill) // 技能編號原樣, 查無由啟動端防禦

	none := cores.NewGame(0, 0, data, tester.FakeOperator{}, tester.FakeRander{}, nil) // 查無關卡 → 空盤面照走
	buildStage(none)
	this.Empty(none.Hand)
	this.Nil(none.PrefixSkill)
}

// TestSettingNum 驗證 settingNum 讀數字設定值; 缺鍵 / 空值 / 非數字回 0。
func (this *SuitePhaseGameStart) TestSettingNum() {
	game := newGame()
	game.GetSheet().Setting.Data["Empty"] = &sheeter.Setting{ID: "Empty"}
	game.GetSheet().Setting.Data["Bad"] = &sheeter.Setting{ID: "Bad", Value: []string{"x"}}

	this.Equal(float64(12), settingNum(game, "RoundMax"))
	this.Equal(float64(0), settingNum(game, "nope"))  // 缺鍵 → 0
	this.Equal(float64(0), settingNum(game, "Empty")) // 空值 → 0
	this.Equal(float64(0), settingNum(game, "Bad"))   // 非數字 → 0
}
