package auth0cliauthorizer

type DeviceConfirmPrompt struct {
	DeviceCode              string `json:"device_code"`
	VerificationUri         string `json:"verification_uri"`
	VerificationUriComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
}
