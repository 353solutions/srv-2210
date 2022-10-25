/*
Types:
- Relational
  - Tables columns & rows
  - SQL
  - ACID
  - PostgreSQL, MySQL/MariaDB, BigQuery, SQLite...

- Key/Value
  - Caching
  - Redis, memcached, ...

- Document
  - Collection/index of documents
  - MongoDB, ElasticSearch, ...

- Graph
  - Nodes/edges
  - Neo4j, dgraph, ...
*/
package db

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"time"

	_ "github.com/lib/pq"
)

var (
	//go:embed sql/insert.sql
	insertSQL string

	//go:embed sql/get.sql
	getSQL string

	//go:embed sql/update.sql
	updateSQL string
)

type DB struct {
	conn *sql.DB
}

func Connect(ctx context.Context, dsn string) (*DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db := DB{conn}
	if err := db.Health(ctx); err != nil {
		conn.Close()
		return nil, err
	}

	return &db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Health(ctx context.Context) error {
	return db.conn.PingContext(ctx)
	// TODO: Run run actual query (such as "SELECT 1")
}

type Ride struct {
	ID     string
	Driver string
	Kind   string
	Start  time.Time
	End    time.Time
	// End      sql.NullTime
	Distance float64
}

func (db *DB) Add(ctx context.Context, r Ride) error {
	_, err := db.conn.ExecContext(ctx, insertSQL,
		r.ID, r.Driver, r.Kind, r.Start, r.End, r.Distance)
	return err
}

var ErrNotFound = errors.New("not found")

func (db *DB) Get(ctx context.Context, id string) (Ride, error) {
	r := db.conn.QueryRowContext(ctx, getSQL, id)
	var rd Ride
	err := r.Scan(&rd.ID, &rd.Driver, &rd.Kind, &rd.Start, &rd.End, &rd.Distance)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Ride{}, ErrNotFound
	case err != nil:
		return Ride{}, err
	}

	return rd, nil
}

func (db *DB) Update(ctx context.Context, r Ride) error {
	_, err := db.conn.ExecContext(ctx, updateSQL,
		r.ID, r.Driver, r.Kind, r.Start, r.End, r.Distance)
	return err
}
