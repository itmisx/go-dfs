package app

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"go-dfs/internal/defines"
	"go-dfs/internal/pkg"
	"go-dfs/internal/schema"
	"math/rand"
	"net"
	"net/http/httputil"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

// Group , group info struct
type Group struct {
	Name           string                   `json:"name"`            // storage server group name
	Status         int8                     `json:"status"`          // group storage status
	Cap            uint64                   `json:"cap"`             // available cap of the group
	StorageServers map[string]StorageServer `json:"storage_servers"` // storage server members of the group
}

// StorageServer , storage server info struct
type StorageServer struct {
	Group      string `json:"group"`
	Scheme     string `json:"scheme"`      //主机http协议类型,https or http
	ServerAddr string `json:"server_addr"` //主机信息，ip:port
	Status     int8   `json:"status"`      //主机状态, status，0：offline 1：alive 2: file sync 3: no enough space
	Cap        uint64 `json:"cap"`         //最大可用容量
	UpdateTime int64  `json:"update_time"` //更新时间
}

// Tracker tracker server
type Tracker struct {
	ServerConfig pkg.DsfConfigType
}

// NewTracker  ,return *Tracker
func NewTracker() *Tracker {
	tracker := new(Tracker)
	return tracker
}

// Start , 启动tracker
func (t *Tracker) Start(serverConfig pkg.DsfConfigType) {
	t.ServerConfig = serverConfig
	// start tracker crontab
	t.StartTrackerCron()
	// gin init
	router := gin.Default()
	router.Use(t.Download())
	router.POST("/upload", t.Upload)
	router.POST("/delete", t.Delete)
	router.POST("/confirm", t.Confirm)
	router.POST("/report-status", t.HanldeStorageServerReport)
	router.POST("/report-err", t.HandleReportErrorMsg)

	router.Run(":" + t.ServerConfig.BindPort)
}

// Download , 文件下载
func (t *Tracker) Download() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			matched, _ := regexp.MatchString("^/group", c.Request.RequestURI)
			if matched {
				// 选择合适的group
				g := strings.Split(c.Request.RequestURI, "/")[1]
				if g == "" {
					pkg.Helper{}.AjaxReturn(c, 300004, "")
					return
				}
				group, err := t.GetGroup(g)
				if err != nil {
					pkg.Helper{}.AjaxReturn(c, 300004, "")
					return
				}
				// 选择合适的存储
				s, err := t.SelectStorage(c, group)
				// 反向代理
				t.HTTPProxy(c, s.Scheme, s.ServerAddr)
				return
			}
		}
		c.Next()
	}
}

// Upload , 文件上传
func (t *Tracker) Upload(c *gin.Context) {
	// select storage
	group, err := t.SelectGroupForUPload()
	if err != nil {
		pkg.Helper{}.AjaxReturn(c, 100000, "")
		return
	}
	// select storage
	validStorageServer, err := t.SelectStorage(c, group)
	if err != nil {
		pkg.Helper{}.AjaxReturn(c, 100000, "")
		return
	}
	goDfsFilepath, goDfsFilename := t.GenerateFileName(validStorageServer)
	c.Request.Header.Add("Go-Dfs-Filepath", goDfsFilepath)
	c.Request.Header.Add("Go-Dfs-Filename", goDfsFilename)

	// http proxy
	t.HTTPProxy(c, validStorageServer.Scheme, validStorageServer.ServerAddr)
	// distribute
	if c.Writer.Header().Get("Go-Dfs-Upload-Result") == "1" {
		// get the file ext
		goDfsExt := c.Writer.Header().Get("Go-Dfs-Ext")
		// put to the temp file list
		if t.ServerConfig.Tracker.EnableTempFile {
			tempFileListDb, err := pkg.NewLDB(defines.TempFileListDb)
			if err == nil {
				ldata, _ := json.Marshal(schema.TempFile{CreateTime: time.Now().Unix()})
				tempFileListDb.Do(goDfsFilepath+"/"+goDfsFilename+goDfsExt, ldata)
			}
		}
		// recode file list db
		fileListDb, err := pkg.NewLDB(defines.FileListDb)
		if err == nil {
			fileInfo := schema.FileInfo{}
			fileInfo.Size, _ = strconv.ParseUint(c.Writer.Header().Get("Go-Dfs-Size"), 10, 64)
			ldata, _ := json.Marshal(fileInfo)
			fileListDb.Do(goDfsFilepath+"/"+goDfsFilename+goDfsExt, ldata)
		}
		StorageServers := t.GetStorages(group)
		for _, sm := range StorageServers {
			if sm.ServerAddr == validStorageServer.ServerAddr {
				continue
			}
			go func(sm StorageServer) {
				syncFileInfo := schema.SyncFileInfo{
					Src:      validStorageServer.Scheme + "://" + validStorageServer.ServerAddr,
					Dst:      sm.Scheme + "://" + sm.ServerAddr,
					FilePath: goDfsFilepath,
					FileName: goDfsFilename + goDfsExt,
					Action:   defines.FileSyncActionAdd,
					Group:    validStorageServer.Group,
				}
				t.FileSyncAndLog(sm, syncFileInfo)
			}(sm)
		}
	}
}

