package models

import (
	"github.com/jinzhu/gorm"
)

// Channel Model contains all the details for a particular channel session
type Channel struct {
	gorm.Model
	Title            string
	Name             string
	Secret           string
	HostPassphrase   string
	ViewerPassphrase string
	DTMF             string
	Recording        Recording
	Hosts            User
}

// Recording contains the details ßof the recording session
type Recording struct {
	UID int
	SID string
	RID string
}
