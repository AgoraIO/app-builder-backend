package pstn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/samyak-jain/agora_backend/utils"
)

// ResponseData contains all details needed in response
type ResponseData struct {
	AppID   string `json:"agoraApp"`
	Channel string `json:"agoraChannel"`
	Token   string
	UID     string
}

// InboundHandler sets the paramters for the inbound PSTN call
func InboundHandler(w http.ResponseWriter, r *http.Request) {
	log.Print(r.Method)
	agoraAppURL := r.URL.Query().Get("agoraAppURL")

	log.Print(agoraAppURL)

	var response ResponseData
	resp, err := http.Get(agoraAppURL)
	log.Print("resp1")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	log.Print("resp2")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
	}

	channelName := response.Channel
	uid := int(rand.Uint32())
	tokenData, err := utils.GetRtcToken(channelName, uid)
	if err != nil {
		panic(err)
	}

	response.Token = tokenData
	response.UID = fmt.Sprint(uid)

	responseString := `{
		"merge":true,
		"commands":[
			{
				"joinAgora": {
					"app": "{{AppID}}",
					"channel": "{{Channel}}",
					"channelKey": "{{Token}}",
					"uid": {{UID}}
				}
			}
		]
	}`

	finalResult := strings.Replace(responseString, "{{AppID}}", response.AppID, 1)
	finalResult = strings.Replace(finalResult, "{{Channel}}", response.Channel, 1)
	finalResult = strings.Replace(finalResult, "{{Token}}", response.Token, 1)
	finalResult = strings.Replace(finalResult, "{{UID}}", response.UID, 1)

	log.Print(finalResult)
	byt := []byte(finalResult)

	w.Header().Set("Content-Type", "application/json")
	w.Write(byt)
}
