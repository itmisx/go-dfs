package app

import (
	"fmt"
	"go-dfs/internal/defines"
	"go-dfs/internal/pkg"
	"go-dfs/internal/schema"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/shirou/gopsutil/v3/disk"
)

// Storage storage server
type Storage struct {
	ServerConfig pkg.DsfConfigType
	Status       int
	Dirs         []string //file dir map
}
type dbFileInfo struct {
	Name  string
	IsDel bool
}

// NewStorage  ,return *Tracker
func NewStorage() *Storage {
	storage := new(Storage)
	return storage
}

// Start ,start storage server
func (s *Storage) Start(serverConfig pkg.DsfConfigType) {
	s.ServerConfig = serverConfig

	// start gin
	// before dir sync , only reponse file sync , will not support file download
	router := gin.Default()
	// static file handler
	router.Static("", s.ServerConfig.Storage.StoragePath)
	// upload file handler
	router.POST("/upload", s.Upload)
	// sync file handler
	router.POST("/sync-file", s.SyncFile)
	// start storage time task
	s.StartStorageCron()
	// run the gin service
	router.Run(":" + serverConfig.BindPort)
}

// Upload upload file
func (s *Storage) Upload(c *gin.Context) {

	// 创建根目录路径
	// 先判断是否存在，不存在就直接创建
	goDfsFilepath := c.Request.Header.Get("Go-Dfs-Filepath")
	baseDir := s.ServerConfig.Storage.StoragePath + goDfsFilepath
	_, err := os.Stat(baseDir)
	if err != nil {
		err = os.MkdirAll(baseDir, os.ModePerm)
		if err != nil {
			s.ReportErrorMsg("create root dir error")
			pkg.Helper{}.AjaxReturn(c, 1, "")
			return
		}
	}

	//处理上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		fmt.Println(err)
		s.ReportErrorMsg("no file find to upload")
		pkg.Helper{}.AjaxReturn(c, 1, "")
		return
	}
	// 文件大小限制
	size := file.Size
	if s.ServerConfig.Storage.FileSizeLimit > 0 && size > s.ServerConfig.Storage.FileSizeLimit {
		pkg.Helper{}.AjaxReturn(c, 300003, "")
		return
	}
	// 保存上传的文件
	goDfsFilename := c.Request.Header.Get("Go-Dfs-Filename")
	goDfsExt := path.Ext(file.Filename)
	err = c.SaveUploadedFile(file, baseDir+"/"+goDfsFilename+goDfsExt)
	if err != nil {
		pkg.Helper{}.AjaxReturn(c, 1, "")
		return
	}
	goDfsFileURL := goDfsFilepath + "/" + goDfsFilename + goDfsExt
	goDfsFileSize := strconv.FormatInt(file.Size, 10)
	c.Writer.Header().Set("Go-Dfs-Upload-Result", "1")
	c.Writer.Header().Set("Go-Dfs-Ext", goDfsExt)
	c.Writer.Header().Set("Go-Dfs-Size", goDfsFileSize)
	c.Writer.Header().Set("Go-Dfs-File-Url", goDfsFileURL)

	pkg.Helper{}.AjaxReturn(c, 0, gin.H{
		"url":  goDfsFileURL,
		"size": goDfsFileSize,
	})
	return
}

// SyncFile sync file
func (s *Storage) SyncFile(c *gin.Context) {
	var syncFileInfo schema.SyncFileInfo
	c.ShouldBind(&syncFileInfo)
	if syncFileInfo.Action == defines.FileSyncActionAdd {
		// Check if the file path exists
		// if does not exist , auto create
		baseDir := s.ServerConfig.Storage.StoragePath + syncFileInfo.FilePath
		_, err := os.Stat(baseDir)
		if err != nil {
			err = os.MkdirAll(baseDir, os.ModePerm)
			if err != nil {
				s.ReportErrorMsg("when sync file,create root dir error")
				pkg.Helper{}.AjaxReturn(c, 1, "")
				return
			}
		}
		// download the file
		res, err := http.Get(syncFileInfo.Src +
			syncFileInfo.FilePath + "/" + syncFileInfo.FileName)
		if err != nil {
			pkg.Helper{}.AjaxReturn(c, 1, "")
		}
		f, err := os.Create(baseDir + "/" + syncFileInfo.FileName)
		if err != nil {
			s.ReportErrorMsg("when sync file,create file error")
			pkg.Helper{}.AjaxReturn(c, 1, "")
		}
		l, err := io.Copy(f, res.Body)
		if err != nil || l == 0 {
			s.ReportErrorMsg("when sync file,copy file error")
			pkg.Helper{}.AjaxReturn(c, 1, "")
		}
		pkg.Helper{}.AjaxReturn(c, 0, "")
		return
	} else if syncFileInfo.Action == defines.FileSyncActionDelete {
		fullpath := s.ServerConfig.Storage.StoragePath + "/" + syncFileInfo.FilePath + "/" + syncFileInfo.FileName
		_, err := os.Stat(fullpath)
		if err != nil {
			pkg.Helper{}.AjaxReturn(c, 0, "")
			return
		}
		err = os.Remove(fullpath)
		if err != nil {
			pkg.Helper{}.AjaxReturn(c, 300005, "")
			return
		}
		pkg.Helper{}.AjaxReturn(c, 0, "")
		return

	}
}

// StartStorageCron storage crontab
// ReportStatus
// sync file system
func (s *Storage) StartStorageCron() {
	cr := cron.New(cron.WithSeconds())
	cr.AddFunc("*/5 * * * * *", func() {
		fmt.Println("ReportStatus")
		s.ReportStatus()
	})
	cr.Start()
}

// ReportStatus register status to tracker
func (s *Storage) ReportStatus() {
	var data struct {
		Group       string `json:"group"`
		Scheme      string `json:"scheme"`
		ServiceIP   string `json:"service_ip"`
		ServicePort string `json:"service_port"`
		Cap         uint64 `json:"cap"`
	}
	data.Group = s.ServerConfig.Storage.Group
	data.Scheme = s.ServerConfig.ServiceScheme
	data.ServiceIP = s.ServerConfig.ServiceIP
	data.ServicePort = s.ServerConfig.ServicePort

	path := s.ServerConfig.Storage.StoragePath
	v, err := disk.Usage(path)
	if err != nil {
		data.Cap = 0
	} else {
		data.Cap = v.Free
	}
	// 获取容量信息
	// 获取负载信息
	for _, url := range s.ServerConfig.Storage.Trackers {
		pkg.Helper{}.PostJSON(url+"/report-status", data, nil, 10*time.Second)
	}
}

// ReportErrorMsg report error msg
func (s *Storage) ReportErrorMsg(msg string) {
	type errMsg struct {
		Group       string
		ServiceIP   string
		ServicePort string
		Port        string
		Msg         string
	}
	for _, url := range s.ServerConfig.Storage.Trackers {
		pkg.Helper{}.PostJSON(url+"/report-err",
			errMsg{
				Group:       s.ServerConfig.Storage.Group,
				ServiceIP:   s.ServerConfig.ServiceIP,
				ServicePort: s.ServerConfig.ServicePort,
				Msg:         msg,
			},
			nil, 10*time.Second)
	}
}