// FileSyncAndLog , 文件新增同步
func (t *Tracker) FileSyncAndLog(sm StorageServer, syncFileInfo schema.SyncFileInfo) {
	ldbData, err := json.Marshal(syncFileInfo)
	if err != nil {
		return
	}
	leveldb, err := pkg.NewLDB(defines.FileSyncLogDb)
	if err != nil {
		return
	}
	if sm.Status != 1 {
		// 写入日志，定时继续同步
		leveldb.Do(syncFileInfo.FileName+"-"+defines.FileSyncActionAdd, ldbData)
	} else if sm.Status == 1 {
		URL := sm.Scheme + "://" + sm.ServerAddr
		res, err := pkg.Helper{}.PostJSON(URL+"/sync-file", syncFileInfo, nil, 10*time.Second)
		if err != nil || len(res) == 0 {
			// 写入日志，定时继续同步
			leveldb.Do(syncFileInfo.FileName+"-"+defines.FileSyncActionAdd, ldbData)
			return
		}
		var syncRes struct {
			Code int64
		}
		err = json.Unmarshal(res, &syncRes)
		if err != nil {
			leveldb.Do(syncFileInfo.Dst+"-"+syncFileInfo.FileName+"-"+defines.FileSyncActionAdd, ldbData)
			return
		}
		if syncRes.Code > 0 {
			leveldb.Do(syncFileInfo.Dst+"-"+syncFileInfo.FileName+"-"+defines.FileSyncActionAdd, ldbData)
			return
		}
	}
}

// Delete , 文件删除
func (t *Tracker) Delete(c *gin.Context) {
	var DelInfo struct {
		File string `json:"file"`
	}
	c.ShouldBind(&DelInfo)
	errCode := t.DeleteSync(DelInfo.File)
	pkg.Helper{}.AjaxReturn(c, errCode, "")

}

// Confirm , 文件确认
// 避免文件被删除
func (t *Tracker) Confirm(c *gin.Context) {
	var ConfirmFile struct {
		File string `json:"file"`
	}
	c.ShouldBind(&ConfirmFile)
	tempFileListDb, err := pkg.NewLDB(defines.TempFileListDb)
	if err != nil {
		pkg.Helper{}.AjaxReturn(c, 300006, "")
		return
	}
	_, err1 := tempFileListDb.Do(ConfirmFile.File, nil)
	if err1 != nil {
		pkg.Helper{}.AjaxReturn(c, 300006, "")
		return
	}
	pkg.Helper{}.AjaxReturn(c, 0, "")
}

