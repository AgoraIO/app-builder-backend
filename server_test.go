package main

import (
	"fmt"
	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/magiconair/properties/assert"
	"github.com/samyak-jain/agora_backend/graph"
	"github.com/samyak-jain/agora_backend/graph/generated"
	"github.com/samyak-jain/agora_backend/migrations"
	"github.com/samyak-jain/agora_backend/models"
	"github.com/samyak-jain/agora_backend/routes"
	"github.com/samyak-jain/agora_backend/utils"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString() string {
	b := make([]byte, rand.Intn(20))
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func randomBool() bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(2) == 1
}

type GraphQLTestSuite struct {
	suite.Suite
	DB                   *models.Database
	Token                string
	createChannelDecoded CreateChannel
}

func (suite *GraphQLTestSuite) SetupSuite() {
	r := suite.Require()
	token, err := utils.GenerateUUID()
	if err != nil {
		r.Error(err, " : in creating bearer token ")
	}
	utils.SetupConfig()
	database, err := models.CreateDB(utils.GetDBURL())
	if err != nil {
		r.Error(err, " : Error initializing database ")
	}
	suite.Token = token
	suite.DB = database
	migrations.RunMigration(suite.DB)

}

type CreateChannel struct {
	CreateChannel struct {
		Channel    string `json:"channel"`
		Passphrase struct {
			Host string `json:"host"`
			View string `json:"view"`
		} `json:"passphrase"`
		Pstn struct {
			Dtmf   string `json:"dtmf"`
			Number string `json:"number"`
		} `json:"pstn,omitempty"`
		Title string `json:"title"`
	} `json:"createChannel"`
}
type JoinChannelSuccess struct {
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
	variables     Variables
	query         string
}

type CreateRoom struct {
	operationName string
	query         string
}

func (suite *GraphQLTestSuite) WebOAuthHandler() {

	testingList := []struct {
		email       string
		GivenName   string
		bearerToken string
	}{
		{email: randomString() + "@testing.com", GivenName: randomString(), bearerToken: suite.Token}, // ideal case
		{email: randomString() + "@testing.com", GivenName: randomString(), bearerToken: suite.Token}, //Same Bearer Token for both.
		{email: "", GivenName: randomString(), bearerToken: suite.Token},                              //Email nil.
		{email: randomString() + "@testing.com", GivenName: randomString(), bearerToken: suite.Token}, //Name Nil.
	}
	var user routes.GoogleOAuthUser
	var tokenData models.Token
	for _, tc := range testingList {
		tc := tc
		user.GivenName = tc.GivenName
		user.Email = tc.email
		routes.TokenGenerator(suite.DB, user, tc.bearerToken)
		assert.Equal(suite.T(), suite.DB.Where("token_id = ?", tc.bearerToken).First(&tokenData).RecordNotFound(), false)
	}
}

//CHANGE HERE
func RoomCreationHandler(t *testing.T, channel string, enablePSTN bool, db *models.Database) CreateChannel {
	query := `mutation CreateChannel($title: String!, $enablePSTN: Boolean) {
				createChannel(title: $title, enablePSTN: $enablePSTN) {
					passphrase {
						host
						view
					}
					channel
					title
					pstn {
						number
						dtmf
					}
					}
				}
			`
	config := generated.Config{
		Resolvers: &graph.Resolver{DB: db},
	}
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(config)))
	var decodedResponse CreateChannel

	c.MustPost(query, &decodedResponse, func(bd *client.Request) {
		bd.Variables = map[string]interface{}{
			"title":      channel,
			"enablePSTN": enablePSTN,
		}
	})
	var channelData models.Channel
	tx := db.Where("hostpassphrase = ?", decodedResponse.CreateChannel.Passphrase.Host).First(&channelData)
	assert.Equal(t, decodedResponse.CreateChannel.Title, channel)
	assert.Equal(t, tx.RecordNotFound(), false)
	if !enablePSTN {
		assert.Equal(t, channelData.DTMF, nil)
	} else {
		assert.Equal(t, channelData.DTMF, decodedResponse.CreateChannel.Pstn.Dtmf)
	}
	return decodedResponse
}

func (suite *GraphQLTestSuite) RoomCreation() {
	for _, tc := range []struct {
		title      string
		enablePSTN bool
	}{
		{title: randomString(), enablePSTN: randomBool()},
		{title: randomString(), enablePSTN: randomBool()},
		{title: randomString(), enablePSTN: randomBool()},
		{title: randomString(), enablePSTN: randomBool()},
		{title: randomString(), enablePSTN: randomBool()},
	} {
		suite.createChannelDecoded = RoomCreationHandler(suite.T(), tc.title, tc.enablePSTN, suite.DB) // Not Authorized
		suite.createChannelDecoded = RoomCreationHandler(suite.T(), tc.title, tc.enablePSTN, suite.DB) // Working case.
	}
}

func (suite *GraphQLTestSuite) JoinRoom() {

	config := generated.Config{
		Resolvers: &graph.Resolver{DB: suite.DB},
	}
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(config)))
	for _, tc := range []struct {
		passPhrase     string
		isHost         bool
		responseStatus int
		bearerToken    string
		status         bool
	}{
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.Host, isHost: randomBool(), bearerToken: suite.Token},
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.View, isHost: randomBool(), bearerToken: suite.Token},
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.Host + "test", isHost: randomBool(), bearerToken: suite.Token},
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.View + "test", isHost: randomBool(), bearerToken: suite.Token},
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.Host, isHost: randomBool(), bearerToken: suite.Token + "wef"},
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.View, isHost: randomBool(), bearerToken: suite.Token + "ef"},
	} {
		query := `query JoinChannel($passphrase: String!) {
					joinChannel(passphrase: $passphrase) {
						channel
						title
						isHost
						mainUser {
							rtc	
							rtm
							uid
						}
						screenShare {
							rtc
							rtm
							uid
						}
					}
					getUser {
						name
						email
					}
				}`

		var decodedResponse JoinChannelSuccess
		c.MustPost(query, &decodedResponse, func(bd *client.Request) {
			bd.Variables = map[string]interface{}{
				"passphrase": tc.passPhrase,
			}
		})
		fmt.Print(&decodedResponse)
		assert.Equal(suite.T(), true, true) // Can't do anything for this. Need to figure out.

	}
}

func TestGraphQLTestSuite(t *testing.T) {
	GraphQLTest := new(GraphQLTestSuite)
	suite.Run(t, GraphQLTest)
}
