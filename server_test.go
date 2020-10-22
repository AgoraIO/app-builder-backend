package main

import (
	"encoding/json"
	_ "github.com/coreos/etcd/client"
	"github.com/samyak-jain/agora_backend/models"
	"github.com/samyak-jain/agora_backend/routes"
	"github.com/samyak-jain/agora_backend/utils"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type CreateChannel struct {
	Data struct {
		CreateChannel struct {
			Passphrase struct {
				Host     string `json:"host"`
				View     string `json:"view"`
				Typename string `json:"__typename"`
			} `json:"passphrase"`
			Channel  string      `json:"channel"`
			Title    string      `json:"title"`
			Pstn     interface{} `json:"pstn"`
			Typename string      `json:"__typename"`
		} `json:"createChannel"`
	} `json:"data"`
}

type JoinChannelSuccess struct {
	Data struct {
		JoinChannel struct {
			Channel  string `json:"channel"`
			Title    string `json:"title"`
			IsHost   bool   `json:"isHost"`
			MainUser struct {
				Rtc      string `json:"rtc"`
				Rtm      string `json:"rtm"`
				UID      int64  `json:"uid"`
				Typename string `json:"__typename"`
			} `json:"mainUser"`
			ScreenShare struct {
				Rtc      string      `json:"rtc"`
				Rtm      interface{} `json:"rtm"`
				UID      int64       `json:"uid"`
				Typename string      `json:"__typename"`
			} `json:"screenShare"`
			Typename string `json:"__typename"`
		} `json:"joinChannel"`
		GetUser struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Typename string `json:"__typename"`
		} `json:"getUser"`
	} `json:"data"`
}

type JoinChannelFailed struct {
	Errors []struct {
		Message string   `json:"message"`
		Path    []string `json:"path"`
	} `json:"errors"`
	Data interface{} `json:"data"`
}


type Variables struct {
	passphrase string
}

type JoinRoomCreate struct {
	operationName string
	variables Variables
	query string
}

type CreateRoom struct {
	operationName string
	query string
}


var bearerTokenGlobal string
var createChannelDecoded CreateChannel



func TestWebOAuthHandler(t *testing.T) {
	bearerToken, err := utils.GenerateUUID()
	if err != nil {
		t.Error(err," : in WebOauthHandler ")
	}
	bearerTokenGlobal = bearerToken
	database, err := models.CreateDB(utils.GetDBURL())
	if err != nil {
		t.Error("DB Creation Failed!")
	}
	testingList := []struct{
		email string
		GivenName string
		bearerToken string
	}{
		{email:"test@testing.com",GivenName: "Testing Acc", bearerToken: bearerToken}, // ideal case
		{email:"test1@testing.com",GivenName: "Testing Acc 1", bearerToken: bearerToken}, //Same Bearer Token for both.
		{email:"",GivenName: "Testing Acc 1", bearerToken: bearerToken}, //Email nil.
		{email:"test2@testing.com",GivenName: "", bearerToken: bearerToken}, //Name Nil.
	}
	var user routes.GoogleOAuthUser
	for _,tc := range testingList {
		tc:=tc
		user.GivenName = tc.GivenName
		user.Email = tc.email
		routes.TokenGenerator(database,user,tc.bearerToken)

	}
}

//CHANGE HERE
func RoomCreationHandler(method string, url string,t *testing.T, status int, bearerToken string) CreateChannel {
	query := `mutation CreateChannel($title: String!, $enablePSTN: Boolean) {
				createChannel(title: $title, enablePSTN: $enablePSTN) {
					passphrase {
						host
						view
						__typename
					}
					channel
					title
					pstn {
						number
						dtmf
						__typename
					}
					__typename
					}
				}
			}`

	payload := CreateRoom{"JoinChannel",query}
	marshaledJson, _ := json.Marshal(payload)
	jsonPayload := strings.NewReader(string(marshaledJson))
	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, jsonPayload)
	if err != nil {
		t.Error(err," : in CreateRoomHandler ")
	}
	req.Header.Add("authorization", "Bearer "+bearerToken)
	req.Header.Add("content-type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		t.Error(err," : in CreateRoomHandler ")
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err," : in CreateRoomHandler ")
	}
	var decodedResponse CreateChannel
	if res.StatusCode != status{
		t.Fatal("Create Room Failed! Got", res.Status, " expected ", status)
	}
	if status==200{
		json.Unmarshal(body, &decodedResponse)
	}
	return decodedResponse
}

