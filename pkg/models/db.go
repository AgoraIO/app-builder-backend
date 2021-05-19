package models

import (
	"github.com/jmoiron/sqlx"
)

// Database contains a pointer to the database object
type Database struct {
	*sqlx.DB
}

// CreateDB is used to initialize a new database connection
func CreateDB(dbURL string) (*Database, error) {
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	return &Database{db}, nil
}
