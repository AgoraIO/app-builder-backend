package models

import "database/sql"

// Channel Model contains all the details for a particular channel session
type Channel struct {
	ID               int64          `db:"id"`
	Title            string         `db:"title"`
	ChannelName      string         `db:"channel_name"`
	ChannelSecret    string         `db:"channel_secret"`
	HostPassphrase   string         `db:"host_passphrase"`
	ViewerPassphrase string         `db:"viewer_passphrase"`
	DTMF             string         `db:"dtmf"`
	RecordingUID     sql.NullInt32  `db:"recording_uid"`
	RecordingSID     sql.NullString `db:"recording_sid"`
	RecordingRID     sql.NullString `db:"recording_rid"`
}
