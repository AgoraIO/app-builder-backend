package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc"
	"github.com/rs/zerolog/log"
	"github.com/samyak-jain/agora_backend/utils"

	"github.com/samyak-jain/agora_backend/models"
	"golang.org/x/oauth2"
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

// TokenTemplate is a struct that will be used to template the token into the html that will be served for Desktop and Mobile
type TokenTemplate struct {
	Token  string
	Scheme string
}

// Details contains all the OAuth related information parsed from the request
type Details struct {
	Code        string
	RedirectURL string
	BackendURL  string
	OAuthSite   string
}

func parseState(r *http.Request) (*Details, error) {
	code := r.FormValue("code")
	if len(code) <= 0 {
		log.Error().Str("code", code).Msg("Code is empty")
		return nil, errors.New("Code is empty")
	}

	state := r.FormValue("state")
	if len(state) <= 0 {
		log.Error().Str("state", state).Msg("State is empty")
		return nil, errors.New("State is empty")
	}

	decodedState, err := url.QueryUnescape(state)
	if err != nil {
		log.Error().Err(err).Msg("Could not url decode state")
		return nil, err
	}

	parsedState, err := url.ParseQuery(decodedState)
	if err != nil {
		log.Error().Err(err).Msg("Could not parse deocoded state")
		return nil, err
	}

	redirect := parsedState.Get("redirect")
	if len(redirect) <= 0 {
		log.Error().Str("redirect", redirect).Msg("Redirect URL is empty")
		return nil, errors.New("Redirect URL is empty")
	}

	backendURL := parsedState.Get("backend")
	if len(backendURL) <= 0 {
		log.Error().Str("backend", backendURL).Msg("Backend URL is empty")
		return nil, errors.New("Backend URL is empty")
	}

	// Remove trailing slash from URL
	runeBackendURL := []rune(backendURL)
	if runeBackendURL[len(runeBackendURL)-1] == '/' {
		runeBackendURL = runeBackendURL[:len(runeBackendURL)-1]
	}

	finalBackendURL := string(runeBackendURL)

	site := parsedState.Get("site")

	// Let's assume by default that we are using Google OAuth
	if site == "" {
		site = "google"
	}

	return &Details{
		Code:        code,
		RedirectURL: redirect,
		BackendURL:  finalBackendURL,
		OAuthSite:   site,
	}, nil
}

// Handler is the handler that will do most of the heavy lifting for OAuth
func Handler(w http.ResponseWriter, r *http.Request, db *models.Database, platform string) (*string, *string, error) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Error().Err(err).Msg("Could not parse form request")
		return nil, nil, err
	}

	oauthDetails, err := parseState(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, err
	}
	var provider *oidc.Provider

	oauthConfig, userInfoURL, err := GetOAuthConfig(oauthDetails.OAuthSite, oauthDetails.BackendURL+"/oauth/"+platform)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, err
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Str("code", oauthDetails.Code).Interface("config", oauthConfig).Msg("Could not exchange code for access token")
		return nil, nil, err
	}

	if oauthDetails.OAuthSite == "apple" {

	}

	response, err := http.Get(*userInfoURL + token.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Str("code", oauthDetails.Code).Str("token", token.AccessToken).Msg("Could not fetch user info details")
		return nil, nil, err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("Could not read response body")
		return nil, nil, err
	}

	var user GoogleOAuthUser
	err = json.Unmarshal(contents, &user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Str("body", string(contents)).Msg("Could not parse response body")
	}

	ctx := context.Background()
	userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		log.Error().Err(err).Str("token", token.AccessToken).Msg("Could not get user info from token")
	}

	if !userInfo.EmailVerified {
		w.WriteHeader(http.StatusBadRequest)
		log.Error().Err(err).Str("email", userInfo.Email).Msg("Email is not verified")
		return nil, nil, errors.New("Email is not verified")
	}

	bearerToken, err := utils.GenerateUUID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("Could not generate bearer token")
		return nil, nil, err
	}

	var userData models.User
	if db.Where("email = ?", user.Email).First(&userData).RecordNotFound() {
		db.Create(&models.User{
			Name:  user.GivenName,
			Email: user.Email,
			Tokens: []models.Token{{
				TokenID: bearerToken,
			}},
		})
	} else {
		db.Model(&userData).Association("Tokens").Append(models.Token{
			TokenID: bearerToken,
		})
	}

	return &oauthDetails.RedirectURL, &bearerToken, nil
}
