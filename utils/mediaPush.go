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
	"errors"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/spf13/viper"
)

type MediaPush struct {
	http.Client
	Logger *Logger
}
type AudioOptions struct {
	CodecProfile  string `json:"codecProfile"`
	SampleRate    int    `json:"sampleRate"`
	Bitrate       int    `json:"bitrate"`
	AudioChannels int    `json:"audioChannels"`
}
type VideoOptions struct {
	Canvas       Canvas   `json:"canvas"`
	Layout       []Layout `json:"layout"`
	CodecProfile string   `json:"codecProfile"`
	FrameRate    int      `json:"frameRate"`
	Gop          int      `json:"gop"`
	Bitrate      int      `json:"bitrate"`
	SeiOptions   struct {
	} `json:"seiOptions"`
}
type Canvas struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}
type Region struct {
	XPos   int `json:"xPos"`
	YPos   int `json:"yPos"`
	ZIndex int `json:"zIndex"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
type Layout struct {
	RtcStreamUID        int    `json:"rtcStreamUid"`
	Region              Region `json:"region"`
	FillMode            string `json:"fillMode"`
	PlaceholderImageURL string `json:"placeholderImageUrl"`
}
type TranscodeOptions struct {
	RtcChannel   string       `json:"rtcChannel"`
	AudioOptions AudioOptions `json:"audioOptions"`
	VideoOptions VideoOptions `json:"videoOptions"`
}
type ConverterClientRequest struct {
	Name             string           `json:"name"`
	TranscodeOptions TranscodeOptions `json:"transcodeOptions"`
	RtmpURL          string           `json:"rtmpUrl"`
}
type ConverterRequest struct {
	Converter ConverterClientRequest `json:"converter"`
}

func (m *MediaPush) RTMPConverters(channelName string, uuid int, streamKey string) error {
	name, err := GenerateUUID()
	if err != nil {
		return errors.New("error while generating uuid")
	}
	requestBody, _ := json.Marshal(&ConverterRequest{
		// TODO: this hardcoded value will be replaced
		ConverterClientRequest{
			Name: name,
			TranscodeOptions: TranscodeOptions{
				RtcChannel: channelName,
				AudioOptions: AudioOptions{
					CodecProfile:  "LC-AAC",
					SampleRate:    48000,
					Bitrate:       viper.GetInt("AUDIO_BIT_RATE"),
					AudioChannels: viper.GetInt("AUDIO_CHANNEL"),
				},
				VideoOptions: VideoOptions{
					Canvas: Canvas{
						Width:  viper.GetInt("CANVAS_WIDTH"),
						Height: viper.GetInt("CANVAS_HEIGHT"),
					},
					Layout: []Layout{
						{
							RtcStreamUID: uuid,
							Region: Region{
								XPos:   0,
								YPos:   0,
								ZIndex: 1,
								Width:  viper.GetInt("LAYOUT_WIDTH"),
								Height: viper.GetInt("LAYOUT_HEIGHT"),
							},
							FillMode:            "fill",
							PlaceholderImageURL: viper.GetString("PLACEHOLDER_IMAGE_URL"),
						},
					},
					CodecProfile: "high",
					FrameRate:    viper.GetInt("VIDEO_FRAME_RATE"),
					Gop:          viper.GetInt("VIDEO_GOP"),
					Bitrate:      viper.GetInt("VIDEO_BIT_RATE"),
					SeiOptions:   struct{}{},
				},
			},
			RtmpURL: viper.GetString("RTMP_URL") + streamKey,
		},
	})

	req, err := http.NewRequest("POST", "https://api.agora.io/ap/v1/projects/"+viper.GetString("APP_ID")+"/rtmp-converters",
		bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(viper.GetString("CUSTOMER_ID"), viper.GetString("CUSTOMER_CERTIFICATE"))

	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}
	m.Logger.Debug().Interface("Request", string(reqDump)).Msg("Request")

	resp, err := m.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	m.Logger.Debug().Interface("Response", resp.Status).Msg("Converter Response")
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

	m.Logger.Debug().Interface("Result", result).Msg("Converter Result")

	return nil
}

func (m *MediaPush) LiveStreams() (string, error) {
	req, err := http.NewRequest("POST", "https://api.agora.io/ap/v1/projects/"+viper.GetString("PROJECT_APP_ID")+"/cfls/live-streams",
		nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(viper.GetString("CUSTOMER_ID"), viper.GetString("CUSTOMER_CERTIFICATE"))

	resp, err := m.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", errors.New("not able to get data from live stream response")
	}
	m.Logger.Debug().Interface("Result", result).Msg("Live Stream Result")

	streamkey, ok := data["streamKey"].(string)
	if !ok {
		return "", errors.New("not able to get streamkey from live stream response")
	}
	return streamkey, nil
}
