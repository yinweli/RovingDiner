package games

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteValidate(t *testing.T) {
	suite.Run(t, new(SuiteValidate))
}

// SuiteValidate 驗證命令語言詞彙表的 Validate / ValidateExpr(validate.go)逐名查全域表。
// 操作命令查 verb / 命令對象 / 命令對象 [...] 參數數量(M28); 屬性修改命令查左值可寫性(M7 生效);
// 內嵌算術式與運算式詞彙查名(M28 R4A)。
type SuiteValidate struct {
	suite.Suite
}

func (this *SuiteValidate) TestValidateOperate() {
	// 已登錄命令 + 合法命令對象 → 通過(M9 詞條到位後生效)
	this.NoError(Validate(this.parse("handAdd(none, 10031, 1)")))
	// 未登錄命令 → 報錯
	this.Error(Validate(this.parse("noSuchCommand(none)")))
	// 未登錄命令對象 → 報錯
	this.Error(Validate(this.parse("handAdd(noSuchSelector)")))
}

// TestValidateArity 驗證命令對象 [...] 參數數量校驗(M28; 依【二十四、命令對象清單 | 參數規則】裸寫 / 數量不符即報錯)。
func (this *SuiteValidate) TestValidateArity() {
	// 數量相符 → 通過(參數為算術式, 數量看 token 個數不看值)
	this.NoError(Validate(this.parse("guestExit(guestPick[1], 1, 1)")))
	this.NoError(Validate(this.parse("handToDrop(handAll[0])")))
	this.NoError(Validate(this.parse("handToDrop(handPick[1 + 1, 0])")))
	// 裸寫缺參數 → 報錯
	this.Error(Validate(this.parse("guestExit(guestPick, 1, 1)")))
	this.Error(Validate(this.parse("handToDrop(handAll)")))
	// 數量不足 / 過多 → 報錯
	this.Error(Validate(this.parse("handToDrop(handPick[2])")))
	this.Error(Validate(this.parse("handAdd(none[1], 10031, 1)")))
}

func (this *SuiteValidate) TestValidateAssign() {
	// 全域可寫屬性 → 通過
	this.NoError(Validate(this.parse("morale = 10")))
	// 全域唯讀屬性 → 報錯(nextPhase 僅 phaseJump 可設)
	this.Error(Validate(this.parse("nextPhase = '玩家行動'")))
	// 全域未知屬性 → 報錯
	this.Error(Validate(this.parse("nope = 1")))

	// 引用可寫屬性 → 通過
	this.NoError(Validate(this.parse("self.calm += 1")))
	this.NoError(Validate(this.parse("drawLast.cost = 1")))
	// 引用唯讀屬性 → 報錯(cardID 唯讀)
	this.Error(Validate(this.parse("drawLast.cardID = 1")))
	// 引用基底非物件引用 → 報錯(M28; morale 為數值屬性、不可作 <基底>.<屬性> 基底)
	this.Error(Validate(this.parse("morale.cost = 1")))
	// 引用基底未登錄 → 報錯(M28; 先前僅查屬性可寫性會漏過)
	this.Error(Validate(this.parse("nope.cost = 1")))
}

// TestValidateEmbed 驗證內嵌算術式的詞彙走訪(M28 R4A): 右值 / 操作命令參數 / 命令對象參數的
// 未知識別子皆報錯且位置回算至命令座標; 鎖定 / 解鎖無右值不走訪。
func (this *SuiteValidate) TestValidateEmbed() {
	this.NoError(Validate(this.parse("morale @"))) // @ 無右值
	this.NoError(Validate(this.parse("morale += min(energy, 2)")))

	err := Validate(this.parse("morale = morale2 + 1")) // 右值未知識別子
	this.Require().Error(err)
	syntaxError := &exprs.SyntaxError{}
	this.Require().True(errors.As(err, &syntaxError))
	this.Equal(9, syntaxError.Pos) // 回算至命令座標(morale2 起點)

	this.Error(Validate(this.parse("handAdd(none, 10031, morale2)")))    // 操作命令參數
	this.Error(Validate(this.parse("guestExit(guestPick[nope], 1, 1)"))) // 命令對象參數
}

// TestValidateExpr 驗證運算式詞彙走訪(M28 R4A): 識別子 / 查詢函式 / 內建函式 / 引用屬性的
// 名稱與參數數量逐項查讀側詞彙表; 無括號識別子與零參數呼叫等價。
func (this *SuiteValidate) TestValidateExpr() {
	// 合法: 屬性 + 查詢函式 + 內建函式 + 引用屬性 / 引用查詢函式
	this.NoError(ValidateExpr(this.parseExpr("morale + tableGuest(1) - min(1, 2)")))
	this.NoError(ValidateExpr(this.parseExpr("self.calm > 0 AND drawLast.cost <= tableCount('>=', 2)")))
	this.NoError(ValidateExpr(this.parseExpr("self.effectGroup(5) + morale()"))) // 零參數呼叫 = 識別子
	// 未知識別子(帶位置)
	err := ValidateExpr(this.parseExpr("1 + morale2"))
	this.Require().Error(err)
	syntaxError := &exprs.SyntaxError{}
	this.Require().True(errors.As(err, &syntaxError))
	this.Equal(4, syntaxError.Pos)
	// 查詢函式裸寫 / 數量不符
	this.Error(ValidateExpr(this.parseExpr("tableGuest")))
	this.Error(ValidateExpr(this.parseExpr("tableGuest(1, 2)")))
	// 內建函式低於最少參數
	this.Error(ValidateExpr(this.parseExpr("min(1)")))
	// 引用基底非引用 / 未知, 引用屬性未知 / 數量不符
	this.Error(ValidateExpr(this.parseExpr("morale.calm")))
	this.Error(ValidateExpr(this.parseExpr("nope.calm")))
	this.Error(ValidateExpr(this.parseExpr("self.nope")))
	this.Error(ValidateExpr(this.parseExpr("self.effectGroup")))
	this.Error(ValidateExpr(this.parseExpr("self.calm(1)")))
}

// parse 解析命令來源為 Command; 解析失敗即測試失敗(供 Validate 斷言聚焦於語意檢查)。
func (this *SuiteValidate) parse(source string) Command {
	command, err := Parse(source)
	this.Require().NoError(err)
	return command
}

// parseExpr 解析運算式來源為 *exprs.Expr; 解析失敗即測試失敗(供 ValidateExpr 斷言聚焦於詞彙檢查)。
func (this *SuiteValidate) parseExpr(source string) *exprs.Expr {
	expr, err := exprs.Parse(source)
	this.Require().NoError(err)
	return expr
}
