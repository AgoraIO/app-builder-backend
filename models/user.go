package models

import (
	"github.com/jinzhu/gorm"
)

// User model contains all relevant details of a particular user
type User struct {
	gorm.Model
	Name   string
	Email  string  `gorm:"primary_key"`
	Tokens []Token `gorm:"foreignkey:UserEmail;association_foreignkey:Email"`
}

// Token stores the token of a user
type Token struct {
	gorm.Model
	TokenID   string
	UserEmail string
}
