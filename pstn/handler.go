package pstn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"

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

	var response ResponseData
	resp, err := http.Get(agoraAppURL)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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
					"app": "{{.AppID}}",
					"channel": "{{.Channel}}",
					"channelKey": "{{.Token}}",
					"uid": {{.UID}}
				}
			}
		]
	}`

	tmpl, err := template.New("test").Parse(responseString)
	if err != nil {
		panic(err)
	}

	var templateResult bytes.Buffer
	err = tmpl.Execute(&templateResult, response)
	if err != nil {
		panic(err)
	}

	byt := []byte(templateResult.String())

	w.Header().Set("Content-Type", "application/json")
	w.Write(byt)
}
