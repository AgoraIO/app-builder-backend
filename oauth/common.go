package oauth

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/samyak-jain/agora_backend/utils"

	"github.com/samyak-jain/agora_backend/models"
)

// User contains all the information that we get as a response from oauth
type User struct {
	ID            string
	Name          string `json:"given_name"`
	EmailVerified bool   `json:"verified_email"`
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

	oauthConfig, provider, err := GetOAuthConfig(oauthDetails.OAuthSite, oauthDetails.BackendURL+"/oauth/"+platform)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, err
	}

	userInfo, err := GetUserInfo(*oauthConfig, *oauthDetails, provider)
	if err != nil {
		return nil, nil, err
	}

	if !userInfo.EmailVerified {
		w.WriteHeader(http.StatusBadRequest)
		log.Error().Str("Sub", userInfo.ID).Interface("OAuth Details", oauthDetails).Interface("OAuth Config", oauthConfig).Msg("Email is not verified")
		return nil, nil, errors.New("Email is not verified")
	}

	bearerToken, err := utils.GenerateUUID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("Could not generate bearer token")
		return nil, nil, err
	}

	var userData models.User
	if db.Where("id = ?", userInfo.ID).First(&userData).RecordNotFound() {
		db.Create(&models.User{
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
