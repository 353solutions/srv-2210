package main

import (
	"context"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/353solutions/unter"
	"github.com/353solutions/unter/cache"
	"github.com/353solutions/unter/db"
	"github.com/353solutions/unter/logger"
)

/* CRUD: Create, Retrieve, Update, Delete
- C: POST
- R: GET
- U: PUT, PATCH
- D: DELETE
*/

/*
API:
POST /rides
	car_id
    driver
    kind (private, shared)
-> ID

POST /rides/{id}/end
    distance

GET /rides/{id}

GET /rides?start=<time>&end=<time>
*/

var (
	getCalls = expvar.NewInt("get.calls")
)

type Server struct {
	db    *db.DB
	cache *cache.Cache
	log   *log.Logger
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	ok := true
	resp := map[string]any{
		"db":    nil,
		"cache": nil,
	}
	if err := s.db.Health(r.Context()); err != nil {
		resp["db"] = err.Error()
		ok = false
	}
	if err := s.cache.Health(r.Context()); err != nil {
		resp["cache"] = err.Error()
		ok = false
	}

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
	}
	sendJSON(w, resp)
}

func kindFromString(s string) (unter.Kind, error) {
	switch s {
	case unter.Shared.String():
		return unter.Shared, nil
	case unter.Private.String():
		return unter.Private, nil
	}

	return 0, fmt.Errorf("unknown kind: %s", s)
}

func (s *Server) startHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Unmarshal & Validate data
	// {"driver": "Bond", "kind": "private"}
	var req struct {
		Driver string
		Kind   string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	k, err := kindFromString(req.Kind)
	if err != nil {
		http.Error(w, "bad kind", http.StatusBadRequest)
		return
	}

	rd := unter.Ride{
		ID:     unter.NewID(),
		Driver: req.Driver,
		Kind:   k,
		Start:  time.Now().UTC(),
	}
	if err := rd.Validate(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Step 2: Work
	// FIXME s.db.Add(rd)
	dbr := db.Ride{
		ID:     rd.ID,
		Driver: rd.Driver,
		Kind:   rd.Kind.String(),
		Start:  rd.Start,
	}
	if err := s.db.Add(r.Context(), dbr); err != nil {
		http.Error(w, "can't insert", http.StatusInternalServerError)
		return
	}

	// Step 3: Marshal & send response
	resp := map[string]any{
		"id":     rd.ID,
		"action": "start",
	}
	/*
		 if err := json.NewEncoder(w).Encode(resp); err != nil {
			// Can't change response code
		}
	*/

	//	w.WriteHeader(http.StatusCreated)
	if err := sendJSON(w, resp); err != nil {
		http.Error(w, "can't marshal to JSON", http.StatusInternalServerError)
		return

	}
}

/*
Exercise:
POST /rides/{id}/end

	{"distance": 1.3}
*/
func (s *Server) endHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Distance float64
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if req.Distance == 0 {
		http.Error(w, "missing distance", http.StatusBadRequest)
		return

	}

	vars := mux.Vars(r)
	id := vars["id"]
	rd, err := s.db.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rd.Distance = req.Distance
	rd.End = time.Now().UTC()
	s.db.Update(r.Context(), rd)
	// TODO: invalidate cache

	resp := map[string]any{
		"id":     id,
		"action": "end",
	}

	if err := sendJSON(w, resp); err != nil {
		http.Error(w, "can't marshal to JSON", http.StatusInternalServerError)
		return
	}
}

// any = interface{} (go >= 1.18)
func sendJSON(w http.ResponseWriter, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	return nil
}

type GetResponse struct {
	ID       string     `json:"id,omitempty"`
	Driver   string     `json:"driver,omitempty"`
	Kind     string     `json:"kind,omitempty"`
	Start    time.Time  `json:"start,omitempty"`
	End      *time.Time `json:"end,omitempty"`
	Distance float64    `json:"distance,omitempty"`
}

func ctxLogger(logger *log.Logger, ctx context.Context) *log.Logger {
	rid := RequestID(ctx)
	return log.New(
		logger.Writer(),
		fmt.Sprintf("%s <%s> ", logger.Prefix(), rid),
		logger.Flags(),
	)
}

