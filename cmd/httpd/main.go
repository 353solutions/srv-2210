package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	//go:embed html/info.html
	infoHTML     string
	infoTemplate *template.Template
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

// const maxMsgSize = 3_000_000 // 3MB

/* auth middleware
Writer -> POST, PUT
Reader -> GET
*/

func (s *Server) startHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Unmarshal & Validate data
	// {"driver": "Bond", "kind": "private"}
	var req struct {
		Driver string
		Kind   string
	}
	// rdr := http.MaxBytesReader(w, r.Body, maxMsgSize)
	// if err := json.NewDecoder(rdr).Decode(&req); err != nil {
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
		Driver: strings.ToValidUTF8(req.Driver, ""),
		Kind:   k,
		Start:  time.Now().UTC(),
	}
	if err := rd.Validate(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	v := RequestValues(r.Context())
	if v == nil || v.User.Login != rd.Driver {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if !HasRole(v.User, Writer, Admin) {
		http.Error(w, "not allowed", http.StatusUnauthorized)
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
	if req.Distance < 0 {
		http.Error(w, "negative distance", http.StatusBadRequest)
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

func (s *Server) infoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	rd, err := s.db.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// fee := unter.RideFee(rd.End.Sub(rd.Start), rd.Distance, rd.Kind == unter.Shared.String())
	w.Header().Set("Content-Type", "text/html")
	// exercise: replace printf with html/template
	// fmt.Fprintf(w, infoHTML, rd.ID, rd.Driver, rd.Start, rd.End, rd.Kind, rd.Distance, fee)
	infoTemplate.Execute(w, rd)
}

var version = "1.2.3"

func buildRouter(s *Server) *http.ServeMux {
	r := mux.NewRouter()
	r.HandleFunc("/health", s.healthHandler).Methods("GET")
	r.HandleFunc("/rides", s.startHandler).Methods("POST")
	r.HandleFunc("/rides/{id}", s.getHandler).Methods("GET")
	r.HandleFunc("/rides/{id}/end", s.endHandler).Methods("POST")
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/info/{id}", s.infoHandler)

	mux := http.NewServeMux()
	h := topMiddleware(s.log, r)
	h = http.MaxBytesHandler(h, 3_000_000)
	mux.Handle("/", h)
	return mux
}

func main() {
	var err error
	infoTemplate, err = template.New("info").Parse(infoHTML)
	if err != nil {
		log.Printf("ERROR: can't compile template - %s", err)
		os.Exit(1)
	}

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

	// logger.Printf("INFO: config=%#v", cfg)
	logger.Printf("INFO: config: Addr: %#v, CacheAddr: %#v, LogFile: %#v", cfg.Addr, cfg.CacheAddr, cfg.LogFile)

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
		Addr:         cfg.Addr,
		Handler:      mux,
		ReadTimeout:  time.Second,
		WriteTimeout: 2 * time.Second,
	}

	logger.Printf("INFO: server starting on %s", cfg.Addr)
	errCh := make(chan error)
	go func() {
		errCh <- srv.ListenAndServeTLS("cert.pem", "key.pem")
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
