package staticmap

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tianyk/short-url/internal/store"
)

func TestParse(t *testing.T) {
	parsed, err := parse([]byte("\ufeff# comment\n\nme = https://example.com/search?q=test&type=all#result\nMyPage=系统维护中\nrelease-2026=download\nempty=\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"me":           "https://example.com/search?q=test&type=all#result",
		"MyPage":       "系统维护中",
		"release-2026": "download",
		"empty":        "",
	}
	if len(parsed) != len(want) {
		t.Fatalf("expected %d entries, got %d", len(want), len(parsed))
	}
	for key, value := range want {
		if parsed[key] != value {
			t.Fatalf("expected %q=%q, got %q", key, value, parsed[key])
		}
	}
}

func TestParseRejectsInvalidLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "missing separator", content: "docs https://example.com"},
		{name: "space in key", content: "my page=value"},
		{name: "slash in key", content: "my/page=value"},
		{name: "unicode key", content: "我的=value"},
		{name: "dot segment", content: ".=value"},
		{name: "parent segment", content: "..=value"},
		{name: "api namespace", content: "api=value"},
		{name: "static namespace", content: "static=value"},
		{name: "heartbeat route", content: "_heartbeat=value"},
		{name: "too long", content: strings.Repeat("a", maxKeyLength+1) + "=value"},
		{name: "duplicate key", content: "docs=one\ndocs=two"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parse([]byte(test.content)); err == nil {
				t.Fatal("expected parse error")
			}
		})
	}
}

func TestNewAllowsMissingFile(t *testing.T) {
	mapping, err := New(filepath.Join(t.TempDir(), Filename))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := mapping.Lookup("docs"); ok {
		t.Fatal("expected missing mapping")
	}
}

func TestMappingFileCanShareLevelDBDirectory(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "store")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, Filename)
	if err := os.WriteFile(path, []byte("docs=https://example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	storage, err := store.Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			t.Error(err)
		}
	}()

	mapping, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	if value, ok := mapping.Lookup("docs"); !ok || value != "https://example.com" {
		t.Fatalf("expected mapping beside LevelDB, got value=%q ok=%v", value, ok)
	}
}

func TestWatchReloadsCreatesAndRemovesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), Filename)
	if err := os.WriteFile(path, []byte("docs=old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	mapping, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done, err := mapping.StartWatching(ctx)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cancel()
		if err := <-done; err != nil {
			t.Errorf("watcher stopped: %v", err)
		}
	})

	if err := os.WriteFile(path, []byte("docs=new\nnotice=hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	waitForValue(t, mapping, "docs", "new", true)
	waitForValue(t, mapping, "notice", "hello", true)

	// A malformed update must leave the last valid snapshot active.
	if err := os.WriteFile(path, []byte("invalid line\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	time.Sleep(reloadDelay * 2)
	if value, ok := mapping.Lookup("docs"); !ok || value != "new" {
		t.Fatalf("expected previous snapshot after invalid reload, got value=%q ok=%v", value, ok)
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	waitForValue(t, mapping, "docs", "", false)
}

func TestWatchLoadsFileCreatedAfterStartup(t *testing.T) {
	path := filepath.Join(t.TempDir(), Filename)
	mapping, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done, err := mapping.StartWatching(ctx)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cancel()
		if err := <-done; err != nil {
			t.Errorf("watcher stopped: %v", err)
		}
	})

	temporaryPath := filepath.Join(filepath.Dir(path), "replacement.tmp")
	if err := os.WriteFile(temporaryPath, []byte("docs=created\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		t.Fatal(err)
	}
	waitForValue(t, mapping, "docs", "created", true)
}

func waitForValue(t *testing.T, mapping *FileMap, key, want string, wantOK bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		value, ok := mapping.Lookup(key)
		if ok == wantOK && (!wantOK || value == want) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	value, ok := mapping.Lookup(key)
	t.Fatalf("timed out waiting for %q: want value=%q ok=%v, got value=%q ok=%v", key, want, wantOK, value, ok)
}

func TestNewRejectsOversizedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), Filename)
	if err := os.WriteFile(path, []byte(strings.Repeat("x", maxFileSize+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := New(path); err == nil {
		t.Fatal("expected oversized file error")
	}
}
