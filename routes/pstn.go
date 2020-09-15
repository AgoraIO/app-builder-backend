package routes

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/spf13/viper"

	"github.com/samyak-jain/agora_backend/models"
	"github.com/samyak-jain/agora_backend/utils"
)

// DTMFHandler handles DTMF
func (o *Router) DTMFHandler(w http.ResponseWriter, r *http.Request) {
	dtmf := r.URL.Query().Get("id")

	var channelData *models.Channel
	if err := o.DB.Where("DTMF = ?", dtmf).First(&channelData).Error; err != nil {
		log.Print(err)
	}

	uid := int(rand.Uint32())
	tokenData, err := utils.GetRtcToken(channelData.Name, uid)
	if err != nil {
		log.Print(err)
	}

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

	finalResult := strings.Replace(responseString, "{{AppID}}", viper.GetString("appID"), 1)
	finalResult = strings.Replace(finalResult, "{{Channel}}", channelData.Name, 1)
	finalResult = strings.Replace(finalResult, "{{Token}}", tokenData, 1)
	finalResult = strings.Replace(finalResult, "{{UID}}", fmt.Sprint(uid), 1)

	log.Print(finalResult)
	byt := []byte(finalResult)

	w.Header().Set("Content-Type", "application/json")
	w.Write(byt)
}
