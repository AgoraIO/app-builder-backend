// ********************************************
// Copyright © 2021 Agora Lab, Inc., all rights reserved.
// AppBuilder and all associated components, source code, APIs, services, and documentation
// (the “Materials”) are owned by Agora Lab, Inc. and its licensors.  The Materials may not be
// accessed, used, modified, or distributed for any purpose without a license from Agora Lab, Inc.
// Use without a license or in violation of any license terms and conditions (including use for
// any purpose competitive to Agora Lab, Inc.’s business) is strictly prohibited.  For more
// information visit https://appbuilder.agora.io.
// *********************************************

package models

import (
	"database/sql"
	"time"
)

// UserAccount model contains all relevant details of a particular user
type UserAccount struct {
	ID         int64          `db:"id"`
	UserName   sql.NullString `db:"user_name"`
	Email      string         `db:"email"`
	Identifier string         `db:"identifier"`
}

type Auth struct {
	ID           int64     `db:"id"`
	Code         string    `db:"code"`
	AccessToken  string    `db:"access_token"`
	RefreshToken string    `db:"refresh_token"`
	TokenType    string    `db:"token_type"`
	Expiry       time.Time `db:"expiry"`
}

// Token stores the token of a user
type Token struct {
	ID      int64  `db:"id"`
	TokenID string `db:"token_id"`
	UserID  int64  `db:"user_id"`
}

// GetAllTokens fetches the token id of all the tokens of that user
// func (u *User) GetAllTokens() []string {
// 	var tokens []string

// 	for index := range u.Tokens {
// 		tokens = append(tokens, u.Tokens[index].TokenID)
// 	}

// 	return tokens
// }
