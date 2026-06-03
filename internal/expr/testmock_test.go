package expr

// mockResolver 是測試用的假 Resolver:attr / attrRef 以 map 注入,未登記者回傳失敗。
// 需驗證引數傳遞或自訂行為時,設定 onAttr / onAttrRef 覆寫對應方法。
// 預設 AttrRef 對空物件存取屬性回傳失敗,模擬【營業規格書 | 二十七、運算式 | 5】「引用為空物件 → 評估失敗」。
type mockResolver struct {
	attr      map[string]Value
	attrRef   map[string]Value
	onAttr    func(name string, arg []Value) (value Value, ok bool)
	onAttrRef func(target Value, name string, arg []Value) (value Value, ok bool)
}

func (this *mockResolver) Attr(name string, arg []Value) (value Value, ok bool) {
	if this.onAttr != nil {
		return this.onAttr(name, arg)
	} // if

	value, ok = this.attr[name]
	return value, ok
}

func (this *mockResolver) AttrRef(target Value, name string, arg []Value) (value Value, ok bool) {
	if this.onAttrRef != nil {
		return this.onAttrRef(target, name, arg)
	} // if

	if target.IsNone() {
		return Value{}, false // 空物件存取屬性 → 失敗
	} // if

	value, ok = this.attrRef[name]
	return value, ok
}
