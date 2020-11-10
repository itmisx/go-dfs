package schema

// SyncFileInfo file sync info
type SyncFileInfo struct {
	SrcScheme string `json:"src_scheme"`
	SrcHost   string `json:"src_host"` //file source host
	DstScheme string `json:"dst_scheme"`
	DstHost   string `json:"dst_host"` //file dst host
	FilePath  string `json:"file_path"`
	FileName  string `json:"file_name"`
	Action    string `json:"action"` // add or delete
	Group     string `json:"group"`
}

// SyncLogType , file sync log struct
type SyncLogType struct {
	Time int64 `json:"time"`
	SyncFileInfo
}
