package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/353solutions/unter/cache"
	"github.com/353solutions/unter/db"

	"github.com/stretchr/testify/require"
)

func setupServer(t *testing.T) *Server {
	require := require.New(t)

	cfg, err := loadConfig()
	require.NoError(err, "config")
	db, err := db.Connect(context.Background(), cfg.DSN)
	require.NoError(err, "db")
	t.Cleanup(func() { db.Close() })

	cache, err := cache.Connect(context.Background(), cfg.CacheAddr, time.Second)
	require.NoError(err, "cache")
	t.Cleanup(func() { cache.Close() })

	s := Server{
		db:    db,
		cache: cache,
		log:   log.Default(),
	}
	return &s
}

func Test_healthHandler(t *testing.T) {
	require := require.New(t)
	s := setupServer(t)
	mux := buildRouter(s)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)

	// Bypass routing, might be OK in some cases
	// s.healthHandler(w, r)
	h, p := mux.Handler(r)
	t.Logf("pattern: %v", p)
	h.ServeHTTP(w, r)

	resp := w.Result()
	require.Equal(http.StatusOK, resp.StatusCode)

	var reply struct {
		DB    any
		Cache any
	}
	err := json.NewDecoder(resp.Body).Decode(&reply)
	require.NoError(err, "decode json")
	require.Equal(nil, reply.DB, "db")
	require.Equal(nil, reply.DB, "cache")
}

// exercise:
func Test_startHandler(t *testing.T) {
	require := require.New(t)
	s := setupServer(t)

	w := httptest.NewRecorder()
	// buf := strings.NewReader(`{"driver": "q", "kind": "start"}`)
	msg := map[string]any{
		"driver": "q",
		"kind":   "shared",
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(msg)
	require.NoError(err, "json encode")
	r := httptest.NewRequest(http.MethodPost, "/rides", &buf)
	// r.Header.Set()
	// r.BasicAuth("bond", "007")

	// Bypass routing, might be OK in some cases
	s.startHandler(w, r)

	resp := w.Result()
	require.Equal(http.StatusOK, resp.StatusCode)

	var reply struct {
		ID     string
		Action string
	}
	err = json.NewDecoder(resp.Body).Decode(&reply)
	require.NoError(err, "decode json")
	require.NotEmpty(reply.ID, "id")
	require.Equal("start", reply.Action, "action")
}
