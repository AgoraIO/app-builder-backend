package models

import (
	"github.com/jinzhu/gorm"

	// Importing postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Datastore represents an interface of any model
type Datastore interface {
}

// Database contains a pointer to the database object
type Database struct {
	*gorm.DB
}

// CreateDB is used to initialize a new database connection
func CreateDB(dbURL string) (*Database, error) {
	db, err := gorm.Open("postgres", dbURL)
	return &Database{db}, err
}
