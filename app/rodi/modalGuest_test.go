package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteModalGuest(t *testing.T) {
	suite.Run(t, new(SuiteModalGuest))
}

// SuiteModalGuest 驗證顧客 modal(modalGuest.go): 身分配對對齊 / 門檻技能與觸發標記 / 免疫 / 效果表 /
// 位置顯示。
type SuiteModalGuest struct {
	suite.Suite
}

// TestModalGuestBody 驗證內容行(全量釘字串): 標題嵌含實例段識別碼、飽食門檻升序 / 耐心門檻降序、
// 已觸發標 v、免疫與效果列。
func (this *SuiteModalGuest) TestModalGuestBody() {
	game := testGame()
	guest := cores.NewGuest(game, 501) // #1; 資料: 滿意 4/10 士氣 5/8 耐心 3 飽食上限 6 SateSeal 鎖 1
	game.Seat.Place(1, guest)
	guest.GetSateHit().Add(5)
	guest.GetEffectImmune().Add(2)
	guest.GetSkillImmune().Add(6)
	effect := cores.NewEffect(game, 401, cores.NewRefGuest(guest), 2) // #2
	effect.SetExpire(4)
	game.Effect.Push(effect)
	game.Effect.Push(cores.NewEffect(game, 402, cores.Ref{}, 1)) // self 非本顧客, 不入列

	title, row := modalGuest{guest: guest}.Body(game)
	this.Equal("顧客檢視 501@老饕#1", title)
	this.Equal([]string{
		"滿意值 4/10   士氣值 5/8",
		"飽食值 0/6    封印飽食技能[鎖1]",
		"耐心值 3      封印耐心技能[鎖0]",
		"座位 1 (桌1)  凍結起始回合 0",
		"+- 飽食技能",
		"門檻值  技能      觸發",
		"5       301@開朗  v",
		"7       301@開朗",
		"+- 耐心技能",
		"門檻值  技能      觸發",
		"9       301@開朗",
		"7       301@開朗",
		"+- 免疫",
		"類型  群組  計數",
		"效果  2     1",
		"技能  6     1",
		"+- 效果",
		"效果        結束回合  當前層數",
		"401@加耐#2  4         2",
	}, row)
}

// TestSeatText 驗證位置顯示: 在座顯 座位 N (桌M)、離座依容器標 排隊 / 遊蕩 / 卡牌化、
// 都不在(防禦)即離場。
func (this *SuiteModalGuest) TestSeatText() {
	game := testGame()
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(3, guest) // 座 3 = 桌 2
	this.Equal("座位 3 (桌2)", seatText(game, guest))

	wait := cores.NewGuest(game, 501)
	game.Wait.Insert(wait)
	this.Equal("座位 排隊", seatText(game, wait))

	roam := cores.NewGuest(game, 501)
	game.Roam.Push(roam)
	this.Equal("座位 遊蕩", seatText(game, roam))

	cardify := cores.NewGuest(game, 501)
	game.Cardify.Push(cardify)
	this.Equal("座位 卡牌化", seatText(game, cardify))

	this.Equal("座位 離場", seatText(game, cores.NewGuest(game, 501))) // 防禦: 不在任一容器
}
