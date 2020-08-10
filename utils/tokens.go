package utils

import (
	"time"

	"github.com/samyak-jain/agora_backend/utils/rtctoken"
	"github.com/samyak-jain/agora_backend/utils/rtmtoken"
)

var config AgoraConfig = GetAgoraConfig()

// GetRtcToken generates token for Agora RTC SDK
func GetRtcToken(channel string, uid int) (string, error) {
	var RtcRole rtctoken.Role = rtctoken.RolePublisher

	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp := currentTimestamp + 24

	return rtctoken.BuildTokenWithUID(config.AppID, config.AppCertificate, channel, uint32(uid), RtcRole, expireTimestamp)
}

// GetRtmToken generates a token for Agora RTM SDK
func GetRtmToken(user string) (string, error) {

	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp := currentTimestamp + 24

	return rtmtoken.BuildToken(config.AppID, config.AppCertificate, user, rtmtoken.RoleRtmUser, expireTimestamp)
}
