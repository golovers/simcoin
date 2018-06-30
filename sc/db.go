package sc

import (
	"errors"
	"sync"
)

var lastBlockKey = []byte("lastblock")

// Database wraps all database operations. All methods are safe for concurrent use.
type Database interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Close()
}

/*
 * This is a test memory database. Do not use for any production it does not get persisted
 */
type MemDatabase struct {
	db   map[string][]byte
	lock sync.RWMutex
}

func NewMemDatabase() (*MemDatabase, error) {
	return &MemDatabase{
		db: make(map[string][]byte),
	}, nil
}

func NewMemDatabaseWithCap(size int) (*MemDatabase, error) {
	return &MemDatabase{
		db: make(map[string][]byte, size),
	}, nil
}

func (db *MemDatabase) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.db[string(key)] = copyBytes(value)
	return nil
}

func (db *MemDatabase) Has(key []byte) (bool, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	_, ok := db.db[string(key)]
	return ok, nil
}

func (db *MemDatabase) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if entry, ok := db.db[string(key)]; ok {
		return copyBytes(entry), nil
	}
	return nil, errors.New("not found")
}

func (db *MemDatabase) Keys() [][]byte {
	db.lock.RLock()
	defer db.lock.RUnlock()

	keys := [][]byte{}
	for key := range db.db {
		keys = append(keys, []byte(key))
	}
	return keys
}

func (db *MemDatabase) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	delete(db.db, string(key))
	return nil
}

func (db *MemDatabase) Close() {}

// BlockIterator represent block ite
type BlockIterator struct {
	current *Block
	db      Database
}

func NewBlockIterator(db Database) BlockIterator {
	return BlockIterator{
		db: db,
	}
}

func (it *BlockIterator) Next() *Block {
	var b *Block
	var v []byte
	var err error
	if it.current == nil {
		v, err = it.db.Get(lastBlockKey)
	} else if len(it.current.PrevHash) == 0 { // genesis
		return nil
	} else {
		v, err = it.db.Get(it.current.PrevHash)
	}
	if err != nil {
		panic(err)
	}
	toObject(v, &b)
	it.current = b
	return b
}
