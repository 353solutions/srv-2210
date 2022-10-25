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

	_ "github.com/lib/pq"
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
