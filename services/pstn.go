// ********************************************
// Copyright © 2021 Agora Lab, Inc., all rights reserved.
// AppBuilder and all associated components, source code, APIs, services, and documentation
// (the “Materials”) are owned by Agora Lab, Inc. and its licensors.  The Materials may not be
// accessed, used, modified, or distributed for any purpose without a license from Agora Lab, Inc.
// Use without a license or in violation of any license terms and conditions (including use for
// any purpose competitive to Agora Lab, Inc.’s business) is strictly prohibited.  For more
// information visit https://appbuilder.agora.io.
// *********************************************

package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/samyak-jain/agora_backend/pkg/models"
	"github.com/samyak-jain/agora_backend/utils"
	"github.com/spf13/viper"
)

type AuthAccount struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	PartnerID string `json:"partnerID"`
	AccountID string `json:"accountID"`
}

type BridgeConfig struct {
	ConferenceID        string `json:"conferenceID"`
	MinimumParticipants int    `json:"minParticipants"`
	ExitChimes          string `json:"exitChimes"`
	ConfigParameterURL  string `json:"confParamsUrl"`
}

type BridgeRequest struct {
	SetBridge BridgeConfig `json:"setBridge"`
}

type PSTNRequest struct {
	AuthAccount AuthAccount     `json:"authAccount"`
	RequestList []BridgeRequest `json:"requestList"`
}

type Request struct {
	Request PSTNRequest `json:"request"`
}

func CreateBridge(logger *utils.Logger, confID string, backendURL string) {
	request := Request{
		Request: PSTNRequest{
			AuthAccount: AuthAccount{
				Email:     viper.GetString("PSTN_EMAIL"),
				Password:  viper.GetString("PSTN_PASSWORD"),
				PartnerID: "turbobridge",
				AccountID: viper.GetString("PSTN_ACCOUNT"),
			},
			RequestList: []BridgeRequest{
				{
					SetBridge: BridgeConfig{
						ConferenceID:        confID,
						MinimumParticipants: 1,
						ExitChimes:          "none",
						ConfigParameterURL:  backendURL + "/pstn",
					},
				},
			},
		},
	}

	requestBody, err := json.Marshal(&request)
	if err != nil {
		logger.Error().Err(err).Interface("Request", request).Msg("Unable to Marshal JSON")
		return
	}

	logger.Debug().Str("Create Bridge parameters", string(requestBody)).Msg("Create Bridge")

	req, err := http.NewRequest("POST", "https://api-dev.turbobridge.com/4.3/Bridge", bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Error().Err(err).Msg("Unable to Create Request")
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error().Err(err).Interface("Request", req).Msg("Unable to Create Bridge")
		return
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode != 200 {
		logger.Error().Int("Status Code", resp.StatusCode).Interface("Response", result).Msg("Error response in create bridge")
	} else {
		logger.Info().Interface("Response", result).Msg("Create Bridge Response")
	}

}

type AgoraFields struct {
	AppID          string  `json:"app"`
	ChannelName    string  `json:"channel"`
	Token          string  `json:"channelKey"`
	UID            int     `json:"uid"`
	EncryptionMode *string `json:"encryption,omitempty"`
	ChannelSecret  *string `json:"secret,omitempty"`
}

type CallData struct {
	UID int `json:"uid"`
}

type PSTNResponse struct {
	Type     string      `json:"type"`
	App      string      `json:"app"`
	Fields   AgoraFields `json:"fields"`
	CallData CallData    `json:"callDataPerm"`
}

func (router *ServiceRouter) PSTN(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	conferenceID := query.Get("confID")

	router.Logger.Debug().Str("Conference ID", conferenceID).Msg("Got conference ID")

	var channelData models.Channel
	err := router.DB.Get(&channelData, "SELECT channel_name, channel_secret FROM channels WHERE dtmf=$1", conferenceID)
	if err != nil {
		router.Logger.Error().Err(err).Str("Conference ID", conferenceID).Msg("Could not fetch relevant channel from DB")
		return
	}

	user, err := utils.GenerateUserCredentials(channelData.ChannelName, false, true)
	if err != nil {
		router.Logger.Error().Err(err).Msg("Could not generate main user credentials")
		return
	}

	isEncrpytionEnabled := viper.GetBool("ENCRYPTION_ENABLED")

	router.Logger.Debug().Bool("Encryption Enabled", isEncrpytionEnabled).Msg("Is Encrpytion enabled?")

	var response PSTNResponse
	encMode := "aes-128-xts"
	w.Header().Set("Content-Type", "application/json")

	if isEncrpytionEnabled {
		response = PSTNResponse{
			Type: "callStream",
			App:  "agora",
			Fields: AgoraFields{
				AppID:          viper.GetString("APP_ID"),
				ChannelName:    channelData.ChannelName,
				Token:          user.Rtc,
				UID:            user.UID,
				EncryptionMode: &encMode,
				ChannelSecret:  &channelData.ChannelSecret,
			},
			CallData: CallData{
				UID: user.UID,
			},
		}

	} else {
		response = PSTNResponse{
			Type: "callStream",
			App:  "agora",
			Fields: AgoraFields{
				AppID:       viper.GetString("APP_ID"),
				ChannelName: channelData.ChannelName,
				Token:       user.Rtc,
				UID:         user.UID,
			},
			CallData: CallData{
				UID: user.UID,
			},
		}
	}

	responseString, err := json.Marshal(response)
	if err != nil {
		router.Logger.Error().Err(err).Interface("Response", response).Msg("Unable to Marshal JSON")
		return
	}

	router.Logger.Info().Str("Response", string(responseString)).Msg("Response String")

	json.NewEncoder(w).Encode(response)
}

type ConferenceInfo struct {
	ConferenceID string `json:"conferenceID"`
}

type ConferenceRequest struct {
	ConferenceInfo ConferenceInfo `json:"getConferenceInfo"`
}

type GetConferenceRequest struct {
	AuthAccount AuthAccount         `json:"authAccount"`
	RequestList []ConferenceRequest `json:"requestList"`
}

type ConferencePSTNRequest struct {
	Request GetConferenceRequest `json:"request"`
}

type CustomData struct {
	UID string `json:"uid"`
}

type Call struct {
	CustomData CustomData `json:"dataPerm"`
	CallID     string     `json:"callID"`
}

type Calls struct {
	Call []Call `json:"call"`
}

type Conference struct {
	Calls Calls `json:"calls"`
}

type Result struct {
	Conference Conference `json:"conference"`
}

type RequestItem struct {
	Result Result `json:"result"`
}

type ConferenceResponse struct {
	RequestItem []RequestItem `json:"requestItem"`
}

type ConferencePSTNResponse struct {
	Response ConferenceResponse `json:"responseList"`
}

// MutePSTN is a helper method to mute and unmute a PSTN User
func MutePSTN(logger *utils.Logger, uid int, muteState bool, confID string) {
	request := ConferencePSTNRequest{
		Request: GetConferenceRequest{
			AuthAccount: AuthAccount{
				Email:     viper.GetString("PSTN_EMAIL"),
				Password:  viper.GetString("PSTN_PASSWORD"),
				PartnerID: "turbobridge",
				AccountID: viper.GetString("PSTN_ACCOUNT"),
			},
			RequestList: []ConferenceRequest{
				{
					ConferenceInfo: ConferenceInfo{
						ConferenceID: confID,
					},
				},
			},
		},
	}

	requestBody, err := json.Marshal(&request)
	if err != nil {
		logger.Error().Err(err).Interface("Request", request).Msg("Unable to Marshal JSON")
		return
	}

	logger.Debug().Str("Conference Info Parameters", string(requestBody)).Msg("Get Conference Info")

	req, err := http.NewRequest("POST", "https://api-dev.turbobridge.com/4.3/LCM", bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Error().Err(err).Msg("Unable to Create Request")
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error().Err(err).Interface("Request", req).Msg("Unable to get conference info")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Error().Int("Status Code", resp.StatusCode).Msg("Error response in create bridge")
		return
	}

	var result ConferencePSTNResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to decode JSON response")
	}

	for _, call := range result.Response.RequestItem[0].Result.Conference.Calls.Call {
		if call.CustomData.UID == strconv.Itoa(uid) {
			SetMuteState(logger, call.CallID, confID, muteState)
			return
		}
	}

	logger.Error().Interface("Conference Response", result).Msg("No matching UID found")
}

