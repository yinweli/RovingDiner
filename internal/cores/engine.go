package cores

import (
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// engine 驅動引擎骨架(M5):委派實作 exprs.Resolver、讀全域詞彙表。
// Runtime / 三 port / *sheeter.Sheeter 等狀態於 M6 起按需補(【營業實作規格書 | 二、套件結構】engine.go)。
type engine struct{}

// Attr 委派全域屬性詞彙表,以自身為 context 求值。M5 表為空故恆回傳未命中。
func (this *engine) Attr(name string, arg []exprs.Value) (result exprs.Value, ok bool) {
	read, known := tableAttr[name]

	if known == false {
		return exprs.Value{}, false
	} // if

	return read(this, arg)
}

// AttrRef 委派全域引用屬性詞彙表。M5 表為空故恆回傳未命中。
func (this *engine) AttrRef(ref exprs.Ref, name string, arg []exprs.Value) (result exprs.Value, ok bool) {
	read, known := tableAttrRef[name]

	if known == false {
		return exprs.Value{}, false
	} // if

	return read(this, ref, arg)
}

// 編譯期確認 engine 滿足 exprs.Resolver(條件對象求值的接縫)。
var _ exprs.Resolver = (*engine)(nil)
