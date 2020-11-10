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

func TestFunc(*testing.T) {
	leveldb, err := pkg.NewLDB(defines.FileSyncLogDb)
	if err != nil {
		return
	}
	leveldb.Do("abc", []byte("abcd"))
	leveldb.Do("abcd", []byte("abcd1"))
	leveldb.Do("abcde", []byte("abcd1"))
	leveldb.Do("abcdef", []byte("abcdef"))
	iter := leveldb.Db().NewIterator(nil, nil)
	// for ok := iter.Seek([]byte("abcd")); ok; ok = iter.Next() {
	// 	fmt.Printf("%s:%s\n", iter.Key(), iter.Value())
	// }
	for iter.Next() {
		fmt.Printf("%s:%s\n", iter.Key(), iter.Value())
	}
	iter.Release()
}

func TestDiskUsage(t *testing.T) {
	path := "./dfs1"
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

func TestChan(*testing.T) {
	c := make(chan int, 1)
	c <- 1
	<-c
	fmt.Println("pass")
}
