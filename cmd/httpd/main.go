package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/353solutions/unter"
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
*/

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO:
	fmt.Fprintln(w, "OK")
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	id := unter.NewID()
	fmt.Fprintln(w, id)
}

func main() {
	// routing
	// - if route ends with / it's a prefix match
	// - otherwise exact match
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/rides", startHandler)

	addr := ":8080"
	log.Printf("INFO: server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("ERROR: can't start - %s", err)
		os.Exit(1)
		// log.Fatalf("ERROR: can't start - %s", err)
	}
}
