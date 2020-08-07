package models

// User model contains all relevant details of a particular user
type User struct {
	token  string
	name   string
	email  string
	isHost bool
}
