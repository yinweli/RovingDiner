package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteCursor(t *testing.T) {
	suite.Run(t, new(SuiteCursor))
}

// SuiteCursor 驗證游標工具(cursor.go): 夾界 / 單列移動 / cursor-follow 窗格起點。
type SuiteCursor struct {
	suite.Suite
}

// TestClampIndex 驗證夾界: 範圍內原樣、越界拉回、空列表回 0。
func (this *SuiteCursor) TestClampIndex() {
	this.Equal(2, clampIndex(2, 5))
	this.Equal(0, clampIndex(-1, 5))
	this.Equal(4, clampIndex(9, 5))
	this.Equal(0, clampIndex(3, 0)) // 空列表
}

// TestMoveIndex 驗證單列移動: 左右增減、端點夾住不迴繞、上下不動作。
func (this *SuiteCursor) TestMoveIndex() {
	this.Equal(2, moveIndex(1, "right", 5))
	this.Equal(0, moveIndex(1, "left", 5))
	this.Equal(0, moveIndex(0, "left", 5))  // 左端夾住
	this.Equal(4, moveIndex(4, "right", 5)) // 右端夾住
	this.Equal(1, moveIndex(1, "up", 5))    // 上下不動作
	this.Equal(2, moveIndex(9, "left", 4))  // 越界先夾再移
}

// TestStripFirst 驗證窗格起點: 游標可視即從 0 起、超出最小右捲(扣左緣記號 2 格與游標非末項的右緣保留)、
// 單 cell 過寬回游標自身(cursor 由呼叫端先夾界)。
func (this *SuiteCursor) TestStripFirst() {
	width := []int{10, 10, 10, 10}
	this.Equal(0, stripFirst(width, 2, 60, 3))     // 全部放得下
	this.Equal(0, stripFirst(width, 2, 24, 1))     // 游標 1: 0..1 共 22 + 右緣保留 2 = 24 剛好
	this.Equal(2, stripFirst(width, 2, 24, 2))     // 游標 2: 起點 1 仍超(左緣 2 + 22 + 右緣 2 = 26)→ 起點 2
	this.Equal(3, stripFirst(width, 2, 12, 3))     // 窄到只容一格時退到游標自身, 末項不保右緣
	this.Equal(0, stripFirst([]int{30}, 2, 12, 0)) // 單 cell 過寬: 起點 = 游標自身(交給右緣截斷)
	this.Equal(0, stripFirst(nil, 2, 12, 0))       // 空列表
}
