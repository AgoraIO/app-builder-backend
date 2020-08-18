package pstn

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

// ResponseData contains all details needed in response
type ResponseData struct {
	AppID   string
	Channel string
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

	responseString := `{
		"merge":true,
		"commands":[
			{
				"joinAgora": {
					"app": "{{.AppID}}",
					"channel": "{{.Channel}}",
					"channelKey": "{{.Token}}",
					"uid": {{.Uid}}
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
