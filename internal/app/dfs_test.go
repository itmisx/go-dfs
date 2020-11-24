package app

import (
	"fmt"
	"go-dfs/internal/defines"
	"go-dfs/internal/pkg"
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/v3/disk"
)

func TestDfs(t *testing.T) {
	// 启动tracker
	config1 := pkg.DsfConfigType{
		ServiceType: "tracker",
		BindPort:    "9000",
		DefaultLang: "zh_cn",
	}
	config1.Tracker.NodeID = 1
	config1.Tracker.EnableTempFile = true
	go Start(&config1)
	// 启动storage
	config2 := pkg.DsfConfigType{
		ServiceType:   "storage",
		ServiceScheme: "http",
		ServiceIP:     "127.0.0.1",
		ServicePort:   "9001",
		BindPort:      "9001",
		DefaultLang:   "zh_cn",
	}
	config2.Storage.Group = "group1"
	config2.Storage.StoragePath = "./dfs/1"
	config2.Storage.Trackers = []string{"http://127.0.0.1:9000"}
	go Start(&config2)

	// 启动storage
	config3 := pkg.DsfConfigType{
		ServiceType:   "storage",
		ServiceScheme: "http",
		ServiceIP:     "127.0.0.1",
		ServicePort:   "9002",
		BindPort:      "9002",
		DefaultLang:   "zh_cn",
	}
	config3.Storage.Group = "group1"
	config3.Storage.StoragePath = "./dfs/2"
	config3.Storage.Trackers = []string{"http://127.0.0.1:9000"}
	go Start(&config3)

	<-make(chan bool)
}

func TestUUID(t *testing.T) {
	for i := 100; i > 0; i-- {
		fmt.Println(pkg.Helper{}.UUID())
	}
}

func TestLevelDb(*testing.T) {
	fmt.Println("**************temp-file****************")
	leveldb3, err := pkg.NewLDB(defines.TempFileListDb)
	if err != nil {
		fmt.Println(err)
		return
	}
	iter3 := leveldb3.Db().NewIterator(nil, nil)
	for iter3.Next() {
		fmt.Printf("%s\n", iter3.Key())
	}
	iter3.Release()
}

func TestDiskUsage(t *testing.T) {
	path := "./dfs"
	if runtime.GOOS == "windows" {
		path = "C:"
	}
	v, err := disk.Usage(path)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v.Path != path {
		t.Errorf("error %v", err)
	}
	fmt.Printf("%+v", v)
}
