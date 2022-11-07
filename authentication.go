package auth0cliauthorizer

import "time"

type Authentication struct {
	User   User   `json:"user"`
	Tokens Tokens `json:"tokens"`
}

type User struct {
	Nickname            string `json:"nickname"`
	Name                string `json:"name"`
	Picture             string `json:"picture"`
	Email               string `json:"email"`
	EmailVerified       bool   `json:"email_verified"`
	Sub                 string `json:"sub"`
	GivenName           string `json:"given_name"`
	FamilyName          string `json:"family_name"`
	MiddleName          string `json:"middle_name"`
	PreferredUsername   string `json:"preferred_username"`
	Profile             string `json:"profile"`
	Website             string `json:"website"`
	Gender              string `json:"gender"`
	Birthdate           string `json:"birthdate"`
	Zoneinfo            string `json:"zoneinfo"`
	Locale              string `json:"locale"`
	PhoneNumber         string `json:"phone_number"`
	PhoneNumberVerified bool   `json:"phone_number_verified"`
	Address             struct {
		Country string `json:"country"`
	} `json:"address"`
	UpdatedAt string `json:"updated_at"`

	Permissions []string `json:"permissions"`
}

type Tokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IdToken      string    `json:"id_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}
