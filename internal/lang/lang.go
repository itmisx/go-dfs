package lang

// T 语言码翻译
func T(l string, code int64) string {
	if l != "zh-cn" && l != "en-us" && l != "ja-jp" {
		l = "zh-cn"
	}
	return LangCodeDefine[l][code]
}
