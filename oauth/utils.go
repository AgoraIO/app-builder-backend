package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/coreos/go-oidc"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/slack"
)

// GetOAuthConfig makes the oauth2 config for the relevant site
func (r *Router) GetOAuthConfig(site string, redirectURI string) (*oauth2.Config, *oidc.Provider, error) {
	var provider *oidc.Provider
	var err error

	ctx := context.Background()

	switch site {
	case "google":
		provider, err = oidc.NewProvider(ctx, "https://accounts.google.com")
		if err != nil {
			r.Logger.Error().Err(err).Msg("Google Provider failed")
			return nil, nil, err
		}
	case "microsoft":
		provider, err = oidc.NewProvider(ctx, "https://login.microsoftonline.com/common")
		if err != nil {
			r.Logger.Error().Err(err).Msg("Microsoft Provider failed")
			return nil, nil, err
		}
	case "slack":
		return &oauth2.Config{
			ClientID:     viper.GetString("CLIENT_ID"),
			ClientSecret: viper.GetString("CLIENT_SECRET"),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			Endpoint:     slack.Endpoint,
			RedirectURL:  redirectURI,
		}, nil, nil
	case "apple":
		provider, err = oidc.NewProvider(ctx, "https://appleid.apple.com")
		if err != nil {
			r.Logger.Error().Err(err).Msg("Apple Provider failed")
			return nil, nil, err
		}
	default:
		r.Logger.Error().Msg("Unknown state parameter passed")
		return nil, nil, errors.New("Unknow state parameter passed")
	}

	return &oauth2.Config{
		ClientID:     viper.GetString("CLIENT_ID"),
		ClientSecret: viper.GetString("CLIENT_SECRET"),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURI,
	}, provider, nil
}

// GetUserInfo fetches the User Info from the Open ID Endpoint
func (r *Router) GetUserInfo(oauthConfig oauth2.Config, oauthDetails Details, provider *oidc.Provider) (*User, error) {
	token, err := oauthConfig.Exchange(oauth2.NoContext, oauthDetails.Code)
	if err != nil {
		r.Logger.Error().Err(err).Interface("OAuth Details", oauthDetails).Interface("config", oauthConfig).Msg("OAuth Token Exchange failed")
		return nil, err
	}

	if provider == nil {
		if oauthDetails.OAuthSite == "slack" {
			// Adding this since Slack does not publicly publish it's .well-known discovery URL.
			// So we will have to manually hard code the UserInfo URL until we find that URL
			userInfoURL := "https://slack.com/api/users.profile.get"

			type authedSlackUser struct {
				ID string
			}

			authedUser, ok := token.Extra("authed_user").(authedSlackUser)
			if !ok {
				r.Logger.Error().Str("OAuth Details", oauthDetails.Code).Interface("OAuth Exchange", token).Msg("No UserID in Slack OAuth Response")
				return nil, errors.New("No UserID in Slack OAuth Response")
			}

			response, err := http.Get(userInfoURL + token.AccessToken)
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
				r.Logger.Error().Str("Error", user.Error).Str("id", authedUser.ID).Str("body", string(contents)).Msg("Could not fetch Userinfo for slack")
				return nil, errors.New(user.Error)
			}

			return &User{ID: authedUser.ID, Name: user.Profile.Name, Email: user.Profile.Email, EmailVerified: true}, nil
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
func (r *Router) AllowListValidator(email string) (bool, error) {
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
