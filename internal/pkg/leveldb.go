package pkg

import (
	"errors"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

// LevelDBPool , leveldb pool
var LevelDBPool sync.Map

// LevelDBOpenChan ,leveldb channel
var LevelDBOpenChan chan int = make(chan int, 1)

// LevelDb ,level db funcs
type LevelDb struct {
	DbName string
}

// NewLDB , new levelDb
func NewLDB(name string) (ldb *LevelDb, err error) {
	LevelDBOpenChan <- 1
	defer func() {
		<-LevelDBOpenChan
	}()
	ldb = &LevelDb{}
	ldb.DbName = name
	_, ok := LevelDBPool.Load(name)
	if ok {
		return ldb, nil
	}
	NewDB, err := leveldb.OpenFile("./leveldb/"+name, nil)
	if err != nil {
		return ldb, err
	}
	LevelDBPool.Store(name, NewDB)
	return ldb, nil
}

// Db , return global LevelDB
func (l *LevelDb) Db() *leveldb.DB {
	v, ok := LevelDBPool.Load(l.DbName)
	if ok {
		db, ok := v.(*leveldb.DB)
		if ok {
			return db
		}
		panic("nil db")
	}
	panic("nil db")
}

// Do ,leveldb operation
func (l *LevelDb) Do(key string, value ...[]byte) ([]byte, error) {
	if key == "" {
		return nil, errors.New("key can not be empty")
	}
	if len(value) == 0 { // get value
		return l.Db().Get([]byte(key), nil)
	} else if value[0] == nil { // delete value
		err := l.Db().Delete([]byte(key), nil)
		return nil, err
	} else { //set value
		err := l.Db().Put([]byte(key), value[0], nil)
		return nil, err
	}
}