// DeleteSync 删除同步
func (t *Tracker) DeleteSync(file string) (errCode int64) {
	g := ""
	f := strings.Split(file, "/")
	if len(f) < 2 {
		errCode = 300004
		return
	}
	g = f[0]
	if g == "" {
		g = f[1]
	}
	if g == "" {
		errCode = 300004
		return
	}
	group, err := t.GetGroup(g)
	if err != nil {
		errCode = 300004
		return
	}
	// delete file from everyone storage server of the group
	leveldb, err := pkg.NewLDB(defines.FileSyncLogDb)
	// delete the file record from file list db
	leveldb1, err1 := pkg.NewLDB(defines.FileListDb)
	if err1 == nil {
		leveldb1.Do(file, nil)
	}
	for _, s := range group.StorageServers {
		syncFileInfo := schema.SyncFileInfo{
			Dst:      s.Scheme + "://" + s.ServerAddr,
			FilePath: path.Dir(file),
			FileName: path.Base(file),
			Action:   defines.FileSyncActionDelete,
			Group:    s.Group,
		}
		ldbData, _ := json.Marshal(syncFileInfo)
		res, err := pkg.Helper{}.PostJSON(s.Scheme+"://"+s.ServerAddr+"/sync-file", syncFileInfo, nil, 10*time.Second)
		if err != nil || len(res) == 0 {
			leveldb.Do(syncFileInfo.FileName+"-"+defines.FileSyncActionDelete, ldbData)
			return
		}
		var syncRes struct {
			Code int64
		}
		err = json.Unmarshal(res, &syncRes)
		if err != nil {
			leveldb.Do(syncFileInfo.Dst+"-"+syncFileInfo.FileName+"-"+defines.FileSyncActionDelete, ldbData)
			return
		}
		if syncRes.Code > 0 {
			leveldb.Do(syncFileInfo.Dst+"-"+syncFileInfo.FileName+"-"+defines.FileSyncActionDelete, ldbData)
			return
		}
	}
	return 0
}

// HanldeStorageServerReport , 处理存储服务器的上报信息
// update the status and the maximum capacity of the group
func (t *Tracker) HanldeStorageServerReport(c *gin.Context) {
	// parse request param
	var params struct {
		Scheme      string `json:"scheme"`
		Group       string `json:"group"`
		ServiceIP   string `json:"service_ip"`
		ServicePort string `json:"service_port"`
		Cap         uint64 `json:"cap"`
	}
	c.ShouldBind(&params)
	// pack
	if params.ServiceIP == "" || params.ServicePort == "" {
		return
	}
	storageServer := StorageServer{
		Scheme:     params.Scheme,
		ServerAddr: net.JoinHostPort(params.ServiceIP, params.ServicePort),
		Status:     1,
		Cap:        params.Cap,
		Group:      params.Group,
		UpdateTime: time.Now().Unix(),
	}
	// read group info from leveldb
	leveldb, err := pkg.NewLDB(defines.StorageGroupDb)
	if err != nil {
		fmt.Println(err)
		return
	}
	g, err := leveldb.Do(storageServer.Group)
	if g == nil { // new group
		newGroup := Group{
			Name:           params.Group,
			Status:         1,
			Cap:            storageServer.Cap,
			StorageServers: make(map[string]StorageServer),
		}
		// add new member
		newGroup.StorageServers[storageServer.ServerAddr] = storageServer
		// save to group
		ldbData, err := json.Marshal(newGroup)
		if err != nil {
			return
		}
		leveldb.Do(storageServer.Group, ldbData)
		return
	}
	var group Group
	err = json.Unmarshal(g, &group)
	//modify group
	newGroup := Group{
		Name:           params.Group,
		Status:         1,
		StorageServers: make(map[string]StorageServer),
	}
	// the minimum capacity is the maximum cap of the group
	var caps []uint64
	for k, v := range group.StorageServers {
		newGroup.StorageServers[k] = v
		caps = append(caps, v.Cap)
	}
	sort.Slice(caps, func(i, j int) bool { return caps[i] < caps[j] })
	newGroup.Cap = caps[0]
	// add new member
	newGroup.StorageServers[net.JoinHostPort(params.ServiceIP, params.ServicePort)] = storageServer
	// save to group
	ldbData, err := json.Marshal(newGroup)
	if err != nil {
		return
	}
	leveldb.Do(storageServer.Group, ldbData)
	return
}

