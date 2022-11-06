package auth0cliauthorizer

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type deviceCodeResponseDTO struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationUri         string `json:"verification_uri"`
	VerificationUriComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type tokenResponseDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type accessTokenContentDTO struct {
	jwt.RegisteredClaims
	Azp         string   `json:"azp"`
	Scope       string   `json:"scope"`
	Permissions []string `json:"permissions"`
}

type idTokenContentDTO struct {
	commonUserInfoDTO
	jwt.RegisteredClaims
	UpdatedAt time.Time `json:"updated_at"`
}

type userInfoResponseDTO struct {
	commonUserInfoDTO
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
}

type commonUserInfoDTO struct {
	Nickname      string `json:"nickname"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

type refreshTokenResponseDTO struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	IdToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

type errorResponseDTO struct {
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}
