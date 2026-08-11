package shorturl

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/tianyk/short-url/internal/store"
	"github.com/tianyk/short-url/proto"
)

type fakeStore struct {
	mu               sync.RWMutex
	values           map[string][]byte
	deleteBatchSizes []int
	deleteNotify     chan struct{}
	closed           bool
}

func newFakeStore() *fakeStore {
	return &fakeStore{values: make(map[string][]byte)}
}

func (s *fakeStore) Get(key []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[string(key)]
	if !ok {
		return nil, store.ErrNotFound
	}
	return append([]byte(nil), value...), nil
}

func (s *fakeStore) Put(key, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[string(key)] = append([]byte(nil), value...)
	return nil
}

func (s *fakeStore) Delete(key []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, string(key))
	return nil
}

func (s *fakeStore) DeleteMany(keys [][]byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteBatchSizes = append(s.deleteBatchSizes, len(keys))
	for _, key := range keys {
		delete(s.values, string(key))
	}
	if s.deleteNotify != nil {
		select {
		case s.deleteNotify <- struct{}{}:
		default:
		}
	}
	return nil
}

func (s *fakeStore) Iterate(prefix []byte, fn func(key, value []byte) error) error {
	s.mu.RLock()
	entries := make(map[string][]byte)
	for key, value := range s.values {
		if len(key) >= len(prefix) && key[:len(prefix)] == string(prefix) {
			entries[key] = append([]byte(nil), value...)
		}
	}
	s.mu.RUnlock()

	for key, value := range entries {
		if err := fn([]byte(key), value); err != nil {
			return err
		}
	}
	return nil
}

