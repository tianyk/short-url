// Package staticmap loads manually maintained short URL mappings from disk.
package staticmap

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/fsnotify/fsnotify"
)

const (
	// Filename is the fixed mapping filename expected in the LevelDB data directory.
	Filename = "short-url-map.conf"

	maxFileSize = 1024 * 1024
	reloadDelay = 200 * time.Millisecond
)

type entries map[string]string

// FileMap is an immutable, atomically replaced snapshot of the mapping file.
type FileMap struct {
	path    string
	current atomic.Pointer[entries]
}

// New loads path if it exists. A missing file produces an empty mapping.
func New(path string) (*FileMap, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve static mapping path: %w", err)
	}

	m := &FileMap{path: filepath.Clean(absolutePath)}
	empty := make(entries)
	m.current.Store(&empty)

	present, count, err := m.reload()
	if err != nil {
		return nil, err
	}
	if present {
		log.Printf("loaded static short URL mappings: path=%s entries=%d", m.path, count)
	} else {
		log.Printf("static short URL mapping file not found; LevelDB-only lookup remains active: path=%s", m.path)
	}
	return m, nil
}

// Path returns the absolute mapping file path.
func (m *FileMap) Path() string {
	return m.path
}

// Lookup returns a value from the latest successfully loaded snapshot.
func (m *FileMap) Lookup(key string) (string, bool) {
	snapshot := m.current.Load()
	if snapshot == nil {
		return "", false
	}
	value, ok := (*snapshot)[key]
	return value, ok
}

// StartWatching registers the directory watcher before returning and reloads until ctx is canceled.
func (m *FileMap) StartWatching(ctx context.Context) (<-chan error, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create static mapping watcher: %w", err)
	}
	if err := watcher.Add(filepath.Dir(m.path)); err != nil {
		_ = watcher.Close()
		return nil, fmt.Errorf("watch static mapping directory: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		defer close(done)
		defer watcher.Close()
		done <- m.watch(ctx, watcher)
	}()
	return done, nil
}

func (m *FileMap) watch(ctx context.Context, watcher *fsnotify.Watcher) error {
	var timer *time.Timer
	var timerC <-chan time.Time
	scheduleReload := func() {
		if timer == nil {
			timer = time.NewTimer(reloadDelay)
		} else {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(reloadDelay)
		}
		timerC = timer.C
	}
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return errors.New("static mapping watcher stopped unexpectedly")
			}
			if filepath.Clean(event.Name) == m.path && event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				scheduleReload()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return errors.New("static mapping watcher error channel closed unexpectedly")
			}
			log.Printf("static mapping watcher error: %s", err)
		case <-timerC:
			timerC = nil
			present, count, err := m.reload()
			if err != nil {
				log.Printf("reload static short URL mappings failed; keeping previous mappings: %s", err)
				continue
			}
			if present {
				log.Printf("reloaded static short URL mappings: path=%s entries=%d", m.path, count)
			} else {
				log.Printf("static short URL mapping file removed; cleared static mappings: path=%s", m.path)
			}
		}
	}
}

func (m *FileMap) reload() (present bool, count int, err error) {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			empty := make(entries)
			m.current.Store(&empty)
			return false, 0, nil
		}
		return false, 0, fmt.Errorf("read static mapping file %s: %w", m.path, err)
	}
	if len(data) > maxFileSize {
		return true, 0, fmt.Errorf("static mapping file %s exceeds %d bytes", m.path, maxFileSize)
	}
	if !utf8.Valid(data) {
		return true, 0, fmt.Errorf("static mapping file %s is not valid UTF-8", m.path)
	}

	parsed, err := parse(data)
	if err != nil {
		return true, 0, fmt.Errorf("parse static mapping file %s: %w", m.path, err)
	}
	m.current.Store(&parsed)
	return true, len(parsed), nil
}

func parse(data []byte) (entries, error) {
	result := make(entries)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Buffer(make([]byte, 64*1024), maxFileSize)

	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		if lineNumber == 1 {
			line = strings.TrimPrefix(line, "\ufeff")
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		separator := strings.IndexByte(line, '=')
		if separator < 0 {
			return nil, fmt.Errorf("line %d: expected key=value", lineNumber)
		}
		key := strings.TrimSpace(line[:separator])
		value := strings.TrimSpace(line[separator+1:])
		if !validKey(key) {
			return nil, fmt.Errorf("line %d: key %q must contain 4-8 lowercase letters or digits", lineNumber, key)
		}
		if _, exists := result[key]; exists {
			return nil, fmt.Errorf("line %d: duplicate key %q", lineNumber, key)
		}
		result[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan mapping file: %w", err)
	}
	return result, nil
}

func validKey(key string) bool {
	if len(key) < 4 || len(key) > 8 {
		return false
	}
	for _, character := range key {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}
