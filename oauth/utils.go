package oauth

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"
	"github.com/samyak-jain/agora_backend/pkg/video_conferencing/models"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
	"golang.org/x/oauth2/slack"
)

// GetOAuthConfig makes the oauth2 config for the relevant site
func (r *RouterOAuth) GetOAuthConfig(site string, redirectURI string) (*oauth2.Config, *oidc.Provider, error) {
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
		client_secret, err = GenerateClientSecret(viper.GetString("APPLE_PRIVATE_KEY"), viper.GetString("APPLE_TEAM_ID"), client_id, viper.GetString("APPLE_KEY_ID"))
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
func (r *RouterOAuth) GetUserInfo(oauthConfig oauth2.Config, oauthDetails Details, provider *oidc.Provider) (*User, error) {

	var tokenData models.Auth
	var token *oauth2.Token
	err := r.DB.Get(&tokenData, "SELECT id, code, access_token, refresh_token, token_type, expiry FROM credentials WHERE code=$1", oauthDetails.Code)
	if err != nil {
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

			response, err := http.Get("https://graph.microsoft.com/oidc/userinfo" + token.AccessToken)
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

// AllowListValidator takes an email and searches the Allow List for a match
func (r *RouterOAuth) AllowListValidator(email string) (bool, error) {
	for _, value := range viper.GetStringSlice("ALLOW_LIST") {

		pattern := wildCardToRegexp(value)
		r.Logger.Debug().Str("Allow List Pattern", value).Str("Email", email).Str("Regex Pattern", pattern).Msg("Allow List Debug Information")

		match, err := regexp.MatchString(pattern, email)
		if err != nil {
			r.Logger.Error().Err(err).Str("Pattern", value).Str("Email", email).Msg("Could not match wildcard")
			return false, err
		}

		if match {
			r.Logger.Info().Str("Email", email).Str("Match", value).Msg("Allow list email matched")
			return true, nil
		}
	}

	r.Logger.Info().Str("Email", email).Msg("No match found for email in Allow List")
	return false, nil
}

// Converts a wildcard string to RegExp Pattern
// Taken from https://stackoverflow.com/a/64520572/4127046
func wildCardToRegexp(pattern string) string {
	var result strings.Builder
	for i, literal := range strings.Split(pattern, "*") {

		// Replace * with .*
		if i > 0 {
			result.WriteString(".*")
		}

		// Quote any regular expression meta characters in the
		// literal text.
		result.WriteString(regexp.QuoteMeta(literal))
	}
	return result.String()
}

/*
From: https://github.com/Timothylock/go-signin-with-apple/blob/828dfdd59ab1d83cc630247ec12f2efa2e9cd039/apple/secret.go
GenerateClientSecret generates the client secret used to make requests to the validation server.
The secret expires after 6 months
signingKey - Private key from Apple obtained by going to the keys section of the developer section
teamID - Your 10-character Team ID
clientID - Your Services ID, e.g. com.aaronparecki.services
keyID - Find the 10-char Key ID value from the portal
*/
func GenerateClientSecret(signingKey, teamID, clientID, keyID string) (string, error) {
	block, _ := pem.Decode([]byte(signingKey))
	if block == nil {
		return "", errors.New("empty block after decoding")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	// Create the Claims
	now := time.Now()
	claims := &jwt.StandardClaims{
		Issuer:    teamID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(time.Hour*24*180 - time.Second).Unix(), // 180 days
		Audience:  "https://appleid.apple.com",
		Subject:   clientID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["alg"] = "ES256"
	token.Header["kid"] = keyID

	return token.SignedString(privKey)
}
