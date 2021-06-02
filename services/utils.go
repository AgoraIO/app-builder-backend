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
	"crypto/x509"
	"encoding/pem"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/samyak-jain/agora_backend/pkg/models"
	"github.com/samyak-jain/agora_backend/utils"
	"github.com/spf13/viper"
)

// ServiceRouter refers to all the oauth endpoints
type ServiceRouter struct {
	DB     *models.Database
	Logger *utils.Logger
}

// AllowListValidator takes an email and searches the Allow List for a match
func (r *ServiceRouter) AllowListValidator(email string) (bool, error) {
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
GenerateAppleClientSecret generates the client secret used to make requests to the validation server.
The secret expires after 6 months
signingKey - Private key from Apple obtained by going to the keys section of the developer section
teamID - Your 10-character Team ID
clientID - Your Services ID, e.g. com.aaronparecki.services
keyID - Find the 10-char Key ID value from the portal
*/
func GenerateAppleClientSecret(signingKey, teamID, clientID, keyID string) (string, error) {
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
