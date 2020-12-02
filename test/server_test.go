package test

import (
	"github.com/samyak-jain/agora_backend/migrations"
	"github.com/samyak-jain/agora_backend/models"
	"github.com/samyak-jain/agora_backend/utils"
	"github.com/stretchr/testify/suite"
)

type GraphQLTestSuite struct {
	suite.Suite
	DB    *models.Database
	Token string
}

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
	variables     Variables
	query         string
}

type CreateRoom struct {
	operationName string
	query         string
}

var bearerTokenGlobal string
var createChannelDecoded CreateChannel

func (suite *GraphQLTestSuite) SetupTest() {
	r := suite.Require()

	utils.SetupConfig()
	database, err := models.CreateDB(utils.GetDBURL())
	suite.DB = database

	migrations.RunMigration(suite.DB)

	r.NoError(err, "Error initializing database")
}

// func TestWebOAuthHandler(t *testing.T) {
// 	bearerToken, err := utils.GenerateUUID()
// 	if err != nil {
// 		t.Error(err, " : in WebOauthHandler ")
// 	}
// 	bearerTokenGlobal = bearerToken
// 	database, err := models.CreateDB(utils.GetDBURL())
// 	if err != nil {
// 		t.Error("DB Creation Failed!")
// 	}
// 	testingList := []struct {
// 		email       string
// 		GivenName   string
// 		bearerToken string
// 	}{
// 		{email: "test@testing.com", GivenName: "Testing Acc", bearerToken: bearerToken},    // ideal case
// 		{email: "test1@testing.com", GivenName: "Testing Acc 1", bearerToken: bearerToken}, //Same Bearer Token for both.
// 		{email: "", GivenName: "Testing Acc 1", bearerToken: bearerToken},                  //Email nil.
// 		{email: "test2@testing.com", GivenName: "", bearerToken: bearerToken},              //Name Nil.
// 	}
// 	var user routes.GoogleOAuthUser
// 	for _, tc := range testingList {
// 		tc := tc
// 		user.GivenName = tc.GivenName
// 		user.Email = tc.email
// 		routes.TokenGenerator(database, user, tc.bearerToken)

// 	}
// }

// //CHANGE HERE
// func RoomCreationHandler(method string, url string, t *testing.T, status int, bearerToken string) CreateChannel {
// 	query := `mutation CreateChannel($title: String!, $enablePSTN: Boolean) {
// 				createChannel(title: $title, enablePSTN: $enablePSTN) {
// 					passphrase {
// 						host
// 						view
// 					}
// 					channel
// 					title
// 					pstn {
// 						number
// 						dtmf
// 					}
// 					}
// 				}
// 			}`

// 	database, err := models.CreateDB(utils.GetDBURL())
// 	if err != nil {
// 		t.Fatal("DB Connection Failed!")
// 	}
// 	var GraphQLTestSuite = GraphQLTestSuite{
// 		DB:          database,
// 		bearerToken: bearerToken,
// 	}
// 	t.Log(GraphQLTestSuite)
// 	GraphQLTestSuite.SetupSuite()
// 	config := generated.Config{
// 		Resolvers: &graph.Resolver{DB: GraphQLTestSuite.DB},
// 	}
// 	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(config)))
// 	var decodedResponse JoinRoomCreate
// 	c.MustPost(query, &decodedResponse)
// 	fmt.Print(&decodedResponse)
// 	//return decodedResponse
// 	return CreateChannel{}
// }

// func TestRoomCreation(t *testing.T) {
// 	url := "http://localhost:8080/query"
// 	method := "POST"
// 	createChannelDecoded = RoomCreationHandler(method, url, t, 401, bearerTokenGlobal+"wef")
// 	createChannelDecoded = RoomCreationHandler(method, url, t, 200, bearerTokenGlobal)
// }

// func (suite *GraphQLTestSuite) JoinRoomHandler(Passphrase string, t *testing.T, bearerToken string) {
// 	query := fmt.Sprintf(`query JoinChannel(%s: String!) {
// 					joinChannel(passphrase: %s) {
// 						channel
// 						title
// 						isHost
// 						mainUser {
// 							rtc
// 							rtm
// 							uid
// 						}
// 						screenShare {
// 							rtc
// 							rtm
// 							uid
// 						}
// 					}
// 					getUser {
// 						name
// 						email
// 					}
// 				}`, Passphrase, Passphrase)

// 	config := generated.Config{
// 		Resolvers: &graph.Resolver{DB: suite.DB},
// 	}
// 	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(config)))
// 	var decodedResponse JoinChannelSuccess
// 	c.MustPost(query, &decodedResponse)
// 	fmt.Print(&decodedResponse)
// }

// func (suite *GraphQLTestSuite) TestSubtest() {
// 	suite.TestSubtestRunCount++
// 	for _, tc := range []struct {
// 		passPhrase     string
// 		isHost         bool
// 		responseStatus int
// 		bearerToken    string
// 		status         bool
// 	}{
// 		{passPhrase: createChannelDecoded.Data.CreateChannel.Passphrase.Host, isHost: true, bearerToken: bearerTokenGlobal, responseStatus: 200, status: true},
// 		{passPhrase: createChannelDecoded.Data.CreateChannel.Passphrase.View, isHost: false, bearerToken: bearerTokenGlobal, responseStatus: 200, status: true},
// 		{passPhrase: createChannelDecoded.Data.CreateChannel.Passphrase.Host + "test", isHost: true, bearerToken: bearerTokenGlobal, responseStatus: 200, status: false},
// 		{passPhrase: createChannelDecoded.Data.CreateChannel.Passphrase.View + "test", isHost: false, bearerToken: bearerTokenGlobal, responseStatus: 200, status: false},
// 		{passPhrase: createChannelDecoded.Data.CreateChannel.Passphrase.Host, isHost: true, bearerToken: bearerTokenGlobal + "wef", responseStatus: 401, status: false},
// 		{passPhrase: createChannelDecoded.Data.CreateChannel.Passphrase.View, isHost: false, bearerToken: bearerTokenGlobal + "ef", responseStatus: 401, status: false},
// 	} {
// 		suite.Run(tc.passPhrase, func() {
// 			query := fmt.Sprintf(`query JoinChannel(%s: String!) {
// 					joinChannel(passphrase: %s) {
// 						channel
// 						title
// 						isHost
// 						mainUser {
// 							rtc
// 							rtm
// 							uid
// 						}
// 						screenShare {
// 							rtc
// 							rtm
// 							uid
// 						}
// 					}
// 					getUser {
// 						name
// 						email
// 					}
// 				}`, tc.passPhrase, tc.passPhrase)

// 			config := generated.Config{
// 				Resolvers: &graph.Resolver{DB: suite.DB},
// 			}
// 			c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(config)))
// 			var decodedResponse JoinChannelSuccess
// 			c.MustPost(query, &decodedResponse)
// 			fmt.Print(&decodedResponse)
// 			suite.Equal(true, true)
// 		})

// 	}
// }

// func TestJoinRoom(t *testing.T) {
// 	suite := new(GraphQLTestSuite)

// 	database, err := models.CreateDB(utils.GetDBURL())
// 	if err != nil {
// 		t.Fatal("DB Connection Failed!")
// 	}
// 	suite.DB = database
// 	t.Log(suite)
// 	suite.Run("Testing Join Room", suite.TestSubtest)
// }
