package models

import "database/sql"

// Channel Model contains all the details for a particular channel session
type Channel struct {
	ID               int64          `db:"id"`
	Title            string         `db:"title"`
	ChannelName      string         `db:"name"`
	ChannelSecret    sql.NullString `db:"secret"`
	HostPassphrase   string         `db:"host"`
	ViewerPassphrase sql.NullString `db:"view"`
	DTMF             sql.NullString `db:"dtmf"`
	RecordingUID     sql.NullInt32  `db:"uid"`
	RecordingSID     sql.NullString `db:"sid"`
	RecordingRID     sql.NullString `db:"rid"`
}
