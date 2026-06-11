package rodi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteComponent(t *testing.T) {
	suite.Run(t, new(SuiteComponent))
}

// SuiteComponent 驗證組件框架(component.go): 父層組合依掛載順序堆疊、寬度預算逐組件下發。
type SuiteComponent struct {
	suite.Suite
}

// TestComposeView 驗證父層組合: 掛載順序 = 堆疊順序、每組件收到同一寬度預算; 空組件列表回空字串。
func (this *SuiteComponent) TestComposeView() {
	world := newMirror()
	this.Equal("a:80\nb:80", composeView(world, 80, []component{fakeComponent{text: "a"}, fakeComponent{text: "b"}}))
	this.Equal("", composeView(world, 80, nil))
}

// fakeComponent 測試替身: 渲染自身文字與收到的寬度預算, 供 TestComposeView 驗證下發。
type fakeComponent struct {
	text string
}

func (this fakeComponent) View(world *mirror, width int) string {
	return fmt.Sprintf("%v:%v", this.text, width)
}
