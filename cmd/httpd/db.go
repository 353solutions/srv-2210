package main

import (
	"fmt"
	"sync"

	"github.com/353solutions/unter"
)

type DB struct {
	mu    sync.Mutex
	rides map[string]unter.Ride
}

func NewDB() *DB {
	db := DB{
		rides: make(map[string]unter.Ride),
	}
	return &db
}

func (db *DB) Add(r unter.Ride) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.rides[r.ID] = r
}

func (db *DB) Get(id string) (unter.Ride, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	r, ok := db.rides[id]
	if !ok {
		return unter.Ride{}, fmt.Errorf("ride %q not found", id)
	}

	return r, nil
}
