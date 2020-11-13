package defines

// define db
const (
	// StorageGroupDb,storage server group
	StorageGroupDb = "storage_group_db"
	// SyncDb , db for file sync log
	FileSyncLogDb = "file_sync_log"
	// FileListDb
	FileListDb = "file_list_db"
	// LogDb
	LogDb = "log_db"
	// 文件回收站
	TempFileListDb = "temp_file_list_db"
)

// define action
const (
	// FileSyncActionAdd ,action for add
	FileSyncActionAdd = "add"
	// FileSyncActionDelete ,action for delete
	FileSyncActionDelete = "delete"
)
