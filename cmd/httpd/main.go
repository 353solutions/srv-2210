package main

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/353solutions/unter/db"
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

type Server struct {
	db *db.DB
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO:
	fmt.Fprintln(w, "OK")
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
	/* FIXME
	rd, err := s.db.Get(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rd.Distance = req.Distance
	rd.End = time.Now().UTC()
	s.db.Add(rd)
	*/

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

// GET /rides/<id>
func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	rd, err := s.db.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
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

type idKeyType int

var idKey idKeyType = 1

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
func addLogging(h http.Handler) http.Handler {
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

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("ERROR: can't load config - %s", err)
		os.Exit(1)
	}
	log.Printf("INFO: config=%#v", cfg)
	r := mux.NewRouter()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	db, err := db.Connect(ctx, cfg.DSN)
	if err != nil {
		log.Printf("ERROR: can't connect to database- %s", err)
		os.Exit(1)
	}

	s := Server{
		db: db, // injection
	}
	// routing
	// - if route ends with / it's a prefix match
	// - otherwise exact match
	r.HandleFunc("/health", s.healthHandler).Methods("GET")
	r.HandleFunc("/rides", s.startHandler).Methods("POST")
	r.HandleFunc("/rides/{id}", s.getHandler).Methods("GET")
	r.HandleFunc("/rides/{id}/end", s.endHandler).Methods("POST")
	http.Handle("/", addLogging(r))

	srv := http.Server{
		Addr:    cfg.Addr,
		Handler: http.DefaultServeMux,
	}

	log.Printf("INFO: server starting on %s", cfg.Addr)
	errCh := make(chan error)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	exitCode := 0
	select {
	case sig := <-sigCh:
		log.Printf("INFO: caught signal %v, shutting down", sig)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("ERROR: %s", err)
			exitCode = 1
		}
	}

	log.Printf("INFO: server down")
	// cleanup
	// s.db.Close() ...
	os.Exit(exitCode)
}
