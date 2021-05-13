package models

import (
	"database/sql"
)

// User model contains all relevant details of a particular user
type User struct {
	ID         int64          `db:"id"`
	UserName   sql.NullString `db:"user_name"`
	Email      string         `db:"email"`
	Identifier string         `db:"identifier"`
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
