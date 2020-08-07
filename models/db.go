package models

import (
	"log"

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
	if err != nil {
		log.Panic(err)
	}

	// TODO: Setup Production Migrations
	// genericDB := db.DB()
	// driver, err := postgres.WithInstance(genericDB, &postgres.Config{})
	// if err != nil {
	// 	return nil, err
	// }

	// m, err := migrate.NewWithDatabaseInstance("file://", "postgres", driver)
	// if err := m.Up(); err != nil {
	// 	return nil, err
	// }

	return &Database{db}, nil
}
