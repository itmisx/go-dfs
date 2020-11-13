package schema

// FileInfo , file list
type FileInfo struct {
	Size uint64 `json:"size"`
}

// TempFile 文件过期控制
type TempFile struct {
	CreateTime int64 `json:"create_time"`
}