type ChangeConferenceCallDetails struct {
	ConferenceID string `json:"conferenceID"`
	CallID       string `json:"callID"`
	Command      string `json:"command"`
	Value        string `json:"value"`
}

type ChangeConferenceRequest struct {
	ChangeConferenceCallDetails ChangeConferenceCallDetails `json:"changeConferenceCall"`
}

type ChangeConferenceRequestList struct {
	AuthAccount AuthAccount               `json:"authAccount"`
	RequestList []ChangeConferenceRequest `json:"requestList"`
}

type ChangeConferenceCall struct {
	Request ChangeConferenceRequestList `json:"request"`
}

func SetMuteState(logger *utils.Logger, callID string, confID string, muteState bool) {
	var numberMuteState string
	if muteState {
		numberMuteState = "1"
	} else {
		numberMuteState = "0"
	}

	request := ChangeConferenceCall{
		Request: ChangeConferenceRequestList{
			AuthAccount: AuthAccount{
				Email:     viper.GetString("PSTN_EMAIL"),
				Password:  viper.GetString("PSTN_PASSWORD"),
				PartnerID: "turbobridge",
				AccountID: viper.GetString("PSTN_ACCOUNT"),
			},
			RequestList: []ChangeConferenceRequest{
				{
					ChangeConferenceCallDetails: ChangeConferenceCallDetails{
						ConferenceID: confID,
						CallID:       callID,
						Command:      "setMute",
						Value:        numberMuteState,
					},
				},
			},
		},
	}

	requestBody, err := json.Marshal(&request)
	if err != nil {
		logger.Error().Err(err).Interface("Request", request).Msg("Unable to Marshal JSON")
		return
	}

	logger.Debug().Str("Change Conference parameters", string(requestBody)).Msg("Mute user in conference")

	req, err := http.NewRequest("POST", "https://api-dev.turbobridge.com/4.3/LCM", bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Error().Err(err).Msg("Unable to Create Request")
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error().Err(err).Interface("Request", req).Msg("Unable to Change Conference")
		return
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode != 200 {
		logger.Error().Int("Status Code", resp.StatusCode).Interface("Response", result).Msg("Error response in change conference")
	} else {
		logger.Info().Interface("Response", result).Msg("Change Conference Response")
	}

}
