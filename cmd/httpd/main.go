package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/353solutions/unter"
	"github.com/gorilla/mux"
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
	db = NewDB()
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
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

func startHandler(w http.ResponseWriter, r *http.Request) {
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
	db.Add(rd)

	// Step 3: Marshal & send response
	resp := map[string]any{
		"id": rd.ID,
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
func getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	rd, err := db.Get(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	resp := GetResponse{
		ID:       rd.ID,
		Driver:   rd.Driver,
		Kind:     rd.Kind.String(),
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

func main() {
	r := mux.NewRouter()
	// routing
	// - if route ends with / it's a prefix match
	// - otherwise exact match
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/rides", startHandler).Methods("POST")
	r.HandleFunc("/rides/{id}", getHandler).Methods("GET")
	http.Handle("/", r)

	addr := ":8080"
	log.Printf("INFO: server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("ERROR: can't start - %s", err)
		os.Exit(1)
		// log.Fatalf("ERROR: can't start - %s", err)
	}
}
