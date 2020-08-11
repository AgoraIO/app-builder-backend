package models

import (
	"github.com/jinzhu/gorm"
)

// User model contains all relevant details of a particular user
type User struct {
	gorm.Model
	Token int `gorm:"primary_key"`
	Name  string
	Email string
}
