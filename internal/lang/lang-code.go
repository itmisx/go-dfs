package lang

// LangCodeDefine 语言码定义
var LangCodeDefine = map[string]map[int64]string{
	"zh-cn": {
		100000: "系统出错",
		300001: "成功",
		300002: "上传失败",
		300003: "文件太大",
		300004: "文件名称不合法",
	},
	"en-us": {
		100000: "system error",
		300001: "success",
		300002: "upload failed",
		300003: "file size too big",
		300004: "File name is invalid",
	},
	"ja-jp": {
		100000: "システムエラー",
		300001: "成功",
		300002: "アップロードに失敗しました",
		300003: "ファイルサイズが大きすぎます",
		300004: "ファイル名が無効です",
	},
}
