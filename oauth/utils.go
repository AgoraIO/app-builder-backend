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
		provider, err = oidc.NewProvider(ctx, "https://appleid.apple.com")
		if err != nil {
			log.Error().Err(err).Msg("Apple Provider failed")
			return nil, nil, err
		}
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

// GetUserInfo fetches the User Info from the Open ID Endpoint
func GetUserInfo(oauthConfig oauth2.Config, oauthDetails Details, provider *oidc.Provider) (*User, error) {
	token, err := oauthConfig.Exchange(oauth2.NoContext, oauthDetails.Code)
	if err != nil {
		log.Error().Err(err).Interface("OAuth Details", oauthDetails).Interface("config", oauthConfig).Msg("OAuth Token Exchange failed")
		return nil, err
	}

	if oauthDetails.OAuthSite == "apple" {
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			log.Error().Interface("token", token).Msg("Could not get id_token from apple token")
			return nil, errors.New("Could not get id_token from apple token")
		}
		idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: oauthConfig.ClientID})
		idToken, err := idTokenVerifier.Verify(oauth2.NoContext, rawIDToken)
		if err != nil {
			log.Error().Str("rawIDToken", rawIDToken).Interface("idTokenVerifier", idTokenVerifier).Interface("OAuth Config", oauthConfig).Interface("OAuth Details", oauthDetails).Msg("Could not verify id_token")
			return nil, errors.New("Could not verify id_token")
		}

		return &User{ID: idToken.Subject, EmailVerified: true}, nil

	}

	tokenSource := oauthConfig.TokenSource(oauth2.NoContext, token)
	userInfo, err := provider.UserInfo(oauth2.NoContext, tokenSource)
	if err != nil {
		log.Error().Err(err).Str("code", oauthDetails.Code).Interface("config", oauthConfig).Interface("token", token).Msg("Fetching UserInfo Failed")
		return nil, err
	}

	return &User{
		ID:            userInfo.Subject,
		Name:          userInfo.Profile,
		EmailVerified: userInfo.EmailVerified,
	}, nil
}
