package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	DB *sql.DB
	Q  *Queries
}

func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	// Ensure FK is on
	_, _ = db.Exec(`PRAGMA foreign_keys = ON;`)

	return &Store{DB: db, Q: &Queries{db: db}}, nil
}

func (s *Store) Close() error { return s.DB.Close() }
func (s *Store) Ping() error  { return s.DB.Ping() }
