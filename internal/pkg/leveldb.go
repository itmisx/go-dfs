package pkg

import (
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
)

// LevelDB , global leveldb resource
var LevelDB *leveldb.DB

// LevelDBOpenChan ,leveldb channel
var LevelDBOpenChan chan int = make(chan int, 1)

// LevelDb ,level db funcs
type LevelDb struct {
}

// NewLDB , new levelDb
func NewLDB(name string) (ldb *LevelDb, err error) {
	ldb = &LevelDb{}
	// db file chan lock, avoid concurrrent access the db file
	LevelDBOpenChan <- 1
	defer func() {
		<-LevelDBOpenChan
	}()
	if LevelDB == nil {
		LevelDB, err = leveldb.OpenFile("./leveldb/"+name, nil)
		if err != nil {
			return ldb, err
		}
		return ldb, nil
	}
	return ldb, nil
}

// Db , return global LevelDB
func (l *LevelDb) Db() *leveldb.DB {
	return LevelDB
}

// Do ,leveldb operation
func (l *LevelDb) Do(key string, value ...[]byte) ([]byte, error) {
	if key == "" {
		return nil, errors.New("key can not be empty")
	}
	if len(value) == 0 { // get value
		return LevelDB.Get([]byte(key), nil)
	} else if value[0] == nil { // delete value
		err := LevelDB.Delete([]byte(key), nil)
		return nil, err
	} else { //set value
		err := LevelDB.Put([]byte(key), value[0], nil)
		return nil, err
	}
}

// Close , relase db resource
// func (l *LevelDb) Close() {
// 	l.Db.Close()
// }
