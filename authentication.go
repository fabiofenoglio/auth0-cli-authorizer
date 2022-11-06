package auth0cliauthorizer

import "time"

type Authentication struct {
	User   User   `json:"user"`
	Tokens Tokens `json:"tokens"`
}

type User struct {
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
}

type Tokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}
