package oauth

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/samyak-jain/agora_backend/models"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleOAuthUser contains all the information that we get as a response from oauth in google
type GoogleOAuthUser struct {
	GivenName     string `json:"given_name"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string
	Locale        string
	ID            string
	Email         string
}

// Router refers to all the oauth endpoints
type Router struct {
	DB *models.Database
}

// Handler is the handler that will do most of the heavy lifting for OAuth
func Handler(w http.ResponseWriter, r *http.Request, db *models.Database) (*string, *int, error) {
	err := r.ParseForm()
	if err != nil {
		log.Panic(err)
		w.WriteHeader(http.StatusBadRequest)
	}

	code := r.FormValue("code")
	state := r.FormValue("state")

	decodedState, err := url.QueryUnescape(state)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, err
	}

	stateURL, err := url.Parse(decodedState)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, err
	}

	redirect := stateURL.Query().Get("redirect")
	var oauthConfig *oauth2.Config
	var userInfoURL string

	switch site := stateURL.Query().Get("site"); site {
	case "google":
		oauthConfig = &oauth2.Config{
			ClientID:     viper.GetString("CLIENT_ID"),
			ClientSecret: viper.GetString("CLIENT_SECRET"),
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
			RedirectURL:  "http://localhost:8080",
		}
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

	case "microsoft":
		oauthConfig = &oauth2.Config{}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, err
	}

	token, err := oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, err
	}

	response, err := http.Get(userInfoURL + token.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, err
	}

	var user GoogleOAuthUser
	err = json.Unmarshal(contents, &user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, err
	}

	if !user.VerifiedEmail {
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, errors.New("Email is not verified")
	}

	bearerToken, err := strconv.Atoi(uuid.NewV4().String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, err
	}

	db.Save(&models.User{
		Name:  user.GivenName,
		Token: bearerToken,
		Email: user.Email,
	})

	return &redirect, &bearerToken, nil
}
