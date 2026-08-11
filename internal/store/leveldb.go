package store

import (
	"errors"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	leveldbErrors "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type leveldbStore struct {
	db         *leveldb.DB
	readOpt    *opt.ReadOptions
	iterateOpt *opt.ReadOptions
	writeOpt   *opt.WriteOptions
}

// Open opens a LevelDB-backed Store at path.
func Open(path string) (Store, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("open leveldb: %w", err)
	}

	return &leveldbStore{
		db:         db,
		readOpt:    &opt.ReadOptions{DontFillCache: false},
		iterateOpt: &opt.ReadOptions{DontFillCache: true},
		writeOpt:   &opt.WriteOptions{Sync: false, NoWriteMerge: false},
	}, nil
}

func (s *leveldbStore) Get(key []byte) ([]byte, error) {
	value, err := s.db.Get(key, s.readOpt)
	if errors.Is(err, leveldbErrors.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("leveldb get: %w", err)
	}
	return value, nil
}

func (s *leveldbStore) Put(key, value []byte) error {
	if err := s.db.Put(key, value, s.writeOpt); err != nil {
		return fmt.Errorf("leveldb put: %w", err)
	}
	return nil
}

func (s *leveldbStore) Delete(key []byte) error {
	if err := s.db.Delete(key, s.writeOpt); err != nil {
		return fmt.Errorf("leveldb delete: %w", err)
	}
	return nil
}

func (s *leveldbStore) DeleteMany(keys [][]byte) error {
	if len(keys) == 0 {
		return nil
	}

	batch := new(leveldb.Batch)
	for _, key := range keys {
		batch.Delete(key)
	}
	if err := s.db.Write(batch, s.writeOpt); err != nil {
		return fmt.Errorf("leveldb delete many: %w", err)
	}
	return nil
}

func (s *leveldbStore) Iterate(prefix []byte, fn func(key, value []byte) error) error {
	iter := s.db.NewIterator(util.BytesPrefix(prefix), s.iterateOpt)
	defer iter.Release()

	for iter.Next() {
		if err := fn(iter.Key(), iter.Value()); err != nil {
			return err
		}
	}
	if err := iter.Error(); err != nil {
		return fmt.Errorf("leveldb iterate: %w", err)
	}
	return nil
}

func (s *leveldbStore) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close leveldb: %w", err)
	}
	return nil
}