func (s *fakeStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

func newTestService(t *testing.T, storage store.Store) *ShortURLService {
	t.Helper()
	shortURLService, err := New(storage)
	if err != nil {
		t.Fatal(err)
	}
	return shortURLService
}

func TestCreateAndFindShortURL(t *testing.T) {
	shortURLService := newTestService(t, newFakeStore())
	message := &proto.ShortUrlMessage{
		LongUrl: "https://example.com",
		Expire:  time.Now().Add(time.Minute).Unix(),
	}

	urlID, err := shortURLService.CreateShortURL(message)
	if err != nil {
		t.Fatal(err)
	}
	longURL, err := shortURLService.FindLongURL(urlID, "")
	if err != nil {
		t.Fatal(err)
	}
	if longURL != message.LongUrl {
		t.Fatalf("expected %q, got %q", message.LongUrl, longURL)
	}
}

func TestFindShortURLFailures(t *testing.T) {
	shortURLService := newTestService(t, newFakeStore())

	privateID, err := shortURLService.CreateShortURL(&proto.ShortUrlMessage{
		LongUrl:  "https://example.com/private",
		Expire:   time.Now().Add(time.Minute).Unix(),
		Password: "abcd",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := shortURLService.FindLongURL(privateID, "wrong"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}

	expiredID, err := shortURLService.CreateShortURL(&proto.ShortUrlMessage{
		LongUrl: "https://example.com/expired",
		Expire:  time.Now().Add(-time.Second).Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := shortURLService.FindLongURL(expiredID, ""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteExpiredURLs(t *testing.T) {
	storage := newFakeStore()
	shortURLService := newTestService(t, storage)
	expired := &proto.ShortUrlMessage{LongUrl: "expired", Expire: time.Now().Add(-time.Second).Unix()}
	active := &proto.ShortUrlMessage{LongUrl: "active", Expire: time.Now().Add(time.Minute).Unix()}
	expiredData, _ := expired.Marshal()
	activeData, _ := active.Marshal()
	_ = storage.Put([]byte("s-expired"), expiredData)
	_ = storage.Put([]byte("s-active"), activeData)
	_ = storage.Put([]byte(urlIDCounterKey), []byte(initialURLID))

	if err := shortURLService.deleteExpiredURLs(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.Get([]byte("s-expired")); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected expired entry to be deleted, got %v", err)
	}
	if _, err := storage.Get([]byte("s-active")); err != nil {
		t.Fatalf("expected active entry to remain: %v", err)
	}
	if _, err := storage.Get([]byte(urlIDCounterKey)); err != nil {
		t.Fatalf("expected counter to remain: %v", err)
	}
}

func TestDeleteExpiredURLsInBoundedBatches(t *testing.T) {
	storage := newFakeStore()
	shortURLService := newTestService(t, storage)
	expiredData, err := (&proto.ShortUrlMessage{
		LongUrl: "expired",
		Expire:  time.Now().Add(-time.Second).Unix(),
	}).Marshal()
	if err != nil {
		t.Fatal(err)
	}

	const remainder = 17
	for i := 0; i < cleanupBatchSize*2+remainder; i++ {
		if err := storage.Put([]byte("s-expired-"+strconv.Itoa(i)), expiredData); err != nil {
			t.Fatal(err)
		}
	}

	if err := shortURLService.deleteExpiredURLs(context.Background()); err != nil {
		t.Fatal(err)
	}

	storage.mu.RLock()
	batchSizes := append([]int(nil), storage.deleteBatchSizes...)
	storage.mu.RUnlock()
	want := []int{cleanupBatchSize, cleanupBatchSize, remainder}
	if len(batchSizes) != len(want) {
		t.Fatalf("expected batch sizes %v, got %v", want, batchSizes)
	}
	for i := range want {
		if batchSizes[i] != want[i] {
			t.Fatalf("expected batch sizes %v, got %v", want, batchSizes)
		}
	}
}

func TestDeleteExpiredURLsFromLevelDBInMultipleBatches(t *testing.T) {
	storage, err := store.Open(filepath.Join(t.TempDir(), "store"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			t.Error(err)
		}
	}()

	shortURLService := newTestService(t, storage)
	expiredData, err := (&proto.ShortUrlMessage{
		LongUrl: "expired",
		Expire:  time.Now().Add(-time.Second).Unix(),
	}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < cleanupBatchSize*2+17; i++ {
		if err := storage.Put([]byte("s-expired-"+strconv.Itoa(i)), expiredData); err != nil {
			t.Fatal(err)
		}
	}

	if err := shortURLService.deleteExpiredURLs(context.Background()); err != nil {
		t.Fatal(err)
	}
	remaining := 0
	if err := storage.Iterate([]byte(shortURLPrefix), func(key, value []byte) error {
		remaining++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("expected all expired entries to be deleted, got %d", remaining)
	}
}

func TestDeleteExpiredURLsStopsWithCanceledContext(t *testing.T) {
	storage := newFakeStore()
	shortURLService := newTestService(t, storage)
	expired := &proto.ShortUrlMessage{LongUrl: "expired", Expire: time.Now().Add(-time.Second).Unix()}
	expiredData, _ := expired.Marshal()
	_ = storage.Put([]byte("s-expired"), expiredData)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := shortURLService.deleteExpiredURLs(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if _, err := storage.Get([]byte("s-expired")); err != nil {
		t.Fatalf("expected expired entry to remain after cancellation: %v", err)
	}
}

func TestExistingDataFormatRemainsCompatible(t *testing.T) {
	storage := newFakeStore()
	legacy := &proto.ShortUrlMessage{LongUrl: "https://example.com/legacy", Expire: time.Now().Add(time.Minute).Unix()}
	legacyData, _ := legacy.Marshal()
	_ = storage.Put([]byte("s-abcd"), legacyData)
	_ = storage.Put([]byte("URL-ID"), []byte("46656"))

	shortURLService := newTestService(t, storage)
	longURL, err := shortURLService.FindLongURL("abcd", "")
	if err != nil || longURL != legacy.LongUrl {
		t.Fatalf("read legacy entry: url=%q err=%v", longURL, err)
	}
	if _, err := shortURLService.CreateShortURL(&proto.ShortUrlMessage{LongUrl: "new"}); err != nil {
		t.Fatal(err)
	}
	counter, err := storage.Get([]byte("URL-ID"))
	if err != nil {
		t.Fatal(err)
	}
	value, err := strconv.ParseInt(string(counter), 10, 64)
	if err != nil || value <= 46656 {
		t.Fatalf("expected compatible counter greater than 46656, got %q", counter)
	}
}

func TestConcurrentCreateProducesUniqueIDs(t *testing.T) {
	shortURLService := newTestService(t, newFakeStore())
	const count = 200
	ids := make(chan string, count)
	errs := make(chan error, count)
	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := shortURLService.CreateShortURL(&proto.ShortUrlMessage{LongUrl: "test"})
			ids <- id
			errs <- err
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)

	seen := make(map[string]bool, count)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate URL ID %q", id)
		}
		seen[id] = true
	}
}

func TestEncodeURLID(t *testing.T) {
	if len(encodeChars) != 36 {
		t.Fatalf("expected 36 encode characters, got %d", len(encodeChars))
	}
	characters := make(map[byte]bool, len(encodeChars))
	for i := range encodeChars {
		if characters[encodeChars[i]] {
			t.Fatalf("duplicate encode character %q", encodeChars[i])
		}
		characters[encodeChars[i]] = true
	}

	seen := make(map[string]bool)
	for id := int64(46656); id < 46656+100000; id++ {
		urlID := encodeURLID(id)
		if len(urlID) < 4 || len(urlID) > 8 || !URLIDRegexp.MatchString(urlID) {
			t.Fatalf("invalid URL ID %q", urlID)
		}
		if seen[urlID] {
			t.Fatalf("duplicate URL ID %q", urlID)
		}
		seen[urlID] = true
	}
}

func TestRunExpirationCleanupStopsWithContext(t *testing.T) {
	shortURLService := newTestService(t, newFakeStore())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		shortURLService.RunExpirationCleanup(ctx, time.Hour)
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expiration cleanup did not stop after context cancellation")
	}
}

func TestRunExpirationCleanupRunsImmediately(t *testing.T) {
	storage := newFakeStore()
	storage.deleteNotify = make(chan struct{}, 1)
	expiredData, err := (&proto.ShortUrlMessage{
		LongUrl: "expired",
		Expire:  time.Now().Unix(),
	}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := storage.Put([]byte("s-expired"), expiredData); err != nil {
		t.Fatal(err)
	}

	shortURLService := newTestService(t, storage)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		shortURLService.RunExpirationCleanup(ctx, time.Hour)
	}()

	select {
	case <-storage.deleteNotify:
	case <-time.After(time.Second):
		t.Fatal("expiration cleanup did not run immediately")
	}
	cancel()
	<-done
	if _, err := storage.Get([]byte("s-expired")); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected expired entry to be deleted, got %v", err)
	}
}

func TestRunExpirationCleanupRejectsNonPositiveInterval(t *testing.T) {
	shortURLService := newTestService(t, newFakeStore())
	shortURLService.RunExpirationCleanup(context.Background(), 0)
	shortURLService.RunExpirationCleanup(context.Background(), -time.Second)
}
