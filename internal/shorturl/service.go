package shorturl

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/binary"
	stderrors "errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/tianyk/short-url/internal/store"
	"github.com/tianyk/short-url/proto"
)

const (
	// urlIDCounterKey stores the last allocated numeric URL ID.
	urlIDCounterKey = "URL-ID"
	// initialURLID (= 36^3) is used when no persisted URL ID counter exists;
	// it keeps encoded codes at least 4 characters long.
	initialURLID = "46656"
	// shortURLPrefix separates short URL records from other store entries; changing it breaks existing data lookup.
	shortURLPrefix = "s-"
	// cleanupBatchSize limits the number of expired keys deleted per write batch.
	cleanupBatchSize = 1000
	// digitMap permutes decimal digits before encoding and must remain stable for ID compatibility.
	digitMap = "0415263897"
	// encodeChars defines the persisted base-36 alphabet and must remain stable for ID compatibility.
	encodeChars = "9c5rwjupmqvabs8kgz6o0ydi2xtl3n4efh17"
)

var (
	ErrNotFound  = stderrors.New("short URL not found")
	ErrForbidden = stderrors.New("short URL password is invalid")

	URLIDRegexp = regexp.MustCompile("^[0-9a-z]+$")
)

// randomOffsets preserves the existing short-code allocation algorithm.
var randomOffsets = [...]int64{
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	11, 13, 17, 19, 23, 29, 31, 37, 41, 43,
}

// ShortURLService contains short URL business logic and depends only on Store.
type ShortURLService struct {
	storage   store.Store
	idMu      sync.Mutex
	urlIDRand *rand.Rand
}

// New explicitly initializes a short URL service.
func New(storage store.Store) (*ShortURLService, error) {
	if storage == nil {
		return nil, fmt.Errorf("store is required")
	}

	var seedBytes [8]byte
	if _, err := cryptorand.Read(seedBytes[:]); err != nil {
		return nil, fmt.Errorf("seed URL ID generator: %w", err)
	}

	return &ShortURLService{
		storage:   storage,
		urlIDRand: rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(seedBytes[:])))),
	}, nil
}

func shortURLKey(urlID string) string {
	return shortURLPrefix + urlID
}

func substituteDigits(id int64) int64 {
	s := strconv.FormatInt(id, 10)
	b := []byte(s)
	for i, c := range b {
		b[i] = digitMap[c-'0']
	}
	mapped, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return id
	}
	return mapped
}

func encodeURLID(id int64) string {
	mapped := substituteDigits(id)
	var buf [16]byte
	i := len(buf)
	for mapped > 0 {
		i--
		buf[i] = encodeChars[mapped%36]
		mapped /= 36
	}
	s := string(buf[i:])
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}

func (s *ShortURLService) newID() (int64, error) {
	s.idMu.Lock()
	defer s.idMu.Unlock()

	currentIDBytes, err := s.storage.Get([]byte(urlIDCounterKey))
	if err != nil {
		if stderrors.Is(err, store.ErrNotFound) {
			currentIDBytes = []byte(initialURLID)
		} else {
			return 0, errors.Wrap(err, "get last URL ID")
		}
	}

	currentID, err := strconv.ParseInt(string(currentIDBytes), 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "parse last URL ID")
	}

	offset := randomOffsets[s.urlIDRand.Intn(len(randomOffsets))]
	nextID := currentID + offset
	if err := s.storage.Put([]byte(urlIDCounterKey), []byte(strconv.FormatInt(nextID, 10))); err != nil {
		return 0, errors.Wrap(err, "set next URL ID")
	}

	return nextID, nil
}

// CreateShortURL creates and persists a short URL.
func (s *ShortURLService) CreateShortURL(input *proto.ShortUrlMessage) (string, error) {
	id, err := s.newID()
	if err != nil {
		return "", err
	}

	urlID := encodeURLID(id)
	log.Printf("[%s <=> %s] %d", urlID, input.LongUrl, input.Expire)

	data, err := input.Marshal()
	if err != nil {
		return "", errors.Wrap(err, "marshal short URL")
	}
	if err := s.storage.Put([]byte(shortURLKey(urlID)), data); err != nil {
		return "", errors.Wrap(err, "set short URL")
	}

	return urlID, nil
}

// FindLongURL returns the destination associated with a short URL.
func (s *ShortURLService) FindLongURL(urlID, password string) (string, error) {
	cacheKey := []byte(shortURLKey(urlID))
	data, err := s.storage.Get(cacheKey)
	if err != nil {
		if stderrors.Is(err, store.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", errors.Wrap(err, "get long URL")
	}

	msg := new(proto.ShortUrlMessage)
	if err := msg.Unmarshal(data); err != nil {
		return "", errors.Wrap(err, "unmarshal short URL")
	}

	if msg.Password != "" && msg.Password != password {
		return "", ErrForbidden
	}

	if now := time.Now().Unix(); msg.Expire > 0 && msg.Expire <= now {
		log.Printf("expire %s", urlID)
		_ = s.storage.Delete(cacheKey)
		return "", ErrNotFound
	}

	return msg.LongUrl, nil
}

// deleteExpiredURLs removes expired short URLs until ctx is canceled.
func (s *ShortURLService) deleteExpiredURLs(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	now := time.Now().Unix()
	expiredKeys := make([][]byte, 0, cleanupBatchSize)
	deletedCount := 0
	flush := func() error {
		if len(expiredKeys) == 0 {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.storage.DeleteMany(expiredKeys); err != nil {
			return errors.Wrap(err, "delete expired short URLs")
		}
		deletedCount += len(expiredKeys)
		expiredKeys = expiredKeys[:0]
		return nil
	}

	err := s.storage.Iterate([]byte(shortURLPrefix), func(key, value []byte) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		msg := new(proto.ShortUrlMessage)
		if err := msg.Unmarshal(value); err != nil {
			log.Printf("skip malformed short URL record key=%s: %s", key, err)
			return nil
		}
		if msg.Expire > 0 && msg.Expire <= now {
			expiredKeys = append(expiredKeys, append([]byte(nil), key...))
			if len(expiredKeys) == cleanupBatchSize {
				return flush()
			}
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "collect expired short URLs")
	}
	if err := flush(); err != nil {
		return err
	}

	if deletedCount > 0 {
		log.Printf("expired URL cleanup complete: deleted=%d", deletedCount)
	}
	return nil
}

// RunExpirationCleanup periodically deletes expired short URLs until ctx is canceled.
func (s *ShortURLService) RunExpirationCleanup(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		log.Printf("expired URL cleanup disabled: interval must be positive, got %s", interval)
		return
	}

	log.Printf("start expired URL cleanup")
	defer log.Printf("stop expired URL cleanup")

	for {
		if err := s.deleteExpiredURLs(ctx); err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("expired URL cleanup failed: %s", err)
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-timer.C:
		}
	}
}
