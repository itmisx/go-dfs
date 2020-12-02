package main

import (
	"fmt"
	"go-dfs/internal/app"
	"go-dfs/internal/defines"
	"go-dfs/internal/pkg"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		// 存储组信息
		if os.Args[1] == "group-info" {
			leveldb, err := pkg.NewLDB(defines.StorageGroupDb)
			if err != nil {
				fmt.Println(err)
				return
			}
			iter := leveldb.Db().NewIterator(nil, nil)
			for iter.Next() {
				fmt.Printf("----------%s----------\n", iter.Key())
				fmt.Printf("%s\n", iter.Value())
			}
			iter.Release()
		}
		// 文件同步列表
		if os.Args[1] == "file-sync-log" {
			leveldb, err := pkg.NewLDB(defines.FileSyncLogDb)
			if err != nil {
				fmt.Println(err)
				return
			}
			iter := leveldb.Db().NewIterator(nil, nil)
			for iter.Next() {
				fmt.Printf("%s:", iter.Key())
				fmt.Printf("%s\n", iter.Value())
			}
			iter.Release()
		}
		// 文件列表
		if os.Args[1] == "file-list" {
			leveldb, err := pkg.NewLDB(defines.FileListDb)
			if err != nil {
				fmt.Println(err)
				return
			}
			iter := leveldb.Db().NewIterator(nil, nil)
			for iter.Next() {
				fmt.Printf("%s:", iter.Key())
				fmt.Printf("%s\n", iter.Value())
			}
			iter.Release()
		}
		// 异常日志
		if os.Args[1] == "log" {
			leveldb, err := pkg.NewLDB(defines.LogDb)
			if err != nil {
				fmt.Println(err)
				return
			}
			iter := leveldb.Db().NewIterator(nil, nil)
			for iter.Next() {
				fmt.Printf("%s:", iter.Key())
				fmt.Printf("%s\n", iter.Value())
			}
			iter.Release()
		}
		return
	}
	// 启动服务
	app.Start(nil)
}
