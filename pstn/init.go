package pstn

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

// SetPSTN sets the PSTN
func SetPSTN(channel string) error {
	log.Print(channel)

	const reqURL = "https://dids.turbobridge.com/api/1.0/PhoneNumber"
	in := `{
		"request": {
		  "authAdministrator": {
			"username": "{{PSTN_USERNAME}}",
			"password": "{{PSTN_PASS}}"
		  },
		  "requestList": [
			{
			  "setPhoneNumber": {
				"phoneNumberID": 6126,
				"voiceAPI": {
				  "vars": {
					"agoraAppURL": "https://dev.turbobridge.com/voiceAPI/agoraInboundReflect.php?agoraApp=b8c2ef0f986541a8992451c07d30fb4b&agoraChannel={{Channel}}"
				  }
				}
			  }
			}
		  ]
		}
	  }`

	in = strings.Replace(in, "{{PSTN_USERNAME}}", viper.GetString("PSTN_USERNAME"), 1)
	in = strings.Replace(in, "{{PSTN_PASS}}", viper.GetString("PSTN_PASSWORD"), 1)
	in = strings.Replace(in, "{{Channel}}", channel, 1)

	log.Print(in)

	byteData := []byte(in)

	resp, err := http.Post(reqURL, "application/json", bytes.NewBuffer(byteData))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	finalResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var dat map[string]interface{}

	json.Unmarshal(finalResp, &dat)

	log.Print(dat)

	return nil
}
