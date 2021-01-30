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
	DB                  *models.Database
	passPhrase          string
	isHost              bool
	responseStatus      int
	bearerToken         string
	status              bool
	TestSubtestRunCount int
}

func (suite *GraphQLTestSuite) SetupSuite() {
	//r := suite.Require()
	utils.SetupConfig()
	database, err := models.CreateDB(utils.GetDBURL())
	//r.NoError(err, "Error initializing database")
	fmt.Print(err)
	fmt.Print(database)
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

var bearerTokenGlobal string
var createChannelDecoded CreateChannel

func TestWebOAuthHandler(t *testing.T) {

	bearerToken, err := utils.GenerateUUID()
	if err != nil {
		t.Error(err, " : in WebOauthHandler ")
	}
	bearerTokenGlobal = bearerToken
	database, err := models.CreateDB(utils.GetDBURL())
	if err != nil {
		t.Error("DB Creation Failed!")
	}
	testingList := []struct {
		email       string
		GivenName   string
		bearerToken string
	}{
		{email: "test@testing.com", GivenName: "Testing Acc", bearerToken: bearerToken},    // ideal case
		{email: "test1@testing.com", GivenName: "Testing Acc 1", bearerToken: bearerToken}, //Same Bearer Token for both.
		{email: "", GivenName: "Testing Acc 1", bearerToken: bearerToken},                  //Email nil.
		{email: "test2@testing.com", GivenName: "", bearerToken: bearerToken},              //Name Nil.
	}
	var user routes.GoogleOAuthUser
	var tokenData models.Token
	for _, tc := range testingList {
		tc := tc
		user.GivenName = tc.GivenName
		user.Email = tc.email
		routes.TokenGenerator(database, user, tc.bearerToken)
		assert.Equal(t, database.Where("token_id = ?", tc.bearerToken).First(&tokenData).RecordNotFound(), false)
	}
}

//CHANGE HERE
func RoomCreationHandler(t *testing.T, bearerToken string) CreateChannel {
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
	var GraphQLTestSuite = GraphQLTestSuite{
		//DB:          database,
		bearerToken: bearerToken,
	}
	GraphQLTestSuite.SetupSuite()
	config := generated.Config{
		Resolvers: &graph.Resolver{DB: GraphQLTestSuite.DB},
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

func TestRoomCreation(t *testing.T) {
	createChannelDecoded = RoomCreationHandler(t, bearerTokenGlobal+"wef") // Not Authorized
	createChannelDecoded = RoomCreationHandler(t, bearerTokenGlobal)       // Working case.
}

func TestJoinRoom(t *testing.T) {
	database, err := models.CreateDB(utils.GetDBURL())
	if err != nil {
		t.Fatal("DB Connection Failed!")
	}
	var GraphQLTestSuite = GraphQLTestSuite{
		DB:          database,
		bearerToken: bearerTokenGlobal,
	}
	GraphQLTestSuite.SetupSuite()
	config := generated.Config{
		Resolvers: &graph.Resolver{DB: GraphQLTestSuite.DB},
	}
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(config)))
	for _, tc := range []struct {
		passPhrase     string
		isHost         bool
		responseStatus int
		bearerToken    string
		status         bool
	}{
		{passPhrase: createChannelDecoded.CreateChannel.Passphrase.Host, isHost: true, bearerToken: bearerTokenGlobal, status: true},
		{passPhrase: createChannelDecoded.CreateChannel.Passphrase.View, isHost: false, bearerToken: bearerTokenGlobal, status: true},
		{passPhrase: createChannelDecoded.CreateChannel.Passphrase.Host + "test", isHost: true, bearerToken: bearerTokenGlobal, status: false},
		{passPhrase: createChannelDecoded.CreateChannel.Passphrase.View + "test", isHost: false, bearerToken: bearerTokenGlobal, status: false},
		{passPhrase: createChannelDecoded.CreateChannel.Passphrase.Host, isHost: true, bearerToken: bearerTokenGlobal + "wef", status: false},
		{passPhrase: createChannelDecoded.CreateChannel.Passphrase.View, isHost: false, bearerToken: bearerTokenGlobal + "ef", status: false},
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
