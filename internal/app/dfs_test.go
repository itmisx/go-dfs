package app

import (
	"fmt"
	"go-dfs/internal/defines"
	"go-dfs/internal/pkg"
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/v3/disk"
)

func TestTracker(t *testing.T) {
	// 启动tracker
	config1 := pkg.DsfConfigType{
		ServerType:  "tracker",
		HTTPPort:    "9000",
		DefaultLang: "zh_cn",
	}
	config1.Tracker.NodeID = 1
	go Start(&config1)
	// 启动storage
	config2 := pkg.DsfConfigType{
		ServerType:  "storage",
		HTTPPort:    "9001",
		DefaultLang: "zh_cn",
	}
	config2.Storage.HTTPScheme = "http"
	config2.Storage.Group = "group1"
	config2.Storage.StoragePath = "./dfs/1"
	config2.Storage.TrackerServers = []string{"http://127.0.0.1:9000"}
	go Start(&config2)

	// 启动storage
	config3 := pkg.DsfConfigType{
		ServerType:  "storage",
		HTTPPort:    "9002",
		DefaultLang: "zh_cn",
	}
	config3.Storage.HTTPScheme = "http"
	config3.Storage.Group = "group1"
	config3.Storage.StoragePath = "./dfs/2"
	config3.Storage.TrackerServers = []string{"http://127.0.0.1:9000"}
	go Start(&config3)

	<-make(chan bool)
}

func TestUUID(t *testing.T) {
	for i := 100; i > 0; i-- {
		fmt.Println(pkg.Helper{}.UUID())
	}
}

func TestLevelDb(*testing.T) {
	leveldb, err := pkg.NewLDB(defines.FileListDb)
	if err != nil {
		return
	}
	iter := leveldb.Db().NewIterator(nil, nil)
	for iter.Next() {
		fmt.Printf("%s\n", iter.Key())
	}
	iter.Release()
	iter1 := leveldb.Db().NewIterator(nil, nil)
	for iter1.Next() {
		fmt.Printf("%s\n", iter1.Key())
	}
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
