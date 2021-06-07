// ********************************************
// Copyright © 2021 Agora Lab, Inc., all rights reserved.
// AppBuilder and all associated components, source code, APIs, services, and documentation
// (the “Materials”) are owned by Agora Lab, Inc. and its licensors.  The Materials may not be
// accessed, used, modified, or distributed for any purpose without a license from Agora Lab, Inc.
// Use without a license or in violation of any license terms and conditions (including use for
// any purpose competitive to Agora Lab, Inc.’s business) is strictly prohibited.  For more
// information visit https://appbuilder.agora.io.
// *********************************************

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
	UID     int32
	RID     string
	SID     string
	Logger  *Logger
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
	MaxResolutionUID string `json:"maxResolutionUid,omitempty"`
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

type RecordingFileConfig struct {
	AVFileType []string `json:"avFileType"`
}

type ClientRequest struct {
	Token               string              `json:"token"`
	RecordingConfig     RecordingConfig     `json:"recordingConfig"`
	RecordingFileConfig RecordingFileConfig `json:"recordingFileConfig"`
	StorageConfig       StorageConfig       `json:"storageConfig"`
}

type StartRecordRequest struct {
	Cname         string        `json:"cname"`
	UID           string        `json:"uid"`
	ClientRequest ClientRequest `json:"clientRequest"`
}

// Acquire runs the acquire endpoint for Cloud Recording
func (rec *Recorder) Acquire() error {
	creds, err := GenerateUserCredentials(rec.Channel, false, false)
	if err != nil {
		return err
	}

	rec.UID = int32(creds.UID)
	rec.Token = creds.Rtc

	requestBody, err := json.Marshal(&AcquireRequest{
		Cname: rec.Channel,
		UID:   strconv.Itoa(int(rec.UID)),
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

	rec.Logger.Debug().Interface("Result", result).Msg("Recording Result")

	return nil
}

// Start starts the recording
func (rec *Recorder) Start(channelTitle string, secret *string) error {
	// currentTime := strconv.FormatInt(time.Now().Unix(), 10)
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return err
	}
	currentTimeStamp := time.Now().In(location)
	currentDate := currentTimeStamp.Format("20060102")
	currentTime := currentTimeStamp.Format("150405")

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

	recordingRequest := StartRecordRequest{
		Cname: rec.Channel,
		UID:   strconv.Itoa(int(rec.UID)),
		ClientRequest: ClientRequest{
			Token: rec.Token,
			StorageConfig: StorageConfig{
				Vendor:    viper.GetInt("RECORDING_VENDOR"),
				Region:    viper.GetInt("RECORDING_REGION"),
				Bucket:    viper.GetString("BUCKET_NAME"),
				AccessKey: viper.GetString("BUCKET_ACCESS_KEY"),
				SecretKey: viper.GetString("BUCKET_ACCESS_SECRET"),
				FileNamePrefix: []string{
					channelTitle, currentDate, currentTime,
				},
			},
			RecordingFileConfig: RecordingFileConfig{
				AVFileType: []string{"hls", "mp4"},
			},
			RecordingConfig: recordingConfig,
		},
	}

	rec.Logger.Info().Interface("Start Request", recordingRequest).Msg("Recording request")

	requestBody, err := json.Marshal(&recordingRequest)
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

	rec.Logger.Debug().Interface("Result", result).Msg("Recording Result")

	return nil
}

type UpdateRecordRequest struct {
	Cname         string            `json:"cname"`
	UID           string            `json:"uid"`
	ClientRequest TranscodingConfig `json:"clientRequest"`
}

func ChangeRecordingMode(channel string, uid int, rid string, sid string, mode int, maxUID string, logger *Logger) error {
	recordingRequest := UpdateRecordRequest{
		Cname: channel,
		UID:   strconv.Itoa(uid),
		ClientRequest: TranscodingConfig{
			MixedVideoLayout: mode,
			MaxResolutionUID: maxUID,
		},
	}

	logger.Info().Interface("Change Recording", recordingRequest).Msg("Change Recording Mode")

	requestBody, err := json.Marshal(&recordingRequest)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.agora.io/v1/apps/"+viper.GetString("APP_ID")+"/cloud_recording/resourceid/"+rid+"/sid/"+sid+"/mode/mix/update",
		bytes.NewBuffer(requestBody))
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

	logger.Info().Interface("response", result).Msg("Update Cloud Recording Response")

	return nil

}

// Stop stops the cloud recording
func Stop(channel string, uid int, rid string, sid string, logger *Logger) error {
	recordingRequest := AcquireRequest{
		Cname:         channel,
		UID:           strconv.Itoa(uid),
		ClientRequest: AcquireClientRequest{},
	}

	logger.Info().Interface("Stop Request", recordingRequest).Msg("Stop Recording Request")

	requestBody, err := json.Marshal(&recordingRequest)

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

	logger.Info().Interface("response", result).Msg("Stop Cloud Recording Response")

	return nil
}

// FirstN is to return the first N characters of a string
func FirstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}
