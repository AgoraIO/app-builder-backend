package models

import (
	"github.com/jinzhu/gorm"
)

// Channel Model contains all the details for a particular channel session
type Channel struct {
	gorm.Model
	Title            string
	ChannelName      string
	ChannelSecret    string
	HostPassphrase   string
	ViewerPassphrase string
	DTMF             string
	RecordingUID     int
	RecordingSID     string
	RecordingRID     string
}
