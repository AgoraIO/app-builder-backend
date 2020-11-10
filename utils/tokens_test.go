package utils

import (
	"testing"
)

//GenerateUserCredentials

func TestGenerateUserCredentials(t *testing.T) {
	testingList := []struct {
		channel string
		rtmBool bool
	}{
		{"test", true},
		{"test1", false},
		{"test2", true},
		{"test3", false},
	}

	for _, tc := range testingList {
		result, err := GenerateUserCredentials(tc.channel, tc.rtmBool)
		t.Log(result.Rtc, result.Rtm, result.UID)
		if err != nil {
			t.Fatal("Failed at Token Generation with this error: ", err)
		}
	}
}
