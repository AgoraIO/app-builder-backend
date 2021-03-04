package oauth

import (
	"context"
	"errors"

	"github.com/coreos/go-oidc"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/slack"
)

// GetOAuthConfig makes the oauth2 config for the relevant site
func GetOAuthConfig(site string, redirectURI string) (*oauth2.Config, *string, error) {
	var provider *oidc.Provider
	var err error

	ctx := context.Background()

	switch site {
	case "google":
		provider, err = oidc.NewProvider(ctx, "https://accounts.google.com")
		if err != nil {
			log.Error().Err(err).Msg("Google Provider failed")
			return nil, nil, err
		}
	case "microsoft":
		provider, err = oidc.NewProvider(ctx, "https://login.microsoftonline.com/common")
		if err != nil {
			log.Error().Err(err).Msg("Microsoft Provider failed")
			return nil, nil, err
		}
	case "slack":
		userInfoURL := "https://slack.com/api/users.info"
		return &oauth2.Config{
			ClientID:     viper.GetString("CLIENT_ID"),
			ClientSecret: viper.GetString("CLIENT_SECRET"),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			Endpoint:     slack.Endpoint,
			RedirectURL:  redirectURI,
		}, &userInfoURL, nil
	case "apple":
		userInfoURL := "https://slack.com/api/users.info"
		return &oauth2.Config{
			ClientID:     viper.GetString("CLIENT_ID"),
			ClientSecret: viper.GetString("CLIENT_SECRET"),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://appleid.apple.com/auth/authorize",
				TokenURL: "https://appleid.apple.com/auth/token",
			},
			RedirectURL: redirectURI,
		}, &userInfoURL, nil
	default:
		log.Error().Msg("Unknown state parameter passed")
		return nil, nil, errors.New("Unknow state parameter passed")
	}

	var claims struct {
		UserInfoEndpoint string `json:"userinfo_endpoint"`
	}

	if err := provider.Claims(&claims); err != nil || claims.UserInfoEndpoint == "" {
		log.Error().Err(err).Interface("provider", provider).Msg("Could not get userinfo from claims")
		return nil, nil, errors.New("Could not get userinfo from claims")
	}

	return &oauth2.Config{
		ClientID:     viper.GetString("CLIENT_ID"),
		ClientSecret: viper.GetString("CLIENT_SECRET"),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURI,
	}, &claims.UserInfoEndpoint, nil
}

// ExchangeToken gets the access token from the code.
func ExchangeToken(oauthConfig oauth2.Config, oauthDetails Details) (*oauth2.Token, *string, error) {
	token, err := oauthConfig.Exchange(oauth2.NoContext, oauthDetails.Code)
	if err != nil {

	}

	// Special case for Apple OAuth.
	// Apple OAuth flow does not have a serparate UserInfo API.
	// All UserInfo is fetched from the id_token which is a JWT containing the info
	if oauthDetails.OAuthSite == "apple" {
		id_token := token.Extra("id_token")
	}
}
