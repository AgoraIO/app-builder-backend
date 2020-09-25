package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// Recorder manages cloud recording
type Recorder struct {
	http.Client
	Channel string
	Token   string
	UID     int
	RID     string
	SID     string
}

// Acquire runs the acquire endpoint for Cloud Recording
func (rec *Recorder) Acquire() error {
	creds, err := GenerateUserCredentials(rec.Channel, false)
	if err != nil {
		return err
	}

	rec.UID = creds.UID
	rec.Token = creds.Rtc

	requestBody := fmt.Sprintf(`
		{
			"cname": "%s",
			"uid": "%d",
			"clientRequest": {
				"resourceExpiredHour": 24
			}
		}
	`, rec.Channel, rec.UID)

	req, err := http.NewRequest("POST", "https://api.agora.io/v1/apps/"+viper.GetString("APP_ID")+"/cloud_recording/acquire",
		bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(viper.GetString("CUSTOMER_ID"), viper.GetString("CUSTOMER_CERTIFICATE"))

	resp, err := rec.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

	rec.RID = result["resourceId"]

	return nil
}

// Start starts the recording
func (rec *Recorder) Start() error {
	currentTime := strconv.FormatInt(time.Now().Unix(), 10)

	requestBody := fmt.Sprintf(`
		{
			"cname": "%s",
			"uid": "%d",
			"clientRequest": {
				"token": "%s",
				"recordingConfig": {
					"maxIdleTime": 30,
					"streamTypes": 2,
					"channelType": 1,
					"transcodingConfig": {
						"height": 720, 
						"width": 1280,
						"bitrate": 2260, 
						"fps": 15, 
						"mixedVideoLayout": 1,
						"backgroundColor": "#000000"
					}
				},
				"storageConfig": {
					"vendor": 1, 
					"region": 0,
					"bucket": "%s",
					"accessKey": "%s",
					"secretKey": "%s",
					"fileNamePrefix": ["%s", "%s"]
				}
			}
		}
	`, rec.Channel, rec.UID, rec.Token, viper.GetString("BUCKET_NAME"),
		viper.GetString("BUCKET_ACCESS_KEY"), viper.GetString("BUCKET_ACCESS_SECRET"),
		rec.Channel, currentTime)

	req, err := http.NewRequest("POST", "https://api.agora.io/v1/apps/"+viper.GetString("APP_ID")+"/cloud_recording/resourceid/"+rec.RID+"/mode/mix/start",
		bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(viper.GetString("CUSTOMER_ID"), viper.GetString("CUSTOMER_CERTIFICATE"))

	resp, err := rec.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	rec.SID = result["sid"]

	return nil
}

// Stop stops the cloud recording
func Stop(channel string, uid int, rid string, sid string) error {
	requestBody := fmt.Sprintf(`
		{
			"cname": "%s",
			"uid": "%d",
			"clientRequest": {
			}
		}
	`, channel, uid)

	req, err := http.NewRequest("POST", "https://api.agora.io/v1/apps/"+viper.GetString("APP_ID")+"/cloud_recording/resourceid/"+rid+"/sid/"+sid+"/mode/mix/stop",
		bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(viper.GetString("CUSTOMER_ID"), viper.GetString("CUSTOMER_CERTIFICATE"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

	return nil
}
