package shorturl

import (
	"io"
	"log"
	"path/filepath"
	"testing"

	"github.com/tianyk/short-url/internal/store"
	"github.com/tianyk/short-url/proto"
)

func benchmarkService(b *testing.B, storage store.Store) *ShortURLService {
	b.Helper()
	service, err := New(storage)
	if err != nil {
		b.Fatal(err)
	}
	return service
}

func benchmarkLevelDBService(b *testing.B) *ShortURLService {
	b.Helper()
	storage, err := store.OpenLevelDB(filepath.Join(b.TempDir(), "store"))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		if err := storage.Close(); err != nil {
			b.Error(err)
		}
	})
	return benchmarkService(b, storage)
}

func runNewIDBenchmark(b *testing.B, service *ShortURLService, parallel bool) {
	b.Helper()
	b.ResetTimer()
	if parallel {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if _, err := service.newID(); err != nil {
					b.Error(err)
				}
			}
		})
		return
	}
	for i := 0; i < b.N; i++ {
		if _, err := service.newID(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewID(b *testing.B) {
	b.Run("MemoryStore/Sequential", func(b *testing.B) {
		runNewIDBenchmark(b, benchmarkService(b, newFakeStore()), false)
	})
	b.Run("MemoryStore/Parallel", func(b *testing.B) {
		runNewIDBenchmark(b, benchmarkService(b, newFakeStore()), true)
	})
	b.Run("LevelDB/Sequential", func(b *testing.B) {
		runNewIDBenchmark(b, benchmarkLevelDBService(b), false)
	})
	b.Run("LevelDB/Parallel", func(b *testing.B) {
		runNewIDBenchmark(b, benchmarkLevelDBService(b), true)
	})
}

func runCreateBenchmark(b *testing.B, service *ShortURLService, parallel bool) {
	b.Helper()
	message := &proto.ShortUrlMessage{LongUrl: "https://example.com/benchmark"}
	b.ResetTimer()
	if parallel {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if _, err := service.CreateShortURL(message); err != nil {
					b.Error(err)
				}
			}
		})
		return
	}
	for i := 0; i < b.N; i++ {
		if _, err := service.CreateShortURL(message); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateShortURL(b *testing.B) {
	previousOutput := log.Writer()
	log.SetOutput(io.Discard)
	b.Cleanup(func() { log.SetOutput(previousOutput) })

	b.Run("MemoryStore/Sequential", func(b *testing.B) {
		runCreateBenchmark(b, benchmarkService(b, newFakeStore()), false)
	})
	b.Run("MemoryStore/Parallel", func(b *testing.B) {
		runCreateBenchmark(b, benchmarkService(b, newFakeStore()), true)
	})
	b.Run("LevelDB/Sequential", func(b *testing.B) {
		runCreateBenchmark(b, benchmarkLevelDBService(b), false)
	})
	b.Run("LevelDB/Parallel", func(b *testing.B) {
		runCreateBenchmark(b, benchmarkLevelDBService(b), true)
	})
}
