package utils

import (
	"github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/AccessToken"
	rtctokenbuilder "github.com/samyak-jain/agora_backend/utils/rtctoken"
	rtmtokenbuilder "github.com/samyak-jain/agora_backend/utils/rtmtoken"
	"math/rand"
	"testing"
)

func Test_RtcTokenBuilder(t *testing.T) {
	appID := "970CA35de60c44645bbae8a215061b33"
	appCertificate := "5CFd2fd1755d40ecb72977518be15d3b"
	channelName := "7d72365eb983485397e3e3f9d460bdda"
	uid := rand.Uint32()
	expiredTs := uint32(14471)
	result, err := rtctokenbuilder.BuildTokenWithUID(appID, appCertificate, channelName, uid, rtctokenbuilder.RoleSubscriber, expiredTs)

	if err != nil {
		t.Error(err)
	}

	token := accesstoken.AccessToken{}
	token.FromString(result)
	if token.Message[accesstoken.KJoinChannel] != expiredTs {
		t.Error("no kJoinChannel ts")
	}

	if token.Message[accesstoken.KPublishVideoStream] != 0 {
		t.Error("should not have publish video stream privilege")
	}
}

func Test_RtmTokenBuilder(t *testing.T) {
	appID := "970CA35de60c44645bbae8a215061b33"
	appCertificate := "5CFd2fd1755d40ecb72977518be15d3b"
	userAccount := "test_user"
	expiredTs := uint32(1446455471)
	result, err := rtmtokenbuilder.BuildToken(appID, appCertificate, userAccount, rtmtokenbuilder.RoleRtmUser, expiredTs)

	if err != nil {
		t.Error(err)
	}

	token := accesstoken.AccessToken{}
	token.FromString(result)
	if token.Message[accesstoken.KLoginRtm] != expiredTs {
		t.Error("no kLoginRtm ts")
	}
}
