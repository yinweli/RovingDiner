package exprs

import (
	"fmt"
)

// newError 建立帶位置的語法錯誤;pos 為來源字串的 rune 索引(0 起算)。
// lexer 與未來的 parser / 命令解析器共用此建構式,使各文法的錯誤格式一致。
func newError(pos int, msg string) error {
	return &SyntaxError{Pos: pos, Msg: msg}
}

// SyntaxError 是運算式語法錯誤,帶出錯位置,供企劃輔助工具標示位置
// (詳見【營業實作規格書 | 附錄：企劃驗證器】)。
// 設計為 exprs 與未來命令解析器共用的「帶位置錯誤」型別,讓兩種文法的錯誤格式一致。
type SyntaxError struct {
	Pos int    // 出錯處在來源字串的 rune 索引(0 起算;供工具標示用)
	Msg string // 錯誤說明(企劃白話,不含內部前綴)
}

// Error 以「第 N 字附近」開頭回報;N 為 Pos 的 1 起算位置,對齊企劃對「第幾個字」的直覺。
func (this *SyntaxError) Error() string {
	return fmt.Sprintf("第 %d 字附近：%s", this.Pos+1, this.Msg)
}
