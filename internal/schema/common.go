package schema

// Response 响应统一格式
type Response struct {
	Code int64
	Data map[string]interface{}
	Msg  string
}
