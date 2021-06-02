// ********************************************
// Copyright © 2021 Agora Lab, Inc., all rights reserved.
// AppBuilder and all associated components, source code, APIs, services, and documentation
// (the “Materials”) are owned by Agora Lab, Inc. and its licensors.  The Materials may not be
// accessed, used, modified, or distributed for any purpose without a license from Agora Lab, Inc.
// Use without a license or in violation of any license terms and conditions (including use for
// any purpose competitive to Agora Lab, Inc.’s business) is strictly prohibited.  For more
// information visit https://appbuilder.agora.io.
// *********************************************

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
