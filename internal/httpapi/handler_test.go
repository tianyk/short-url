package httpapi

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tianyk/short-url/internal/config"
	"github.com/tianyk/short-url/internal/shorturl"
	"github.com/tianyk/short-url/internal/store"
	"github.com/tianyk/short-url/proto"
)

type fakeStaticMapping map[string]string

func (m fakeStaticMapping) Lookup(key string) (string, bool) {
	value, ok := m[key]
	return value, ok
}

func TestStaticMappingRunsBeforeStoredURLValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storage, err := store.Open(filepath.Join(t.TempDir(), "store"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			t.Error(err)
		}
	}()

	service, err := shorturl.New(storage)
	if err != nil {
		t.Fatal(err)
	}
	storedID, err := service.CreateShortURL(&proto.ShortUrlMessage{
		LongUrl: "https://example.com/from-leveldb",
		Expire:  time.Now().Add(time.Minute).Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}

	mapping := fakeStaticMapping{
		"me":           "https://example.com/from-static",
		"release-2026": "进度 100%",
		storedID:       "https://example.com/override",
		"_heartbeat":   "must not override the exact route",
	}
	controller := NewShortURLController("https://short.example", service, mapping)
	router := gin.New()
	router.NoRoute(NotFoundHandler)
	RegisterRoutes(router, config.Config{}, controller)

	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantBody     string
		wantLocation string
	}{
		{
			name:         "short static key bypasses stored URL validation",
			path:         "/me",
			wantStatus:   http.StatusMovedPermanently,
			wantLocation: "https://example.com/from-static",
		},
		{
			name:       "static text is returned verbatim",
			path:       "/release-2026",
			wantStatus: http.StatusOK,
			wantBody:   "进度 100%",
		},
		{
			name:         "static mapping overrides LevelDB",
			path:         "/" + storedID,
			wantStatus:   http.StatusMovedPermanently,
			wantLocation: "https://example.com/override",
		},
		{
			name:       "static miss continues to original validation",
			path:       "/no",
			wantStatus: http.StatusNotFound,
			wantBody:   "NotFound",
		},
		{
			name:       "exact heartbeat route takes priority",
			path:       "/_heartbeat",
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("expected status %d, got %d", test.wantStatus, response.Code)
			}
			if test.wantBody != "" && response.Body.String() != test.wantBody {
				t.Fatalf("expected body %q, got %q", test.wantBody, response.Body.String())
			}
			if location := response.Header().Get("Location"); location != test.wantLocation {
				t.Fatalf("expected Location %q, got %q", test.wantLocation, location)
			}
		})
	}

	delete(mapping, storedID)
	request := httptest.NewRequest(http.MethodGet, "/"+storedID, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusMovedPermanently {
		t.Fatalf("expected LevelDB fallback status %d, got %d", http.StatusMovedPermanently, response.Code)
	}
	if location := response.Header().Get("Location"); location != "https://example.com/from-leveldb" {
		t.Fatalf("expected LevelDB fallback Location, got %q", location)
	}
}
