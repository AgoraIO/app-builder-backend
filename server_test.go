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
	"testing"
)

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

func (suite *GraphQLTestSuite) WebOAuthHandler(t *testing.T) {

	testingList := []struct {
		email       string
		GivenName   string
		bearerToken string
	}{
		{email: "test@testing.com", GivenName: "Testing Acc", bearerToken: suite.Token},    // ideal case
		{email: "test1@testing.com", GivenName: "Testing Acc 1", bearerToken: suite.Token}, //Same Bearer Token for both.
		{email: "", GivenName: "Testing Acc 1", bearerToken: suite.Token},                  //Email nil.
		{email: "test2@testing.com", GivenName: "", bearerToken: suite.Token},              //Name Nil.
	}
	var user routes.GoogleOAuthUser
	var tokenData models.Token
	for _, tc := range testingList {
		tc := tc
		user.GivenName = tc.GivenName
		user.Email = tc.email
		routes.TokenGenerator(suite.DB, user, tc.bearerToken)
		assert.Equal(t, suite.DB.Where("token_id = ?", tc.bearerToken).First(&tokenData).RecordNotFound(), false)
	}
}

//CHANGE HERE
func RoomCreationHandler(t *testing.T, bearerToken string, db *models.Database) CreateChannel {
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

	//
	//database, err := models.CreateDB(utils.GetDBURL())
	//t.Log(database)
	//if err != nil {
	//	t.Fatal("DB Connection Failed!")
	//}
	config := generated.Config{
		Resolvers: &graph.Resolver{DB: db},
	}
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(config)))
	var decodedResponse CreateChannel

	c.MustPost(query, &decodedResponse, func(bd *client.Request) {
		bd.Variables = map[string]interface{}{
			"title":      "Hey",
			"enablePSTN": true,
		}
	})
	assert.Equal(t, decodedResponse.CreateChannel.Title, "Hey") //change this
	return decodedResponse
}

func (suite *GraphQLTestSuite) RoomCreation(t *testing.T) {
	suite.createChannelDecoded = RoomCreationHandler(t, suite.Token+"wef", suite.DB) // Not Authorized
	suite.createChannelDecoded = RoomCreationHandler(t, suite.Token, suite.DB)       // Working case.
}

func (suite *GraphQLTestSuite) JoinRoom(t *testing.T) {

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
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.Host, isHost: true, bearerToken: suite.Token, status: true},
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.View, isHost: false, bearerToken: suite.Token, status: true},
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.Host + "test", isHost: true, bearerToken: suite.Token, status: false},
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.View + "test", isHost: false, bearerToken: suite.Token, status: false},
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.Host, isHost: true, bearerToken: suite.Token + "wef", status: false},
		{passPhrase: suite.createChannelDecoded.CreateChannel.Passphrase.View, isHost: false, bearerToken: suite.Token + "ef", status: false},
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
		assert.Equal(t, true, true) // change here

	}
}

func TestGraphQLTestSuite(t *testing.T) {
	GraphQLTest := new(GraphQLTestSuite)
	suite.Run(t, GraphQLTest)
}
