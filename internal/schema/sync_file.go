package schema

// SyncFileInfo file sync info
type SyncFileInfo struct {
	Src      string `json:"src"` //file source host
	Dst      string `json:"dst"` //file dst host
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
	Action   string `json:"action"` // add or delete
	Group    string `json:"group"`
}
