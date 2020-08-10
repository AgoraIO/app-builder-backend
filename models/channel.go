package models

import (
	"github.com/jinzhu/gorm"
)

// Channel Model contains all the details for a particular channel session
type Channel struct {
	gorm.Model
	Name             string
	HostPassword     string
	ViewerPassword   string
	HostPassphrase   string
	ViewerPassphrase string
	Creator          User
}