func TestRoomCreation(t *testing.T) {
	url := "http://localhost:8080/query"
	method := "POST"
	createChannelDecoded = RoomCreationHandler(method,url,t, 401, bearerTokenGlobal+"wef")
	createChannelDecoded = RoomCreationHandler(method,url,t, 200, bearerTokenGlobal)
}


func JoinRoomHandler(url string, method string, Passphrase string, t *testing.T, resStatus int, bearerToken string, status bool) {
	query := `query JoinChannel($passphrase: String!) {
					joinChannel(passphrase: $passphrase) {
						channel
						title
						isHost
						mainUser {
							rtc	
							rtm
							uid
							__typename
						}
						screenShare {
							rtc
							rtm
							uid
							__typename
						}
						__typename
					}
					getUser {
						name
						email
						__typename
					}
				}`
	variable := Variables{
		passphrase: Passphrase,
	}
	payload := JoinRoomCreate{"JoinChannel",variable,query}
	marshaledJson, _ := json.Marshal(payload)
	jsonPayload := strings.NewReader(string(marshaledJson))
	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, jsonPayload)
	if err != nil {
		t.Error(err," : in JoinRoomHandler ")
	}
	req.Header.Add("authorization", "Bearer "+bearerToken)
	req.Header.Add("content-type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		t.Error(err," : in JoinRoomHandler ")
	}
	defer res.Body.Close()
	if err != nil {
		t.Error(err," : in JoinRoomHandler ")
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err," : in CreateRoomHandler ")
	}
	if res.StatusCode != resStatus{
		t.Fatal("Join Room Failed!")
	}else if resStatus ==200{

		if status{
			var decodedResponse JoinChannelSuccess
			err = json.Unmarshal(body, &decodedResponse)
			if err != nil{
				t.Fatal("Create Room Failed! Got", res.Status, " expected ",resStatus)
			}
		}else{
			var decodedResponse JoinChannelFailed
			err = json.Unmarshal(body, &decodedResponse)
			if err != nil{
				t.Fatal("Create Room Failed! Got", res.Status, " expected ",resStatus)
			}
		}
	}
}

func TestJoinRoom(t *testing.T) {
	url := "http://localhost:8080/query"
	method := "POST"

	testingList := []struct{
		passPhrase string
		isHost bool
		responseStatus int
		bearerToken string
		status bool
	}{
		{passPhrase:createChannelDecoded.Data.CreateChannel.Passphrase.Host,isHost: true, bearerToken: bearerTokenGlobal, responseStatus:200, status:true},
		{passPhrase:createChannelDecoded.Data.CreateChannel.Passphrase.View,isHost: false, bearerToken: bearerTokenGlobal, responseStatus:200, status:true},
		{passPhrase:createChannelDecoded.Data.CreateChannel.Passphrase.Host+"test",isHost: true, bearerToken: bearerTokenGlobal, responseStatus:200, status:false},
		{passPhrase:createChannelDecoded.Data.CreateChannel.Passphrase.View+"test",isHost: false, bearerToken: bearerTokenGlobal, responseStatus:200, status:false},
		{passPhrase:createChannelDecoded.Data.CreateChannel.Passphrase.Host,isHost: true, bearerToken: bearerTokenGlobal+"wef", responseStatus:401, status:false},
		{passPhrase:createChannelDecoded.Data.CreateChannel.Passphrase.View,isHost: false, bearerToken: bearerTokenGlobal+"ef", responseStatus:401, status:false},
	}
	for _,tc := range testingList {
		JoinRoomHandler(url, method, tc.passPhrase, t, tc.responseStatus, tc.bearerToken, tc.status)
	}
}