// HandleReportErrorMsg ,接收storage server的错误上报并保存
func (t *Tracker) HandleReportErrorMsg(c *gin.Context) {
	var params struct {
		Group string `json:"group"`
		Port  string `json:"Port"`
		Msg   string `json:"msg"`
	}
	c.ShouldBind(&params)
	if params.Msg != "" {
		host := net.JoinHostPort(c.Request.RemoteAddr, params.Port)
		leveldb, err := pkg.NewLDB(defines.LogDb)
		if err != nil {
			leveldb.Do("storage-"+params.Group+"-"+host, []byte(params.Msg))
		}
	}
}

// StartTrackerCron ，启动tracker定时任务
// 检测storage server的状态并标记group的状态
// 文件同步失败的补偿
func (t *Tracker) StartTrackerCron() {
	cr := cron.New(cron.WithSeconds())
	cr.AddFunc("0 0 * * * *", func() {
		tempFileListDb, err := pkg.NewLDB(defines.TempFileListDb)
		if err != nil {
			return
		}
		iter := tempFileListDb.Db().NewIterator(nil, nil)
		for iter.Next() {
			tempFile := schema.TempFile{}
			err := json.Unmarshal(iter.Value(), &tempFile)
			if err != nil {
				continue
			}
			// 超过半个小时，删除文件
			if time.Now().Unix()-tempFile.CreateTime > 1800 {
				t.DeleteSync(string(iter.Key()))
			}
		}
		iter.Release()
	})
	// 计算各个存储组的状态以及最大可用容量
	// do/10second
	cr.AddFunc("*/10 * * * * *", func() {
		groups := t.GetGroups()
		for _, g := range groups {
			validStorages := make([]StorageServer, 0)
			for _, s := range g.StorageServers {
				if s.Status == 1 {
					if time.Now().Unix()-s.UpdateTime > 30 {
						s.Status = 0
						// update the storage server status
						g.StorageServers[s.ServerAddr] = s
					} else {
						validStorages = append(validStorages, s)
					}
				}
			}
			if len(validStorages) <= 0 {
				g.Status = 0
				g.Cap = 0
			} else {
				g.Status = 1
				sort.Slice(validStorages, func(i, j int) bool { return validStorages[i].Cap < validStorages[j].Cap })
				g.Cap = validStorages[0].Cap
			}
			// save the group info into the leveldb
			ldb, err := pkg.NewLDB(defines.StorageGroupDb)
			if err != nil {
				return
			}
			ldbData, _ := json.Marshal(g)
			ldb.Do(g.Name, ldbData)
		}
	})
	// 文件同步补偿
	// do/hour
	cr.AddFunc("0 0 * * * *", func() {
		ldb, err := pkg.NewLDB(defines.FileSyncLogDb)
		if err != nil {
			return
		}
		iter := ldb.Db().NewIterator(nil, nil)
		for iter.Next() {
			v := iter.Value()
			fileSyncInfo := schema.SyncFileInfo{}
			err := json.Unmarshal(v, &fileSyncInfo)
			if err != nil {
				continue
			}
			g, err := t.GetGroup(fileSyncInfo.Group)
			if err == nil {
				if g.StorageServers[fileSyncInfo.Dst].Status == 1 && g.StorageServers[fileSyncInfo.Src].Status == 1 {
					URL := fileSyncInfo.Dst
					res, err := pkg.Helper{}.PostJSON(URL+"/sync-file", fileSyncInfo, nil, 10*time.Second)
					if err != nil || len(res) == 0 {
						continue
					}
					var syncRes struct {
						Code int64
					}
					err = json.Unmarshal(res, &syncRes)
					if err != nil {
						continue
					}
					if syncRes.Code > 0 {
						continue
					}
					// if succeed , del the record
					ldb.Do(fileSyncInfo.FileName+"-"+fileSyncInfo.Action, nil)
				}
			}
		}
		iter.Release()
	})
	// 启动定时任务
	cr.Start()
}

