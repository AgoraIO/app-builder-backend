package utils

import (
	"time"

	"github.com/samyak-jain/agora_backend/utils/rtctoken"
	"github.com/samyak-jain/agora_backend/utils/rtmtoken"
	"github.com/spf13/viper"
)

// GetRtcToken generates token for Agora RTC SDK
func GetRtcToken(channel string, uid int) (string, error) {
	var RtcRole rtctoken.Role = rtctoken.RolePublisher

	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp := currentTimestamp + 86400

	return rtctoken.BuildTokenWithUID(viper.GetString("appID"), viper.GetString("appCertificate"), channel, uint32(uid), RtcRole, expireTimestamp)
}

// GetRtmToken generates a token for Agora RTM SDK
func GetRtmToken(user string) (string, error) {

	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp := currentTimestamp + 86400

	return rtmtoken.BuildToken(viper.GetString("appID"), viper.GetString("appCertificate"), user, rtmtoken.RoleRtmUser, expireTimestamp)
}
