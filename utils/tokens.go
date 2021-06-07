// ********************************************
// Copyright © 2021 Agora Lab, Inc., all rights reserved.
// AppBuilder and all associated components, source code, APIs, services, and documentation
// (the “Materials”) are owned by Agora Lab, Inc. and its licensors.  The Materials may not be
// accessed, used, modified, or distributed for any purpose without a license from Agora Lab, Inc.
// Use without a license or in violation of any license terms and conditions (including use for
// any purpose competitive to Agora Lab, Inc.’s business) is strictly prohibited.  For more
// information visit https://appbuilder.agora.io.
// *********************************************

package utils

import (
	"fmt"
	"time"

	"github.com/samyak-jain/agora_backend/pkg/models"
	"github.com/samyak-jain/agora_backend/utils/rtctoken"
	"github.com/samyak-jain/agora_backend/utils/rtmtoken"
	"github.com/spf13/viper"
)

// GetRtcToken generates token for Agora RTC SDK
func GetRtcToken(channel string, uid int) (string, error) {
	var RtcRole rtctoken.Role = rtctoken.RolePublisher

	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp := currentTimestamp + 86400

	return rtctoken.BuildTokenWithUID(viper.GetString("APP_ID"), viper.GetString("APP_CERTIFICATE"), channel, uint32(uid), RtcRole, expireTimestamp)
}

// GetRtmToken generates a token for Agora RTM SDK
func GetRtmToken(user string) (string, error) {

	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp := currentTimestamp + 86400

	return rtmtoken.BuildToken(viper.GetString("APP_ID"), viper.GetString("APP_CERTIFICATE"), user, rtmtoken.RoleRtmUser, expireTimestamp)
}

// GenerateUserCredentials generates uid, rtc and rtc token
func GenerateUserCredentials(channel string, rtm bool, pstn bool) (*models.UserCredentials, error) {
	initialUID := RandomRange(10000000, 99999999)
	var uid int
	if pstn {
		uid = initialUID + 100000000
	} else {
		uid = initialUID + 200000000
	}

	rtcToken, err := GetRtcToken(channel, uid)
	if err != nil {
		return nil, err
	}

	if !rtm {
		return &models.UserCredentials{
			Rtc: rtcToken,
			UID: uid,
		}, nil
	}

	rtmToken, err := GetRtmToken(fmt.Sprint(uid))
	if err != nil {
		return nil, err
	}

	return &models.UserCredentials{
		Rtc: rtcToken,
		Rtm: &rtmToken,
		UID: uid,
	}, nil
}
