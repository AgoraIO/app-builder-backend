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
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/coreos/go-oidc"
	"github.com/rs/zerolog/log"
	"github.com/samyak-jain/agora_backend/pkg/models"
	"github.com/samyak-jain/agora_backend/utils"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
	"golang.org/x/oauth2/slack"
)

// User contains all the information that we get as a response from oauth
type User struct {
	ID            string `json:"sub"`
	Name          string `json:"given_name"`
	Email         string
	EmailVerified bool `json:"verified_email"`
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
func (router *ServiceRouter) Handler(w http.ResponseWriter, r *http.Request) (*string, *string, *string, error) {
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

	var userData models.UserAccount
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
		var userName sql.NullString
		if userInfo.Name == "" {
			userName = sql.NullString{Valid: false}
		} else {
			userName = sql.NullString{String: userInfo.Name, Valid: true}
		}
		err = statement.Get(&userID, &models.UserAccount{
			Identifier: userInfo.ID,
			UserName:   userName,
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
func (o *ServiceRouter) OAuth(w http.ResponseWriter, r *http.Request) {
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

// GetOAuthConfig makes the oauth2 config for the relevant site
func (r *ServiceRouter) GetOAuthConfig(site string, redirectURI string) (*oauth2.Config, *oidc.Provider, error) {
	var provider *oidc.Provider
	var err error

	ctx := context.Background()

	var client_id string
	var client_secret string

	switch site {
	case "google":
		provider, err = oidc.NewProvider(ctx, "https://accounts.google.com")
		client_id = viper.GetString("GOOGLE_CLIENT_ID")
		client_secret = viper.GetString("GOOGLE_CLIENT_SECRET")
		if err != nil {
			r.Logger.Error().Err(err).Msg("Google Provider failed")
			return nil, nil, err
		}
	case "microsoft":
		return &oauth2.Config{
			ClientID:     viper.GetString("MICROSOFT_CLIENT_ID"),
			ClientSecret: viper.GetString("MICROSOFT_CLIENT_SECRET"),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "offline_access"},
			Endpoint:     microsoft.AzureADEndpoint("common"),
			RedirectURL:  redirectURI,
		}, nil, nil
	case "slack":
		return &oauth2.Config{
			ClientID:     viper.GetString("SLACK_CLIENT_ID"),
			ClientSecret: viper.GetString("SLACK_CLIENT_SECRET"),
			Scopes:       []string{"users.profile:read"},
			Endpoint:     slack.Endpoint,
			RedirectURL:  redirectURI,
		}, nil, nil
	case "apple":
		provider, err = oidc.NewProvider(ctx, "https://appleid.apple.com")
		if err != nil {
			r.Logger.Error().Err(err).Msg("Apple Provider failed")
			return nil, nil, err
		}
		client_id = viper.GetString("APPLE_CLIENT_ID")
		client_secret, err = GenerateAppleClientSecret(viper.GetString("APPLE_PRIVATE_KEY"), viper.GetString("APPLE_TEAM_ID"), client_id, viper.GetString("APPLE_KEY_ID"))
		if err != nil {
			r.Logger.Error().Err(err).Msg("Could not generate Apple Client Secret")
			return nil, nil, err
		}

	default:
		r.Logger.Error().Msg("Unknown state parameter passed")
		return nil, nil, errors.New("Unknow state parameter passed")
	}

	if client_id == "" || client_secret == "" {
		r.Logger.Error().Str("ID", client_id).Str("Secret", client_secret).Msg("No Client ID or Client Secret")
		return nil, nil, errors.New("Invalid Config")
	}

	return &oauth2.Config{
		ClientID:     client_id,
		ClientSecret: client_secret,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURI,
	}, provider, nil
}

// GetUserInfo fetches the User Info from the Open ID Endpoint
func (r *ServiceRouter) GetUserInfo(oauthConfig oauth2.Config, oauthDetails Details, provider *oidc.Provider) (*User, error) {

	var tokenData models.Auth
	var token *oauth2.Token
	err := r.DB.Get(&tokenData, "SELECT id, code, access_token, refresh_token, token_type, expiry FROM credentials WHERE code=$1", oauthDetails.Code)
	if err != nil {
		r.Logger.Debug().Msg("Code not found in database")

		token, err = oauthConfig.Exchange(oauth2.NoContext, oauthDetails.Code)
		if err != nil {
			r.Logger.Error().Err(err).Interface("OAuth Details", oauthDetails).Interface("config", oauthConfig).Msg("OAuth Token Exchange failed")
			return nil, err
		}

		_, err = r.DB.NamedExec("INSERT INTO credentials (code, access_token, refresh_token, token_type, expiry) VALUES (:code, :access_token, :refresh_token, :token_type, :expiry)", &models.Auth{
			Code:         oauthDetails.Code,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			TokenType:    token.TokenType,
			Expiry:       token.Expiry,
		})
		if err != nil {
			r.Logger.Error().Err(err).Msg("Cannot insert credentials")
		}
	} else {
		token.AccessToken = tokenData.AccessToken
		token.RefreshToken = tokenData.RefreshToken
		token.Expiry = tokenData.Expiry
		token.TokenType = tokenData.TokenType

		tokenSource := oauthConfig.TokenSource(oauth2.NoContext, token)
		newToken, err := tokenSource.Token()
		if err != nil {
			return nil, err
		}

		if newToken.AccessToken != token.AccessToken {
			r.DB.NamedExec("UPDATE credentials SET access_token = ':access_token' WHERE code = ':code'", &models.Auth{
				AccessToken: newToken.AccessToken,
				Code:        oauthDetails.Code,
			})
		}

		token = newToken
	}

	if provider == nil {
		if oauthDetails.OAuthSite == "slack" {
			// Adding this since Slack does not publicly publish it's .well-known discovery URL.
			// So we will have to manually hard code the UserInfo URL until we find that URL
			userInfoURL := "https://slack.com/api/users.profile.get"

			type authedSlackUser struct {
				ID string
			}

			authedUser, ok := token.Extra("user_id").(string)
			if !ok {
				r.Logger.Error().Str("OAuth Details", oauthDetails.Code).Interface("OAuth Exchange", token).Msg("No UserID in Slack OAuth Response")
				return nil, errors.New("No UserID in Slack OAuth Response")
			}

			client := oauthConfig.Client(oauth2.NoContext, token)

			data := url.Values{}
			data.Set("user", authedUser)
			response, err := client.PostForm(userInfoURL, data)
			if err != nil {
				r.Logger.Error().Err(err).Str("OAuth Details", oauthDetails.Code).Str("token", token.AccessToken).Msg("Could not fetch user info details")
				return nil, err
			}
			defer response.Body.Close()

			contents, err := ioutil.ReadAll(response.Body)
			if err != nil {
				r.Logger.Error().Interface("Response Body", response.Body).Err(err).Msg("Could not read response body")
				return nil, err
			}

			type SlackUser struct {
				Name  string `json:"display_name_normalized"`
				Email string
			}

			type SlackResponse struct {
				Ok      bool
				Profile *SlackUser `json:"profile,omitempty"`
				Error   string     `json:"error,omitempty"`
			}

			var user SlackResponse
			err = json.Unmarshal(contents, &user)
			if err != nil {
				r.Logger.Error().Err(err).Str("body", string(contents)).Msg("Could not parse response body")
				return nil, err
			}

			if user.Error != "" {
				r.Logger.Error().Str("Error", user.Error).Str("id", authedUser).Str("body", string(contents)).Msg("Could not fetch Userinfo for slack")
				return nil, errors.New(user.Error)
			}

			return &User{ID: authedUser, Name: user.Profile.Name, Email: user.Profile.Email, EmailVerified: true}, nil
		}

		if oauthDetails.OAuthSite == "microsoft" {
			client := &http.Client{}
			req, err := http.NewRequest("GET", "https://graph.microsoft.com/oidc/userinfo", nil)
			if err != nil {
				log.Error().Err(err).Str("code", oauthDetails.Code).Str("token", token.AccessToken).Msg("Could not fetch user info details")
				return nil, err
			}

			bearer := "Bearer " + token.AccessToken
			req.Header.Add("Authorization", bearer)

			response, err := client.Do(req)
			if err != nil {
				log.Error().Err(err).Str("code", oauthDetails.Code).Str("token", token.AccessToken).Msg("Could not fetch user info details")
				return nil, err
			}

			defer response.Body.Close()

			contents, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Error().Err(err).Msg("Could not read response body")
				return nil, err
			}

			var user *User
			err = json.Unmarshal(contents, &user)
			if err != nil {
				log.Error().Err(err).Str("body", string(contents)).Msg("Could not parse response body")
				return nil, err
			}

			user.EmailVerified = true
			return user, nil
		}

		r.Logger.Error().Interface("OAuth Config", oauthConfig).Interface("OAuth Details", oauthDetails).Msg("Provider should not be nil")
		return nil, errors.New("Provider should not be nil")
	}

	if oauthDetails.OAuthSite == "apple" {
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			r.Logger.Error().Interface("token", token).Msg("Could not get id_token from apple token")
			return nil, errors.New("Could not get id_token from apple token")
		}
		idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: oauthConfig.ClientID})
		idToken, err := idTokenVerifier.Verify(oauth2.NoContext, rawIDToken)
		if err != nil {
			r.Logger.Error().Str("rawIDToken", rawIDToken).Interface("idTokenVerifier", idTokenVerifier).Interface("OAuth Config", oauthConfig).Interface("OAuth Details", oauthDetails).Msg("Could not verify id_token")
			return nil, errors.New("Could not verify id_token")
		}

		// Get Email from idToken
		var claims struct {
			Email string `json:"email"`
		}

		if err := idToken.Claims(&claims); err != nil {
			return &User{ID: idToken.Subject, EmailVerified: true}, nil
		}

		return &User{ID: idToken.Subject, Email: claims.Email, EmailVerified: true}, nil

	}

	tokenSource := oauthConfig.TokenSource(oauth2.NoContext, token)
	userInfo, err := provider.UserInfo(oauth2.NoContext, tokenSource)
	if err != nil {
		r.Logger.Error().Err(err).Str("code", oauthDetails.Code).Interface("config", oauthConfig).Interface("token", token).Msg("Fetching UserInfo Failed")
		return nil, err
	}

	return &User{
		ID:            userInfo.Subject,
		Name:          userInfo.Profile,
		Email:         userInfo.Email,
		EmailVerified: userInfo.EmailVerified,
	}, nil
}
