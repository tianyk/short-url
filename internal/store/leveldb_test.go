package store

import (
	"errors"
	"path/filepath"
	"sort"
	"testing"
)

func TestLevelDBStoreContract(t *testing.T) {
	storage, err := Open(filepath.Join(t.TempDir(), "store"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			t.Error(err)
		}
	}()
	levelStorage, ok := storage.(*leveldbStore)
	if !ok {
		t.Fatalf("expected LevelDB store, got %T", storage)
	}
	if !levelStorage.iterateOpt.DontFillCache {
		t.Fatal("expected iteration not to populate the block cache")
	}

	if _, err := storage.Get([]byte("missing")); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	entries := map[string]string{
		"s-one": "one",
		"s-two": "two",
		"other": "ignored",
	}
	for key, value := range entries {
		if err := storage.Put([]byte(key), []byte(value)); err != nil {
			t.Fatal(err)
		}
	}

	value, err := storage.Get([]byte("s-one"))
	if err != nil || string(value) != "one" {
		t.Fatalf("get value: value=%q err=%v", value, err)
	}

	var keys []string
	if err := storage.Iterate([]byte("s-"), func(key, value []byte) error {
		keys = append(keys, string(key))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(keys)
	if len(keys) != 2 || keys[0] != "s-one" || keys[1] != "s-two" {
		t.Fatalf("unexpected prefix iteration result: %v", keys)
	}

	if err := storage.Delete([]byte("s-one")); err != nil {
		t.Fatal(err)
	}
	if err := storage.DeleteMany([][]byte{[]byte("s-two"), []byte("missing")}); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.Get([]byte("s-one")); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected deleted key to be missing, got %v", err)
	}
	if _, err := storage.Get([]byte("s-two")); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected batch-deleted key to be missing, got %v", err)
	}
	if value, err := storage.Get([]byte("other")); err != nil || string(value) != "ignored" {
		t.Fatalf("expected unrelated key to remain: value=%q err=%v", value, err)
	}
}

func TestLevelDBStorePersistsExistingFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store")
	storage, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := storage.Put([]byte("URL-ID"), []byte("46656")); err != nil {
		t.Fatal(err)
	}
	if err := storage.Put([]byte("s-abcd"), []byte("protobuf-data")); err != nil {
		t.Fatal(err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	counter, err := reopened.Get([]byte("URL-ID"))
	if err != nil || string(counter) != "46656" {
		t.Fatalf("read counter after reopen: value=%q err=%v", counter, err)
	}
	data, err := reopened.Get([]byte("s-abcd"))
	if err != nil || string(data) != "protobuf-data" {
		t.Fatalf("read value after reopen: value=%q err=%v", data, err)
	}
}

func TestLevelDBIteratePropagatesCallbackError(t *testing.T) {
	storage, err := Open(filepath.Join(t.TempDir(), "store"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	if err := storage.Put([]byte("s-one"), []byte("one")); err != nil {
		t.Fatal(err)
	}

	expected := errors.New("stop")
	err = storage.Iterate([]byte("s-"), func(key, value []byte) error {
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected callback error, got %v", err)
	}
}
