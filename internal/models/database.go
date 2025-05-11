package models

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	// Create expressions table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS expressions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expression TEXT NOT NULL,
			status TEXT NOT NULL,
			result REAL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) DB() *sql.DB {
	return d.db
}