// GET /rides/<id>
func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	// exercise: Add a middleware that will expose metircs for number of calls, number of errors
	// extra: total bytes sent
	getCalls.Add(1)
	vars := mux.Vars(r)
	id := vars["id"]

	log := ctxLogger(s.log, r.Context())

	data, err := s.cache.Get(r.Context(), id)
	if err == nil {
		log.Printf("INFO: cache hit - %s", id)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}

	rd, err := s.db.Get(r.Context(), id)
	switch {
	case errors.Is(err, db.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, "can't get", http.StatusInternalServerError)
		return
	}

	resp := GetResponse{
		ID:       rd.ID,
		Driver:   rd.Driver,
		Kind:     rd.Kind,
		Start:    rd.Start,
		Distance: rd.Distance,
	}
	if !rd.End.Equal(time.Time{}) {
		resp.End = &rd.End
	}

	data, err = json.Marshal(resp)
	if err == nil {
		s.cache.Set(r.Context(), id, data)
	}

	if err := sendJSON(w, resp); err != nil {
		http.Error(w, "can't marshal to JSON", http.StatusInternalServerError)
		return
	}
}

/* encoding/json
Go -> JSON []byte: Marshal
JSON -> Go []byte: Unmarshal
Go -> JSON io.Writer: Encoder
JSON -> Go io.Reader: Decoder
*/

type keyType int

var idKey keyType = 1

func RequestID(ctx context.Context) string {
	rid := ctx.Value(idKey)
	if rid == nil {
		return "XXX"
	}

	s, ok := rid.(string)
	if !ok {
		return "XXX"
	}
	return s
}

// middleware
func addLogging(log *log.Logger, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// before
		rid := uuid.NewString()
		ctx := context.WithValue(r.Context(), idKey, rid)
		r = r.Clone(ctx)

		log.Printf("%s called (rid = %s)", r.URL.Path, rid)
		start := time.Now()

		h.ServeHTTP(w, r)

		// after
		duration := time.Since(start)
		// exercise: Log the return HTTP status
		log.Printf("%s ended in %v (rid = %s)", r.URL.Path, duration, rid)
	}

	return http.HandlerFunc(fn)
}

var version = "1.2.3"

func buildRouter(s *Server) *http.ServeMux {
	r := mux.NewRouter()
	r.HandleFunc("/health", s.healthHandler).Methods("GET")
	r.HandleFunc("/rides", s.startHandler).Methods("POST")
	r.HandleFunc("/rides/{id}", s.getHandler).Methods("GET")
	r.HandleFunc("/rides/{id}/end", s.endHandler).Methods("POST")

	mux := http.NewServeMux()
	mux.Handle("/", addLogging(s.log, r))
	return mux
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("ERROR: can't load config - %s", err)
		os.Exit(1)
	}

	expvar.NewString("version").Set(version)
	host, err := os.Hostname()
	if err != nil {
		host = "UNKNOWN"
	}
	expvar.NewString("host").Set(host)
	// service name....

	logger, err := logger.New("[UNTER] ", cfg.LogFile)
	if err != nil {
		log.Printf("ERROR: can't load logger - %s", err)
		os.Exit(1)

	}

	logger.Printf("INFO: config=%#v", cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	db, err := db.Connect(ctx, cfg.DSN)
	if err != nil {
		logger.Printf("ERROR: can't connect to database - %s", err)
		os.Exit(1)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cache, err := cache.Connect(ctx, cfg.CacheAddr, time.Minute)
	if err != nil {
		logger.Printf("ERROR: can't connect to cache - %s", err)
		os.Exit(1)
	}

	s := Server{
		db:    db, // injection
		cache: cache,
		log:   logger,
	}
	// routing
	// - if route ends with / it's a prefix match
	// - otherwise exact match
	mux := buildRouter(&s)

	srv := http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	logger.Printf("INFO: server starting on %s", cfg.Addr)
	errCh := make(chan error)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	exitCode := 0
	select {
	case sig := <-sigCh:
		logger.Printf("INFO: caught signal %v, shutting down", sig)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("ERROR: %s", err)
			exitCode = 1
		}
	}

	logger.Printf("INFO: server down")
	// cleanup
	// s.db.Close() ...
	os.Exit(exitCode)
}