// HTTPProxy ,http 反向代理
func (t *Tracker) HTTPProxy(c *gin.Context, Scheme, Host string) bool {
	remote, err := url.Parse(Scheme + "://" + Host)
	if err != nil {
		return false
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(c.Writer, c.Request)
	return true
}

// GetGroup , 获取文件匹配的存储组
func (t *Tracker) GetGroup(group string) (g Group, err error) {
	ldb, err := pkg.NewLDB(defines.StorageGroupDb)
	if err != nil {
		fmt.Println(err)
		return Group{}, err
	}
	v, err := ldb.Do(group)
	err = json.Unmarshal(v, &g)
	if err != nil {
		return Group{}, err
	}
	return g, nil
}

// GetGroups , 获取所有的存储组
func (t *Tracker) GetGroups() (groups []Group) {
	ldb, err := pkg.NewLDB(defines.StorageGroupDb)
	if err != nil {
		fmt.Println(err)
		return
	}
	iter := ldb.Db().NewIterator(nil, nil)
	for iter.Next() {
		var g Group
		err := json.Unmarshal(iter.Value(), &g)
		if err != nil {
			continue
		}
		groups = append(groups, g)
	}
	iter.Release()
	if len(groups) <= 0 {
		return nil
	}
	return groups
}

// GetOneValidGroup , 获取一个有效的存储组
// 不包括已经离线的
func (t *Tracker) GetOneValidGroup() (Group, error) {
	var groups []Group
	var validGroups []Group
	groups = t.GetGroups()
	if groups != nil {
		return Group{}, errors.New("there is no available group")
	}
	for _, v := range groups {
		if v.Status == 1 {
			validGroups = append(validGroups, v)
		}
	}
	if len(validGroups) <= 0 {
		return Group{}, errors.New("there is no online group")
	}
	return validGroups[rand.Intn(len(validGroups))], nil
}

// SelectGroupForUPload , 选择存储组
// 选择可用空间最大的
func (t *Tracker) SelectGroupForUPload() (Group, error) {
	var groups Groups
	gs := t.GetGroups()
	for _, v := range gs {
		if v.Status == 1 {
			groups = append(groups, v)
		}
	}
	sort.Sort(groups)
	if len(groups) <= 0 {
		return Group{}, errors.New("thers is no available group")
	}
	return groups[len(groups)-1], nil
}

// GetStorages , 获取存储组的存储服务器列表
func (t *Tracker) GetStorages(group Group) (StorageServers []StorageServer) {
	for _, v := range group.StorageServers {
		if v.ServerAddr == ":" {
			continue
		}
		StorageServers = append(StorageServers, v)
	}
	if len(StorageServers) <= 0 {
		return nil
	}
	return StorageServers
}

//SelectStorage , 选择存储服务器，根据 ip hash
func (t *Tracker) SelectStorage(c *gin.Context, group Group) (StorageServer, error) {
	var validStorages []StorageServer
	s := t.GetStorages(group)
	for _, v := range s {
		if v.Status == 1 {
			validStorages = append(validStorages, v)
		}
	}
	if validStorages == nil {
		return StorageServer{}, errors.New("thers is no available storage server")
	}
	// calculate ip hash , find the storage server
	signByte := []byte(c.ClientIP())
	hash := md5.New()
	hash.Write(signByte)
	md5Hex := hash.Sum(nil)
	hashIndex := int(md5Hex[len(md5Hex)-1]) % len(validStorages)
	vsm := validStorages[hashIndex]
	return vsm, nil
}

// GenerateFileName , 生成文件名
// 包括路径和文件名字
func (t *Tracker) GenerateFileName(sm StorageServer) (filepath, filename string) {

	year := strconv.Itoa(time.Now().Year())
	month := strconv.Itoa(int(time.Now().Month()))
	day := strconv.Itoa(time.Now().Day())
	hour := strconv.Itoa(time.Now().Hour())
	goDfsFilepath := fmt.Sprintf("/"+sm.Group+"/%s/%s/%s/%s", year, month, day, hour)

	// generate dfs filename by snowflake id
	uuid := pkg.Helper{}.UUID()
	goDfsFilename := strconv.FormatInt(uuid, 10)

	return goDfsFilepath, goDfsFilename
}

// Groups 重定义group slice
// 实现自定义排序
type Groups []Group

// Len is the number of elements in the collection.
func (g Groups) Len() int {
	return len(g)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (g Groups) Less(i, j int) bool {
	return g[i].Cap < g[i].Cap
}

// Swap swaps the elements with indexes i and j.
func (g Groups) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}
