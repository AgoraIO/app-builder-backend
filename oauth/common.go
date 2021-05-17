package oauth

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path"

	"github.com/rs/zerolog/log"
	"github.com/samyak-jain/agora_backend/pkg/video_conferencing/models"
	"github.com/samyak-jain/agora_backend/utils"
	"github.com/spf13/viper"
)

// User contains all the information that we get as a response from oauth
type User struct {
	ID            string
	Name          string `json:"given_name"`
	Email         string
	EmailVerified bool `json:"verified_email"`
}

// RouterOAuth refers to all the oauth endpoints
type RouterOAuth struct {
	DB     *models.Database
	Logger *utils.Logger
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
	Platform    string
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

	platform := parsedState.Get("platform")

	// Lat's assume by default that we are on Web
	if platform == "" {
		platform = "platform"
	}

	return &Details{
		Code:        code,
		RedirectURL: redirect,
		BackendURL:  finalBackendURL,
		OAuthSite:   site,
		Platform:    platform,
	}, nil
}

// Handler is the handler that will do most of the heavy lifting for OAuth
func (router *RouterOAuth) Handler(w http.ResponseWriter, r *http.Request) (*string, *string, *string, error) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		router.Logger.Error().Err(err).Msg("Could not parse form request")
		return nil, nil, nil, err
	}

	oauthDetails, err := parseState(r)
	router.Logger.Debug().Interface("OAuth Details", oauthDetails).Msg("OAuth Debug Information")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, nil, err
	}

	oauthConfig, provider, err := router.GetOAuthConfig(oauthDetails.OAuthSite, oauthDetails.BackendURL+"/oauth")
	router.Logger.Debug().Interface("OAuth Config", oauthConfig).Interface("Provider", provider).Msg("OAuth Configuration Debug Information")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, nil, err
	}

	userInfo, err := router.GetUserInfo(*oauthConfig, *oauthDetails, provider)
	router.Logger.Debug().Interface("User Info", userInfo).Msg("Debug User Information")
	if err != nil {
		return nil, nil, nil, err
	}

	ok, err := router.AllowListValidator(userInfo.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Str("email", userInfo.Email).Str("Sub", userInfo.ID).Interface("OAuth Details", oauthDetails).Interface("OAuth Config", oauthConfig).Msg("Email cannot be validated in Allow List")
		return nil, nil, nil, err
	}

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Error().Str("Email", userInfo.Email).Msg("Email not found in Allow List")
		return nil, nil, nil, errors.New("Email not found in Allow List")
	}

	if !userInfo.EmailVerified {
		w.WriteHeader(http.StatusBadRequest)
		log.Error().Str("Sub", userInfo.ID).Interface("OAuth Details", oauthDetails).Interface("OAuth Config", oauthConfig).Msg("Email is not verified")
		return nil, nil, nil, errors.New("Email is not verified")
	}

	bearerToken, err := utils.GenerateUUID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("Could not generate bearer token")
		return nil, nil, nil, err
	}

	var userData models.User
	err = router.DB.Get(&userData, "SELECT id, identifier, user_name, email FROM users WHERE email=$1", userInfo.Email)

	if err != nil {
		tx := router.DB.MustBegin()

		statement, err := tx.PrepareNamed("INSERT INTO users (identifier, user_name, email) VALUES (:identifier, :user_name, :email) RETURNING id")

		if err != nil {
			router.Logger.Error().Err(err).Str("identifier", userInfo.ID).Msg("Could not insert user")
			tx.Rollback()
			return nil, nil, nil, err
		}

		var userID int64
		err = statement.Get(&userID, &models.User{
			Identifier: userInfo.ID,
			UserName:   sql.NullString{String: userInfo.Name, Valid: true},
			Email:      userInfo.Email,
		})
		if err != nil {
			router.Logger.Error().Err(err).Str("identifier", userInfo.ID).Msg("Could not fetch User Database ID")
			tx.Rollback()
			return nil, nil, nil, err
		}

		_, err = tx.NamedExec("INSERT INTO tokens (token_id, user_id) VALUES (:token_id, :user_id)", &models.Token{
			TokenID: bearerToken,
			UserID:  userID,
		})

		if err != nil {
			router.Logger.Error().Err(err).Str("identifier", userInfo.ID).Str("token", bearerToken).Msg("Could not insert token")
			tx.Rollback()
			return nil, nil, nil, err
		}

		tx.Commit()
	} else {
		// userData := userDataList[0]
		_, err = router.DB.NamedExec("INSERT INTO tokens (token_id, user_id) VALUES (:token_id, :user_id)", &models.Token{
			TokenID: bearerToken,
			UserID:  userData.ID,
		})

		if err != nil {
			router.Logger.Error().Err(err).Str("identifier", userInfo.ID).Str("token", bearerToken).Msg("Could not insert token")
			return nil, nil, nil, err
		}
	}

	return &oauthDetails.RedirectURL, &bearerToken, &oauthDetails.Platform, nil
}

// OAuth is a REST route that is called when the oauth provider redirects to here and provides the code
func (o *RouterOAuth) OAuth(w http.ResponseWriter, r *http.Request) {
	redirect, token, platform, err := o.Handler(w, r)
	if err != nil || platform == nil {
		log.Print(err)
		fmt.Fprint(w, err)
		return
	}

	if *platform == "web" {
		newURL, err := url.Parse(*redirect)
		if err != nil {
			log.Error().Err(err).Str("redirect_url", *redirect).Msg("Failed to parse redirect url")
			fmt.Fprint(w, err)
			return
		}

		newURL.Path = path.Join(newURL.Path, *token)

		fmt.Printf("%+v\n", newURL)

		http.Redirect(w, r, newURL.String(), http.StatusSeeOther)
	} else if *platform == "mobile" {
		t, err := template.ParseFiles("web/mobile.html")
		if err != nil {
			fmt.Fprint(w, "Internal Server Error")
			return
		}

		t.Execute(w, TokenTemplate{
			Token:  *token,
			Scheme: viper.GetString("SCHEME"),
		})
	} else if *platform == "desktop" {
		t, err := template.ParseFiles("web/desktop.html")
		if err != nil {
			fmt.Fprint(w, "Internal Server Error")
			return
		}

		t.Execute(w, TokenTemplate{
			Token: *token,
		})
	}
}
