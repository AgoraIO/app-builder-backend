package utils

import (
	"bytes"
	"encoding/json"
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

type AcquireClientRequest struct {
	ResourceExpiredHour int `json:"resourceExpiredHour,omitempty"`
}

type AcquireRequest struct {
	Cname         string               `json:"cname"`
	UID           string               `json:"uid"`
	ClientRequest AcquireClientRequest `json:"clientRequest"`
}

type TranscodingConfig struct {
	Height           int    `json:"height"`
	Width            int    `json:"width"`
	Bitrate          int    `json:"bitrate"`
	Fps              int    `json:"fps"`
	MixedVideoLayout int    `json:"mixedVideoLayout"`
	BackgroundColor  string `json:"backgroundColor"`
}

type RecordingConfig struct {
	MaxIdleTime       int               `json:"maxIdleTime"`
	StreamTypes       int               `json:"streamTypes"`
	ChannelType       int               `json:"channelType"`
	DecryptionMode    int               `json:"decryptionMode,omitempty"`
	Secret            string            `json:"secret,omitempty"`
	TranscodingConfig TranscodingConfig `json:"transcodingConfig"`
}

type StorageConfig struct {
	Vendor         int      `json:"vendor"`
	Region         int      `json:"region"`
	Bucket         string   `json:"bucket"`
	AccessKey      string   `json:"accessKey"`
	SecretKey      string   `json:"secretKey"`
	FileNamePrefix []string `json:"fileNamePrefix"`
}

type ClientRequest struct {
	Token           string          `json:"token"`
	RecordingConfig RecordingConfig `json:"recordingConfig"`
	StorageConfig   StorageConfig   `json:"storageConfig"`
}

type StartRecordRequest struct {
	Cname         string        `json:"cname"`
	UID           string        `json:"uid"`
	ClientRequest ClientRequest `json:"clientRequest"`
}

// Acquire runs the acquire endpoint for Cloud Recording
func (rec *Recorder) Acquire() error {
	creds, err := GenerateUserCredentials(rec.Channel, false)
	if err != nil {
		return err
	}

	rec.UID = creds.UID
	rec.Token = creds.Rtc

	requestBody, err := json.Marshal(&AcquireRequest{
		Cname: rec.Channel,
		UID:   string(rec.UID),
		ClientRequest: AcquireClientRequest{
			ResourceExpiredHour: 24,
		},
	})

	req, err := http.NewRequest("POST", "https://api.agora.io/v1/apps/"+viper.GetString("APP_ID")+"/cloud_recording/acquire",
		bytes.NewBuffer(requestBody))
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
func (rec *Recorder) Start(secret *string) error {
	currentTime := strconv.FormatInt(time.Now().Unix(), 10)

	transcodingConfig := TranscodingConfig{
		Height:           720,
		Width:            1280,
		Bitrate:          2260,
		Fps:              15,
		MixedVideoLayout: 1,
		BackgroundColor:  "#000000",
	}
	var recordingConfig RecordingConfig
	if secret != nil && *secret != "" {

		recordingConfig = RecordingConfig{
			MaxIdleTime:       30,
			StreamTypes:       2,
			ChannelType:       1,
			DecryptionMode:    1,
			Secret:            *secret,
			TranscodingConfig: transcodingConfig,
		}
	} else {
		recordingConfig = RecordingConfig{
			MaxIdleTime:       30,
			StreamTypes:       2,
			ChannelType:       1,
			TranscodingConfig: transcodingConfig,
		}
	}

	requestBody, err := json.Marshal(&StartRecordRequest{
		Cname: rec.Channel,
		UID:   string(rec.UID),
		ClientRequest: ClientRequest{
			Token: rec.Token,
			StorageConfig: StorageConfig{
				Vendor:    viper.GetInt("RECORDING_VENDOR"),
				Region:    viper.GetInt("RECORDING_REGION"),
				Bucket:    viper.GetString("BUCKET_NAME"),
				AccessKey: viper.GetString("BUCKET_ACCESS_KEY"),
				SecretKey: viper.GetString("BUCKET_ACCESS_SECRET"),
				FileNamePrefix: []string{
					rec.Channel, currentTime,
				},
			},
			RecordingConfig: recordingConfig,
		},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.agora.io/v1/apps/"+viper.GetString("APP_ID")+"/cloud_recording/resourceid/"+rec.RID+"/mode/mix/start",
		bytes.NewBuffer(requestBody))
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
	requestBody, err := json.Marshal(&AcquireRequest{
		Cname:         channel,
		UID:           string(uid),
		ClientRequest: AcquireClientRequest{},
	})

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
